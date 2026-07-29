// Package anonymous 实现匿名取件 service（仿 vastsa/FileCodeBox UX）。
//
// 核心流程：
//  1. 上传方：上传文件后，调用 GenerateCode 获取 6 位取件码（建立 pickup_code → share_code 映射）
//  2. 取件方：输入 6 位码 + 可选密码，按码取文件
//  3. 系统：校验（DB 为准）→ 扣减次数（DB 原子）→ 返回下载信息
//
// 设计要点（DB 为唯一真相源）：
//   - Redis 仅存 pickup_code → share_code 映射 + 展示信息（文件名等）
//   - 过期时间、剩余次数、密码哈希全部以 file_codes 表为准
//   - 取件码字符表去掉易混淆字符 0/O/1/I/L
package anonymous

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zy84338719/fileCodeBox/backend/internal/pkg/logger"
	"github.com/zy84338719/fileCodeBox/backend/internal/pkg/utils"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db/dao"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db/model"
	"go.uber.org/zap"
)

// 6 位取件码字符表（去掉易混淆字符 0/O/1/I/L）
const codeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
const codeLength = 6

// Redis key 模板（仅存映射 + 展示信息，真实状态查 DB）
const (
	keyPickupCodeMapping = "anon:code:%s" // pickup_code -> share_code
	keyPickupCodeMeta    = "anon:meta:%s" // pickup_code -> 展示信息(share_code|file_name|file_size|content_type|require_auth)
)

// 错误哨兵
var (
	ErrCodeNotFound  = errors.New("pickup code not found")
	ErrCodeExpired   = errors.New("pickup code expired")
	ErrCodeExhausted = errors.New("pickup code exhausted")
	ErrPasswordWrong = errors.New("password wrong")
)

// Service 匿名取件 service。
// Redis 仅存映射 + 展示信息；过期/次数/密码等真实状态全部以 file_codes 表为准。
type Service struct {
	rdb          *redis.Client
	fileCodeRepo *dao.FileCodeRepository
}

// NewService 创建 service。fileCodeRepo 为 nil 时内部自建（Retrieve 需查 DB）。
func NewService(rdb *redis.Client, fileCodeRepo *dao.FileCodeRepository) *Service {
	if fileCodeRepo == nil {
		fileCodeRepo = dao.NewFileCodeRepository()
	}
	return &Service{rdb: rdb, fileCodeRepo: fileCodeRepo}
}

// CodeMeta 取件码展示信息（仅存于 Redis，真实状态查 DB）
type CodeMeta struct {
	ShareCode   string // 真实 file_code
	FileName    string
	FileSize    int64
	ContentType string
	RequireAuth bool // 仅展示"是否需要密码"
}

// GenerateCode 生成 6 位取件码，建立 pickup_code → share_code 映射。
// expireAt 决定 Redis key 的 TTL（应与 DB 记录过期时间对齐）。
func (s *Service) GenerateCode(ctx context.Context, meta CodeMeta, expireAt time.Time) (string, error) {
	if s.rdb == nil {
		return "", errors.New("Redis 未配置，匿名取件功能不可用")
	}
	ttl := time.Until(expireAt)
	if ttl <= 0 {
		return "", errors.New("expireAt 已过期")
	}
	for i := 0; i < 10; i++ {
		code := randomCode()
		ok, err := s.rdb.SetNX(ctx, fmt.Sprintf(keyPickupCodeMapping, code), meta.ShareCode, ttl).Result()
		if err != nil {
			return "", err
		}
		if !ok {
			continue // 已存在，重试
		}
		metaStr := fmt.Sprintf("%s|%s|%d|%s|%t",
			meta.ShareCode, meta.FileName, meta.FileSize, meta.ContentType, meta.RequireAuth)
		if err := s.rdb.Set(ctx, fmt.Sprintf(keyPickupCodeMeta, code), metaStr, ttl).Err(); err != nil {
			s.rdb.Del(ctx, fmt.Sprintf(keyPickupCodeMapping, code))
			return "", err
		}
		return code, nil
	}
	return "", errors.New("failed to generate unique code after 10 retries")
}

// Retrieve 按取件码取件（校验 + 扣减次数，DB 为准）。
// 返回展示信息 CodeMeta。每次成功调用扣减一次剩余次数。
func (s *Service) Retrieve(ctx context.Context, code, password string) (*CodeMeta, error) {
	if s.rdb == nil {
		return nil, errors.New("Redis 未配置，匿名取件功能不可用")
	}
	// 1. 取 share_code（仅映射）
	shareCode, err := s.rdb.Get(ctx, fmt.Sprintf(keyPickupCodeMapping, code)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrCodeNotFound
	}
	if err != nil {
		return nil, err
	}

	// 2. 展示信息（兼容旧格式：解析失败用空值）
	metaStr, _ := s.rdb.Get(ctx, fmt.Sprintf(keyPickupCodeMeta, code)).Result()
	meta := parseMeta(metaStr, shareCode)

	// 3. 查 DB 真实状态
	fc, err := s.fileCodeRepo.GetByCode(ctx, shareCode)
	if err != nil {
		return nil, ErrCodeNotFound
	}

	// 4. 校验过期（时间 + 次数）
	if fc.IsExpired() {
		return nil, ErrCodeExpired
	}

	// 5. 校验密码（bcrypt，DB 为准）
	if fc.RequireAuth {
		if !utils.CheckPassword(fc.PasswordHash, password) {
			return nil, ErrPasswordWrong
		}
	}

	// 6. 原子扣减次数（DB 为准，防并发超卖）
	ok, err := s.fileCodeRepo.DecrementExpiredCount(ctx, shareCode)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrCodeExhausted
	}

	return meta, nil
}

// Cancel 作废取件码（删 Redis 映射）
func (s *Service) Cancel(ctx context.Context, code string) error {
	return s.cleanup(ctx, code)
}

// Peek 按取件码查询分享信息（不扣次数、不校验密码），仅供展示。
// 返回展示信息 CodeMeta + DB 记录（含剩余次数、过期时间等）。
func (s *Service) Peek(ctx context.Context, code string) (*CodeMeta, *model.FileCode, error) {
	if s.rdb == nil {
		return nil, nil, errors.New("Redis 未配置，匿名取件功能不可用")
	}
	shareCode, err := s.rdb.Get(ctx, fmt.Sprintf(keyPickupCodeMapping, code)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil, ErrCodeNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	metaStr, _ := s.rdb.Get(ctx, fmt.Sprintf(keyPickupCodeMeta, code)).Result()
	meta := parseMeta(metaStr, shareCode)
	fc, err := s.fileCodeRepo.GetByCode(ctx, shareCode)
	if err != nil {
		return nil, nil, ErrCodeNotFound
	}
	return meta, fc, nil
}

// ============ 内部辅助 ============

// randomCode 生成 6 位随机码（crypto/rand）
func randomCode() string {
	result := make([]byte, codeLength)
	max := big.NewInt(int64(len(codeAlphabet)))
	for i := range result {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			// 极端情况下回退到时间戳
			n = big.NewInt(int64(time.Now().UnixNano()) % int64(len(codeAlphabet)))
		}
		result[i] = codeAlphabet[n.Int64()]
	}
	return string(result)
}

// parseMeta 解析展示信息（兼容旧格式：字段不足时用空值）
// 新格式: share_code|file_name|file_size|content_type|require_auth
func parseMeta(metaStr, shareCode string) *CodeMeta {
	meta := &CodeMeta{ShareCode: shareCode}
	if metaStr == "" {
		return meta
	}
	parts := splitBy(metaStr, '|', 5)
	if len(parts) >= 2 {
		meta.FileName = parts[1]
	}
	if len(parts) >= 3 {
		fmt.Sscanf(parts[2], "%d", &meta.FileSize)
	}
	if len(parts) >= 4 {
		meta.ContentType = parts[3]
	}
	if len(parts) >= 5 {
		meta.RequireAuth = parts[4] == "true"
	}
	return meta
}

// splitBy 简易 split（避免引入 strings.Split 提升可读性）
func splitBy(s string, sep byte, max int) []string {
	result := make([]string, 0, max)
	start, count := 0, 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep && count < max-1 {
			result = append(result, s[start:i])
			start = i + 1
			count++
		}
	}
	result = append(result, s[start:])
	return result
}

func (s *Service) cleanup(ctx context.Context, code string) error {
	return s.rdb.Del(ctx,
		fmt.Sprintf(keyPickupCodeMapping, code),
		fmt.Sprintf(keyPickupCodeMeta, code),
	).Err()
}

// AnonymousShareParams 匿名分享端到端参数
type AnonymousShareParams struct {
	FilePath       string
	FileName       string
	FileSize       int64
	ContentType    string
	ExpireAt       *time.Time // 必填，决定 DB 过期时间 + Redis TTL
	MaxPickupCount int        // -1=无限, 0=默认无限, >0=限制
	Password       string     // 空=无密码
}

// CreateAnonymousShare 端到端：建 file_codes 记录 + 生成取件码。
// 供 gen handler 调用，避免 handler 直接操作 DAO；密码在此 bcrypt 哈希后存 DB。
// 返回 6 位取件码。
func (s *Service) CreateAnonymousShare(ctx context.Context, p AnonymousShareParams) (string, error) {
	if p.ExpireAt == nil {
		return "", errors.New("ExpireAt 必填")
	}
	maxCount := p.MaxPickupCount
	if maxCount == 0 {
		maxCount = -1 // 默认无限
	}

	hash, err := utils.HashPassword(p.Password)
	if err != nil {
		return "", err
	}

	fc := &model.FileCode{
		Code:          randomShareCode(),
		FilePath:      p.FilePath,
		UUIDFileName:  p.FileName,
		Size:          p.FileSize,
		ExpiredAt:     p.ExpireAt,
		ExpiredCount:  maxCount,
		RequireAuth:   p.Password != "",
		PasswordHash:  hash,
		UploadType:    "anonymous",
	}
	if err := s.fileCodeRepo.Create(ctx, fc); err != nil {
		return "", err
	}

	pickupCode, err := s.GenerateCode(ctx, CodeMeta{
		ShareCode:   fc.Code,
		FileName:    p.FileName,
		FileSize:    p.FileSize,
		ContentType: p.ContentType,
		RequireAuth: p.Password != "",
	}, *p.ExpireAt)
	if err != nil {
		// 回滚 DB 记录（物理文件未落库，无需清）
		if dErr := s.fileCodeRepo.Delete(ctx, fc.ID); dErr != nil {
			logger.Warn("rollback file_code on generate code failed", zap.Uint("id", fc.ID), zap.Error(dErr))
		}
		return "", err
	}
	return pickupCode, nil
}

// randomShareCode 8 位 file_code（crypto/rand，小写字母+数字）
func randomShareCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	const length = 8
	max := big.NewInt(int64(len(charset)))
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			n = big.NewInt(int64(time.Now().UnixNano()) % int64(len(charset)))
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}
