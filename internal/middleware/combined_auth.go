package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/services"
)

// CombinedAdminAuth 尝试基于 JWT 的用户认证并确保角色为 admin
func CombinedAdminAuth(manager *config.ConfigManager, userService interface {
	ValidateToken(string) (interface{}, error)
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先尝试JWT用户认证
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenParts := strings.SplitN(authHeader, " ", 2)
			if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
				claimsInterface, err := userService.ValidateToken(tokenParts[1])
				if err == nil {
					if claims, ok := claimsInterface.(*services.AuthClaims); ok {
						if claims.Role == "admin" {
							c.Set("user_id", claims.UserID)
							c.Set("username", claims.Username)
							c.Set("role", claims.Role)
							c.Set("session_id", claims.SessionID)
							c.Set("auth_type", "jwt")
							c.Next()
							return
						}
						// role mismatch
						if manager == nil || (manager.Base != nil && !manager.Base.Production) {
							logrus.WithFields(logrus.Fields{"role": claims.Role}).Debug("combined auth: token role is not admin")
						}
					}
				} else {
					if manager == nil || (manager.Base != nil && !manager.Base.Production) {
						logrus.WithError(err).Debug("combined auth: ValidateToken returned error")
					}
				}
				// JWT 验证失败或非管理员角色，继续到失败处理（不回退到旧的静态令牌）
			}
		}

		common.UnauthorizedResponse(c, "需要管理员权限")
		c.Abort()
	}
}
