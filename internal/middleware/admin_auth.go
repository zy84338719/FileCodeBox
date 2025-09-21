package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/config"
)

// AdminAuth 管理员认证中间件（基于用户权限）
func AdminAuth(manager *config.ConfigManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			common.UnauthorizedResponse(c, "用户权限信息不存在")
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok || roleStr != "admin" {
			common.ForbiddenResponse(c, "需要管理员权限")
			c.Abort()
			return
		}

		c.Next()
	}
}
