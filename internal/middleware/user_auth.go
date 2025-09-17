package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/services"
)

// UserAuth 用户认证中间件
func UserAuth(manager *config.ConfigManager, userService interface {
	ValidateToken(string) (interface{}, error)
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			common.UnauthorizedResponse(c, "缺少认证信息")
			c.Abort()
			return
		}

		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			common.UnauthorizedResponse(c, "认证格式错误")
			c.Abort()
			return
		}

		claimsInterface, err := userService.ValidateToken(tokenParts[1])
		if err != nil {
			common.UnauthorizedResponse(c, "认证失败: "+err.Error())
			c.Abort()
			return
		}

		if claims, ok := claimsInterface.(*services.AuthClaims); ok {
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("role", claims.Role)
			c.Set("session_id", claims.SessionID)
		} else {
			common.UnauthorizedResponse(c, "token格式错误")
			c.Abort()
			return
		}

		c.Next()
	}
}
