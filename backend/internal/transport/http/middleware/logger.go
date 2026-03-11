package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	log "github.com/zy84338719/fileCodeBox/internal/pkg/logger"
	"go.uber.org/zap"
)

func Logger() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		path := string(c.URI().Path())
		method := string(c.Method())

		defer func() {
			latency := time.Since(start)
			status := c.Response.StatusCode()
			clientIP := c.ClientIP()

			log.Info("HTTP Request",
				zap.Int("status", status),
				zap.String("method", method),
				zap.String("path", path),
				zap.String("ip", clientIP),
				zap.Duration("latency", latency),
			)
		}()

		c.Next(ctx)
	}
}
