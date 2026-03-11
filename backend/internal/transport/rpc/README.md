# internal/transport/rpc/ - RPC 协议适配

此目录存放 Kitex RPC 服务的协议适配代码。

## 目录结构

```
rpc/
└── handler/    # RPC 请求处理器实现
```

## 职责

- 实现 Kitex 生成的服务接口
- 调用 `internal/app/` 中的业务服务
- 处理 RPC 请求/响应转换
- RPC 特定的错误处理

## 与 gen/rpc 的关系

| 目录 | 内容 | 修改 |
|------|------|------|
| `gen/rpc/` | Kitex 生成的代码 | 禁止修改 |
| `internal/transport/rpc/` | 服务实现代码 | 手写 |

## 示例

```go
package rpc

import (
    "github.com/zy84338719/fileCodeBox/backend/gen/rpc/user"
    userSvc "github.com/zy84338719/fileCodeBox/backend/internal/app/user"
)

// UserServiceImpl 实现 Kitex 生成的 UserService 接口
type UserServiceImpl struct {
    svc *userSvc.Service
}

func NewUserServiceImpl() *UserServiceImpl {
    return &UserServiceImpl{
        svc: userSvc.NewService(),
    }
}
```

## 依赖规则

- 可以依赖：`internal/app/`、`internal/pkg/`、`gen/rpc/`
- 不应直接依赖：`internal/repo/`
