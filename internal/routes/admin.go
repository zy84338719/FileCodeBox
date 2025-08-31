package routes

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/handlers"
	"github.com/zy84338719/filecodebox/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupAdminRoutes 设置管理员相关路由
func SetupAdminRoutes(
	router *gin.Engine,
	adminHandler *handlers.AdminHandler,
	storageHandler *handlers.StorageHandler,
	cfg *config.Config,
) {
	// 管理相关路由
	adminGroup := router.Group("/admin")
	{
		// 管理页面
		adminGroup.GET("/", func(c *gin.Context) {
			ServeAdminPage(c, cfg)
		})

		// 登录不需要认证
		adminGroup.POST("/login", adminHandler.Login)

		// 需要认证的路由
		authGroup := adminGroup.Group("/")
		authGroup.Use(middleware.AdminAuth(cfg))
		{
			authGroup.GET("/dashboard", adminHandler.Dashboard)
			authGroup.GET("/stats", adminHandler.GetStats)
			authGroup.GET("/files", adminHandler.GetFiles)
			authGroup.GET("/files/:code", adminHandler.GetFile)
			authGroup.DELETE("/files/:code", adminHandler.DeleteFile)
			authGroup.PUT("/files/:code", adminHandler.UpdateFile)
			authGroup.GET("/files/download", adminHandler.DownloadFile)
			authGroup.GET("/config", adminHandler.GetConfig)
			authGroup.PUT("/config", adminHandler.UpdateConfig)
			authGroup.POST("/clean", adminHandler.CleanExpiredFiles)

			// 用户管理相关路由
			authGroup.GET("/users", adminHandler.GetUsers)
			authGroup.GET("/users/:id", adminHandler.GetUser)
			authGroup.POST("/users", adminHandler.CreateUser)
			authGroup.PUT("/users/:id", adminHandler.UpdateUser)
			authGroup.DELETE("/users/:id", adminHandler.DeleteUser)
			authGroup.PUT("/users/:id/status", adminHandler.UpdateUserStatus)
			authGroup.GET("/users/:id/files", adminHandler.GetUserFiles)

			// 存储管理相关路由
			authGroup.GET("/storage", storageHandler.GetStorageInfo)
			authGroup.POST("/storage/switch", storageHandler.SwitchStorage)
			authGroup.GET("/storage/test/:type", storageHandler.TestStorageConnection)
			authGroup.PUT("/storage/config", storageHandler.UpdateStorageConfig)
		}
	}
}

// ServeAdminPage 服务管理页面
func ServeAdminPage(c *gin.Context, cfg *config.Config) {
	adminPath := filepath.Join(".", cfg.ThemesSelect, "admin.html")

	content, err := os.ReadFile(adminPath)
	if err != nil {
		c.String(http.StatusNotFound, "Admin page not found")
		return
	}

	c.Header("Cache-Control", "no-cache")
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, string(content))
}
