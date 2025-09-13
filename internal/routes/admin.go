package routes

import (
	"fmt"
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
	cfg *config.ConfigManager,
	userService interface {
		ValidateToken(string) (interface{}, error)
	},
) {
	// 管理相关路由
	adminGroup := router.Group("/admin")

	// 管理页面和静态文件 - 不需要认证就能访问HTML和静态资源
	{
		// 管理页面
		adminGroup.GET("/", func(c *gin.Context) {
			ServeAdminPage(c, cfg)
		})

		// 模块化管理后台静态文件
		themeDir := fmt.Sprintf("./%s", cfg.ThemesSelect)
		adminGroup.Static("/css", fmt.Sprintf("%s/admin/css", themeDir))
		adminGroup.Static("/js", fmt.Sprintf("%s/admin/js", themeDir))
		adminGroup.Static("/templates", fmt.Sprintf("%s/admin/templates", themeDir))
		adminGroup.Static("/assets", fmt.Sprintf("%s/assets", themeDir))
		adminGroup.Static("/components", fmt.Sprintf("%s/components", themeDir))
	}

	// 需要用户认证且为管理员的API路由组
	authGroup := adminGroup.Group("")
	authGroup.Use(middleware.UserAuth(cfg, userService))
	authGroup.Use(middleware.AdminAuth(cfg)) // 验证用户是否为管理员
	{
		// 仪表板和统计
		authGroup.GET("/dashboard", adminHandler.Dashboard)
		authGroup.GET("/stats", adminHandler.GetStats)

		// 文件管理
		authGroup.GET("/files", adminHandler.GetFiles)
		authGroup.GET("/files/:code", adminHandler.GetFile)
		authGroup.DELETE("/files/:code", adminHandler.DeleteFile)
		authGroup.PUT("/files/:code", adminHandler.UpdateFile)
		authGroup.GET("/files/download", adminHandler.DownloadFile)

		// 系统配置
		authGroup.GET("/config", adminHandler.GetConfig)
		authGroup.PUT("/config", adminHandler.UpdateConfig)

		// 系统维护
		setupMaintenanceRoutes(authGroup, adminHandler)

		// 用户管理
		setupUserRoutes(authGroup, adminHandler)

		// 存储管理
		setupStorageRoutes(adminGroup, storageHandler)

		// MCP 服务器管理
		setupMCPRoutes(adminGroup, adminHandler)
	}
}

// ServeAdminPage 服务管理页面
func ServeAdminPage(c *gin.Context, cfg *config.ConfigManager) {
	// 使用新的模块化管理页面
	adminPath := filepath.Join(".", cfg.ThemesSelect, "admin", "index.html")

	content, err := os.ReadFile(adminPath)
	if err != nil {
		c.String(http.StatusNotFound, "Admin page not found")
		return
	}

	c.Header("Cache-Control", "no-cache")
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, string(content))
}

// setupMaintenanceRoutes 设置系统维护路由
func setupMaintenanceRoutes(authGroup *gin.RouterGroup, adminHandler *handlers.AdminHandler) {
	// 系统维护基础操作
	authGroup.POST("/maintenance/clean-expired", adminHandler.CleanExpiredFiles)
	authGroup.POST("/maintenance/clean-temp", adminHandler.CleanTempFiles)
	authGroup.POST("/maintenance/clean-invalid", adminHandler.CleanInvalidRecords)

	// 数据库维护
	authGroup.POST("/maintenance/db/optimize", adminHandler.OptimizeDatabase)
	authGroup.GET("/maintenance/db/analyze", adminHandler.AnalyzeDatabase)
	authGroup.POST("/maintenance/db/backup", adminHandler.BackupDatabase)

	// 缓存管理
	authGroup.POST("/maintenance/cache/clear-system", adminHandler.ClearSystemCache)
	authGroup.POST("/maintenance/cache/clear-upload", adminHandler.ClearUploadCache)
	authGroup.POST("/maintenance/cache/clear-download", adminHandler.ClearDownloadCache)

	// 系统监控
	authGroup.GET("/maintenance/system-info", adminHandler.GetSystemInfo)
	authGroup.GET("/maintenance/monitor/storage", adminHandler.GetStorageStatus)
	authGroup.GET("/maintenance/monitor/performance", adminHandler.GetPerformanceMetrics)

	// 安全管理
	authGroup.POST("/maintenance/security/scan", adminHandler.ScanSecurity)
	authGroup.GET("/maintenance/security/permissions", adminHandler.CheckPermissions)
	authGroup.GET("/maintenance/security/integrity", adminHandler.CheckIntegrity)

	// 日志管理
	authGroup.POST("/maintenance/logs/clear-system", adminHandler.ClearSystemLogs)
	authGroup.POST("/maintenance/logs/clear-access", adminHandler.ClearAccessLogs)
	authGroup.POST("/maintenance/logs/clear-error", adminHandler.ClearErrorLogs)
	authGroup.GET("/maintenance/logs/export", adminHandler.ExportLogs)
	authGroup.GET("/maintenance/logs/stats", adminHandler.GetLogStats)
	authGroup.GET("/maintenance/logs", adminHandler.GetSystemLogs)

	// 任务管理
	authGroup.GET("/maintenance/tasks", adminHandler.GetRunningTasks)
	authGroup.POST("/maintenance/tasks/:id/cancel", adminHandler.CancelTask)
	authGroup.POST("/maintenance/tasks/:id/retry", adminHandler.RetryTask)

	// 系统控制
	authGroup.POST("/maintenance/restart", adminHandler.RestartSystem)
}

// setupUserRoutes 设置用户管理路由
func setupUserRoutes(authGroup *gin.RouterGroup, adminHandler *handlers.AdminHandler) {
	authGroup.GET("/users", adminHandler.GetUsers)
	authGroup.GET("/users/:id", adminHandler.GetUser)
	authGroup.POST("/users", adminHandler.CreateUser)
	authGroup.PUT("/users/:id", adminHandler.UpdateUser)
	authGroup.DELETE("/users/:id", adminHandler.DeleteUser)
	authGroup.PUT("/users/:id/status", adminHandler.UpdateUserStatus)
	authGroup.GET("/users/:id/files", adminHandler.GetUserFiles)
	authGroup.GET("/users/export", adminHandler.ExportUsers)
}

// setupStorageRoutes 设置存储管理路由
func setupStorageRoutes(authGroup *gin.RouterGroup, storageHandler *handlers.StorageHandler) {
	authGroup.GET("/storage", storageHandler.GetStorageInfo)
	authGroup.POST("/storage/switch", storageHandler.SwitchStorage)
	authGroup.GET("/storage/test/:type", storageHandler.TestStorageConnection)
	authGroup.PUT("/storage/config", storageHandler.UpdateStorageConfig)
}

// setupMCPRoutes 设置MCP服务器管理路由
func setupMCPRoutes(authGroup *gin.RouterGroup, adminHandler *handlers.AdminHandler) {
	authGroup.GET("/mcp/config", adminHandler.GetMCPConfig)
	authGroup.PUT("/mcp/config", adminHandler.UpdateMCPConfig)
	authGroup.GET("/mcp/status", adminHandler.GetMCPStatus)
	authGroup.POST("/mcp/restart", adminHandler.RestartMCPServer)
	authGroup.POST("/mcp/control", adminHandler.ControlMCPServer)
	authGroup.POST("/mcp/test", adminHandler.TestMCPConnection)
}
