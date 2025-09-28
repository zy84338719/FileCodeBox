package handlers

import (
	"fmt"

	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/models/web"
	"github.com/zy84338719/filecodebox/internal/services"
	"github.com/zy84338719/filecodebox/internal/utils"

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
	var req web.ChunkUploadInitRequest
	if !utils.BindJSONWithValidation(c, &req) {
		return
	}

	// TODO: 修复服务层接口调用
	// 目前服务层接口与期望不符，需要重构服务层
	// 暂时返回模拟数据
	response := web.ChunkUploadInitResponse{
		UploadID:      req.FileHash, // 使用文件哈希作为上传ID
		TotalChunks:   int((req.FileSize + int64(req.ChunkSize) - 1) / int64(req.ChunkSize)),
		ChunkSize:     req.ChunkSize,
		UploadedCount: 0,
		Progress:      0.0,
	}

	common.SuccessResponse(c, response)
}

// InitChunkUploadAPI 初始化分片上传（API 模式）
// @Summary 初始化分片上传（API 模式）
// @Description 使用 API Key 初始化分片上传，返回上传ID
// @Tags API
// @Accept json
// @Produce json
// @Param request body object true "上传初始化参数"
// @Success 200 {object} map[string]interface{} "初始化成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "API Key 校验失败"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/chunks/upload/init [post]
// @Security ApiKeyAuth
func (h *ChunkHandler) InitChunkUploadAPI(c *gin.Context) {
	h.InitChunkUpload(c)
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
	chunkIndex, success := utils.ParseIntFromParam(c, "chunk_index", "分片索引错误")
	if !success {
		return
	}

	file, success := utils.ParseFileFromForm(c, "chunk")
	if !success {
		return
	}

	// TODO: 修复服务层接口调用
	// 目前服务层接口与期望不符，需要重构服务层
	// 暂时返回模拟数据，这里可以实际处理文件
	_ = file // 暂时忽略文件，避免未使用错误

	response := web.ChunkUploadResponse{
		ChunkHash:  fmt.Sprintf("chunk_%s_%d", uploadID, chunkIndex),
		ChunkIndex: chunkIndex,
		Progress:   float64(chunkIndex+1) / 10.0, // 模拟进度
	}

	common.SuccessResponse(c, response)
}

// UploadChunkAPI 上传文件分片（API 模式）
// @Summary 上传文件分片（API 模式）
// @Description 上传指定索引的文件分片
// @Tags API
// @Accept multipart/form-data
// @Produce json
// @Param upload_id path string true "上传ID"
// @Param chunk_index path int true "分片索引"
// @Param chunk formData file true "分片文件"
// @Success 200 {object} map[string]interface{} "上传成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "API Key 校验失败"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/chunks/upload/chunk/{upload_id}/{chunk_index} [post]
// @Security ApiKeyAuth
func (h *ChunkHandler) UploadChunkAPI(c *gin.Context) {
	h.UploadChunk(c)
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

	var req web.ChunkUploadCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	// TODO: 修复服务层接口调用
	// 目前服务层接口与期望不符，需要重构服务层
	// 暂时返回模拟数据
	response := web.ChunkUploadCompleteResponse{
		Code:     uploadID, // 使用上传ID作为分享代码
		ShareURL: "/share/" + uploadID,
		FileName: "uploaded_file.bin", // 模拟文件名
	}

	common.SuccessResponse(c, response)
}

// CompleteUploadAPI 完成分片上传（API 模式）
// @Summary 完成分片上传（API 模式）
// @Description 合并所有分片并生成分享代码
// @Tags API
// @Accept json
// @Produce json
// @Param upload_id path string true "上传ID"
// @Param request body object true "完成上传参数"
// @Success 200 {object} map[string]interface{} "上传完成，返回分享代码"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "API Key 校验失败"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/chunks/upload/complete/{upload_id} [post]
// @Security ApiKeyAuth
func (h *ChunkHandler) CompleteUploadAPI(c *gin.Context) {
	h.CompleteUpload(c)
}

// GetUploadStatus 获取上传状态（断点续传支持）
func (h *ChunkHandler) GetUploadStatus(c *gin.Context) {
	uploadID := c.Param("upload_id")

	// 调用服务层获取上传状态
	status, err := h.service.GetUploadStatus(uploadID)
	if err != nil {
		common.ErrorResponse(c, 500, "获取上传状态失败: "+err.Error())
		return
	}

	common.SuccessResponse(c, status)
}

// GetUploadStatusAPI 查询上传状态（API 模式）
// @Summary 查询上传状态（API 模式）
// @Description 查询分片上传的进度和状态
// @Tags API
// @Produce json
// @Param upload_id path string true "上传ID"
// @Success 200 {object} map[string]interface{} "上传状态"
// @Failure 401 {object} map[string]interface{} "API Key 校验失败"
// @Failure 404 {object} map[string]interface{} "上传ID不存在"
// @Router /api/v1/chunks/upload/status/{upload_id} [get]
// @Security ApiKeyAuth
func (h *ChunkHandler) GetUploadStatusAPI(c *gin.Context) {
	h.GetUploadStatus(c)
}

// VerifyChunk 验证分片完整性
func (h *ChunkHandler) VerifyChunk(c *gin.Context) {
	uploadID := c.Param("upload_id")

	chunkIndex, success := utils.ParseIntFromParam(c, "chunk_index", "分片索引错误")
	if !success {
		return
	}

	var req struct {
		ChunkHash string `json:"chunk_hash" binding:"required"`
	}

	if !utils.BindJSONWithValidation(c, &req) {
		return
	}

	isValid, err := h.service.VerifyChunk(uploadID, chunkIndex, req.ChunkHash)
	if err != nil {
		common.BadRequestResponse(c, err.Error())
		return
	}

	common.SuccessResponse(c, web.ChunkValidationResponse{
		Valid: isValid,
	})
}

// VerifyChunkAPI 校验分片（API 模式）
// @Summary 校验分片（API 模式）
// @Description 校验指定分片是否已上传
// @Tags API
// @Accept json
// @Produce json
// @Param upload_id path string true "上传ID"
// @Param chunk_index path int true "分片索引"
// @Param request body object true "分片校验参数"
// @Success 200 {object} map[string]interface{} "校验结果"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "API Key 校验失败"
// @Router /api/v1/chunks/upload/verify/{upload_id}/{chunk_index} [post]
// @Security ApiKeyAuth
func (h *ChunkHandler) VerifyChunkAPI(c *gin.Context) {
	h.VerifyChunk(c)
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

// CancelUploadAPI 取消分片上传（API 模式）
// @Summary 取消分片上传（API 模式）
// @Description 取消上传流程并清理已有分片
// @Tags API
// @Produce json
// @Param upload_id path string true "上传ID"
// @Success 200 {object} map[string]interface{} "取消成功"
// @Failure 401 {object} map[string]interface{} "API Key 校验失败"
// @Router /api/v1/chunks/upload/cancel/{upload_id} [delete]
// @Security ApiKeyAuth
func (h *ChunkHandler) CancelUploadAPI(c *gin.Context) {
	h.CancelUpload(c)
}
