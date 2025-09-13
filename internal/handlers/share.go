package handlers

import (
	"strconv"

	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/models/web"
	"github.com/zy84338719/filecodebox/internal/services/share"
	"github.com/zy84338719/filecodebox/internal/storage"

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

	expireValue, err := strconv.Atoi(expireValueStr)
	if err != nil {
		common.BadRequestResponse(c, "过期时间参数错误")
		return
	}

	// 对于forever模式，允许expireValue为0
	// 对于count模式，expireValue必须大于0
	// 对于时间模式，expireValue必须大于0
	if expireValue < 0 || (expireStyle != "forever" && expireValue == 0) {
		common.BadRequestResponse(c, "过期时间参数错误")
		return
	}

	if text == "" {
		common.BadRequestResponse(c, "文本内容不能为空")
		return
	}

	// 检查是否需要登录才能下载
	requireAuth := requireAuthStr == "true"

	// 构建请求
	req := web.ShareTextRequest{
		Text:        text,
		ExpireValue: expireValue,
		ExpireStyle: expireStyle,
		RequireAuth: requireAuth,
	}

	// 检查是否为认证用户上传
	var userID *uint
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(uint)
		userID = &id
	}

	fileResult, err := h.service.ShareTextWithAuth(req.Text, req.ExpireValue, req.ExpireStyle, userID)
	if err != nil {
		common.BadRequestResponse(c, err.Error())
		return
	}

	response := web.ShareResponse{
		Code:      fileResult.Code,
		ShareURL:  fileResult.ShareURL,
		FileName:  "文本分享",
		ExpiredAt: fileResult.ExpiredAt,
	}

	common.SuccessWithMessage(c, "分享成功", response)
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
	// 绑定表单参数
	var req web.ShareFileRequest
	if err := c.ShouldBind(&req); err != nil {
		common.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	// 默认值处理和验证
	if req.ExpireValue < 0 {
		common.BadRequestResponse(c, "过期时间参数不能为负数")
		return
	}

	// 对于非forever模式，ExpireValue不能为0
	if req.ExpireStyle != "forever" && req.ExpireValue == 0 {
		req.ExpireValue = 1 // 默认值
	}

	if req.ExpireStyle == "" {
		req.ExpireStyle = "day"
	}

	file, err := c.FormFile("file")
	if err != nil {
		common.BadRequestResponse(c, "获取文件失败")
		return
	}

	// 检查是否为认证用户上传
	var userID *uint
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(uint)
		userID = &id
	}

	// 构建服务层请求（这里需要适配服务层的接口）
	serviceReq := models.ShareFileRequest{
		File:        file,
		ExpireValue: req.ExpireValue,
		ExpireStyle: req.ExpireStyle,
		RequireAuth: req.RequireAuth,
		ClientIP:    c.ClientIP(),
		UserID:      userID,
	}

	fileResult, err := h.service.ShareFileWithAuth(serviceReq)
	if err != nil {
		common.BadRequestResponse(c, err.Error())
		return
	}

	response := web.ShareResponse{
		Code:      fileResult.Code,
		ShareURL:  fileResult.ShareURL,
		FileName:  fileResult.FileName,
		ExpiredAt: fileResult.ExpiredAt,
	}

	common.SuccessResponse(c, response)
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

	if fileCode.Text != "" {
		common.SuccessResponse(c, fileCode.Text)
		return
	}

	// 使用存储服务下载文件
	storageServiceInterface := h.service.GetStorageService()
	storageService, ok := storageServiceInterface.(*storage.ConcreteStorageService)
	if !ok {
		common.InternalServerErrorResponse(c, "存储服务类型错误")
		return
	}

	if err := storageService.GetFileResponse(c, fileCode); err != nil {
		common.NotFoundResponse(c, "文件下载失败: "+err.Error())
		return
	}
}

// getDisplayFileName 获取用于显示的文件名
func getDisplayFileName(fileCode *models.FileCode) string {
	if fileCode.UUIDFileName != "" {
		return fileCode.UUIDFileName
	}
	// 向后兼容：如果UUIDFileName为空，则使用Prefix + Suffix
	return fileCode.Prefix + fileCode.Suffix
}
