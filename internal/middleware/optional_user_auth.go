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
	AuthenticateAPIKey(string) (*services.APIKeyAuthResult, error)
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("is_anonymous", true)

		setUserFromClaims := func(claims *services.AuthClaims) {
			if claims == nil {
				return
			}
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("role", claims.Role)
			c.Set("session_id", claims.SessionID)
			c.Set("auth_via_api_key", false)
			c.Set("is_anonymous", false)
		}

		setUserFromAPIKey := func(result *services.APIKeyAuthResult) {
			if result == nil {
				return
			}
			c.Set("user_id", result.UserID)
			c.Set("username", result.Username)
			c.Set("role", result.Role)
			c.Set("api_key_id", result.KeyID)
			c.Set("auth_via_api_key", true)
			c.Set("is_anonymous", false)
		}

		tryBearer := func(token string) bool {
			if userService == nil || strings.TrimSpace(token) == "" {
				return false
			}
			claimsInterface, err := userService.ValidateToken(strings.TrimSpace(token))
			if err != nil {
				return false
			}
			if claims, ok := claimsInterface.(*services.AuthClaims); ok {
				setUserFromClaims(claims)
				return true
			}
			return false
		}

		tryAPIKey := func(key string) bool {
			if userService == nil || strings.TrimSpace(key) == "" {
				return false
			}
			result, err := userService.AuthenticateAPIKey(strings.TrimSpace(key))
			if err != nil {
				return false
			}
			setUserFromAPIKey(result)
			return true
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 {
				scheme := strings.ToLower(strings.TrimSpace(parts[0]))
				credentials := parts[1]
				switch scheme {
				case "bearer":
					if tryBearer(credentials) {
						c.Next()
						return
					}
				case "apikey":
					if tryAPIKey(credentials) {
						c.Next()
						return
					}
				}
			}
		}

		// Fallback to X-API-Key header
		if keyHeader := c.GetHeader("X-API-Key"); keyHeader != "" {
			if tryAPIKey(keyHeader) {
				c.Next()
				return
			}
		}

		c.Next()
	}
}
