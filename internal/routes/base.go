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
func SetupBaseRoutes(router *gin.Engine, cfg *config.Config) {
	// 静态文件服务
	router.Static("/assets", fmt.Sprintf("./%s/assets", cfg.ThemesSelect))

	// Swagger 文档路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API文档和健康检查
	apiHandler := handlers.NewAPIHandler()
	router.GET("/health", apiHandler.GetHealth)

	// 首页和静态页面
	router.GET("/", func(c *gin.Context) {
		ServeIndex(c, cfg)
	})

	router.NoRoute(func(c *gin.Context) {
		ServeIndex(c, cfg)
	})

	// robots.txt
	router.GET("/robots.txt", func(c *gin.Context) {
		c.Header("Content-Type", "text/plain")
		c.String(http.StatusOK, cfg.RobotsText)
	})

	// 获取配置接口
	router.POST("/", func(c *gin.Context) {
		common.SuccessResponse(c, gin.H{
			"name":               cfg.Name,
			"description":        cfg.Description,
			"explain":            cfg.PageExplain,
			"uploadSize":         cfg.UploadSize,
			"expireStyle":        cfg.ExpireStyle,
			"enableChunk":        GetEnableChunk(cfg),
			"openUpload":         cfg.OpenUpload,
			"notify_title":       cfg.NotifyTitle,
			"notify_content":     cfg.NotifyContent,
			"show_admin_address": cfg.ShowAdminAddr,
			"max_save_seconds":   cfg.MaxSaveSeconds,
		})
	})
}

// ServeIndex 服务首页
func ServeIndex(c *gin.Context, cfg *config.Config) {
	indexPath := filepath.Join(".", cfg.ThemesSelect, "index.html")

	content, err := os.ReadFile(indexPath)
	if err != nil {
		c.String(http.StatusNotFound, "Index file not found")
		return
	}

	html := string(content)
	// 替换模板变量
	html = strings.ReplaceAll(html, "{{title}}", cfg.Name)
	html = strings.ReplaceAll(html, "{{description}}", cfg.Description)
	html = strings.ReplaceAll(html, "{{keywords}}", cfg.Keywords)
	html = strings.ReplaceAll(html, "{{page_explain}}", cfg.PageExplain)
	html = strings.ReplaceAll(html, "{{opacity}}", fmt.Sprintf("%.1f", cfg.Opacity))
	html = strings.ReplaceAll(html, `"/assets/`, `"assets/`)
	html = strings.ReplaceAll(html, "{{background}}", cfg.Background)

	c.Header("Cache-Control", "no-cache")
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// GetEnableChunk 获取分片上传配置
func GetEnableChunk(cfg *config.Config) int {
	if cfg.FileStorage == "local" && cfg.EnableChunk == 1 {
		return 1
	}
	return 0
}
