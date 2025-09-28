package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/services"
)

// APIKeyAuthenticator 抽象出 API Key 验证能力，避免直接依赖具体服务类型
type APIKeyAuthenticator interface {
	AuthenticateAPIKey(string) (*services.APIKeyAuthResult, error)
}

// APIKeyAuthOnly 强制 API Key 认证中间件
// 未携带或携带无效密钥的请求将被直接拒绝
func APIKeyAuthOnly(authenticator APIKeyAuthenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		if authenticator == nil {
			common.InternalServerErrorResponse(c, "API key authenticator 未配置")
			c.Abort()
			return
		}

		key := extractAPIKeyFromRequest(c)
		if key == "" {
			common.UnauthorizedResponse(c, "缺少 API Key")
			c.Abort()
			return
		}

		result, err := authenticator.AuthenticateAPIKey(key)
		if err != nil || result == nil {
			common.UnauthorizedResponse(c, "API Key 无效或已过期")
			c.Abort()
			return
		}

		c.Set("user_id", result.UserID)
		c.Set("username", result.Username)
		c.Set("role", result.Role)
		c.Set("api_key_id", result.KeyID)
		c.Set("auth_via_api_key", true)
		c.Set("is_anonymous", false)

		c.Next()
	}
}

// APIKeyAuth 为向后兼容的别名
func APIKeyAuth(authenticator APIKeyAuthenticator) gin.HandlerFunc {
	return APIKeyAuthOnly(authenticator)
}

func extractAPIKeyFromRequest(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "apikey") {
			return strings.TrimSpace(parts[1])
		}
	}

	if key := c.GetHeader("X-API-Key"); key != "" {
		return strings.TrimSpace(key)
	}

	return ""
}
