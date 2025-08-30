package handlers

import (
	"strconv"

	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/services"

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
// @Summary 初始化分片上传
// @Description 初始化文件分片上传，返回上传ID和分片信息
// @Tags 分片上传
// @Accept json
// @Produce json
// @Param request body object true "上传初始化参数" example({"file_name":"test.zip","file_size":1024000,"chunk_size":1024,"file_hash":"abc123"})
// @Success 200 {object} map[string]interface{} "初始化成功，返回上传ID和分片信息"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /chunk/upload/init/ [post]
func (h *ChunkHandler) InitChunkUpload(c *gin.Context) {
	var req struct {
		FileName  string `json:"file_name" binding:"required"`
		FileSize  int64  `json:"file_size" binding:"required"`
		ChunkSize int    `json:"chunk_size" binding:"required"`
		FileHash  string `json:"file_hash" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.service.InitChunkUpload(req.FileName, req.FileSize, req.ChunkSize, req.FileHash)
	if err != nil {
		common.InternalServerErrorResponse(c, "初始化上传失败: "+err.Error())
		return
	}

	common.SuccessResponse(c, result)
}

// UploadChunk 上传分片
// @Summary 上传文件分片
// @Description 上传指定索引的文件分片
// @Tags 分片上传
// @Accept multipart/form-data
// @Produce json
// @Param upload_id path string true "上传ID"
// @Param chunk_index path int true "分片索引"
// @Param chunk formData file true "分片文件"
// @Success 200 {object} map[string]interface{} "上传成功，返回分片哈希"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /chunk/upload/chunk/{upload_id}/{chunk_index} [post]
func (h *ChunkHandler) UploadChunk(c *gin.Context) {
	uploadID := c.Param("upload_id")
	chunkIndexStr := c.Param("chunk_index")

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		common.BadRequestResponse(c, "分片索引错误")
		return
	}

	file, err := c.FormFile("chunk")
	if err != nil {
		common.BadRequestResponse(c, "获取分片文件失败")
		return
	}

	chunkHash, err := h.service.UploadChunk(uploadID, chunkIndex, file)
	if err != nil {
		common.BadRequestResponse(c, err.Error())
		return
	}

	common.SuccessResponse(c, gin.H{
		"chunk_hash": chunkHash,
	})
}

// CompleteUpload 完成上传
// @Summary 完成分片上传
// @Description 完成所有分片上传，合并文件并生成分享代码
// @Tags 分片上传
// @Accept json
// @Produce json
// @Param upload_id path string true "上传ID"
// @Param request body object true "完成上传参数" example({"expire_value":1,"expire_style":"day","require_auth":false})
// @Success 200 {object} map[string]interface{} "上传完成，返回分享代码"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /chunk/upload/complete/{upload_id} [post]
func (h *ChunkHandler) CompleteUpload(c *gin.Context) {
	uploadID := c.Param("upload_id")

	var req struct {
		ExpireValue int    `json:"expire_value" binding:"required"`
		ExpireStyle string `json:"expire_style" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	fileCode, err := h.service.CompleteUpload(uploadID, req.ExpireValue, req.ExpireStyle)
	if err != nil {
		common.BadRequestResponse(c, err.Error())
		return
	}

	common.SuccessResponse(c, gin.H{
		"code": fileCode.Code,
		"name": fileCode.Prefix + fileCode.Suffix,
	})
}

// GetUploadStatus 获取上传状态（断点续传支持）
func (h *ChunkHandler) GetUploadStatus(c *gin.Context) {
	uploadID := c.Param("upload_id")

	status, err := h.service.GetUploadStatus(uploadID)
	if err != nil {
		common.BadRequestResponse(c, err.Error())
		return
	}

	common.SuccessResponse(c, status)
}

// VerifyChunk 验证分片完整性
func (h *ChunkHandler) VerifyChunk(c *gin.Context) {
	uploadID := c.Param("upload_id")
	chunkIndexStr := c.Param("chunk_index")

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		common.BadRequestResponse(c, "分片索引错误")
		return
	}

	var req struct {
		ChunkHash string `json:"chunk_hash" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	isValid, err := h.service.VerifyChunk(uploadID, chunkIndex, req.ChunkHash)
	if err != nil {
		common.BadRequestResponse(c, err.Error())
		return
	}

	common.SuccessResponse(c, gin.H{
		"valid": isValid,
	})
}

// CancelUpload 取消上传
func (h *ChunkHandler) CancelUpload(c *gin.Context) {
	uploadID := c.Param("upload_id")

	err := h.service.CancelUpload(uploadID)
	if err != nil {
		common.BadRequestResponse(c, err.Error())
		return
	}

	common.SuccessWithMessage(c, "上传已取消", nil)
}
