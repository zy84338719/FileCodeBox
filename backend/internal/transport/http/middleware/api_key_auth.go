package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"sync"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/fileCodeBox/internal/repo/db/dao"
	"github.com/zy84338719/fileCodeBox/internal/repo/db/model"
)

const (
	// APIKeyHeaderName API Key 请求头名称
	APIKeyHeaderName = "X-API-Key"
	// APIKeyQueryParamName API Key 查询参数名称
	APIKeyQueryParamName = "api_key"
	// APIKeyAuthHeaderPrefix Authorization 头中 API Key 的前缀
	APIKeyAuthHeaderPrefix = "ApiKey"
)

var (
	apiKeyRepository     *dao.UserAPIKeyRepository
	apiKeyRepositoryOnce sync.Once
)

// InitAPIKeyRepository 初始化 API Key Repository
func InitAPIKeyRepository() {
	apiKeyRepositoryOnce.Do(func() {
		apiKeyRepository = dao.NewUserAPIKeyRepository()
	})
}

// GetAPIKeyRepository 获取 API Key Repository
func GetAPIKeyRepository() *dao.UserAPIKeyRepository {
	if apiKeyRepository == nil {
		InitAPIKeyRepository()
	}
	return apiKeyRepository
}

// APIKeyAuth API Key 认证中间件
// 从 Header 或 Query 参数获取 API Key，验证后将用户信息注入上下文
func APIKeyAuth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		apiKey := extractAPIKey(c)
		if apiKey == "" {
			respondUnauthorized(c, "API Key is required")
			return
		}

		repo := GetAPIKeyRepository()
		if repo == nil {
			respondUnauthorized(c, "API Key service not initialized")
			return
		}

		// 计算 hash 并验证
		hash := computeAPIKeyHash(apiKey)
		key, err := repo.GetActiveByHash(ctx, hash)
		if err != nil {
			respondUnauthorized(c, "Invalid or expired API Key")
			return
		}

		// 获取用户信息以设置角色
		user, err := getUserByID(ctx, key.UserID)
		if err != nil {
			respondUnauthorized(c, "User not found")
			return
		}

		// 更新最后使用时间
		_ = repo.TouchLastUsed(ctx, key.ID)

		// 将用户信息存入上下文
		c.Set(ContextKeyUserID, user.ID)
		c.Set(ContextKeyUsername, user.Username)
		c.Set(ContextKeyUserRole, user.Role)
		c.Set(ContextKeyAPIKeyID, key.ID)
		c.Set(ContextKeyAuthType, "api_key")

		c.Next(ctx)
	}
}

// OptionalAPIKeyAuth 可选 API Key 认证中间件
// 不强制要求 API Key，但如果提供了有效的 key 则解析用户信息
func OptionalAPIKeyAuth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		apiKey := extractAPIKey(c)
		if apiKey == "" {
			// 没有提供 API Key，继续执行（匿名用户）
			c.Next(ctx)
			return
		}

		repo := GetAPIKeyRepository()
		if repo == nil {
			c.Next(ctx)
			return
		}

		// 尝试验证 API Key
		hash := computeAPIKeyHash(apiKey)
		key, err := repo.GetActiveByHash(ctx, hash)
		if err != nil {
			// API Key 无效，继续执行（匿名用户）
			c.Next(ctx)
			return
		}

		// 获取用户信息
		user, err := getUserByID(ctx, key.UserID)
		if err != nil {
			c.Next(ctx)
			return
		}

		// 更新最后使用时间
		_ = repo.TouchLastUsed(ctx, key.ID)

		// 将用户信息存入上下文
		c.Set(ContextKeyUserID, user.ID)
		c.Set(ContextKeyUsername, user.Username)
		c.Set(ContextKeyUserRole, user.Role)
		c.Set(ContextKeyAPIKeyID, key.ID)
		c.Set(ContextKeyAuthType, "api_key")

		c.Next(ctx)
	}
}

// APIKeyAuthWithAdmin API Key 认证 + 管理员权限检查
func APIKeyAuthWithAdmin() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		apiKey := extractAPIKey(c)
		if apiKey == "" {
			respondUnauthorized(c, "API Key is required")
			return
		}

		repo := GetAPIKeyRepository()
		if repo == nil {
			respondUnauthorized(c, "API Key service not initialized")
			return
		}

		// 计算 hash 并验证
		hash := computeAPIKeyHash(apiKey)
		key, err := repo.GetActiveByHash(ctx, hash)
		if err != nil {
			respondUnauthorized(c, "Invalid or expired API Key")
			return
		}

		// 获取用户信息
		user, err := getUserByID(ctx, key.UserID)
		if err != nil {
			respondUnauthorized(c, "User not found")
			return
		}

		// 检查管理员权限
		if user.Role != UserRoleAdmin {
			respondForbidden(c, "Admin access required")
			return
		}

		// 更新最后使用时间
		_ = repo.TouchLastUsed(ctx, key.ID)

		// 将用户信息存入上下文
		c.Set(ContextKeyUserID, user.ID)
		c.Set(ContextKeyUsername, user.Username)
		c.Set(ContextKeyUserRole, user.Role)
		c.Set(ContextKeyAPIKeyID, key.ID)
		c.Set(ContextKeyAuthType, "api_key")

		c.Next(ctx)
	}
}

// extractAPIKey 从请求中提取 API Key
// 支持以下方式:
// 1. Authorization: ApiKey xxx
// 2. X-API-Key: xxx
// 3. ?api_key=xxx (query parameter)
func extractAPIKey(c *app.RequestContext) string {
	// 1. 检查 Authorization 头
	authHeader := string(c.GetHeader("Authorization"))
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], APIKeyAuthHeaderPrefix) {
			return strings.TrimSpace(parts[1])
		}
	}

	// 2. 检查 X-API-Key 头
	if key := string(c.GetHeader(APIKeyHeaderName)); key != "" {
		return strings.TrimSpace(key)
	}

	// 3. 检查 query 参数
	if key := c.Query(APIKeyQueryParamName); key != "" {
		return strings.TrimSpace(key)
	}

	return ""
}

// computeAPIKeyHash 计算 API Key 的 SHA-256 hash
func computeAPIKeyHash(apiKey string) string {
	hasher := sha256.New()
	hasher.Write([]byte(apiKey))
	return hex.EncodeToString(hasher.Sum(nil))
}

// getUserByID 根据 ID 获取用户信息
func getUserByID(ctx context.Context, userID uint) (*model.User, error) {
	// 使用 UserRepository 获取用户
	userRepo := dao.NewUserRepository()
	return userRepo.GetByID(ctx, userID)
}
