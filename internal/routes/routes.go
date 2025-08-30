package routes

import (
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/handlers"
	"github.com/zy84338719/filecodebox/internal/middleware"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(
	router *gin.Engine,
	shareHandler *handlers.ShareHandler,
	chunkHandler *handlers.ChunkHandler,
	adminHandler *handlers.AdminHandler,
	storageHandler *handlers.StorageHandler,
	cfg *config.Config,
) {
	// API文档和健康检查
	apiHandler := handlers.NewAPIHandler()
	router.GET("/api/doc", apiHandler.GetAPIDoc)
	router.GET("/health", apiHandler.GetHealth)

	// 首页和静态页面
	router.GET("/", func(c *gin.Context) {
		serveIndex(c, cfg)
	})

	router.NoRoute(func(c *gin.Context) {
		serveIndex(c, cfg)
	})

	// robots.txt
	router.GET("/robots.txt", func(c *gin.Context) {
		c.Header("Content-Type", "text/plain")
		c.String(http.StatusOK, cfg.RobotsText)
	})

	// 获取配置接口
	router.POST("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "success",
			"detail": gin.H{
				"name":               cfg.Name,
				"description":        cfg.Description,
				"explain":            cfg.PageExplain,
				"uploadSize":         cfg.UploadSize,
				"expireStyle":        cfg.ExpireStyle,
				"enableChunk":        getEnableChunk(cfg),
				"openUpload":         cfg.OpenUpload,
				"notify_title":       cfg.NotifyTitle,
				"notify_content":     cfg.NotifyContent,
				"show_admin_address": cfg.ShowAdminAddr,
				"max_save_seconds":   cfg.MaxSaveSeconds,
			},
		})
	})

	// 分享相关路由
	shareGroup := router.Group("/share")
	shareGroup.Use(middleware.ShareAuth(cfg))
	{
		shareGroup.POST("/text/", shareHandler.ShareText)
		shareGroup.POST("/file/", shareHandler.ShareFile)
		shareGroup.GET("/select/", shareHandler.GetFile)
		shareGroup.POST("/select/", shareHandler.GetFile)
		shareGroup.GET("/download", shareHandler.DownloadFile)
	}

	// 分片上传相关路由
	chunkGroup := router.Group("/chunk")
	chunkGroup.Use(middleware.ShareAuth(cfg))
	{
		chunkGroup.POST("/upload/init/", chunkHandler.InitChunkUpload)
		chunkGroup.POST("/upload/chunk/:upload_id/:chunk_index", chunkHandler.UploadChunk)
		chunkGroup.POST("/upload/complete/:upload_id", chunkHandler.CompleteUpload)

		// 断点续传相关路由
		chunkGroup.GET("/upload/status/:upload_id", chunkHandler.GetUploadStatus)
		chunkGroup.POST("/upload/verify/:upload_id/:chunk_index", chunkHandler.VerifyChunk)
		chunkGroup.DELETE("/upload/cancel/:upload_id", chunkHandler.CancelUpload)
	}

	// 管理相关路由
	adminGroup := router.Group("/admin")
	{
		// 管理页面
		adminGroup.GET("/", func(c *gin.Context) {
			serveAdminPage(c, cfg)
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
			authGroup.GET("/files/:id", adminHandler.GetFile)
			authGroup.DELETE("/files/:id", adminHandler.DeleteFile)
			authGroup.PUT("/files/:id", adminHandler.UpdateFile)
			authGroup.GET("/files/download", adminHandler.DownloadFile)
			authGroup.GET("/config", adminHandler.GetConfig)
			authGroup.PUT("/config", adminHandler.UpdateConfig)
			authGroup.POST("/clean", adminHandler.CleanExpiredFiles)

			// 存储管理相关路由
			authGroup.GET("/storage", storageHandler.GetStorageInfo)
			authGroup.POST("/storage/switch", storageHandler.SwitchStorage)
			authGroup.GET("/storage/test/:type", storageHandler.TestStorageConnection)
			authGroup.PUT("/storage/config", storageHandler.UpdateStorageConfig)
		}
	}
}

// serveIndex 服务首页
func serveIndex(c *gin.Context, cfg *config.Config) {
	indexPath := filepath.Join(".", cfg.ThemesSelect, "index.html")

	content, err := ioutil.ReadFile(indexPath)
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

// serveAdminPage 服务管理页面
func serveAdminPage(c *gin.Context, cfg *config.Config) {
	adminPath := filepath.Join(".", cfg.ThemesSelect, "admin.html")

	content, err := ioutil.ReadFile(adminPath)
	if err != nil {
		c.String(http.StatusNotFound, "Admin page not found")
		return
	}

	c.Header("Cache-Control", "no-cache")
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, string(content))
}

// getEnableChunk 获取分片上传配置
func getEnableChunk(cfg *config.Config) int {
	if cfg.FileStorage == "local" && cfg.EnableChunk == 1 {
		return 1
	}
	return 0
}
