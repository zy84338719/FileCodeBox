package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/config"
)

// ShareAuth 分享认证中间件
func ShareAuth(manager *config.ConfigManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if manager.Transfer.Upload.OpenUpload == 0 {
			common.ForbiddenResponse(c, "上传功能已关闭")
			c.Abort()
			return
		}
		c.Next()
	}
}
