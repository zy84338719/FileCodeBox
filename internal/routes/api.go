package routes

import (
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/handlers"
	"github.com/zy84338719/filecodebox/internal/middleware"
	"github.com/zy84338719/filecodebox/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupAPIRoutes 注册面向 API Key 客户端的精简接口
func SetupAPIRoutes(
	router *gin.Engine,
	shareHandler *handlers.ShareHandler,
	chunkHandler *handlers.ChunkHandler,
	cfg *config.ConfigManager,
	userService *services.UserService,
) {
	if router == nil || shareHandler == nil || cfg == nil || userService == nil {
		return
	}

	apiGroup := router.Group("/api/v1")
	apiGroup.Use(middleware.ShareAuth(cfg))
	apiGroup.Use(middleware.APIKeyAuthOnly(userService))

	{
		shareGroup := apiGroup.Group("/share")
		shareGroup.POST("/text", shareHandler.ShareTextAPI)
		shareGroup.POST("/file", shareHandler.ShareFileAPI)
		shareGroup.GET("/:code", shareHandler.GetFileAPI)
		shareGroup.GET("/:code/download", shareHandler.DownloadFileAPI)
	}

	if chunkHandler != nil {
		chunkGroup := apiGroup.Group("/chunks")
		chunkGroup.POST("/upload/init", chunkHandler.InitChunkUploadAPI)
		chunkGroup.POST("/upload/chunk/:upload_id/:chunk_index", chunkHandler.UploadChunkAPI)
		chunkGroup.POST("/upload/complete/:upload_id", chunkHandler.CompleteUploadAPI)
		chunkGroup.GET("/upload/status/:upload_id", chunkHandler.GetUploadStatusAPI)
		chunkGroup.POST("/upload/verify/:upload_id/:chunk_index", chunkHandler.VerifyChunkAPI)
		chunkGroup.DELETE("/upload/cancel/:upload_id", chunkHandler.CancelUploadAPI)
	}
}
