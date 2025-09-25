package web

import (
	"time"

	"github.com/zy84338719/filecodebox/internal/models"
)

// UserAPIKeyCreateRequest 用户 API Key 创建请求
// 可以指定名称以及过期时间（expires_in_days 或 RFC3339 格式的 expires_at）
type UserAPIKeyCreateRequest struct {
	Name          string  `json:"name"`
	ExpiresInDays *int    `json:"expires_in_days"`
	ExpiresAt     *string `json:"expires_at"`
}

// UserAPIKeyResponse 用户 API Key 响应体
// 不返回真实密钥，仅返回元信息

type UserAPIKeyResponse struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Prefix     string `json:"prefix"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	LastUsedAt string `json:"last_used_at,omitempty"`
	ExpiresAt  string `json:"expires_at,omitempty"`
	RevokedAt  string `json:"revoked_at,omitempty"`
	Revoked    bool   `json:"revoked"`
}

// UserAPIKeyCreateResponse 创建 API Key 的响应，包含明文密钥和元信息

type UserAPIKeyCreateResponse struct {
	Key    string             `json:"key"`
	APIKey UserAPIKeyResponse `json:"api_key"`
}

// UserAPIKeyListResponse 用户 API Key 列表响应
type UserAPIKeyListResponse struct {
	Keys []UserAPIKeyResponse `json:"keys"`
}

// MakeUserAPIKeyResponse 将数据库模型转换为响应结构
func MakeUserAPIKeyResponse(key models.UserAPIKey) UserAPIKeyResponse {
	resp := UserAPIKeyResponse{
		ID:        key.ID,
		Name:      key.Name,
		Prefix:    key.Prefix,
		CreatedAt: key.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: key.UpdatedAt.UTC().Format(time.RFC3339),
		Revoked:   key.Revoked,
	}

	if key.LastUsedAt != nil {
		resp.LastUsedAt = key.LastUsedAt.UTC().Format(time.RFC3339)
	}
	if key.ExpiresAt != nil {
		resp.ExpiresAt = key.ExpiresAt.UTC().Format(time.RFC3339)
	}
	if key.RevokedAt != nil {
		resp.RevokedAt = key.RevokedAt.UTC().Format(time.RFC3339)
	}

	return resp
}
