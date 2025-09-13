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
	html = strings.ReplaceAll(html, `"/assets/`, `"assets/`)
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
	html = strings.ReplaceAll(html, `"/assets/`, `"assets/`)

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
