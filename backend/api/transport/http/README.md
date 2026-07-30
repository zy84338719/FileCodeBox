# internal/transport/http/ - HTTP 协议适配

此目录存放 HTTP 协议相关的适配代码。

## 目录结构

```
http/
├── handler/        # HTTP 请求处理器实现（复杂场景）
└── middleware/     # HTTP 中间件
```

## 内容说明

### handler/
- 复杂业务场景的 handler 实现
- 简单场景直接在 `gen/http/handler/` 中处理即可

### middleware/
- 全局 HTTP 中间件
- 认证、日志、CORS、限流、熔断等

## 示例

```go
// middleware/auth.go
func Auth() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        // 认证逻辑
        c.Next(ctx)
    }
}
```
