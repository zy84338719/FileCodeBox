package resp

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/zy84338719/fileCodeBox/internal/pkg/errors"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func Success(c *app.RequestContext, data interface{}) {
	c.JSON(consts.StatusOK, Response{
		Code:    errors.CodeSuccess,
		Message: errors.GetMessage(errors.CodeSuccess),
		Data:    data,
	})
}

func SuccessWithMessage(c *app.RequestContext, message string, data interface{}) {
	c.JSON(consts.StatusOK, Response{
		Code:    errors.CodeSuccess,
		Message: message,
		Data:    data,
	})
}

func Error(c *app.RequestContext, code int) {
	httpStatus := consts.StatusOK
	if code == errors.CodeUnauthorized {
		httpStatus = consts.StatusUnauthorized
	}
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: errors.GetMessage(code),
	})
}

func ErrorWithMessage(c *app.RequestContext, code int, message string) {
	httpStatus := consts.StatusOK
	if code == errors.CodeUnauthorized {
		httpStatus = consts.StatusUnauthorized
	}
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
	})
}

func ErrorWithData(c *app.RequestContext, code int, message string, data interface{}) {
	httpStatus := consts.StatusOK
	if code == errors.CodeUnauthorized {
		httpStatus = consts.StatusUnauthorized
	}
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func Page(c *app.RequestContext, list interface{}, total int64, page, pageSize int) {
	Success(c, PageData{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func BadRequest(c *app.RequestContext, message string) {
	if message == "" {
		message = errors.GetMessage(errors.CodeBadRequest)
	}
	ErrorWithMessage(c, errors.CodeBadRequest, message)
}

func Unauthorized(c *app.RequestContext, message string) {
	if message == "" {
		message = errors.GetMessage(errors.CodeUnauthorized)
	}
	ErrorWithMessage(c, errors.CodeUnauthorized, message)
}

func NotFound(c *app.RequestContext, message string) {
	if message == "" {
		message = errors.GetMessage(errors.CodeNotFound)
	}
	ErrorWithMessage(c, errors.CodeNotFound, message)
}

func InternalError(c *app.RequestContext, message string) {
	if message == "" {
		message = errors.GetMessage(errors.CodeInternalError)
	}
	ErrorWithMessage(c, errors.CodeInternalError, message)
}
