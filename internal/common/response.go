package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/zy84338719/filecodebox/internal/models/web"
)

// Response 通用响应结构

// SuccessResponse 成功响应
func SuccessResponse(c *gin.Context, detail interface{}) {
	c.JSON(http.StatusOK, web.SuccessResponse{
		Code: http.StatusOK,
		Data: detail,
	})
}

// SuccessWithMessage 带自定义消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, detail interface{}) {
	c.JSON(http.StatusOK, web.SuccessResponse{
		Code:    http.StatusOK,
		Message: message,
		Data:    detail,
	})
}

// ErrorResponse 错误响应
func ErrorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, web.ErrorResponse{
		Code:    code,
		Message: message,
	})
}

// BadRequestResponse 400 错误响应
func BadRequestResponse(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, web.ErrorResponse{
		Code:    http.StatusBadRequest,
		Message: message,
	})
}

// UnauthorizedResponse 401 未授权响应
func UnauthorizedResponse(c *gin.Context, message string) {
	// log the unauthorized response with request path to aid debugging
	logrus.WithField("path", c.Request.URL.Path).Infof("UnauthorizedResponse: %s", message)
	c.JSON(http.StatusUnauthorized, web.ErrorResponse{
		Code:    http.StatusUnauthorized,
		Message: message,
	})
}

// ForbiddenResponse 403 禁止访问响应
func ForbiddenResponse(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, web.ErrorResponse{
		Code:    http.StatusForbidden,
		Message: message,
	})
}

// NotFoundResponse 404 未找到响应
func NotFoundResponse(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, web.ErrorResponse{
		Code:    http.StatusNotFound,
		Message: message,
	})
}

// InternalServerErrorResponse 500 服务器内部错误响应
func InternalServerErrorResponse(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, web.ErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: message,
	})
}

// TooManyRequestsResponse 429 请求过多响应
func TooManyRequestsResponse(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, web.ErrorResponse{
		Code:    http.StatusTooManyRequests,
		Message: message,
	})
}

// SuccessWithCleanedCount 带清理计数的成功响应
func SuccessWithCleanedCount(c *gin.Context, count int) {
	response := web.CleanedCountResponse{
		CleanedCount: count,
	}
	SuccessResponse(c, response)
}

// SuccessWithFileInfo 带文件信息的成功响应
func SuccessWithFileInfo(c *gin.Context, fileInfo interface{}) {
	SuccessResponse(c, fileInfo)
}

// SuccessWithList 带列表数据的成功响应（统一使用分页响应结构）
func SuccessWithList(c *gin.Context, list interface{}, total int, pagination ...PaginationParams) {
	var paginationInfo web.PaginationInfo

	if len(pagination) > 0 {
		// 有分页参数，计算分页信息
		p := pagination[0]
		totalPages := (total + p.PageSize - 1) / p.PageSize
		paginationInfo = web.PaginationInfo{
			Page:       p.Page,
			PageSize:   p.PageSize,
			Total:      int64(total),
			TotalPages: totalPages,
			HasNext:    p.Page < totalPages,
			HasPrev:    p.Page > 1,
		}
	} else {
		// 无分页参数，使用默认分页信息（显示所有数据）
		paginationInfo = web.PaginationInfo{
			Page:       1,
			PageSize:   total,
			Total:      int64(total),
			TotalPages: 1,
			HasNext:    false,
			HasPrev:    false,
		}
	}

	response := web.PaginatedResponse{
		List:       list,
		Pagination: paginationInfo,
	}
	SuccessResponse(c, response)
}

// PaginationParams 分页参数
type PaginationParams struct {
	Page     int
	PageSize int
}

// SuccessWithPagination 带分页数据的成功响应（保留向后兼容性）
func SuccessWithPagination(c *gin.Context, list interface{}, total int, page int, pageSize int) {
	SuccessWithList(c, list, total, PaginationParams{Page: page, PageSize: pageSize})
}

// SuccessWithToken 带令牌的成功响应
func SuccessWithToken(c *gin.Context, token string, userInfo interface{}) {
	response := web.TokenResponse{
		Token:    token,
		UserInfo: userInfo,
	}
	SuccessResponse(c, response)
}

// SuccessWithUploadInfo 带上传信息的成功响应
func SuccessWithUploadInfo(c *gin.Context, shareCode string, downloadLink string) {
	response := web.UploadInfoResponse{
		ShareCode:    shareCode,
		DownloadLink: downloadLink,
	}
	SuccessResponse(c, response)
}
