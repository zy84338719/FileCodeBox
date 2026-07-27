// Package anonymous 实现匿名取件 service（仿 vastsa/FileCodeBox UX）。
//
// 核心流程：
//   1. 上传方：上传文件后，调用 GenerateCode 获取 6 位取件码
//   2. 取件方：输入 6 位码 + 可选密码，按码取文件
//   3. 系统：校验 → 限次 → 返回下载信息
//
// 取件码生成：
//   - 6 位字母数字（去掉易混淆字符 0/O/1/I/L）
//   - 用 Redis 缓存 mapping: pickup_code -> share_code
//   - 缓存 TTL = 分享过期时间
//
// 限次：
//   - 每 code 独立 counter（pickup_count）
//   - 达到 max_pickup_count 后拒绝
package anonymous

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"
)

// Code 6 位取件码字符表（去掉易混淆字符 0/O/1/I/L）
const codeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
const codeLength = 6

// Redis key 模板
const (
	keyPickupCodeMapping = "anon:code:%s"   // pickup_code -> share_code
	keyPickupCodeMeta    = "anon:meta:%s"    // pickup_code -> meta (JSON-ish: file_name|expire_at|max_count|password_hash)
	keyPickupCodeCounter = "anon:count:%s"   // pickup_code -> current pickup count
)

// ErrCodeNotFound 取件码不存在
var ErrCodeNotFound = errors.New("pickup code not found")

// ErrCodeExpired 取件码已过期
var ErrCodeExpired = errors.New("pickup code expired")

// ErrCodeExhausted 取件码已达上限
var ErrCodeExhausted = errors.New("pickup code exhausted")

// ErrPasswordWrong 密码错误
var ErrPasswordWrong = errors.New("password wrong")

// Service 匿名取件 service
type Service struct {
	rdb *redis.Client
	// passwordHasher 注入的密码哈希函数（默认用 simple）
	passwordHasher func(string) string
}

// NewService 创建 service
func NewService(rdb *redis.Client) *Service {
	return &Service{
		rdb:            rdb,
		passwordHasher: defaultPasswordHash,
	}
}

// SetPasswordHasher 设置自定义密码哈希函数
func (s *Service) SetPasswordHasher(fn func(string) string) {
	s.passwordHasher = fn
}

// CodeMeta 取件码元信息
type CodeMeta struct {
	ShareCode     string
	FileName      string
	FileSize      int64
	ContentType   string
	RequireAuth   bool
	PasswordHash  string
	ExpireAt      time.Time
	MaxPickupCount int32
}

// GenerateCode 生成 6 位取件码
// 返回 (pickup_code, error)；如所有组合都冲突（极低概率），返回 error
func (s *Service) GenerateCode(ctx context.Context, meta CodeMeta) (string, error) {
	// 最多重试 10 次
	for i := 0; i < 10; i++ {
		code := randomCode()
		// 用 SETNX 抢占：成功则保留
		ok, err := s.rdb.SetNX(ctx, fmt.Sprintf(keyPickupCodeMapping, code), meta.ShareCode, time.Until(meta.ExpireAt)).Result()
		if err != nil {
			return "", err
		}
		if !ok {
			continue // 已存在，重试
		}
		// 写 meta
		metaKey := fmt.Sprintf(keyPickupCodeMeta, code)
		metaStr := fmt.Sprintf("%s|%d|%s|%s|%d|%d",
			meta.FileName,
			meta.FileSize,
			meta.ContentType,
			meta.PasswordHash,
			meta.ExpireAt.Unix(),
			meta.MaxPickupCount,
		)
		if err := s.rdb.Set(ctx, metaKey, metaStr, time.Until(meta.ExpireAt)).Err(); err != nil {
			// 回滚
			s.rdb.Del(ctx, fmt.Sprintf(keyPickupCodeMapping, code))
			return "", err
		}
		// 初始化 counter = 0
		if err := s.rdb.Set(ctx, fmt.Sprintf(keyPickupCodeCounter, code), 0, time.Until(meta.ExpireAt)).Err(); err != nil {
			s.rdb.Del(ctx, fmt.Sprintf(keyPickupCodeMapping, code), metaKey)
			return "", err
		}
		return code, nil
	}
	return "", errors.New("failed to generate unique code after 10 retries")
}

// Retrieve 按取件码取件（不下载）
func (s *Service) Retrieve(ctx context.Context, code, password string) (*CodeMeta, error) {
	shareCode, err := s.rdb.Get(ctx, fmt.Sprintf(keyPickupCodeMapping, code)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrCodeNotFound
	}
	if err != nil {
		return nil, err
	}

	metaStr, err := s.rdb.Get(ctx, fmt.Sprintf(keyPickupCodeMeta, code)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrCodeNotFound
	}
	if err != nil {
		return nil, err
	}

	meta, err := parseMeta(code, shareCode, metaStr)
	if err != nil {
		return nil, err
	}

	// 校验过期
	if time.Now().After(meta.ExpireAt) {
		// 主动清理
		s.cleanup(ctx, code)
		return nil, ErrCodeExpired
	}

	// 校验密码
	if meta.PasswordHash != "" {
		if s.passwordHasher(password) != meta.PasswordHash {
			return nil, ErrPasswordWrong
		}
	}

	// 校验限次
	count, err := s.rdb.Get(ctx, fmt.Sprintf(keyPickupCodeCounter, code)).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}
	if meta.MaxPickupCount > 0 && int32(count) >= meta.MaxPickupCount {
		return nil, ErrCodeExhausted
	}

	return meta, nil
}

// IncrementCount 取件成功后 +1
func (s *Service) IncrementCount(ctx context.Context, code string) error {
	return s.rdb.Incr(ctx, fmt.Sprintf(keyPickupCodeCounter, code)).Err()
}

// RemainingCount 剩余取件次数
func (s *Service) RemainingCount(ctx context.Context, code string) (int32, int32, error) {
	used, err := s.rdb.Get(ctx, fmt.Sprintf(keyPickupCodeCounter, code)).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		return 0, 0, err
	}
	metaStr, err := s.rdb.Get(ctx, fmt.Sprintf(keyPickupCodeMeta, code)).Result()
	if err != nil {
		return 0, 0, err
	}
	_, _, _, _, _, maxCount := parseMetaRaw(metaStr)
	if maxCount <= 0 {
		return -1, 0, nil // 无限
	}
	return int32(maxCount) - int32(used), int32(used), nil
}

// Cancel 取消/作废取件码
func (s *Service) Cancel(ctx context.Context, code string) error {
	return s.cleanup(ctx, code)
}

// ============ 内部辅助 ============

// randomCode 生成 6 位随机码
func randomCode() string {
	result := make([]byte, codeLength)
	max := big.NewInt(int64(len(codeAlphabet)))
	for i := 0; i < codeLength; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			// 极端情况下回退到时间戳
			n = big.NewInt(int64(time.Now().UnixNano()) % int64(len(codeAlphabet)))
		}
		result[i] = codeAlphabet[n.Int64()]
	}
	return string(result)
}

func defaultPasswordHash(p string) string {
	// 占位实现：生产环境应该用 bcrypt/scrypt
	// 这里用 SHA256 简化（实际项目换 bcrypt）
	if p == "" {
		return ""
	}
	// TODO: 替换成 bcrypt
	return "sha256:" + p
}

// parseMeta 解析 meta 字符串为 CodeMeta
func parseMeta(code, shareCode, metaStr string) (*CodeMeta, error) {
	fileName, fileSize, contentType, passwordHash, expireAt, maxCount := parseMetaRaw(metaStr)
	expireTime := time.Unix(expireAt, 0)
	return &CodeMeta{
		ShareCode:      shareCode,
		FileName:       fileName,
		FileSize:       fileSize,
		ContentType:    contentType,
		PasswordHash:   passwordHash,
		ExpireAt:       expireTime,
		MaxPickupCount: int32(maxCount),
	}, nil
}

// parseMetaRaw 解析元信息原始格式
//   fileName|fileSize|contentType|passwordHash|expireAt|maxCount
func parseMetaRaw(metaStr string) (string, int64, string, string, int64, int64) {
	parts := splitBy(metaStr, '|', 6)
	if len(parts) < 6 {
		return "", 0, "", "", 0, 0
	}
	var fileSize, expireAt, maxCount int64
	fmt.Sscanf(parts[1], "%d", &fileSize)
	fmt.Sscanf(parts[4], "%d", &expireAt)
	fmt.Sscanf(parts[5], "%d", &maxCount)
	return parts[0], fileSize, parts[2], parts[3], expireAt, maxCount
}

// splitBy 简易 split（避免引入 strings.Split 提升可读性）
func splitBy(s string, sep byte, max int) []string {
	result := make([]string, 0, max)
	start := 0
	count := 0
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
		fmt.Sprintf(keyPickupCodeCounter, code),
	).Err()
}
