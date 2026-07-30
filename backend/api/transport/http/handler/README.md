# internal/transport/http/handler/ - HTTP 请求处理器

此目录存放复杂业务场景的 HTTP 请求处理器实现。

## 与 gen/http/handler 的区别

| 目录 | 用途 | 修改 |
|------|------|------|
| `gen/http/handler/` | Hz 生成的骨架代码 | 可编辑调用逻辑 |
| `internal/transport/http/handler/` | 复杂 handler 实现 | 完全手写 |

## 使用场景

- 需要复杂参数处理的接口
- 需要聚合多个服务的接口
- 自定义响应格式的接口
- WebSocket 等特殊协议处理

## 示例

```go
package handler

import (
    "context"
    
    "github.com/cloudwego/hertz/pkg/app"
    "github.com/zy84338719/fileCodeBox/backend/internal/app/user"
    "github.com/zy84338719/fileCodeBox/backend/internal/pkg/resp"
)

type UserHandler struct {
    userSvc *user.Service
}

func NewUserHandler() *UserHandler {
    return &UserHandler{
        userSvc: user.NewService(),
    }
}

func (h *UserHandler) GetUserProfile(ctx context.Context, c *app.RequestContext) {
    userID := c.Param("id")
    
    // 聚合多个服务数据
    userInfo, _ := h.userSvc.GetByID(ctx, userID)
    // orderInfo, _ := h.orderSvc.GetByUserID(ctx, userID)
    
    resp.Success(c, map[string]interface{}{
        "user": userInfo,
        // "orders": orderInfo,
    })
}
```

## 注册路由

在 `gen/http/router/` 中注册自定义 handler：

```go
// 在路由中间件或自定义路由文件中
userHandler := handler.NewUserHandler()
r.GET("/api/v1/user/:id/profile", userHandler.GetUserProfile)
```

## 依赖规则

- 可以依赖：`internal/app/`、`internal/pkg/`
- 不应直接依赖：`internal/repo/`（通过 app 层访问）
