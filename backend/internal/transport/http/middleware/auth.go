package middleware

import (
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/fileCodeBox/backend/internal/pkg/auth"
	"github.com/zy84338719/fileCodeBox/backend/internal/pkg/errors"
)

const (
	// ContextKeyUserID 用户ID上下文键
	ContextKeyUserID = "user_id"
	// ContextKeyUsername 用户名上下文键
	ContextKeyUsername = "username"
	// ContextKeyUserRole 用户角色上下文键
	ContextKeyUserRole = "role"
	// ContextKeyAPIKeyID API Key ID上下文键
	ContextKeyAPIKeyID = "api_key_id"
	// ContextKeyAuthType 认证类型上下文键 (jwt/api_key)
	ContextKeyAuthType = "auth_type"
)

// parseAndSetClaims 解析JWT并将用户信息存入上下文
func parseAndSetClaims(c *app.RequestContext, tokenString string) error {
	claims, err := auth.ParseToken(tokenString)
	if err != nil {
		return err
	}

	c.Set(ContextKeyUserID, claims.UserID)
	c.Set(ContextKeyUsername, claims.Username)
	c.Set(ContextKeyUserRole, claims.Role)
	c.Set(ContextKeyAuthType, "jwt")

	return nil
}

// extractBearerToken 从Authorization头提取Bearer token
func extractBearerToken(c *app.RequestContext) string {
	authHeader := string(c.GetHeader("Authorization"))
	if authHeader == "" {
		return ""
	}
	parts := splitToken(authHeader)
	if len(parts) == 2 && parts[0] == "Bearer" {
		return parts[1]
	}
	return ""
}

// splitToken 分割token字符串
func splitToken(authHeader string) []string {
	result := make([]string, 0, 2)
	current := make([]rune, 0, len(authHeader))

	for _, ch := range authHeader {
		if ch == ' ' {
			if len(current) > 0 {
				result = append(result, string(current))
				current = current[:0]
			}
		} else {
			current = append(current, ch)
		}
	}
	if len(current) > 0 {
		result = append(result, string(current))
	}

	return result
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *app.RequestContext) uint {
	if userID, exists := c.Get(ContextKeyUserID); exists {
		if uid, ok := userID.(uint); ok {
			return uid
		}
	}
	return 0
}

// GetUsername 从上下文获取用户名
func GetUsername(c *app.RequestContext) string {
	if username, exists := c.Get(ContextKeyUsername); exists {
		if uname, ok := username.(string); ok {
			return uname
		}
	}
	return ""
}

// GetUserRole 从上下文获取用户角色
func GetUserRole(c *app.RequestContext) string {
	if role, exists := c.Get(ContextKeyUserRole); exists {
		if r, ok := role.(string); ok {
			return r
		}
	}
	return ""
}

// GetAPIKeyID 从上下文获取API Key ID
func GetAPIKeyID(c *app.RequestContext) uint {
	if keyID, exists := c.Get(ContextKeyAPIKeyID); exists {
		if id, ok := keyID.(uint); ok {
			return id
		}
	}
	return 0
}

// GetAuthType 从上下文获取认证类型
func GetAuthType(c *app.RequestContext) string {
	if authType, exists := c.Get(ContextKeyAuthType); exists {
		if t, ok := authType.(string); ok {
			return t
		}
	}
	return ""
}

// IsAuthenticated 检查是否已认证
func IsAuthenticated(c *app.RequestContext) bool {
	return GetUserID(c) > 0
}

// IsAdmin 检查是否是管理员
func IsAdmin(c *app.RequestContext) bool {
	return GetUserRole(c) == "admin"
}

// respondUnauthorized 返回401响应
func respondUnauthorized(c *app.RequestContext, message string) {
	c.Abort()
	c.JSON(http.StatusUnauthorized, map[string]interface{}{
		"code":    errors.CodeUnauthorized,
		"message": message,
	})
}

// respondForbidden 返回403响应
func respondForbidden(c *app.RequestContext, message string) {
	c.Abort()
	c.JSON(http.StatusForbidden, map[string]interface{}{
		"code":    403,
		"message": message,
	})
}
