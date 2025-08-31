package routes

import (
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/handlers"
	"github.com/zy84338719/filecodebox/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupShareRoutes 设置分享相关路由
func SetupShareRoutes(
	router *gin.Engine,
	shareHandler *handlers.ShareHandler,
	cfg *config.Config,
	userService interface {
		ValidateToken(string) (interface{}, error)
	},
) {
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
