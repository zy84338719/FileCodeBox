package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/services"
)

// OptionalUserAuth 可选用户认证中间件（支持匿名和登录用户）
func OptionalUserAuth(manager *config.ConfigManager, userService interface {
	ValidateToken(string) (interface{}, error)
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Set("is_anonymous", true)
			c.Next()
			return
		}

		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.Set("is_anonymous", true)
			c.Next()
			return
		}

		claimsInterface, err := userService.ValidateToken(tokenParts[1])
		if err != nil {
			c.Set("is_anonymous", true)
			c.Next()
			return
		}

		if claims, ok := claimsInterface.(*services.AuthClaims); ok {
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("role", claims.Role)
			c.Set("session_id", claims.SessionID)
			c.Set("is_anonymous", false)
		} else {
			c.Set("is_anonymous", true)
		}

		c.Next()
	}
}
