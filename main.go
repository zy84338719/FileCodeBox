package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/database"
	"github.com/zy84338719/filecodebox/internal/handlers"
	"github.com/zy84338719/filecodebox/internal/middleware"
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/routes"
	"github.com/zy84338719/filecodebox/internal/services"
	"github.com/zy84338719/filecodebox/internal/storage"
	"github.com/zy84338719/filecodebox/internal/tasks"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// 解析命令行参数
	var showVersion = flag.Bool("version", false, "show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("FileCodeBox %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Built: %s\n", date)
		fmt.Printf("Go Version: %s\n", "go1.21+")
		return
	}

	// 初始化配置
	cfg := config.Init()

	// 初始化日志
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logrus.Info("正在初始化应用...")

	// 初始化数据库
	db, err := database.Init(cfg.DataPath + "/filecodebox.db")
	if err != nil {
		logrus.Fatal("初始化数据库失败:", err)
	}

	// 自动迁移（需要在配置初始化前进行）
	err = db.AutoMigrate(&models.FileCode{}, &models.UploadChunk{}, &models.KeyValue{})
	if err != nil {
		logrus.Fatal("数据库迁移失败:", err)
	}

	// 使用数据库初始化配置
	if err := cfg.InitWithDB(db); err != nil {
		logrus.Fatal("初始化配置失败:", err)
	}

	// 初始化存储
	storageManager := storage.NewStorageManager(cfg)

	// 初始化服务
	shareService := services.NewShareService(db, storageManager, cfg)
	chunkService := services.NewChunkService(db, storageManager, cfg)
	adminService := services.NewAdminService(db, cfg, storageManager)

	// 初始化处理器
	shareHandler := handlers.NewShareHandler(shareService)
	chunkHandler := handlers.NewChunkHandler(chunkService)
	adminHandler := handlers.NewAdminHandler(adminService, cfg)
	storageHandler := handlers.NewStorageHandler(storageManager, cfg)

	// 初始化清理任务
	taskManager := tasks.NewTaskManager(db, storageManager, cfg.DataPath)
	taskManager.Start()
	defer taskManager.Stop()

	// 设置Gin模式
	if cfg.Production {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := gin.New()

	// 添加中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimit(cfg))

	// 静态文件服务
	router.Static("/assets", fmt.Sprintf("./%s/assets", cfg.ThemesSelect))

	// 设置路由
	routes.SetupRoutes(router, shareHandler, chunkHandler, adminHandler, storageHandler, cfg)

	logrus.Info("应用初始化完成")

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:           router,
		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// 启动服务器
	go func() {
		logrus.Infof("服务器启动在 %s:%d", cfg.Host, cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatal("启动服务器失败:", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("正在关闭服务器...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.Fatal("服务器强制关闭:", err)
	}

	logrus.Info("服务器已关闭")
}
