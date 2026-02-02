package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/services"
)

// QRCodeHandler 二维码处理器
type QRCodeHandler struct {
	qrService *services.QRCodeService
}

// NewQRCodeHandler 创建二维码处理器实例
func NewQRCodeHandler() *QRCodeHandler {
	return &QRCodeHandler{
		qrService: services.NewQRCodeService(),
	}
}

// GenerateQRCode 生成二维码
// @Summary 生成二维码
// @Description 根据提供的数据生成二维码图片
// @Tags 二维码
// @Accept json
// @Produce png
// @Param data query string true "二维码数据内容"
// @Param size query int false "二维码尺寸(像素)" default(256)
// @Success 200 {file} binary "PNG格式的二维码图片"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/qrcode/generate [get]
func (h *QRCodeHandler) GenerateQRCode(c *gin.Context) {
	// 获取查询参数
	data := c.Query("data")
	if data == "" {
		common.BadRequestResponse(c, "二维码数据不能为空")
		return
	}

	// 获取尺寸参数
	sizeStr := c.DefaultQuery("size", "256")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size <= 0 {
		size = 256 // 默认尺寸
	}

	// 验证数据有效性
	if !h.qrService.ValidateQRCodeData(data) {
		common.BadRequestResponse(c, "二维码数据无效或过长")
		return
	}

	// 生成二维码
	pngData, err := h.qrService.GenerateQRCode(data, size)
	if err != nil {
		common.InternalServerErrorResponse(c, fmt.Sprintf("生成二维码失败: %v", err))
		return
	}

	// 设置响应头
	c.Header("Content-Type", "image/png")
	c.Header("Content-Length", fmt.Sprintf("%d", len(pngData)))
	c.Header("Cache-Control", "public, max-age=3600") // 缓存1小时

	// 返回PNG图片数据
	c.Data(http.StatusOK, "image/png", pngData)
}

// GenerateQRCodeBase64 生成Base64编码的二维码
// @Summary 生成Base64二维码
// @Description 根据提供的数据生成Base64编码的二维码图片
// @Tags 二维码
// @Accept json
// @Produce json
// @Param data query string true "二维码数据内容"
// @Param size query int false "二维码尺寸(像素)" default(256)
// @Success 200 {object} map[string]interface{} "Base64编码的二维码图片"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/qrcode/base64 [get]
func (h *QRCodeHandler) GenerateQRCodeBase64(c *gin.Context) {
	// 获取查询参数
	data := c.Query("data")
	if data == "" {
		common.BadRequestResponse(c, "二维码数据不能为空")
		return
	}

	// 获取尺寸参数
	sizeStr := c.DefaultQuery("size", "256")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size <= 0 {
		size = 256 // 默认尺寸
	}

	// 验证数据有效性
	if !h.qrService.ValidateQRCodeData(data) {
		common.BadRequestResponse(c, "二维码数据无效或过长")
		return
	}

	// 生成Base64编码的二维码
	base64Data, err := h.qrService.GenerateQRCodeBase64(data, size)
	if err != nil {
		common.InternalServerErrorResponse(c, fmt.Sprintf("生成二维码失败: %v", err))
		return
	}

	response := map[string]interface{}{
		"qr_code": base64Data,
		"data":    data,
		"size":    size,
	}

	common.SuccessResponse(c, response)
}
