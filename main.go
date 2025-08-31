package main

// @title FileCodeBox API
// @version 1.0
// @description FileCodeBox 是一个用于文件分享和代码片段管理的 Web 应用程序
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://github.com/zy84338719/filecodebox/blob/main/LICENSE

// @host localhost:12345
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key

// @securityDefinitions.basic BasicAuth

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/dao"
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

	// swagger imports
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// swagger docs
	_ "github.com/zy84338719/filecodebox/docs"
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
	err = db.AutoMigrate(&models.FileCode{},
		&models.UploadChunk{}, &models.KeyValue{}, &models.User{}, &models.UserSession{})
	if err != nil {
		logrus.Fatal("数据库迁移失败:", err)
	}

	// 使用数据库初始化配置
	if err := cfg.InitWithDB(db); err != nil {
		logrus.Fatal("初始化配置失败:", err)
	}

	// 初始化存储
	storageManager := storage.NewStorageManager(cfg)

	// 初始化 DAO 管理器
	daoManager := dao.NewDAOManager(db)

	// 初始化服务
	userService := services.NewUserService(db, cfg)                                // 先初始化用户服务
	shareService := services.NewShareService(db, storageManager, cfg, userService) // 传入用户服务
	chunkService := services.NewChunkService(db, storageManager, cfg)
	adminService := services.NewAdminService(db, cfg, storageManager)

	// 初始化处理器
	shareHandler := handlers.NewShareHandler(shareService)
	chunkHandler := handlers.NewChunkHandler(chunkService)
	adminHandler := handlers.NewAdminHandler(adminService, cfg)
	storageHandler := handlers.NewStorageHandler(storageManager, cfg)
	userHandler := handlers.NewUserHandler(userService) // 新增用户处理器

	// 初始化清理任务
	taskManager := tasks.NewTaskManager(daoManager, storageManager, cfg.DataPath)
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
	routes.SetupRoutes(router, shareHandler, chunkHandler, adminHandler, storageHandler, userHandler, cfg, userService)

	// 添加 Swagger 路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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

	// 优雅启动和关闭
	go func() {
		logrus.Infof("HTTP服务器启动在 %s:%d", cfg.Host, cfg.Port)
		logrus.Infof("访问地址: http://%s:%d", cfg.Host, cfg.Port)
		logrus.Infof("管理后台: http://%s:%d/admin/", cfg.Host, cfg.Port)
		logrus.Infof("API文档: http://%s:%d/swagger/index.html", cfg.Host, cfg.Port)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Fatalf("HTTP服务器启动失败: %v", err)
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
