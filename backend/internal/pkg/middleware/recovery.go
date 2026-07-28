package middleware

import (
	"context"
	"runtime/debug"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/zy84338719/fileCodeBox/backend/internal/pkg/logger"
	"github.com/zy84338719/fileCodeBox/backend/internal/pkg/resp"
	"go.uber.org/zap"
)

// Recovery 全局 panic 恢复中间件。
//
// 捕获 handler 链中任何 panic，记录堆栈与 trace_id，并返回统一 500 响应，
// 防止单个请求的 panic 导致整个进程崩溃（Hertz 默认不保证全链路 recover）。
//
// 必须注册在其他中间件之前，以覆盖后续所有 handler。
func Recovery() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				// 尽量带上 trace_id 便于排障
				tid, _ := c.Get("trace_id")
				logger.Error("panic recovered",
					zap.Any("panic", r),
					zap.Any("trace_id", tid),
					zap.ByteString("stack", stack),
					zap.String("path", string(c.Request.URI().Path())),
					zap.String("method", string(c.Method())),
				)
				// 已开始写入响应则只能终止，避免重复写 header
				if !c.Response.IsBodyStream() && len(c.Response.Body()) == 0 {
					resp.InternalError(c, "Internal server error")
				}
				c.AbortWithStatus(consts.StatusInternalServerError)
			}
		}()
		c.Next(ctx)
	}
}
