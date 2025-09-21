package routes

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/handlers"
	"github.com/zy84338719/filecodebox/internal/mcp"
	"github.com/zy84338719/filecodebox/internal/middleware"
	"github.com/zy84338719/filecodebox/internal/repository"
	"github.com/zy84338719/filecodebox/internal/services"
	"github.com/zy84338719/filecodebox/internal/static"
	"github.com/zy84338719/filecodebox/internal/storage"

	"sync"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// CreateAndStartServer 创建并启动完整的HTTP服务器
func CreateAndStartServer(
	manager *config.ConfigManager,
	daoManager *repository.RepositoryManager,
	storageManager *storage.StorageManager,
) (*http.Server, error) {
	// 创建并配置路由
	router := CreateAndSetupRouter(manager, daoManager, storageManager)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", manager.Base.Host, manager.Base.Port),
		Handler:           router,
		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// 在后台启动服务器
	go func() {
		logrus.Infof("HTTP服务器启动在 %s:%d", manager.Base.Host, manager.Base.Port)
		logrus.Infof("访问地址: http://%s:%d", manager.Base.Host, manager.Base.Port)
		logrus.Infof("管理后台: http://%s:%d/admin/", manager.Base.Host, manager.Base.Port)
		logrus.Infof("API文档: http://%s:%d/swagger/index.html", manager.Base.Host, manager.Base.Port)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Fatalf("HTTP服务器启动失败: %v", err)
		}
	}()

	return srv, nil
}

// GracefulShutdown 优雅关闭服务器
func GracefulShutdown(srv *http.Server, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("服务器强制关闭: %w", err)
	}

	logrus.Info("服务器已关闭")
	return nil
}

// CreateAndSetupRouter 创建并完全配置Gin引擎
func CreateAndSetupRouter(
	manager *config.ConfigManager,
	daoManager *repository.RepositoryManager,
	storageManager *storage.StorageManager,
) *gin.Engine {
	// 设置Gin模式
	if manager.Base.Production {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := gin.New()

	// 添加中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimit(manager))

	// 如果 daoManager 为 nil，表示尚未初始化数据库，只注册基础和初始化相关的路由
	if daoManager == nil {
		// 基础路由（不传 userHandler）
		SetupBaseRoutes(router, nil, manager)

		// 提供一个不依赖数据库的初始化 POST 接口，由 handlers.InitializeNoDB 处理
		router.POST("/setup/initialize", handlers.InitializeNoDB(manager))

		// 即便数据库尚未初始化，也应当能访问用户登录/注册页面（只返回静态HTML），
		// 以便用户能够在首次部署时完成初始化或查看登录页面。
		router.GET("/user/login", func(c *gin.Context) {
			static.ServeUserPage(c, manager, "login.html")
		})
		router.GET("/user/register", func(c *gin.Context) {
			static.ServeUserPage(c, manager, "register.html")
		})
		// 在未初始化数据库时，也允许访问用户仪表板静态页面，避免被 NoRoute 回退到首页
		router.GET("/user/dashboard", func(c *gin.Context) {
			static.ServeUserPage(c, manager, "dashboard.html")
		})
		router.GET("/user/forgot-password", func(c *gin.Context) {
			static.ServeUserPage(c, manager, "forgot-password.html")
		})

		// 在未初始化数据库时，不直接注册真实的 POST /admin/login 处理器以避免后续动态注册冲突。
		// 这里注册一个轻量的委派处理器：当 admin_handler 被注入到全局 app state（动态注册完成）时，
		// 它会尝试调用真实的 handler；否则返回明确的 JSON 错误，提示调用 /setup/initialize。
		router.POST("/admin/login", func(c *gin.Context) {
			// 如果动态注入了 admin_handler（在 RegisterDynamicRoutes 中注入），
			// 使用全局注入的 handler（通过 handlers.GetInjectedAdminHandler）进行委派。
			if injected := handlers.GetInjectedAdminHandler(); injected != nil {
				injected.Login(c)
				return
			}
			// 否则返回 JSON 提示，说明数据库尚未初始化
			c.JSON(404, gin.H{"code": 404, "message": "admin 登录不可用：数据库尚未初始化，请调用 /setup/initialize 完成初始化"})
		})

		return router
	}

	// 设置路由（自动初始化所有服务和处理器）
	SetupAllRoutesWithDependencies(router, manager, daoManager, storageManager)

	return router
}

// SetupAllRoutesWithDependencies 从依赖项初始化并设置所有路由
func SetupAllRoutesWithDependencies(
	router *gin.Engine,
	manager *config.ConfigManager,
	daoManager *repository.RepositoryManager,
	storageManager *storage.StorageManager,
) {
	// 创建具体的存储服务
	storageService := storage.NewConcreteStorageService(manager)

	// 初始化服务
	userService := services.NewUserService(daoManager, manager)                                        // 先初始化用户服务
	shareServiceInstance := services.NewShareService(daoManager, manager, storageService, userService) // 使用带用户服务的分享服务
	chunkService := services.NewChunkService(daoManager, manager, storageService)                      // 使用新的存储服务
	adminService := services.NewAdminService(daoManager, manager, storageService)                      // 使用新的存储服务

	// 初始化处理器
	shareHandler := handlers.NewShareHandler(shareServiceInstance)
	chunkHandler := handlers.NewChunkHandler(chunkService)
	adminHandler := handlers.NewAdminHandler(adminService, manager)
	storageHandler := handlers.NewStorageHandler(storageManager, manager.Storage, manager)
	userHandler := handlers.NewUserHandler(userService)
	setupHandler := handlers.NewSetupHandler(daoManager, manager)

	// 设置所有路由
	SetupAllRoutes(router, shareHandler, chunkHandler, adminHandler, storageHandler, userHandler, setupHandler, manager, userService)
}

// RegisterDynamicRoutes 在数据库可用后注册需要数据库的路由（不包含基础路由）
func RegisterDynamicRoutes(
	router *gin.Engine,
	manager *config.ConfigManager,
	daoManager *repository.RepositoryManager,
	storageManager *storage.StorageManager,
) {
	// 序列化注册，防止并发导致的重复注册 panic
	dynamicRegisterMu.Lock()
	defer dynamicRegisterMu.Unlock()

	// 使用原子标志防止重复调用
	if registerDynamicOnce() {
		logrus.Info("动态路由已注册（atomic），跳过 RegisterDynamicRoutes")
		return
	}
	// 如果动态路由已经注册（例如 /share/text/ 已存在），则跳过注册以防止重复注册导致 panic
	for _, r := range router.Routes() {
		if r.Method == "POST" && r.Path == "/share/text/" {
			logrus.Info("动态路由已存在，跳过 RegisterDynamicRoutes")
			return
		}
	}
	// 创建具体的存储服务
	storageService := storage.NewConcreteStorageService(manager)

	// 初始化服务
	userService := services.NewUserService(daoManager, manager)
	shareServiceInstance := services.NewShareService(daoManager, manager, storageService, userService)
	chunkService := services.NewChunkService(daoManager, manager, storageService)
	adminService := services.NewAdminService(daoManager, manager, storageService)

	// 重新初始化 MCP 管理器并根据配置启动（确保动态路由注册时MCP管理器可用）
	mcpManager := mcp.NewMCPManager(manager, daoManager, storageManager, shareServiceInstance, adminService, userService)
	handlers.SetMCPManager(mcpManager)
	if manager.MCP.EnableMCPServer == 1 {
		if err := mcpManager.StartMCPServer(manager.MCP.MCPPort); err != nil {
			logrus.Errorf("启动 MCP 服务器失败: %v", err)
		} else {
			logrus.Info("MCP 服务器已启动")
		}
	} else {
		logrus.Info("MCP 服务器未启用")
	}

	// 初始化处理器
	shareHandler := handlers.NewShareHandler(shareServiceInstance)
	chunkHandler := handlers.NewChunkHandler(chunkService)
	adminHandler := handlers.NewAdminHandler(adminService, manager)
	storageHandler := handlers.NewStorageHandler(storageManager, manager.Storage, manager)
	userHandler := handlers.NewUserHandler(userService)
	// 设置分享、用户、分片、管理员等路由（不重复注册基础路由）
	// 注意：SetupAllRoutes 会调用 SetupBaseRoutes，因此我们直接调用 SetupShareRoutes 等单独函数
	// 将 adminHandler 注入到全局 app state，以便占位路由可以查找并委派
	handlers.SetInjectedAdminHandler(adminHandler)
	SetupShareRoutes(router, shareHandler, manager, userService)
	// Use API-only user routes here to avoid duplicate page route registration
	SetupUserAPIRoutes(router, userHandler, manager, userService)
	SetupChunkRoutes(router, chunkHandler, manager)
	SetupAdminRoutes(router, adminHandler, storageHandler, manager, userService)
	// System init routes are no longer needed after DB init
}

// package-level atomic to ensure RegisterDynamicRoutes runs only once
var dynamicRoutesRegistered int32 = 0

func registerDynamicOnce() bool {
	// 如果已经设置，则返回 true
	if atomic.LoadInt32(&dynamicRoutesRegistered) == 1 {
		return true
	}
	// 尝试设置为 1
	return !atomic.CompareAndSwapInt32(&dynamicRoutesRegistered, 0, 1)
}

// mutex to serialize dynamic route registration
var dynamicRegisterMu sync.Mutex

// SetupAllRoutes 设置所有路由（使用已初始化的处理器）
func SetupAllRoutes(
	router *gin.Engine,
	shareHandler *handlers.ShareHandler,
	chunkHandler *handlers.ChunkHandler,
	adminHandler *handlers.AdminHandler,
	storageHandler *handlers.StorageHandler,
	userHandler *handlers.UserHandler,
	setupHandler *handlers.SetupHandler,
	manager *config.ConfigManager,
	userService interface {
		ValidateToken(string) (interface{}, error)
	},
) {

	// 设置基础路由
	SetupBaseRoutes(router, userHandler, manager)

	// 设置系统初始化路由
	SetupSystemInitRoutes(router, setupHandler, userHandler, manager)

	// 设置分享路由
	SetupShareRoutes(router, shareHandler, manager, userService)

	// 设置用户路由
	SetupUserRoutes(router, userHandler, manager, userService)

	// 设置分片上传路由
	SetupChunkRoutes(router, chunkHandler, manager)

	// 设置管理员路由
	SetupAdminRoutes(router, adminHandler, storageHandler, manager, userService)
}

// SetupSystemInitRoutes 设置系统初始化路由
func SetupSystemInitRoutes(
	router *gin.Engine,
	setupHandler *handlers.SetupHandler,
	userHandler *handlers.UserHandler,
	manager *config.ConfigManager,
) {
	// 系统初始化相关路由
	router.GET("/check-init", userHandler.CheckSystemInitialization)
	router.POST("/setup/initialize", setupHandler.Initialize)
}

// SetupRoutes 设置路由 (保持兼容性，使用已初始化的处理器)
func SetupRoutes(
	router *gin.Engine,
	shareHandler *handlers.ShareHandler,
	chunkHandler *handlers.ChunkHandler,
	adminHandler *handlers.AdminHandler,
	storageHandler *handlers.StorageHandler,
	userHandler *handlers.UserHandler,
	cfg *config.ConfigManager,
	userService interface {
		ValidateToken(string) (interface{}, error)
	},
) {
	// 为兼容性创建一个空的setupHandler
	setupHandler := &handlers.SetupHandler{}

	// 使用新的路由设置函数
	SetupAllRoutes(router, shareHandler, chunkHandler, adminHandler, storageHandler, userHandler, setupHandler, cfg, userService)
}
