// Package presign 实现预签名上传 service。
//
// 流程：
//  1. 客户端调 Init 申请预签名 URL + token
//  2. 客户端 PUT 直传文件到 URL
//  3. 客户端调 Complete 通知服务写 share 表
//  4. 服务返回 share_code
//
// 当前实现：fallback 走"自家直传 URL + token"模式（不依赖真实 S3/OSS 预签名）。
// 当 OpenDAL binding 接入后，可直接生成真实预签名 URL。
package presign

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/zy84338719/fileCodeBox/backend/internal/app/share"
	"github.com/zy84338719/fileCodeBox/backend/internal/pkg/utils"
)

// Redis key 模板
const (
	keyUploadMeta = "presign:meta:%s" // upload_id -> InitMeta (JSON)
)

// 错误定义
var (
	ErrUploadNotFound  = errors.New("upload not found")
	ErrTokenInvalid    = errors.New("upload token invalid")
	ErrUploadExpired   = errors.New("upload expired")
	ErrAlreadyComplete = errors.New("upload already completed")
)

// Service 预签名上传 service
type Service struct {
	rdb           *redis.Client
	defaultExpire time.Duration
	signingKey    []byte
	baseURL       string
	// shareService 注入的 share service（Complete 时调用写分享表）
	shareService ShareServiceInterface
}

// ShareServiceInterface share service 接口（避免循环依赖）
// 与 share.Service.ShareFile / CreateShare 签名保持一致
type ShareServiceInterface interface {
	ShareFile(ctx context.Context, req *share.ShareFileReq) (*share.ShareResp, error)
	CreateShare(ctx context.Context, req *share.ShareFileReq) (*share.ShareResp, error)
}

// NewService 创建 service
func NewService(rdb *redis.Client, baseURL string, signingKey string) *Service {
	return &Service{
		rdb:           rdb,
		defaultExpire: 1 * time.Hour,
		signingKey:    []byte(signingKey),
		baseURL:       baseURL,
	}
}

// SetShareService 注入 share service（用于 Complete 时写分享表）
func (s *Service) SetShareService(svc ShareServiceInterface) {
	s.shareService = svc
}

// InitMeta init 元信息
type InitMeta struct {
	UploadID    string    `json:"upload_id"`
	ObjectKey   string    `json:"object_key"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	ContentType string    `json:"content_type"`
	Scheme      string    `json:"scheme"`
	ExpireAt    time.Time `json:"expire_at"`
	Complete    bool      `json:"complete"`
	UserID      uint      `json:"user_id"`
	ExpireValue int32     `json:"expire_value"`
	ExpireStyle string    `json:"expire_style"`
	RequireAuth bool      `json:"require_auth"`
}

// InitResult init 返回
type InitResult struct {
	UploadID      string            `json:"upload_id"`
	UploadURL     string            `json:"upload_url"`
	Method        string            `json:"method"`
	Headers       map[string]string `json:"headers"`
	ExpireSeconds int32             `json:"expire_seconds"`
	ObjectKey     string            `json:"object_key"`
	Scheme        string            `json:"scheme"`
	Token         string            `json:"token"`
}

// Init 申请预签名
func (s *Service) Init(ctx context.Context, meta InitMeta) (*InitResult, error) {
	// 1. 生成 upload_id
	uploadID, err := genID("up")
	if err != nil {
		return nil, err
	}
	meta.UploadID = uploadID
	meta.ExpireAt = time.Now().Add(s.defaultExpire)
	meta.Complete = false

	// 2. 存 meta
	metaJSON, _ := json.Marshal(meta)
	if err := s.rdb.Set(ctx, fmt.Sprintf(keyUploadMeta, uploadID), metaJSON, s.defaultExpire).Err(); err != nil {
		return nil, err
	}

	// 3. 生成 token
	token := s.signToken(uploadID, meta.ObjectKey, meta.ExpireAt)

	// 4. 生成 upload URL
	// 当前 fallback：自家直传 URL
	uploadURL := fmt.Sprintf("%s/api/v1/presign/upload-direct/%s", s.baseURL, uploadID)

	return &InitResult{
		UploadID:  uploadID,
		UploadURL: uploadURL,
		Method:    "PUT",
		Headers: map[string]string{
			"X-Upload-Token": token,
		},
		ExpireSeconds: int32(s.defaultExpire.Seconds()),
		ObjectKey:     meta.ObjectKey,
		Scheme:        meta.Scheme,
		Token:         token,
	}, nil
}

// CompleteResult Complete 返回结果（包含 share_code）
type CompleteResult struct {
	*InitMeta
	ShareCode    string
	ShareURL     string
	FullShareURL string
	OwnerIP      string
}

// Complete 完成通知（调 share service 写分享表）
func (s *Service) Complete(ctx context.Context, uploadID, token, ownerIP string) (*CompleteResult, error) {
	// 1. 读 meta
	metaJSON, err := s.rdb.Get(ctx, fmt.Sprintf(keyUploadMeta, uploadID)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrUploadNotFound
	}
	if err != nil {
		return nil, err
	}

	var meta InitMeta
	if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
		return nil, err
	}

	// 2. 校验过期
	if time.Now().After(meta.ExpireAt) {
		s.rdb.Del(ctx, fmt.Sprintf(keyUploadMeta, uploadID))
		return nil, ErrUploadExpired
	}

	// 3. 校验 token
	expected := s.signToken(meta.UploadID, meta.ObjectKey, meta.ExpireAt)
	if !hmac.Equal([]byte(expected), []byte(token)) {
		return nil, ErrTokenInvalid
	}

	// 4. 校验是否已完成
	if meta.Complete {
		return nil, ErrAlreadyComplete
	}

	// 5. 标记完成
	meta.Complete = true
	updatedJSON, _ := json.Marshal(meta)
	// 保留 meta 一段时间供查询（5 分钟）
	s.rdb.Set(ctx, fmt.Sprintf(keyUploadMeta, uploadID), updatedJSON, 5*time.Minute)

	// 6. 调 share service 写分享表
	shareCode, shareURL, fullShareURL, shareErr := s.createShareRecord(ctx, &meta, ownerIP)
	if shareErr != nil {
		// 写分享表失败不算 fatal（meta 已标记 complete），返回 shareErr 让调用方决定
		return &CompleteResult{
			InitMeta:     &meta,
			ShareCode:    "",
			ShareURL:     "",
			FullShareURL: "",
			OwnerIP:      ownerIP,
		}, fmt.Errorf("create share record: %w", shareErr)
	}

	return &CompleteResult{
		InitMeta:     &meta,
		ShareCode:    shareCode,
		ShareURL:     shareURL,
		FullShareURL: fullShareURL,
		OwnerIP:      ownerIP,
	}, nil
}

// createShareRecord 调 share service 写分享记录
// 返回 (shareCode, shareURL, fullShareURL, error)
func (s *Service) createShareRecord(ctx context.Context, meta *InitMeta, ownerIP string) (string, string, string, error) {
	if s.shareService == nil {
		// share service 未注入：返回 mock 数据（用于单测 / 未配置场景）
		return "mock_" + meta.UploadID, "/share/mock", s.baseURL + "/share/mock", nil
	}

	// 计算过期时间（与 share.ShareTextWithAuth 行为一致）
	expireTime := utils.CalculateExpireTime(int(meta.ExpireValue), meta.ExpireStyle)
	expireCount := utils.CalculateExpireCount(meta.ExpireStyle, int(meta.ExpireValue))

	uploadType := "presign_anonymous"
	var userIDPtr *uint
	if meta.UserID != 0 {
		uid := meta.UserID
		userIDPtr = &uid
		uploadType = "presign_authenticated"
	}

	req := &share.ShareFileReq{
		FilePath:     meta.ObjectKey,
		Size:         meta.FileSize,
		Text:         meta.FileName,
		ExpiredAt:    expireTime,
		ExpiredCount: expireCount,
		RequireAuth:  meta.RequireAuth,
		UserID:       userIDPtr,
		UploadType:   uploadType,
		OwnerIP:      ownerIP,
		UploadID:     meta.UploadID,
	}

	resp, err := s.shareService.CreateShare(ctx, req)
	if err != nil {
		return "", "", "", err
	}

	return resp.Code, resp.ShareURL, resp.FullShareURL, nil
}

// Abort 取消
func (s *Service) Abort(ctx context.Context, uploadID, token string) error {
	metaJSON, err := s.rdb.Get(ctx, fmt.Sprintf(keyUploadMeta, uploadID)).Result()
	if errors.Is(err, redis.Nil) {
		return ErrUploadNotFound
	}
	if err != nil {
		return err
	}
	var meta InitMeta
	if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
		return err
	}
	expected := s.signToken(meta.UploadID, meta.ObjectKey, meta.ExpireAt)
	if !hmac.Equal([]byte(expected), []byte(token)) {
		return ErrTokenInvalid
	}
	return s.rdb.Del(ctx, fmt.Sprintf(keyUploadMeta, uploadID)).Err()
}

// GetMeta 查询（管理后台用）
func (s *Service) GetMeta(ctx context.Context, uploadID string) (*InitMeta, error) {
	metaJSON, err := s.rdb.Get(ctx, fmt.Sprintf(keyUploadMeta, uploadID)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrUploadNotFound
	}
	if err != nil {
		return nil, err
	}
	var meta InitMeta
	if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// ============ 内部 ============

// signToken 生成 HMAC token
//
//	token = hex(hmac-sha256(signingKey, uploadID|objectKey|expireAt.Unix))
func (s *Service) signToken(uploadID, objectKey string, expireAt time.Time) string {
	mac := hmac.New(sha256.New, s.signingKey)
	mac.Write([]byte(fmt.Sprintf("%s|%s|%d", uploadID, objectKey, expireAt.Unix())))
	return hex.EncodeToString(mac.Sum(nil))
}

func genID(prefix string) (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return prefix + "_" + hex.EncodeToString(b), nil
}
