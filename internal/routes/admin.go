package routes

import (
	"net/http"
	"strings"

	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/handlers"
	"github.com/zy84338719/filecodebox/internal/middleware"
	"github.com/zy84338719/filecodebox/internal/static"

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

	// 管理页面和静态文件 - 管理页面本身应当需要管理员认证，静态资源仍然注册为公开以便前端加载
	{
		// 管理页面 - 不在此注册（移到需要认证的路由组），确保只有管理员可以访问前端入口
		// 登录接口保持公开

		// 管理员登录（通过用户名/密码获取 JWT）
		// 如果已经存在相同的 POST /admin/login 路由（例如在未初始化数据库时注册的占位处理器），
		// 则跳过注册以避免 gin 的 "handlers are already registered for path" panic。
		// If a placeholder route was registered earlier (no-DB mode), skip; otherwise register.
		exists := false
		for _, r := range router.Routes() {
			if r.Method == "POST" && r.Path == "/admin/login" {
				exists = true
				break
			}
		}
		if !exists {
			adminGroup.POST("/login", func(c *gin.Context) {
				// 尝试从全局注入获取真实 handler（SetInjectedAdminHandler）
				if injected := handlers.GetInjectedAdminHandler(); injected != nil {
					injected.Login(c)
					return
				}
				c.JSON(404, gin.H{"code": 404, "message": "admin handler not configured"})
			})
		}

	}

	// 将管理后台静态资源与前端入口注册为公开路由，允许未认证用户加载登录页面和相关静态资源
	// 注意：API 路由仍然放在受保护的 authGroup 中
	serveFile := func(parts ...string) func(*gin.Context) {
		return func(c *gin.Context) {
			rel := strings.TrimPrefix(c.Param("filepath"), "/")
			if rel == "" {
				c.Status(http.StatusNotFound)
				return
			}
			joined := append(parts, rel)
			static.ServeThemeFile(c, cfg, joined...)
		}
	}

	// css
	adminGroup.GET("/css/*filepath", serveFile("admin", "css"))
	adminGroup.HEAD("/css/*filepath", serveFile("admin", "css"))

	// js
	adminGroup.GET("/js/*filepath", serveFile("admin", "js"))
	adminGroup.HEAD("/js/*filepath", serveFile("admin", "js"))

	// templates
	adminGroup.GET("/templates/*filepath", serveFile("admin", "templates"))
	adminGroup.HEAD("/templates/*filepath", serveFile("admin", "templates"))

	// assets and components
	adminGroup.GET("/assets/*filepath", serveFile("assets"))
	adminGroup.HEAD("/assets/*filepath", serveFile("assets"))
	adminGroup.GET("/components/*filepath", serveFile("components"))
	adminGroup.HEAD("/components/*filepath", serveFile("components"))

	// 管理前端入口公开：允许未认证用户加载登录页面
	adminGroup.GET("/", func(c *gin.Context) {
		static.ServeAdminPage(c, cfg)
	})
	// HEAD for admin entry
	adminGroup.HEAD("/", func(c *gin.Context) {
		static.ServeAdminPage(c, cfg)
	})

	// 使用复用的中间件实现（JWT 用户认证并要求 admin 角色）
	combinedAuthMiddleware := middleware.CombinedAdminAuth(cfg, userService)

	// 需要管理员认证的API路由组
	authGroup := adminGroup.Group("")
	authGroup.Use(combinedAuthMiddleware)
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

		// 传输日志
		authGroup.GET("/logs/transfer", adminHandler.GetTransferLogs)

		// 用户管理
		setupUserRoutes(authGroup, adminHandler)

		// 存储管理 (需要管理员认证)
		setupStorageRoutes(authGroup, storageHandler)

		// MCP 服务器管理 (需要管理员认证)
		setupMCPRoutes(authGroup, adminHandler)
	}
}

// ServeAdminPage moved to internal/static

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
	authGroup.GET("/maintenance/audit", adminHandler.GetOperationLogs)

	// 任务管理
	authGroup.GET("/maintenance/tasks", adminHandler.GetRunningTasks)
	authGroup.POST("/maintenance/tasks/:id/cancel", adminHandler.CancelTask)
	authGroup.POST("/maintenance/tasks/:id/retry", adminHandler.RetryTask)

	// 系统控制
	authGroup.POST("/maintenance/restart", adminHandler.RestartSystem)
	authGroup.POST("/maintenance/shutdown", adminHandler.ShutdownSystem)
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
	// 批量用户操作
	authGroup.POST("/users/batch-enable", adminHandler.BatchEnableUsers)
	authGroup.POST("/users/batch-disable", adminHandler.BatchDisableUsers)
	authGroup.POST("/users/batch-delete", adminHandler.BatchDeleteUsers)
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
