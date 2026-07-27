package resp

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/uuid"
	"github.com/zy84338719/fileCodeBox/backend/internal/pkg/errcode"
	"github.com/zy84338719/fileCodeBox/backend/internal/pkg/errors"
)

// Response 全站统一响应 envelope
// {
//   "code":      业务码 (0=成功),
//   "message":   业务码默认文案 / 自定义 message,
//   "data":      业务数据 (omitempty),
//   "trace_id":  请求链路 ID (omitempty, 用于排错)
// }
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	TraceID string      `json:"trace_id,omitempty"`
}

type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// traceIDHeader trace_id 通过 X-Trace-Id header 透传
const traceIDHeader = "X-Trace-Id"

func getOrGenTraceID(c *app.RequestContext) string {
	tid := string(c.GetHeader(traceIDHeader))
	if tid == "" {
		tid = uuid.New().String()
	}
	c.Response.Header.Set(traceIDHeader, tid)
	return tid
}

func Success(c *app.RequestContext, data interface{}) {
	c.JSON(consts.StatusOK, Response{
		Code:    errors.CodeSuccess,
		Message: errors.GetMessage(errors.CodeSuccess),
		Data:    data,
		TraceID: getOrGenTraceID(c),
	})
}

func SuccessWithMessage(c *app.RequestContext, message string, data interface{}) {
	c.JSON(consts.StatusOK, Response{
		Code:    errors.CodeSuccess,
		Message: message,
		Data:    data,
		TraceID: getOrGenTraceID(c),
	})
}

// Error 旧版 errors 码的错误响应（兼容老调用方）
func Error(c *app.RequestContext, code int) {
	httpStatus := consts.StatusOK
	if code == errors.CodeUnauthorized {
		httpStatus = consts.StatusUnauthorized
	}
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: errors.GetMessage(code),
		TraceID: getOrGenTraceID(c),
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
		TraceID: getOrGenTraceID(c),
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
		TraceID: getOrGenTraceID(c),
	})
}

// NewErrorByCode 新版业务码错误响应（推荐）
// 自动根据业务码选择 HTTP 状态：
//   401/403/404/429/500/503 等
func NewErrorByCode(c *app.RequestContext, code int) {
	c.JSON(httpStatusForCode(code), Response{
		Code:    code,
		Message: errcode.Message(code),
		TraceID: getOrGenTraceID(c),
	})
}

// NewErrorWithMessage 新版业务码 + 自定义 message
func NewErrorWithMessage(c *app.RequestContext, code int, message string) {
	c.JSON(httpStatusForCode(code), Response{
		Code:    code,
		Message: message,
		TraceID: getOrGenTraceID(c),
	})
}

// NewErrorWithData 新版业务码 + data
func NewErrorWithData(c *app.RequestContext, code int, message string, data interface{}) {
	c.JSON(httpStatusForCode(code), Response{
		Code:    code,
		Message: message,
		Data:    data,
		TraceID: getOrGenTraceID(c),
	})
}

// httpStatusForCode 业务码 → HTTP 状态码 映射
func httpStatusForCode(code int) int {
	switch {
	case code == errcode.CodeSuccess:
		return consts.StatusOK
	case code == errcode.CodeUnauthorized, code == errcode.CodeTokenInvalid, code == errcode.CodeTokenExpired:
		return consts.StatusUnauthorized
	case code == errcode.CodeForbidden, code == errcode.CodePermissionDeny:
		return consts.StatusForbidden
	case code == errcode.CodeNotFound, code == errcode.CodeShareNotFound, code == errcode.CodeFileNotFound, code == errcode.CodePickupCodeNotFound, code == errcode.CodeUserNotFound:
		return consts.StatusNotFound
	case code == errcode.CodeRateLimit:
		return consts.StatusTooManyRequests
	case code == errcode.CodeMethodNotAllowed:
		return consts.StatusMethodNotAllowed
	case code == errcode.CodeTooLarge:
		return consts.StatusRequestEntityTooLarge
	case code == errcode.CodeUnavailable:
		return consts.StatusServiceUnavailable
	case code == errcode.CodeTimeout:
		return consts.StatusRequestTimeout
	case errcode.IsServerError(code):
		return consts.StatusInternalServerError
	case errcode.IsClientError(code):
		return consts.StatusBadRequest
	default:
		return consts.StatusOK
	}
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
