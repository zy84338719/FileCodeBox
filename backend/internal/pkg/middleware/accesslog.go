package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/fileCodeBox/backend/internal/pkg/logger"
	"go.uber.org/zap"
)

// AccessLog 结构化访问日志中间件。
//
// 记录每个请求的 method/path/status/ip/latency/trace_id，便于生产环境排障。
// 与 RequestID 中间件配合：trace_id 由 RequestID 写入 ctx 与 RequestContext，
// 本中间件从 RequestContext 读取并打入日志字段，实现日志与 resp envelope 的
// trace_id 一致，方便串联请求链路。
func AccessLog() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		path := string(c.Request.URI().Path())
		method := string(c.Method())

		c.Next(ctx)

		latency := time.Since(start)
		status := c.Response.StatusCode()
		ip := c.ClientIP()

		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.String("ip", ip),
			zap.Duration("latency", latency),
			zap.Int("size", len(c.Response.Body())),
		}
		// 注入 trace_id（由 RequestID 中间件设置）
		if tid, ok := c.Get("trace_id"); ok {
			fields = append(fields, zap.Any("trace_id", tid))
		}

		// 按 HTTP 状态码分级日志
		switch {
		case status >= 500:
			logger.Error("http request", fields...)
		case status >= 400:
			logger.Warn("http request", fields...)
		default:
			logger.Info("http request", fields...)
		}
	}
}
