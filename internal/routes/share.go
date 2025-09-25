package routes

import (
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/handlers"
	"github.com/zy84338719/filecodebox/internal/middleware"
	"github.com/zy84338719/filecodebox/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupShareRoutes 设置分享相关路由
func SetupShareRoutes(
	router *gin.Engine,
	shareHandler *handlers.ShareHandler,
	cfg *config.ConfigManager,
	userService *services.UserService,
) {
	// 幂等检查：如果 /share/text/ 已注册则跳过（防止重复注册导致 gin panic）
	for _, r := range router.Routes() {
		if r.Method == "POST" && r.Path == "/share/text/" {
			return
		}
	}
	// 分享相关路由
	shareGroup := router.Group("/share")
	shareGroup.Use(middleware.ShareAuth(cfg))
	shareGroup.Use(middleware.OptionalUserAuth(cfg, userService)) // 使用可选用户认证
	{
		shareGroup.POST("/text/", shareHandler.ShareText)
		shareGroup.POST("/file/", shareHandler.ShareFile)
		shareGroup.GET("/select/", shareHandler.GetFile)
		shareGroup.POST("/select/", shareHandler.GetFile)
		shareGroup.GET("/download", shareHandler.DownloadFile)
	}
}
