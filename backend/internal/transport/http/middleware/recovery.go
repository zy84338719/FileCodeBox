package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	resp "github.com/zy84338719/fileCodeBox/internal/pkg/resp"
)

func Recovery() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if err := recover(); err != nil {
				resp.InternalError(c, "Internal server error")
				c.Abort()
			}
		}()
		c.Next(ctx)
	}
}
