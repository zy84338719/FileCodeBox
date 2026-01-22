package main

// @title FileCodeBox API
// @version 1.9.8
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
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zy84338719/filecodebox/internal/cli"
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/database"
	"github.com/zy84338719/filecodebox/internal/handlers"
	"github.com/zy84338719/filecodebox/internal/mcp"
	"github.com/zy84338719/filecodebox/internal/repository"
	"github.com/zy84338719/filecodebox/internal/routes"
	"github.com/zy84338719/filecodebox/internal/services"
	"github.com/zy84338719/filecodebox/internal/static"
	"github.com/zy84338719/filecodebox/internal/storage"
	"github.com/zy84338719/filecodebox/internal/tasks"

	"github.com/sirupsen/logrus"

	// swagger docs
	_ "github.com/zy84338719/filecodebox/docs"
)

func main() {
	// 初始化嵌入的静态文件
	static.SetEmbeddedFS(EmbeddedThemes)

	// 如果有子命令参数，切换到 CLI 模式（使用 Cobra）
	if len(os.Args) > 1 {
		// delay import of CLI to avoid cycles
		cli.Execute()
		return
	}

	// 初始化日志
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	logrus.Info("正在初始化应用...")

	// 使用上下文管理生命周期
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 初始化配置管理器
	manager := config.InitManager()

	// 延迟数据库初始化：不在启动时创建 DB，让用户通过 /setup/initialize 触发

	// 初始化存储（不依赖数据库）
	storageManager := storage.NewStorageManager(manager)

	// 创建并启动最小 HTTP 服务器（daoManager 传 nil）
	var daoManager *repository.RepositoryManager = nil
	srv, err := routes.CreateAndStartServer(manager, daoManager, storageManager)
	if err != nil {
		logrus.Fatalf("创建服务器失败: %v", err)
	}

	// 从 srv.Handler 获取 router（gin 引擎），用于动态注册路由
	routerEngine := srv.Handler

	// 设置 OnDatabaseInitialized 回调：当 /setup/initialize 完成数据库初始化后会调用此回调
	handlers.OnDatabaseInitialized = func(dmgr *repository.RepositoryManager) {
		// 这里创建 DB 相关的服务、任务与 MCP，并动态注册路由
		logrus.Info("收到数据库初始化完成回调，开始挂载动态路由与启动后台服务")

		// 创建具体的存储服务（基于 manager）
		storageService := storage.NewConcreteStorageService(manager)

		// 初始化服务
		userService := services.NewUserService(dmgr, manager)
		shareService := services.NewShareService(dmgr, manager, storageService, userService)
		adminService := services.NewAdminService(dmgr, manager, storageService)

		// 启动任务管理器
		taskManager := tasks.NewTaskManager(dmgr, storageManager, adminService, manager.Base.DataPath)
		taskManager.Start()
		// 注意：taskManager 的停止将在主结束时处理（可以扩展保存引用以便停止）

		// 初始化 MCP 管理器并根据配置启动
		mcpManager := mcp.NewMCPManager(manager, dmgr, storageManager, shareService, adminService, userService)
		handlers.SetMCPManager(mcpManager)
		if manager.MCP.EnableMCPServer == 1 {
			if err := mcpManager.StartMCPServer(manager.MCP.MCPPort); err != nil {
				logrus.Errorf("启动 MCP 服务器失败: %v", err)
			} else {
				logrus.Info("MCP 服务器已启动")
			}
		}

		// 将 DAO 底层 DB 注入 manager
		if dmdb := dmgr.DB(); dmdb != nil {
			manager.SetDB(dmdb)
		}

		// 动态注册需要数据库支持的路由
		if ginEngine, ok := routerEngine.(*gin.Engine); ok {
			routes.RegisterDynamicRoutes(ginEngine, manager, dmgr, storageManager)
		} else {
			logrus.Warn("无法获取 gin 引擎实例，动态路由未注册")
		}
	}

	logrus.Info("应用初始化完成")

	// 如果 data/filecodebox.db 已经存在，尝试提前初始化数据库并注册动态路由，
	// 这样在已初始化环境下，API（如 POST /user/login）能直接可用，而不是返回静态 HTML。
	// 这对于用户已经自行初始化数据库但以 "daoManager == nil" 启动的场景非常重要。
	// 尝试多路径检测数据库文件：优先使用 manager.Database.Name（若配置了），其次尝试基于 Base.DataPath 的常见位置
	var candidates []string
	if manager.Database.Name != "" {
		candidates = append(candidates, manager.Database.Name)
	}
	if manager.Base != nil && manager.Base.DataPath != "" {
		candidates = append(candidates, filepath.Join(manager.Base.DataPath, "filecodebox.db"))
		// 如果 manager.Database.Name 是相对路径，尝试基于 DataPath 拼接
		if manager.Database.Name != "" && !filepath.IsAbs(manager.Database.Name) {
			candidates = append(candidates, filepath.Join(manager.Base.DataPath, manager.Database.Name))
		}
	}

	logrus.Infof("数据库检测候选路径: %v", candidates)
	found := ""
	for _, dbFile := range candidates {
		if dbFile == "" {
			continue
		}
		if _, err := os.Stat(dbFile); err == nil {
			found = dbFile
			break
		}
	}

	if found != "" {
		logrus.Infof("检测到已有数据库文件，尝试提前初始化数据库: %s", found)
		// 如果 Base.DataPath 为空，设置为 found 的目录，避免 database.InitWithManager 在创建目录时使用空字符串
		if manager.Base == nil {
			manager.Base = &config.BaseConfig{}
		}
		if manager.Base.DataPath == "" {
			dir := filepath.Dir(found)
			if dir == "." || dir == "" {
				dir = "./data"
			}
			if abs, err := filepath.Abs(dir); err == nil {
				manager.Base.DataPath = abs
			} else {
				manager.Base.DataPath = dir
			}
			logrus.Infof("为 manager.Base.DataPath 赋值: %s", manager.Base.DataPath)
		}

		if db, err := database.InitWithManager(manager); err == nil {
			if db != nil {
				manager.SetDB(db)
				dmgr := repository.NewRepositoryManager(db)
				if ginEngine, ok := routerEngine.(*gin.Engine); ok {
					routes.RegisterDynamicRoutes(ginEngine, manager, dmgr, storageManager)
					logrus.Info("动态路由已注册（基于已存在的数据库）")
				}
			}
		} else {
			logrus.Warnf("尝试提前初始化数据库失败: %v", err)
		}
	}

	// 等待中断信号，优雅退出
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		logrus.Info("上下文已取消，开始关闭...")
	case sig := <-sigCh:
		logrus.Infof("收到信号 %v，开始关闭...", sig)
	}

	// 优雅关闭 HTTP 服务器
	if err := routes.GracefulShutdown(srv, 30*time.Second); err != nil {
		logrus.Errorf("关闭服务器失败: %v", err)
	}

	// 关闭数据库连接（如果已初始化）
	if dbPtr := manager.GetDB(); dbPtr != nil {
		if sqlDB, err := dbPtr.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				logrus.Errorf("关闭数据库连接失败: %v", err)
			}
		} else {
			logrus.Errorf("获取数据库底层连接失败: %v", err)
		}
	}
}
