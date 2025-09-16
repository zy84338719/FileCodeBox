package routes

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/handlers"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupBaseRoutes 设置基础路由（首页、健康检查、静态文件等）
func SetupBaseRoutes(router *gin.Engine, userHandler *handlers.UserHandler, cfg *config.ConfigManager) {
	// 静态文件服务 - 统一挂载所有前端资源
	themeDir := fmt.Sprintf("./%s", cfg.ThemesSelect)

	// 挂载主要静态资源
	router.Static("/assets", fmt.Sprintf("%s/assets", themeDir))

	// 挂载组件化CSS文件
	router.Static("/css", fmt.Sprintf("%s/css", themeDir))

	// 挂载组件化JS文件
	router.Static("/js", fmt.Sprintf("%s/js", themeDir))

	// 挂载组件目录（如果存在）
	router.Static("/components", fmt.Sprintf("%s/components", themeDir))

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
		ServeIndex(c, cfg)
	})

	// 系统初始化页面
	router.GET("/setup", func(c *gin.Context) {
		ServeSetup(c, cfg)
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
		ServeIndex(c, cfg)
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

// ServeIndex 服务首页
func ServeIndex(c *gin.Context, cfg *config.ConfigManager) {
	indexPath := filepath.Join(".", cfg.ThemesSelect, "index.html")

	content, err := os.ReadFile(indexPath)
	if err != nil {
		c.String(http.StatusNotFound, "Index file not found")
		return
	}

	html := string(content)
	// 替换模板变量
	html = strings.ReplaceAll(html, "{{title}}", cfg.Base.Name)
	html = strings.ReplaceAll(html, "{{description}}", cfg.Base.Description)
	html = strings.ReplaceAll(html, "{{keywords}}", cfg.Base.Keywords)
	html = strings.ReplaceAll(html, "{{page_explain}}", cfg.PageExplain)
	html = strings.ReplaceAll(html, "{{opacity}}", fmt.Sprintf("%.1f", cfg.Opacity))
	// 将相对路径转换为绝对路径，避免在子路径下请求相对路径（例如 /user/login -> /user/js/...）
	html = strings.ReplaceAll(html, "src=\"js/", "src=\"/js/")
	html = strings.ReplaceAll(html, "href=\"css/", "href=\"/css/")
	html = strings.ReplaceAll(html, "src=\"assets/", "src=\"/assets/")
	html = strings.ReplaceAll(html, "href=\"assets/", "href=\"/assets/")
	html = strings.ReplaceAll(html, "src=\"components/", "src=\"/components/")
	html = strings.ReplaceAll(html, "{{background}}", cfg.Background)

	c.Header("Cache-Control", "no-cache")
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// ServeSetup 服务系统初始化页面
func ServeSetup(c *gin.Context, cfg *config.ConfigManager) {
	setupPath := filepath.Join(".", cfg.ThemesSelect, "setup.html")

	content, err := os.ReadFile(setupPath)
	if err != nil {
		c.String(http.StatusNotFound, "Setup page not found")
		return
	}

	html := string(content)
	// 替换模板变量
	html = strings.ReplaceAll(html, "{{title}}", cfg.Base.Name+" - 系统初始化")
	html = strings.ReplaceAll(html, "{{description}}", cfg.Base.Description)
	html = strings.ReplaceAll(html, "{{keywords}}", cfg.Base.Keywords)
	// 将相对资源路径转换为绝对路径
	html = strings.ReplaceAll(html, "src=\"js/", "src=\"/js/")
	html = strings.ReplaceAll(html, "href=\"css/", "href=\"/css/")
	html = strings.ReplaceAll(html, "src=\"assets/", "src=\"/assets/")

	c.Header("Cache-Control", "no-cache")
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// GetEnableChunk 获取分片上传配置
func GetEnableChunk(cfg *config.ConfigManager) int {
	if cfg.Storage.Type == "local" && cfg.Transfer.Upload.EnableChunk == 1 {
		return 1
	}
	return 0
}
