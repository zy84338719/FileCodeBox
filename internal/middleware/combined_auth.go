package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/services"
)

// CombinedAdminAuth 支持两类管理员认证：
// - session-based JWT（由 userService.ValidateToken 验证，返回 *services.AuthClaims，且 Role=="admin"）
// - admin JWT（由 AdminService 生成，使用 manager.User.JWTSecret 签名，claims 为 jwt.MapClaims 且包含 is_admin:true 或 role:"admin"）
func CombinedAdminAuth(manager *config.ConfigManager, userService interface {
	ValidateToken(string) (interface{}, error)
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := c.Request.URL.Path

		// 公开静态/入口路径（允许匿名访问 admin 前端和静态资源）
		if strings.HasPrefix(p, "/admin/css/") || strings.HasPrefix(p, "/admin/js/") || strings.HasPrefix(p, "/admin/templates/") || strings.HasPrefix(p, "/admin/assets/") || strings.HasPrefix(p, "/admin/components/") || p == "/admin" || p == "/admin/" || p == "/admin/login" {
			c.Next()
			return
		}

		// 读取 Authorization header
		authHeader := c.GetHeader("Authorization")
		var tokenStr string
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenStr = parts[1]
			}
		}

		if tokenStr != "" {
			// 仅使用 session-based token（复用普通用户登录逻辑），并加强角色校验
			if userService != nil {
				if claimsIface, err := userService.ValidateToken(tokenStr); err == nil {
					if claims, ok := claimsIface.(*services.AuthClaims); ok {
						// 如果不是管理员直接拒绝
						if claims.Role == "admin" {
							c.Set("user_id", claims.UserID)
							c.Set("username", claims.Username)
							c.Set("role", claims.Role)
							c.Set("session_id", claims.SessionID)
							c.Set("auth_type", "jwt")
							c.Next()
							return
						}
					}
				}
			}
		}

		common.UnauthorizedResponse(c, "需要管理员权限")
		c.Abort()
	}
}
