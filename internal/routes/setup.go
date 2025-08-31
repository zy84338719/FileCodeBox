package routes

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/dao"
	"github.com/zy84338719/filecodebox/internal/handlers"
	"github.com/zy84338719/filecodebox/internal/middleware"
	"github.com/zy84338719/filecodebox/internal/services"
	"github.com/zy84338719/filecodebox/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// CreateAndStartServer 创建并启动完整的HTTP服务器
func CreateAndStartServer(
	cfg *config.Config,
	daoManager *dao.DAOManager,
	storageManager *storage.StorageManager,
) (*http.Server, error) {
	// 创建并配置路由
	router := CreateAndSetupRouter(cfg, daoManager, storageManager)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:           router,
		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// 在后台启动服务器
	go func() {
		logrus.Infof("HTTP服务器启动在 %s:%d", cfg.Host, cfg.Port)
		logrus.Infof("访问地址: http://%s:%d", cfg.Host, cfg.Port)
		logrus.Infof("管理后台: http://%s:%d/admin/", cfg.Host, cfg.Port)
		logrus.Infof("API文档: http://%s:%d/swagger/index.html", cfg.Host, cfg.Port)

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
	cfg *config.Config,
	daoManager *dao.DAOManager,
	storageManager *storage.StorageManager,
) *gin.Engine {
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

	// 设置路由（自动初始化所有服务和处理器）
	SetupAllRoutesWithDependencies(router, cfg, daoManager, storageManager)

	return router
}

// SetupAllRoutesWithDependencies 从依赖项初始化并设置所有路由
func SetupAllRoutesWithDependencies(
	router *gin.Engine,
	cfg *config.Config,
	daoManager *dao.DAOManager,
	storageManager *storage.StorageManager,
) {
	// 初始化服务
	userService := services.NewUserService(daoManager, cfg)                                // 先初始化用户服务
	shareService := services.NewShareService(daoManager, cfg, storageManager, userService) // 传入用户服务
	chunkService := services.NewChunkService(daoManager, cfg, storageManager)
	adminService := services.NewAdminService(daoManager, cfg, storageManager)

	// 初始化处理器
	shareHandler := handlers.NewShareHandler(shareService)
	chunkHandler := handlers.NewChunkHandler(chunkService)
	adminHandler := handlers.NewAdminHandler(adminService, cfg)
	storageHandler := handlers.NewStorageHandler(storageManager, cfg)
	userHandler := handlers.NewUserHandler(userService)

	// 设置所有路由
	SetupAllRoutes(router, shareHandler, chunkHandler, adminHandler, storageHandler, userHandler, cfg, userService)
}

// SetupAllRoutes 设置所有路由（使用已初始化的处理器）
func SetupAllRoutes(
	router *gin.Engine,
	shareHandler *handlers.ShareHandler,
	chunkHandler *handlers.ChunkHandler,
	adminHandler *handlers.AdminHandler,
	storageHandler *handlers.StorageHandler,
	userHandler *handlers.UserHandler,
	cfg *config.Config,
	userService interface {
		ValidateToken(string) (interface{}, error)
	},
) {

	// 设置基础路由
	SetupBaseRoutes(router, cfg)

	// 设置分享路由
	SetupShareRoutes(router, shareHandler, cfg, userService)

	// 设置用户路由
	SetupUserRoutes(router, userHandler, cfg, userService)

	// 设置分片上传路由
	SetupChunkRoutes(router, chunkHandler, cfg)

	// 设置管理员路由
	SetupAdminRoutes(router, adminHandler, storageHandler, cfg)
}

// SetupRoutes 设置路由 (保持兼容性，使用已初始化的处理器)
func SetupRoutes(
	router *gin.Engine,
	shareHandler *handlers.ShareHandler,
	chunkHandler *handlers.ChunkHandler,
	adminHandler *handlers.AdminHandler,
	storageHandler *handlers.StorageHandler,
	userHandler *handlers.UserHandler,
	cfg *config.Config,
	userService interface {
		ValidateToken(string) (interface{}, error)
	},
) {
	// 使用新的路由设置函数
	SetupAllRoutes(router, shareHandler, chunkHandler, adminHandler, storageHandler, userHandler, cfg, userService)
}
