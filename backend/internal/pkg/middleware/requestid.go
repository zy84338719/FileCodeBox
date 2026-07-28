package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
)

// TraceIDHeader 请求链路 ID 透传 header，与 internal/pkg/resp 对齐。
const TraceIDHeader = "X-Trace-Id"

// CtxKeyTraceID context.Context 中 trace_id 的键。
type CtxKeyTraceID struct{}

// RequestID 为每个请求生成 / 透传 trace_id（X-Trace-Id）。
//
// 行为：
//   - 优先沿用请求头 X-Trace-Id（便于上游链路串联）
//   - 缺失时生成 UUID
//   - 写入响应头、app.RequestContext（Set）与 context.Context，
//     使访问日志 / 业务日志 / resp envelope 都能拿到同一个 trace_id。
//
// 与 internal/pkg/resp.getOrGenTraceID 配合：resp 包在构造响应体时
// 会读取同样的 header，保证全链路 trace_id 一致。
func RequestID() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		tid := string(c.GetHeader(TraceIDHeader))
		if tid == "" {
			tid = uuid.New().String()
		}
		// 写响应头，供客户端 / 代理记录
		c.Response.Header.Set(TraceIDHeader, tid)
		// 写入 RequestContext，handler 可通过 c.Get 取用
		c.Set("trace_id", tid)
		// 写入 context.Context，下游 service / logger 可取出
		c.Set("ctx", context.WithValue(ctx, CtxKeyTraceID{}, tid))

		c.Next(ctx)
	}
}

// TraceIDFromContext 从 context.Context 提取 trace_id（service / 日志用）。
func TraceIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(CtxKeyTraceID{}).(string); ok {
		return v
	}
	return ""
}
