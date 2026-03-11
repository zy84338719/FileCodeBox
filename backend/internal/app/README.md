# internal/app/ - 应用层（业务逻辑）

此目录存放业务逻辑代码，是整个应用的核心层。

## 目录结构

```
app/
└── {domain}/           # 按业务领域划分
    ├── service.go      # 服务实现
    ├── domain/         # 领域模型（可选）
    └── dto/            # 数据传输对象（可选）
```

## 职责

- 实现核心业务逻辑
- 协调多个 Repository 完成复杂操作
- 事务管理
- 业务规则验证

## 示例

```go
// internal/app/user/service.go
type Service struct {
    userRepo *dao.UserRepository
}

func NewService() *Service {
    return &Service{
        userRepo: dao.NewUserRepository(),
    }
}

func (s *Service) Create(ctx context.Context, req *CreateUserReq) (*model.UserResp, error) {
    // 业务逻辑...
}
```

## 依赖规则

- 可以依赖：`internal/repo/`、`internal/pkg/`
- 不应依赖：`internal/transport/`、`gen/`
- 应被依赖于：`internal/transport/`、`gen/http/handler/`
