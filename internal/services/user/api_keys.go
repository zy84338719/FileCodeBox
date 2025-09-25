package user

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zy84338719/filecodebox/internal/models"
)

const (
	userAPIKeyPrefix      = "fcbk_"
	userAPIKeyRandomBytes = 32 // 将被编码为64位十六进制字符串
	maxUserAPIKeys        = 5  // 每个用户最多保留的有效密钥数量
)

// APIKeyAuthResult 表示通过 API Key 验证后的用户信息
// 在中间件中用于注入用户上下文

type APIKeyAuthResult struct {
	UserID   uint
	Username string
	Role     string
	KeyID    uint
}

// GenerateUserAPIKey 为用户生成新的 API Key，并返回明文 Key 及记录
func (s *Service) GenerateUserAPIKey(userID uint, name string, expiresAt *time.Time) (string, *models.UserAPIKey, error) {
	// 限制有效密钥数量，避免滥用
	count, err := s.repositoryManager.UserAPIKey.CountActiveByUser(userID)
	if err != nil {
		return "", nil, fmt.Errorf("统计用户 API Key 失败: %w", err)
	}
	if count >= maxUserAPIKeys {
		return "", nil, fmt.Errorf("最多只能保留 %d 个有效 API Key，请先删除旧的密钥", maxUserAPIKeys)
	}

	// 生成随机 Token
	raw, err := s.authService.GenerateRandomToken(userAPIKeyRandomBytes)
	if err != nil {
		return "", nil, fmt.Errorf("生成随机密钥失败: %w", err)
	}
	key := userAPIKeyPrefix + raw
	hash := hashAPIKey(key)
	prefix := keyPrefix(key)

	// 归一化名称
	trimmedName := strings.TrimSpace(name)
	if len(trimmedName) > 100 {
		trimmedName = trimmedName[:100]
	}

	record := &models.UserAPIKey{
		UserID:    userID,
		Name:      trimmedName,
		Prefix:    prefix,
		KeyHash:   hash,
		ExpiresAt: normalizeExpiry(expiresAt),
	}

	if err := s.repositoryManager.UserAPIKey.Create(record); err != nil {
		return "", nil, fmt.Errorf("保存 API Key 失败: %w", err)
	}

	return key, record, nil
}

// ListUserAPIKeys 返回用户的全部 API Key（包含已撤销）
func (s *Service) ListUserAPIKeys(userID uint) ([]models.UserAPIKey, error) {
	return s.repositoryManager.UserAPIKey.ListByUser(userID)
}

// RevokeUserAPIKey 撤销指定的 API Key
func (s *Service) RevokeUserAPIKey(userID, id uint) error {
	return s.repositoryManager.UserAPIKey.RevokeByID(userID, id)
}

// AuthenticateAPIKey 校验用户 API Key，返回认证结果
func (s *Service) AuthenticateAPIKey(key string) (*APIKeyAuthResult, error) {
	if key == "" {
		return nil, errors.New("api key is empty")
	}

	if !strings.HasPrefix(key, userAPIKeyPrefix) {
		return nil, errors.New("api key 格式不正确")
	}

	hash := hashAPIKey(key)
	record, err := s.repositoryManager.UserAPIKey.GetActiveByHash(hash)
	if err != nil {
		return nil, err
	}

	// 检查是否过期
	if record.ExpiresAt != nil && record.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("api key 已过期")
	}

	user, err := s.repositoryManager.User.GetByID(record.UserID)
	if err != nil {
		return nil, err
	}
	if user.Status != "active" {
		return nil, errors.New("账号状态异常，无法使用 api key")
	}

	// 更新最后使用时间（忽略错误）
	_ = s.repositoryManager.UserAPIKey.TouchLastUsed(record.ID)

	return &APIKeyAuthResult{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		KeyID:    record.ID,
	}, nil
}

// hashAPIKey 计算密钥哈希
func hashAPIKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

// keyPrefix 返回展示用前缀
func keyPrefix(key string) string {
	if len(key) <= 12 {
		return key
	}
	return key[:12]
}

// normalizeExpiry 复制并清理过期时间，确保不会保存过去的时间
func normalizeExpiry(input *time.Time) *time.Time {
	if input == nil {
		return nil
	}
	t := input.UTC()
	if t.Before(time.Now().Add(-1 * time.Minute)) {
		return nil
	}
	return &t
}
