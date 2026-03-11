# internal/transport/http/middleware/ - HTTP 中间件

此目录存放 HTTP 全局中间件实现。

## 现有中间件

- `cors.go` - 跨域资源共享配置
- `logger.go` - 请求日志记录
- `recovery.go` - 异常恢复（防止 panic 导致服务崩溃）

## 添加新中间件

```go
package middleware

import (
    "context"
    "github.com/cloudwego/hertz/pkg/app"
)

func NewMiddleware() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        // 前置处理
        c.Next(ctx)
        // 后置处理
    }
}
```

## 注册中间件

在 `cmd/server/bootstrap/bootstrap.go` 中注册：

```go
h.Use(middleware.NewMiddleware())
```

## 常用中间件

- 认证（Auth）
- 限流（RateLimit）
- 熔断（CircuitBreaker）
- 链路追踪（Tracing）
- 请求 ID（RequestID）
