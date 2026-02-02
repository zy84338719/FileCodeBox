package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/zy84338719/filecodebox/internal/handlers"
)

// SetupQRRoutes 注册二维码相关路由
func SetupQRRoutes(router *gin.Engine) {
	if router == nil {
		return
	}

	// 创建二维码处理器
	qrHandler := handlers.NewQRCodeHandler()

	// 公开的二维码API路由组（无需认证）
	qrGroup := router.Group("/api/qrcode")
	{
		// 生成PNG格式二维码
		qrGroup.GET("/generate", qrHandler.GenerateQRCode)

		// 生成Base64编码的二维码
		qrGroup.GET("/base64", qrHandler.GenerateQRCodeBase64)
	}
}
