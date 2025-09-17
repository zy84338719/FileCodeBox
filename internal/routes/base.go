package routes

import (
	"net/http"

	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/handlers"
	"github.com/zy84338719/filecodebox/internal/static"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupBaseRoutes 设置基础路由（首页、健康检查、静态文件等）
func SetupBaseRoutes(router *gin.Engine, userHandler *handlers.UserHandler, cfg *config.ConfigManager) {
	// 静态文件服务 - 统一挂载所有前端资源
	static.RegisterStaticRoutes(router, cfg)

	// Swagger 文档路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API文档和健康检查
	apiHandler := handlers.NewAPIHandler(cfg)
	router.GET("/health", apiHandler.GetHealth)

	// API 配置路由
	api := router.Group("/api")
	{
		api.GET("/config", apiHandler.GetConfig)
	}

	// 首页和静态页面
	router.GET("/", func(c *gin.Context) {
		// 检查系统是否已初始化
		if userHandler != nil {
			initialized, err := userHandler.IsSystemInitialized()
			if err == nil && !initialized {
				// 系统未初始化，重定向到setup页面
				c.Redirect(302, "/setup")
				return
			}
		}
		static.ServeIndex(c, cfg)
	})

	// 系统初始化页面
	router.GET("/setup", func(c *gin.Context) {
		static.ServeSetup(c, cfg)
	})

	// 永远注册 /user/system-info 接口：
	// - 明确返回 JSON（即使在未初始化数据库时也不会返回 HTML）
	// - 如果传入了 userHandler（数据库已初始化），则委托给 userHandler.GetSystemInfo
	// - 否则返回一个轻量的 JSON 响应，避免返回 HTML 导致前端解析失败
	router.GET("/user/system-info", func(c *gin.Context) {
		// 明确设置 JSON 响应头，避免被其他中间件或 NoRoute 覆盖成 HTML
		c.Header("Cache-Control", "no-cache")
		c.Header("Content-Type", "application/json; charset=utf-8")

		if userHandler != nil {
			// Delegate to the real handler which also writes JSON
			userHandler.GetSystemInfo(c)
			// Ensure no further handlers run
			c.Abort()
			return
		}

		// 返回轻量的 JSON 响应（与前端兼容）
		// 返回与后端其他字段类型一致的整数值（0/1），避免前端对布尔/整型的解析差异
		allowReg := 0
		if cfg.User.AllowUserRegistration == 1 {
			allowReg = 1
		}
		c.JSON(200, gin.H{
			"code": 200,
			"data": gin.H{
				"user_system_enabled":     1,
				"allow_user_registration": allowReg,
			},
		})
		c.Abort()
	})

	// 兼容：在未初始化数据库时，允许 POST /setup 用于提交扁平表单风格的初始化请求
	if cfg != nil && cfg.GetDB() == nil {
		router.POST("/setup", handlers.InitializeNoDB(cfg))
	}

	router.NoRoute(func(c *gin.Context) {
		static.ServeIndex(c, cfg)
	})

	// robots.txt
	router.GET("/robots.txt", func(c *gin.Context) {
		c.Header("Content-Type", "text/plain")
		c.String(http.StatusOK, cfg.RobotsText)
	})

	// 获取配置接口（兼容性保留）
	router.POST("/", func(c *gin.Context) {
		common.SuccessResponse(c, gin.H{
			"name":               cfg.Base.Name,
			"description":        cfg.Base.Description,
			"explain":            cfg.PageExplain,
			"uploadSize":         cfg.Transfer.Upload.UploadSize,
			"expireStyle":        cfg.ExpireStyle,
			"enableChunk":        GetEnableChunk(cfg),
			"openUpload":         cfg.Transfer.Upload.OpenUpload,
			"notify_title":       cfg.NotifyTitle,
			"notify_content":     cfg.NotifyContent,
			"show_admin_address": cfg.ShowAdminAddr,
			"max_save_seconds":   cfg.Transfer.Upload.MaxSaveSeconds,
		})
	})
}

// ServeIndex/ServeSetup moved to internal/static for central management

// GetEnableChunk 获取分片上传配置
func GetEnableChunk(cfg *config.ConfigManager) int {
	if cfg.Storage.Type == "local" && cfg.Transfer.Upload.EnableChunk == 1 {
		return 1
	}
	return 0
}
