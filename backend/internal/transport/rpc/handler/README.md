# internal/transport/rpc/handler/ - RPC 请求处理器

此目录存放 Kitex RPC 服务接口的实现代码。

## 职责

- 实现 `gen/rpc/` 中定义的服务接口
- 调用 `internal/app/` 中的业务服务
- 组装 RPC 响应

## 示例实现

```go
package handler

import (
    "context"
    
    "github.com/zy84338719/fileCodeBox/backend/gen/rpc/base"
    "github.com/zy84338719/fileCodeBox/backend/gen/rpc/user"
    userSvc "github.com/zy84338719/fileCodeBox/backend/internal/app/user"
)

type UserServiceImpl struct {
    svc *userSvc.Service
}

func NewUserServiceImpl() *UserServiceImpl {
    return &UserServiceImpl{
        svc: userSvc.NewService(),
    }
}

// CreateUser 实现创建用户 RPC 接口
func (s *UserServiceImpl) CreateUser(ctx context.Context, req *user.CreateUserReq) (*user.CreateUserResp, error) {
    result, err := s.svc.Create(ctx, &userSvc.CreateUserReq{
        Username: req.Username,
        Email:    req.Email,
        Password: req.Password,
        Nickname: req.Nickname,
    })
    
    if err != nil {
        return &user.CreateUserResp{
            BaseResp: &base.BaseResp{Code: 1, Message: err.Error()},
        }, nil
    }
    
    return &user.CreateUserResp{
        BaseResp: &base.BaseResp{Code: 0, Message: "success"},
        User: &user.User{
            Id:       int64(result.ID),
            Username: result.Username,
            Email:    result.Email,
            Nickname: result.Nickname,
        },
    }, nil
}

// GetUser 实现获取用户 RPC 接口
func (s *UserServiceImpl) GetUser(ctx context.Context, req *user.GetUserReq) (*user.GetUserResp, error) {
    result, err := s.svc.GetByID(ctx, uint(req.Id))
    if err != nil {
        return &user.GetUserResp{
            BaseResp: &base.BaseResp{Code: 1, Message: err.Error()},
        }, nil
    }
    
    return &user.GetUserResp{
        BaseResp: &base.BaseResp{Code: 0, Message: "success"},
        User: &user.User{
            Id:       int64(result.ID),
            Username: result.Username,
            Email:    result.Email,
        },
    }, nil
}
```

## 启动 RPC 服务

```go
package main

import (
    "github.com/zy84338719/fileCodeBox/backend/gen/rpc/user/userservice"
    "github.com/zy84338719/fileCodeBox/backend/internal/transport/rpc/handler"
)

func main() {
    svr := userservice.NewServer(handler.NewUserServiceImpl())
    
    if err := svr.Run(); err != nil {
        panic(err)
    }
}
```

## 命名规范

- 文件名：`{service}_impl.go`（如 `user_impl.go`）
- 结构体：`{Service}Impl`（如 `UserServiceImpl`）
