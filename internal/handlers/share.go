package handlers

import (
	"fmt"
	"time"

	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/models/web"
	"github.com/zy84338719/filecodebox/internal/services/share"
	"github.com/zy84338719/filecodebox/internal/storage"
	"github.com/zy84338719/filecodebox/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ShareHandler 分享处理器
type ShareHandler struct {
	service *share.Service
}

func NewShareHandler(service *share.Service) *ShareHandler {
	return &ShareHandler{service: service}
}

// buildFullShareURL 构建完整的分享URL（包含协议和域名）
func (h *ShareHandler) buildFullShareURL(c *gin.Context, path string) string {
	// 获取请求的协议和主机
	protocol := "http"
	if c.Request.TLS != nil {
		protocol = "https"
	}

	host := c.Request.Host
	return fmt.Sprintf("%s://%s%s", protocol, host, path)
}

// ShareText 分享文本
// @Summary 分享文本内容
// @Description 分享文本内容并生成分享代码
// @Tags 分享
// @Accept multipart/form-data
// @Produce json
// @Param text formData string true "文本内容"
// @Param expire_value formData int false "过期值" default(1)
// @Param expire_style formData string false "过期样式" default(day) Enums(minute, hour, day, week, month, year, forever)
// @Param require_auth formData boolean false "是否需要认证" default(false)
// @Success 200 {object} map[string]interface{} "分享成功，返回分享代码"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /share/text/ [post]
func (h *ShareHandler) ShareText(c *gin.Context) {
	text := c.PostForm("text")
	expireValueStr := c.DefaultPostForm("expire_value", "1")
	expireStyle := c.DefaultPostForm("expire_style", "day")
	requireAuthStr := c.DefaultPostForm("require_auth", "false")

	if text == "" {
		common.BadRequestResponse(c, "文本内容不能为空")
		return
	}

	// 解析过期参数
	expireParams, err := utils.ParseExpireParams(expireValueStr, expireStyle, requireAuthStr)
	if err != nil {
		common.BadRequestResponse(c, err.Error())
		return
	}

	// 构建请求
	req := web.ShareTextRequest{
		Text:        text,
		ExpireValue: expireParams.ExpireValue,
		ExpireStyle: expireParams.ExpireStyle,
		RequireAuth: expireParams.RequireAuth,
	}

	// 检查是否为认证用户上传
	userID := utils.GetUserIDFromContext(c)

	fileResult, err := h.service.ShareTextWithAuth(req.Text, req.ExpireValue, req.ExpireStyle, userID)
	if err != nil {
		common.BadRequestResponse(c, err.Error())
		return
	}

	response := web.ShareResponse{
		Code:         fileResult.Code,
		ShareURL:     fileResult.ShareURL,
		FileName:     "文本分享",
		ExpiredAt:    fileResult.ExpiredAt,
		FullShareURL: h.buildFullShareURL(c, fileResult.ShareURL),
		QRCodeData:   h.buildFullShareURL(c, fileResult.ShareURL),
	}

	common.SuccessWithMessage(c, "分享成功", response)
}

// ShareTextAPI 面向 API Key 用户的文本分享入口
// @Summary 分享文本（API 模式）
// @Description 通过 API Key 分享文本内容
// @Tags API
// @Accept multipart/form-data
// @Produce json
// @Param text formData string true "文本内容"
// @Param expire_value formData int false "过期值" default(1)
// @Param expire_style formData string false "过期样式" default(day) Enums(minute, hour, day, week, month, year, forever)
// @Param require_auth formData boolean false "是否需要认证" default(false)
// @Success 200 {object} map[string]interface{} "分享成功，返回分享代码"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "API Key 校验失败"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/share/text [post]
// @Security ApiKeyAuth
func (h *ShareHandler) ShareTextAPI(c *gin.Context) {
	h.ShareText(c)
}

// ShareFile 分享文件
// @Summary 分享文件
// @Description 上传并分享文件，生成分享代码
// @Tags 分享
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "要分享的文件"
// @Param expire_value formData int false "过期值" default(1)
// @Param expire_style formData string false "过期样式" default(day) Enums(minute, hour, day, week, month, year, forever)
// @Param require_auth formData boolean false "是否需要认证" default(false)
// @Success 200 {object} map[string]interface{} "分享成功，返回分享代码和文件信息"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /share/file/ [post]
func (h *ShareHandler) ShareFile(c *gin.Context) {
	// 解析表单参数
	expireValueStr := c.DefaultPostForm("expire_value", "1")
	expireStyle := c.DefaultPostForm("expire_style", "day")
	requireAuthStr := c.DefaultPostForm("require_auth", "false")

	// 解析过期参数
	expireParams, err := utils.ParseExpireParams(expireValueStr, expireStyle, requireAuthStr)
	if err != nil {
		common.BadRequestResponse(c, err.Error())
		return
	}

	userID := utils.GetUserIDFromContext(c)
	if h.service.IsUploadLoginRequired() && userID == nil {
		common.UnauthorizedResponse(c, "当前配置要求登录后才能上传文件")
		return
	}

	// 解析文件
	file, success := utils.ParseFileFromForm(c, "file")
	if !success {
		return
	}

	// 构建服务层请求（这里需要适配服务层的接口）
	serviceReq := models.ShareFileRequest{
		File:        file,
		ExpireValue: expireParams.ExpireValue,
		ExpireStyle: expireParams.ExpireStyle,
		RequireAuth: expireParams.RequireAuth,
		ClientIP:    c.ClientIP(),
		UserID:      userID,
	}

	fileResult, err := h.service.ShareFileWithAuth(serviceReq)
	if err != nil {
		common.BadRequestResponse(c, err.Error())
		return
	}

	response := web.ShareResponse{
		Code:         fileResult.Code,
		ShareURL:     fileResult.ShareURL,
		FileName:     fileResult.FileName,
		ExpiredAt:    fileResult.ExpiredAt,
		FullShareURL: h.buildFullShareURL(c, fileResult.ShareURL),
		QRCodeData:   h.buildFullShareURL(c, fileResult.ShareURL),
	}

	common.SuccessResponse(c, response)
}

// ShareFileAPI 面向 API Key 用户的文件分享入口
// @Summary 分享文件（API 模式）
// @Description 通过 API Key 上传并分享文件
// @Tags API
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "要分享的文件"
// @Param expire_value formData int false "过期值" default(1)
// @Param expire_style formData string false "过期样式" default(day) Enums(minute, hour, day, week, month, year, forever)
// @Param require_auth formData boolean false "是否需要认证" default(false)
// @Success 200 {object} map[string]interface{} "分享成功，返回分享代码和文件信息"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "API Key 校验失败"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/share/file [post]
// @Security ApiKeyAuth
func (h *ShareHandler) ShareFileAPI(c *gin.Context) {
	h.ShareFile(c)
}

// GetFile 获取文件信息
// @Summary 获取分享文件信息
// @Description 根据分享代码获取文件或文本的详细信息
// @Tags 分享
// @Accept json
// @Produce json
// @Param code query string false "分享代码(GET方式)"
// @Param code formData string false "分享代码(POST方式)"
// @Success 200 {object} map[string]interface{} "文件信息"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "分享代码不存在"
// @Router /share/select/ [get]
// @Router /share/select/ [post]
func (h *ShareHandler) GetFile(c *gin.Context) {
	var code string

	if c.Request.Method == "GET" {
		code = c.Query("code")
	} else {
		// POST 请求，尝试从JSON解析
		var req web.ShareCodeRequest
		if err := c.ShouldBindJSON(&req); err == nil {
			code = req.Code
		} else {
			// 如果JSON解析失败，尝试从表单获取
			code = c.PostForm("code")
		}
	}

	if code == "" {
		common.BadRequestResponse(c, "文件代码不能为空")
		return
	}

	// 获取用户ID（如果已登录）
	var userID *uint
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(uint)
		userID = &id
	}

	fileCode, err := h.service.GetFileByCodeWithAuth(code, userID)
	if err != nil {
		common.NotFoundResponse(c, err.Error())
		return
	}

	// 更新使用次数
	if err := h.service.UpdateFileUsage(fileCode.Code); err != nil {
		// 记录错误但不阻止下载
		logrus.WithError(err).Error("更新文件使用次数失败")
	}

	response := web.FileInfoResponse{
		Code:        fileCode.Code,
		Name:        getDisplayFileName(fileCode),
		Size:        fileCode.Size,
		UploadType:  fileCode.UploadType,
		RequireAuth: fileCode.RequireAuth,
	}

	if fileCode.Text != "" {
		// 返回文本内容
		response.Text = fileCode.Text
	} else {
		// 返回文件下载链接
		response.Text = "/share/download?code=" + fileCode.Code
	}

	common.SuccessResponse(c, response)
}

// GetFileAPI 通过 REST 模式查询分享信息（API 模式）
// @Summary 查询分享详情（API 模式）
// @Description 根据分享代码返回分享的文件或文本信息
// @Tags API
// @Produce json
// @Param code path string true "分享代码"
// @Success 200 {object} map[string]interface{} "分享详情"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "API Key 校验失败"
// @Failure 404 {object} map[string]interface{} "分享不存在"
// @Router /api/v1/share/{code} [get]
// @Security ApiKeyAuth
func (h *ShareHandler) GetFileAPI(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		common.BadRequestResponse(c, "文件代码不能为空")
		return
	}

	fileCode, _, ok := h.fetchFileForRequest(c, code)
	if !ok {
		return
	}

	response := web.FileInfoResponse{
		Code:        fileCode.Code,
		Name:        getDisplayFileName(fileCode),
		Size:        fileCode.Size,
		UploadType:  fileCode.UploadType,
		RequireAuth: fileCode.RequireAuth,
	}

	if fileCode.Text != "" {
		response.Text = fileCode.Text
	} else {
		response.Text = "/share/download?code=" + fileCode.Code
	}

	common.SuccessResponse(c, response)
}

// DownloadFile 下载文件
// @Summary 下载分享文件
// @Description 根据分享代码下载文件或获取文本内容
// @Tags 分享
// @Accept json
// @Produce application/octet-stream
// @Produce application/json
// @Param code query string true "分享代码"
// @Success 200 {file} binary "文件内容"
// @Success 200 {object} map[string]interface{} "文本内容"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "分享代码不存在"
// @Router /share/download [get]
func (h *ShareHandler) DownloadFile(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		common.BadRequestResponse(c, "文件代码不能为空")
		return
	}

	fileCode, userID, ok := h.fetchFileForRequest(c, code)
	if !ok {
		return
	}

	if h.tryReturnText(c, fileCode, userID) {
		return
	}

	if !h.streamFileResponse(c, fileCode, userID) {
		return
	}
}

// DownloadFileAPI REST 风格下载接口（API 模式）
// @Summary 下载分享内容（API 模式）
// @Description 根据分享代码下载文件或获取文本内容
// @Tags API
// @Produce application/octet-stream
// @Produce application/json
// @Param code path string true "分享代码"
// @Success 200 {file} binary "文件内容"
// @Success 200 {object} map[string]interface{} "文本内容"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "API Key 校验失败"
// @Failure 404 {object} map[string]interface{} "分享不存在"
// @Router /api/v1/share/{code}/download [get]
// @Security ApiKeyAuth
func (h *ShareHandler) DownloadFileAPI(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		common.BadRequestResponse(c, "文件代码不能为空")
		return
	}

	fileCode, userID, ok := h.fetchFileForRequest(c, code)
	if !ok {
		return
	}

	if h.tryReturnText(c, fileCode, userID) {
		return
	}

	_ = h.streamFileResponse(c, fileCode, userID)
}

func (h *ShareHandler) fetchFileForRequest(c *gin.Context, code string) (*models.FileCode, *uint, bool) {
	var userID *uint
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(uint)
		userID = &id
	}

	fileCode, err := h.service.GetFileByCodeWithAuth(code, userID)
	if err != nil {
		common.NotFoundResponse(c, err.Error())
		return nil, nil, false
	}

	if err := h.service.UpdateFileUsage(fileCode.Code); err != nil {
		logrus.WithError(err).Error("更新文件使用次数失败")
	}

	return fileCode, userID, true
}

func (h *ShareHandler) tryReturnText(c *gin.Context, fileCode *models.FileCode, userID *uint) bool {
	if fileCode.Text == "" {
		return false
	}

	common.SuccessResponse(c, fileCode.Text)
	h.service.RecordDownloadLog(fileCode, userID, c.ClientIP(), 0)
	return true
}

func (h *ShareHandler) streamFileResponse(c *gin.Context, fileCode *models.FileCode, userID *uint) bool {
	storageServiceInterface := h.service.GetStorageService()
	storageService, ok := storageServiceInterface.(*storage.ConcreteStorageService)
	if !ok {
		common.InternalServerErrorResponse(c, "存储服务类型错误")
		return false
	}

	start := time.Now()
	if err := storageService.GetFileResponse(c, fileCode); err != nil {
		common.NotFoundResponse(c, "文件下载失败: "+err.Error())
		return false
	}

	h.service.RecordDownloadLog(fileCode, userID, c.ClientIP(), time.Since(start))
	return true
}

// getDisplayFileName 获取用于显示的文件名
func getDisplayFileName(fileCode *models.FileCode) string {
	if fileCode.UUIDFileName != "" {
		return fileCode.UUIDFileName
	}
	// 向后兼容：如果UUIDFileName为空，则使用Prefix + Suffix
	return fileCode.Prefix + fileCode.Suffix
}
