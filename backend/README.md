# FileCodeBox

> ⚠️ **重要提醒**：首次使用请务必阅读[快速开始](#快速开始)部分，需要先执行代码生成命令！

基于 CloudWeGo Hertz + Kitex 的 Go 微服务脚手架模板。

集成 Hz HTTP 代码生成、Kitex RPC 代码生成、SQLite/MySQL/PostgreSQL、Redis。

## 特性

- **Hz IDL 驱动**: 使用 Proto 文件定义 HTTP API，自动生成路由和模型
- **Kitex RPC**: 支持 Protobuf RPC 服务定义和代码生成
- **分层架构**: gen（生成代码）/ internal（手写代码）/ configs（配置文件）清晰分离
- **多数据库支持**: SQLite（纯 Go 无 CGO）/ MySQL / PostgreSQL
- **Redis**: 缓存支持，封装常用操作
- **统一响应**: 标准化的 JSON 响应格式
- **中间件**: CORS、Recovery、Logger
- **配置管理**: YAML 配置 + 环境变量

## 项目结构

```
FileCodeBox/
├── cmd/server/                 # 服务入口
│   ├── main.go                 # 程序入口
│   └── bootstrap/              # 初始化代码
│       └── bootstrap.go
├── configs/                    # 配置文件（仅 YAML）
│   └── config.yaml
├── gen/                        # 自动生成代码（禁止手动修改）
│   ├── http/                   # Hz 生成的 HTTP 代码
│   │   ├── handler/            # 请求处理器
│   │   ├── router/             # 路由注册
│   │   └── model/              # 请求/响应模型
│   └── rpc/                    # Kitex 生成的 RPC 代码
├── idl/                        # 接口定义文件
│   ├── api/api.proto           # HTTP 注解定义
│   ├── http/                   # HTTP 服务 IDL
│   │   └── health.proto        # 健康检查
│   └── rpc/                    # RPC 服务 IDL
│       └── health.proto        # RPC 探活
├── internal/                   # 项目私有代码（手写）
│   ├── app/                    # 应用层：业务逻辑
│   │   └── user/               # 用户服务示例
│   ├── transport/              # 传输层：协议适配
│   │   ├── http/               # HTTP 适配
│   │   │   ├── handler/        # 复杂 handler 实现
│   │   │   └── middleware/     # 中间件
│   │   └── rpc/                # RPC 适配
│   │       └── handler/        # RPC 服务实现
│   ├── repo/                   # 数据层：数据访问
│   │   ├── db/                 # 数据库
│   │   │   ├── database.go     # 连接初始化
│   │   │   ├── model/          # GORM 模型
│   │   │   └── dao/            # 数据访问对象
│   │   ├── redis/              # Redis 缓存
│   │   └── external/           # 外部服务调用
│   ├── conf/                   # 配置结构体和加载逻辑
│   └── pkg/                    # 内部工具库
│       ├── errors/             # 错误码
│       ├── logger/             # 日志封装
│       └── resp/               # HTTP 响应封装
├── scripts/                    # 脚本
│   └── gen.sh                  # 代码生成脚本
├── docs/                       # 文档
├── Makefile
├── Dockerfile
└── go.mod
```

## 环境要求

- Go 1.21+
- hz（HTTP 代码生成）
- kitex（RPC 代码生成，可选）

## 快速开始

> ⚠️ **重要提示**：首次运行前必须执行代码生成！

```bash
# 1. 安装工具
make tools-install

# 2. 安装依赖
go mod tidy

# 3. 首次运行必须生成代码（重要！）
make gen-http-update IDL=common.proto

# 4. 运行服务
make run

# 5. 构建
make build
```

### 首次运行说明

由于模板中的 `gen/` 目录只包含框架结构，具体的 Handler 和 Model 代码需要通过 IDL 文件生成。因此**首次运行前必须执行代码生成命令**：

```bash
# 生成基础 HTTP 服务代码
make gen-http-update IDL=common.proto

# 或者生成特定的 HTTP 服务
make gen-http-update IDL=http/health.proto

# 批量生成所有 HTTP 服务
make gen-http-update-all
```

执行完代码生成后，项目才能正常编译和运行。

## 代码生成

> 📝 **注意**：新项目首次使用时，必须先执行代码生成命令，否则无法编译！

### HTTP 代码生成 (Hz)

```bash
# 首次初始化项目（推荐）
make gen-http-new IDL=common.proto

# 更新已有项目
make gen-http-update IDL=common.proto

# 批量更新所有 HTTP IDL 文件
make gen-http-update-all

# 强制重新初始化（谨慎使用）
make gen-http-init IDL=common.proto
```

### RPC 代码生成 (Kitex)

```bash
# 生成 RPC 代码
make gen-rpc IDL=rpc/health.proto
```

### 定义新的 HTTP 接口

在 `idl/http/` 目录创建 proto 文件：

```protobuf
// idl/http/example.proto
syntax = "proto3";

package http.example;

option go_package = "github.com/zy84338719/fileCodeBox/backend/gen/http/model/example";

import "api/api.proto";

message HelloReq {
    string name = 1 [(api.query) = "name"];
}

message HelloResp {
    string message = 1 [(api.body) = "message"];
}

service ExampleService {
    rpc Hello(HelloReq) returns(HelloResp) {
        option (api.get) = "/api/v1/hello";
    }
}
```

然后生成代码：

```bash
make gen-http-new IDL=http/example.proto
```

### 定义新的 RPC 接口

在 `idl/rpc/` 目录创建 proto 文件：

```protobuf
// idl/rpc/example.proto
syntax = "proto3";

package rpc.example;

option go_package = "github.com/zy84338719/fileCodeBox/backend/gen/rpc/example";

message ExampleReq {
    string name = 1;
}

message ExampleResp {
    string message = 1;
}

service ExampleService {
    rpc Hello(ExampleReq) returns (ExampleResp);
}
```

然后生成代码：

```bash
make gen-rpc IDL=rpc/example.proto
```

## 分层架构

```
请求 → gen/http/handler（参数解析）
         ↓
     internal/app（业务逻辑）
         ↓
     internal/repo（数据访问）
         ↓
     数据库 / Redis / 外部服务
```

### 实现业务逻辑

Handler 调用 app 层服务：

```go
// gen/http/handler/example/example_service.go
func Hello(ctx context.Context, c *app.RequestContext) {
    var req example.HelloReq
    if err := c.BindAndValidate(&req); err != nil {
        resp.BadRequest(c, err.Error())
        return
    }

    svc := exampleSvc.NewService()
    result, err := svc.Hello(ctx, req.Name)
    if err != nil {
        resp.InternalError(c, err.Error())
        return
    }

    resp.Success(c, result)
}
```

App 层实现业务逻辑：

```go
// internal/app/example/service.go
type Service struct {
    repo *dao.ExampleRepository
}

func NewService() *Service {
    return &Service{repo: dao.NewExampleRepository()}
}

func (s *Service) Hello(ctx context.Context, name string) (string, error) {
    return fmt.Sprintf("Hello, %s!", name), nil
}
```

## 数据库操作

### 定义模型

```go
// internal/repo/db/model/user.go
type User struct {
    gorm.Model
    Username string `gorm:"uniqueIndex;size:50" json:"username"`
    Email    string `gorm:"uniqueIndex;size:100" json:"email"`
}

func (User) TableName() string {
    return "users"
}
```

### DAO 层

```go
// internal/repo/db/dao/user.go
type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository() *UserRepository {
    return &UserRepository{db: db.GetDB()}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
    var user model.User
    err := r.db.WithContext(ctx).First(&user, id).Error
    return &user, err
}
```

## Redis 操作

```go
import "github.com/zy84338719/fileCodeBox/backend/internal/repo/redis"

// 基本操作
redis.Set(ctx, "key", "value", time.Hour)
val, err := redis.Get(ctx, "key")
redis.Del(ctx, "key")

// Hash
redis.HSet(ctx, "user:1", "name", "John")
name, _ := redis.HGet(ctx, "user:1", "name")

// List
redis.LPush(ctx, "queue", "item1", "item2")
items, _ := redis.LRange(ctx, "queue", 0, -1)
```

## 统一响应格式

所有 API 返回统一格式：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

使用方式：

```go
import "github.com/zy84338719/fileCodeBox/backend/internal/pkg/resp"

resp.Success(c, data)
resp.Page(c, list, total, page, pageSize)
resp.BadRequest(c, "参数错误")
resp.Unauthorized(c, "未授权")
resp.NotFound(c, "未找到")
resp.InternalError(c, "内部错误")
```

## 配置说明

```yaml
# configs/config.yaml
server:
  host: "0.0.0.0"
  port: 8888

database:
  driver: "sqlite"
  host: "localhost"
  port: 3306
  user: "root"
  password: ""
  db_name: "fileCodeBox"
  ssl_mode: "disable"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

log:
  level: "info"
  filename: ""
  max_size: 100
  max_backups: 10
  max_age: 30
  compress: true

app:
  name: "FileCodeBox"
  version: "1.0.0"
```

支持通过环境变量覆盖：

```bash
CONFIG_PATH=configs/config.prod.yaml ./server
```

## Makefile 命令

| 命令 | 用途 |
|------|------|
| `make run` | 运行服务 |
| `make build` | 编译 |
| `make test` | 运行测试 |
| `make lint` | 代码检查 |
| `make tidy` | 整理依赖 |
| `make gen-http-new IDL=...` | 生成 HTTP 代码 |
| `make gen-http-update IDL=...` | 更新 HTTP 代码 |
| `make gen-rpc IDL=...` | 生成 RPC 代码 |
| `make gen-rpc-all` | 生成所有 RPC |
| `make tools-install` | 安装 hz + kitex |
| `make docker-build` | 构建 Docker 镜像 |

## 技术栈

- [Hertz](https://github.com/cloudwego/hertz) - HTTP 框架
- [Kitex](https://github.com/cloudwego/kitex) - RPC 框架
- [hz](https://github.com/cloudwego/hertz/cmd/hz) - HTTP 代码生成
- [GORM](https://gorm.io/) - ORM
- [go-redis](https://github.com/redis/go-redis) - Redis 客户端
- [Viper](https://github.com/spf13/viper) - 配置管理
- [Zap](https://github.com/uber-go/zap) - 日志

## License

MIT
