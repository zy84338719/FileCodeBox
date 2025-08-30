package handlers

import (
	"net/http"
	"strconv"

	"github.com/zy84338719/filecodebox/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ShareHandler 分享处理器
type ShareHandler struct {
	service *services.ShareService
}

func NewShareHandler(service *services.ShareService) *ShareHandler {
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
	if err != nil || expireValue <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "过期时间参数错误",
		})
		return
	}

	if text == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文本内容不能为空",
		})
		return
	}

	// 检查是否需要登录才能下载
	requireAuth := requireAuthStr == "true"

	// 构建请求
	req := services.ShareTextRequest{
		Text:        text,
		ExpireValue: expireValue,
		ExpireStyle: expireStyle,
		RequireAuth: requireAuth,
		ClientIP:    c.ClientIP(),
	}

	// 检查是否为认证用户上传
	if userID, exists := c.Get("user_id"); exists {
		uid := userID.(uint)
		req.UserID = &uid
	}

	fileCode, err := h.service.ShareTextWithAuth(req)
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
			"code":        fileCode.Code,
			"upload_type": fileCode.UploadType,
		},
	})
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
	expireValueStr := c.DefaultPostForm("expire_value", "1")
	expireStyle := c.DefaultPostForm("expire_style", "day")
	requireAuthStr := c.DefaultPostForm("require_auth", "false")

	expireValue, err := strconv.Atoi(expireValueStr)
	if err != nil || expireValue <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "过期时间参数错误",
		})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "获取文件失败",
		})
		return
	}

	// 检查是否需要登录才能下载
	requireAuth := requireAuthStr == "true"

	// 构建请求
	req := services.ShareFileRequest{
		File:        file,
		ExpireValue: expireValue,
		ExpireStyle: expireStyle,
		RequireAuth: requireAuth,
		ClientIP:    c.ClientIP(),
	}

	// 检查是否为认证用户上传
	if userID, exists := c.Get("user_id"); exists {
		uid := userID.(uint)
		req.UserID = &uid
	}

	fileCode, err := h.service.ShareFileWithAuth(req)
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
			"code":        fileCode.Code,
			"name":        file.Filename,
			"upload_type": fileCode.UploadType,
		},
	})
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
		var req struct {
			Code string `json:"code"`
		}
		if err := c.ShouldBindJSON(&req); err == nil {
			code = req.Code
		} else {
			// 如果JSON解析失败，尝试从表单获取
			code = c.PostForm("code")
		}
	}

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文件代码不能为空",
		})
		return
	}

	// 获取用户ID（如果已登录）
	var userID *uint
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(uint)
		userID = &id
	}

	fileCode, err := h.service.GetFileByCodeWithAuth(code, true, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": err.Error(),
		})
		return
	}

	// 更新使用次数
	if err := h.service.UpdateFileUsage(fileCode); err != nil {
		// 记录错误但不阻止下载
		logrus.WithError(err).Error("更新文件使用次数失败")
	}

	if fileCode.Text != "" {
		// 返回文本内容
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "success",
			"detail": gin.H{
				"code":         fileCode.Code,
				"name":         fileCode.Prefix + fileCode.Suffix,
				"size":         fileCode.Size,
				"text":         fileCode.Text,
				"upload_type":  fileCode.UploadType,
				"require_auth": fileCode.RequireAuth,
			},
		})
	} else {
		// 返回文件下载信息
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "success",
			"detail": gin.H{
				"code":         fileCode.Code,
				"name":         fileCode.Prefix + fileCode.Suffix,
				"size":         fileCode.Size,
				"text":         "/share/download?code=" + fileCode.Code,
				"upload_type":  fileCode.UploadType,
				"require_auth": fileCode.RequireAuth,
			},
		})
	}
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
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文件代码不能为空",
		})
		return
	}

	// 获取用户ID（如果已登录）
	var userID *uint
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(uint)
		userID = &id
	}

	fileCode, err := h.service.GetFileByCodeWithAuth(code, false, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": err.Error(),
		})
		return
	}

	if fileCode.Text != "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "success",
			"detail":  fileCode.Text,
		})
		return
	}

	// 使用存储接口下载文件
	storageInterface := h.service.GetStorageInterface()
	if err := storageInterface.GetFileResponse(c, fileCode); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "文件下载失败: " + err.Error(),
		})
		return
	}
}
