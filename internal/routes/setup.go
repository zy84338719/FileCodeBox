package routes

import (
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/handlers"

	"github.com/gin-gonic/gin"
)

// SetupAllRoutes 设置所有路由
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
