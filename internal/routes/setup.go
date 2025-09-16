package routes

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/handlers"
	"github.com/zy84338719/filecodebox/internal/middleware"
	"github.com/zy84338719/filecodebox/internal/repository"
	"github.com/zy84338719/filecodebox/internal/services"
	"github.com/zy84338719/filecodebox/internal/storage"

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
			ServeUserPage(c, manager, "login.html")
		})
		router.GET("/user/register", func(c *gin.Context) {
			ServeUserPage(c, manager, "register.html")
		})
		router.GET("/user/forgot-password", func(c *gin.Context) {
			ServeUserPage(c, manager, "forgot-password.html")
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
	// 创建具体的存储服务
	storageService := storage.NewConcreteStorageService(manager)

	// 初始化服务
	userService := services.NewUserService(daoManager, manager)
	shareServiceInstance := services.NewShareService(daoManager, manager, storageService, userService)
	chunkService := services.NewChunkService(daoManager, manager, storageService)
	adminService := services.NewAdminService(daoManager, manager, storageService)

	// 初始化处理器
	shareHandler := handlers.NewShareHandler(shareServiceInstance)
	chunkHandler := handlers.NewChunkHandler(chunkService)
	adminHandler := handlers.NewAdminHandler(adminService, manager)
	storageHandler := handlers.NewStorageHandler(storageManager, manager.Storage, manager)
	userHandler := handlers.NewUserHandler(userService)
	// 设置分享、用户、分片、管理员等路由（不重复注册基础路由）
	// 注意：SetupAllRoutes 会调用 SetupBaseRoutes，因此我们直接调用 SetupShareRoutes 等单独函数
	SetupShareRoutes(router, shareHandler, manager, userService)
	SetupUserRoutes(router, userHandler, manager, userService)
	SetupChunkRoutes(router, chunkHandler, manager)
	SetupAdminRoutes(router, adminHandler, storageHandler, manager, userService)
	// System init routes are no longer needed after DB init
}

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
