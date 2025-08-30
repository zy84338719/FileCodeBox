package handlers

import (
	"filecodebox/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ChunkHandler 分片处理器
type ChunkHandler struct {
	service *services.ChunkService
}

func NewChunkHandler(service *services.ChunkService) *ChunkHandler {
	return &ChunkHandler{service: service}
}

// InitChunkUpload 初始化分片上传
func (h *ChunkHandler) InitChunkUpload(c *gin.Context) {
	var req struct {
		FileName  string `json:"file_name" binding:"required"`
		FileSize  int64  `json:"file_size" binding:"required"`
		ChunkSize int    `json:"chunk_size" binding:"required"`
		FileHash  string `json:"file_hash" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	result, err := h.service.InitChunkUpload(req.FileName, req.FileSize, req.ChunkSize, req.FileHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "初始化上传失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"detail":  result,
	})
}

// UploadChunk 上传分片
func (h *ChunkHandler) UploadChunk(c *gin.Context) {
	uploadID := c.Param("upload_id")
	chunkIndexStr := c.Param("chunk_index")

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "分片索引错误",
		})
		return
	}

	file, err := c.FormFile("chunk")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "获取分片文件失败",
		})
		return
	}

	chunkHash, err := h.service.UploadChunk(uploadID, chunkIndex, file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"detail": gin.H{
			"chunk_hash": chunkHash,
		},
	})
}

// CompleteUpload 完成上传
func (h *ChunkHandler) CompleteUpload(c *gin.Context) {
	uploadID := c.Param("upload_id")

	var req struct {
		ExpireValue int    `json:"expire_value" binding:"required"`
		ExpireStyle string `json:"expire_style" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	fileCode, err := h.service.CompleteUpload(uploadID, req.ExpireValue, req.ExpireStyle)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"detail": gin.H{
			"code": fileCode.Code,
			"name": fileCode.Prefix + fileCode.Suffix,
		},
	})
}

// GetUploadStatus 获取上传状态（断点续传支持）
func (h *ChunkHandler) GetUploadStatus(c *gin.Context) {
	uploadID := c.Param("upload_id")

	status, err := h.service.GetUploadStatus(uploadID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"detail":  status,
	})
}

// VerifyChunk 验证分片完整性
func (h *ChunkHandler) VerifyChunk(c *gin.Context) {
	uploadID := c.Param("upload_id")
	chunkIndexStr := c.Param("chunk_index")

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "分片索引错误",
		})
		return
	}

	var req struct {
		ChunkHash string `json:"chunk_hash" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	isValid, err := h.service.VerifyChunk(uploadID, chunkIndex, req.ChunkHash)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"detail": gin.H{
			"valid": isValid,
		},
	})
}

// CancelUpload 取消上传
func (h *ChunkHandler) CancelUpload(c *gin.Context) {
	uploadID := c.Param("upload_id")

	err := h.service.CancelUpload(uploadID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "上传已取消",
	})
}
