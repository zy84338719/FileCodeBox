package routes

import (
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/handlers"
	"github.com/zy84338719/filecodebox/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupChunkRoutes 设置分片上传相关路由
func SetupChunkRoutes(
	router *gin.Engine,
	chunkHandler *handlers.ChunkHandler,
	cfg *config.ConfigManager,
) {
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
}
