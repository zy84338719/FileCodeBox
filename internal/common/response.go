package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 通用响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Detail  interface{} `json:"detail,omitempty"`
}

// SuccessResponse 成功响应
func SuccessResponse(c *gin.Context, detail interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Detail:  detail,
	})
}

// SuccessWithMessage 带自定义消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, detail map[string]interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: message,
		Detail:  detail,
	})
}

// ErrorResponse 错误响应
func ErrorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}

// BadRequestResponse 400 错误响应
func BadRequestResponse(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    400,
		Message: message,
	})
}

// UnauthorizedResponse 401 未授权响应
func UnauthorizedResponse(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    401,
		Message: message,
	})
}

// ForbiddenResponse 403 禁止访问响应
func ForbiddenResponse(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Code:    403,
		Message: message,
	})
}

// NotFoundResponse 404 未找到响应
func NotFoundResponse(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Code:    404,
		Message: message,
	})
}

// InternalServerErrorResponse 500 服务器内部错误响应
func InternalServerErrorResponse(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    500,
		Message: message,
	})
}

// SuccessWithCleanedCount 带清理计数的成功响应
func SuccessWithCleanedCount(c *gin.Context, count int) {
	SuccessResponse(c, map[string]interface{}{
		"cleaned_count": count,
	})
}

// SuccessWithFileInfo 带文件信息的成功响应
func SuccessWithFileInfo(c *gin.Context, fileInfo interface{}) {
	SuccessResponse(c, fileInfo)
}

// SuccessWithList 带列表数据的成功响应
func SuccessWithList(c *gin.Context, list interface{}, total int) {
	SuccessResponse(c, map[string]interface{}{
		"list":  list,
		"total": total,
	})
}

// SuccessWithPagination 带分页数据的成功响应
func SuccessWithPagination(c *gin.Context, list interface{}, total int, page int, pageSize int) {
	SuccessResponse(c, map[string]interface{}{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"pages":     (total + pageSize - 1) / pageSize,
	})
}

// SuccessWithToken 带令牌的成功响应
func SuccessWithToken(c *gin.Context, token string, userInfo interface{}) {
	SuccessResponse(c, map[string]interface{}{
		"token":     token,
		"user_info": userInfo,
	})
}

// SuccessWithUploadInfo 带上传信息的成功响应
func SuccessWithUploadInfo(c *gin.Context, shareCode string, downloadLink string) {
	SuccessResponse(c, map[string]interface{}{
		"share_code":    shareCode,
		"download_link": downloadLink,
	})
}
