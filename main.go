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
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/database"
	"github.com/zy84338719/filecodebox/internal/handlers"
	"github.com/zy84338719/filecodebox/internal/mcp"
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/repository"
	"github.com/zy84338719/filecodebox/internal/routes"
	"github.com/zy84338719/filecodebox/internal/services"
	"github.com/zy84338719/filecodebox/internal/storage"
	"github.com/zy84338719/filecodebox/internal/tasks"

	"github.com/sirupsen/logrus"

	// swagger docs
	_ "github.com/zy84338719/filecodebox/docs"
)

func main() {
	// 解析命令行参数
	var showVersion = flag.Bool("version", false, "show version information")
	flag.Parse()

	if *showVersion {
		buildInfo := models.GetBuildInfo()
		fmt.Printf("FileCodeBox %s\n", buildInfo.Version)
		fmt.Printf("Commit: %s\n", buildInfo.GitCommit)
		fmt.Printf("Built: %s\n", buildInfo.BuildTime)
		fmt.Printf("Go Version: %s\n", runtime.Version())
		return
	}

	// 初始化日志
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logrus.Info("正在初始化应用...")

	// 初始化新的配置管理器
	manager := config.InitManager()

	// 初始化数据库 - 使用manager而不是cfg
	db, err := database.InitWithManager(manager)
	if err != nil {
		logrus.Fatal("初始化数据库失败:", err)
	}

	// 自动迁移
	err = db.AutoMigrate(&models.FileCode{},
		&models.UploadChunk{}, &models.KeyValue{}, &models.User{}, &models.UserSession{})
	if err != nil {
		logrus.Fatal("数据库迁移失败:", err)
	}

	// 使用数据库初始化配置管理器
	if err := manager.InitWithDB(db); err != nil {
		logrus.Fatal("初始化配置管理器失败:", err)
	}

	// 初始化存储
	storageManager := storage.NewStorageManager(manager)

	// 初始化 DAO 管理器
	daoManager := repository.NewRepositoryManager(db)

	// 创建具体的存储服务
	storageService := storage.NewConcreteStorageService(manager)

	// 初始化服务（为了MCP服务器）
	userService := services.NewUserService(daoManager, manager)
	shareService := services.NewShareService(daoManager, manager, storageService, userService)
	adminService := services.NewAdminService(daoManager, manager, storageService)

	// 初始化清理任务
	taskManager := tasks.NewTaskManager(daoManager, storageManager, manager.Base.DataPath)
	taskManager.Start()
	defer taskManager.Stop()

	// 初始化 MCP 管理器
	mcpManager := mcp.NewMCPManager(manager, daoManager, storageManager, shareService, adminService, userService)

	// 设置全局 MCP 管理器（供 admin handler 使用）
	handlers.SetMCPManager(mcpManager)

	// 创建并配置路由（包含Gin初始化、中间件、路由设置）
	srv, err := routes.CreateAndStartServer(manager, daoManager, storageManager)
	if err != nil {
		logrus.Fatal("创建服务器失败:", err)
	}

	// 根据配置启动 MCP 服务器
	if manager.MCP.EnableMCPServer == 1 {
		if err := mcpManager.StartMCPServer(manager.MCP.MCPPort); err != nil {
			logrus.Fatal("启动 MCP 服务器失败: ", err)
		}
		logrus.Info("MCP 服务器已在主程序启动时自动启动")
	}

	logrus.Info("应用初始化完成")

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("正在关闭服务器...")

	// 优雅关闭
	if err := routes.GracefulShutdown(srv, 30*time.Second); err != nil {
		logrus.Fatal("关闭服务器失败:", err)
	}
	// 关闭数据库连接
	if sqlDB, err := db.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			logrus.Error("关闭数据库连接失败:", err)
		}
	}
}
