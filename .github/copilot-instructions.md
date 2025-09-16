# FileCodeBox AI Coding Agent Instructions

## Project Overview

FileCodeBox 是一个高性能的文件快传系统的 Go 实现，基于现代化的分层架构设计。该项目提供文件/文本分享、分片上传、用户系统、多存储后端支持，以及独特的 Model Context Protocol (MCP) 服务器集成。

## Architecture Patterns

### Layered Architecture
遵循严格的分层架构：`routes → handlers → services → repository → database/storage`

- **Routes** (`internal/routes/`): 按功能模块化路由 - `base.go`, `share.go`, `user.go`, `chunk.go`, `admin.go`
- **Handlers** (`internal/handlers/`): HTTP 处理器，负责请求解析和响应
- **Services** (`internal/services/`): 业务逻辑层，处理核心业务
- **Repository** (`internal/repository/`): 数据访问层，提供统一的 DAO 接口
- **Storage** (`internal/storage/`): 存储抽象层，支持 local/s3/webdav/onedrive

### Configuration Management
采用分层配置管理模式 (`internal/config/`):
```go
// 通过 ConfigManager 统一管理所有配置
manager := config.InitManager()
manager.SetDB(db) // 注入数据库连接（配置读取现在以 config.yaml 和 环境变量为准）
```

配置分为多个模块：`BaseConfig`, `DatabaseConfig`, `StorageConfig`, `UserSystemConfig`, `MCPConfig`

支持环境变量优先级覆盖、数据库持久化存储、热重载机制：
- **环境变量优先级**：PORT、ADMIN_TOKEN 等关键配置始终优先使用环境变量
- **数据库持久化**：配置自动保存到 key_value 表，支持动态更新
- **热重载机制**：通过 ReloadConfig() 方法实现运行时配置更新
- **配置验证**：每个配置模块都有独立的验证方法
- **分层映射**：ToMap() 和 FromMap() 方法支持配置的序列化和反序列化

### Route Architecture
完全模块化的路由系统 (`internal/routes/`):
```go
// 自动化服务器创建和启动
srv, err := routes.CreateAndStartServer(manager, daoManager, storageManager)

// 分层路由管理
SetupAllRoutesWithDependencies(router, manager, daoManager, storageManager)
```

路由按功能分组：`base.go`, `share.go`, `user.go`, `chunk.go`, `admin.go`
自动依赖注入：服务→处理器→路由的完整初始化链

### Service Layer Architecture  
现代化依赖注入模式：
```go
// 服务层自动初始化
userService := services.NewUserService(daoManager, manager)
shareService := services.NewShareService(daoManager, manager, storageService, userService)
adminService := services.NewAdminService(daoManager, manager, storageService)
```

支持服务间依赖、统一错误处理、自动资源清理

### Model Organization
模型按用途分层组织 (`internal/models/`):
- `db/`: 数据库实体模型
- `web/`: API 请求/响应模型  
- `service/`: 服务层传输对象
- `dto/`: 数据传输对象
- `mcp/`: MCP 协议相关模型

主模型文件 (`internal/models/models.go`) 通过类型别名提供向后兼容性：
```go
// 数据库模型别名
type FileCode = db.FileCode
type User = db.User

// 服务模型别名  
type BuildInfo = service.BuildInfo
type ShareFileRequest = service.ShareFileRequest

// DTO 模型别名
type UserUpdateFields = dto.UserUpdateFields

// MCP 模型别名
type SystemConfigResponse = mcp.SystemConfigResponse
```

## Development Workflows

### Build System
```bash
# 开发构建 (含版本信息)
make build

# 发布构建 (优化编译)
make release

# 交叉编译
GOOS=linux GOARCH=amd64 make build-cross

# 开发热重载 (需要 air)
make dev
```

### Testing
```bash
# 运行所有测试
make test

# 运行测试脚本套件 
cd tests && ./run_all_tests.sh

# 特定功能测试
./tests/test_api.sh
./tests/test_chunk.sh
./tests/test_mcp_client.py
```

### Docker Deployment
```bash
# 多架构构建
./scripts/build-docker.sh

# 单架构构建
./scripts/build-docker.sh --single linux/arm64

# Docker Compose 部署
docker-compose up -d
```

## Project-Specific Conventions

### Error Handling
使用统一的响应格式 (`internal/common/`):
```go
common.SuccessResponse(c, data)
common.ErrorResponse(c, code, message)
common.BadRequestResponse(c, message)
```

### API Design
- **Swagger 集成**: 所有 API 都有详细的 Swagger 注释
- **JWT 认证**: 支持管理员和用户双重认证
- **分片上传**: 完整的断点续传实现 (`/chunk/*` 路由)

### Handler Constructor Pattern
所有处理器都遵循统一的构造模式:
```go
func NewXxxHandler(service *services.XxxService, config *config.ConfigManager) *XxxHandler
```

### Service Layer Integration
服务层使用依赖注入模式，通过 `routes.CreateAndStartServer()` 统一初始化:
```go
srv, err := routes.CreateAndStartServer(manager, daoManager, storageManager)
```

## Unique Features

### MCP Server Integration
FileCodeBox 集成了完整的 Model Context Protocol 服务器 (`internal/mcp/`):

- **启用方式**: 环境变量 `ENABLE_MCP_SERVER=true` 或配置文件
- **管理界面**: 完整的管理界面在 `/admin/` 下的 MCP 标签页
- **工具集**: 8个核心工具 (文件分享、用户管理、系统监控等)
- **协议版本**: MCP 2024-11-05

关键文件:
- `internal/mcp/manager.go`: MCP 服务器生命周期管理
- `internal/mcp/filecodebox.go`: FileCodeBox 特定的 MCP 工具实现
- `scripts/test_mcp_client.py`: MCP 客户端测试工具

### Multi-Storage Backend
动态存储后端切换支持:
```go
// 存储策略模式
storageManager := storage.NewStorageManager(configManager)
// 支持: local, s3, webdav, onedrive
```

### Theme System
主题系统位于 `themes/` 目录，支持完整的主题切换:
- 前端: `themes/2025/`
- 管理后台: `themes/2025/admin/` (模块化 JS 架构)

### Advanced Upload Features
- **分片上传**: 支持大文件分片上传和断点续传
- **秒传功能**: 基于文件哈希的重复检测
- **用户系统**: 可选的用户认证和权限控制

## Development Guidelines

### When Modifying Routes
新增路由时在对应的 `internal/routes/` 文件中添加，保持模块化原则。

### When Adding Storage Backends
实现 `storage.StorageInterface` 接口，并在 `storage.NewStorageManager` 中注册。

### When Working with Config
使用 `ConfigManager` 统一访问配置，避免直接访问环境变量。

### When Adding MCP Tools
在 `internal/mcp/filecodebox.go` 中添加新工具，遵循现有的工具注册模式。

### Testing Strategy
- 单元测试: `go test ./...`
- 集成测试: `tests/` 目录下的 shell 脚本
- API 测试: 通过 Swagger UI (`/swagger/index.html`) 或脚本

## Key Files to Understand

- `main.go`: 应用程序入口点，采用现代化依赖注入模式，包含完整的初始化流程（配置管理→数据库→服务→MCP→服务器启动）和优雅关闭机制
- `internal/routes/setup.go`: 完全模块化的路由和服务器初始化系统，提供自动化依赖注入和分层路由管理（CreateAndStartServer、SetupAllRoutesWithDependencies）
- `internal/config/manager.go`: 统一配置管理核心，支持环境变量+数据库双重配置源，热重载机制和完整的配置验证
- `internal/models/models.go`: 模型统一导出层，通过类型别名提供向后兼容性，整合 db/service/dto/mcp 四层模型架构
- `internal/repository/manager.go`: Repository 模式实现，提供统一的数据访问抽象层，替代原有 DAO 模式，支持事务管理
- `docs/changelogs/REFACTOR_SUMMARY.md`: 完整的架构演进记录，包含数据库多类型支持、模块化路由系统、服务依赖自动化的实现细节
- `Dockerfile`: 优化的多阶段构建配置，支持 CGO 编译、最小化运行时镜像、非 root 用户安全实践
  - **构建阶段**：使用 golang:1.24-alpine 作为构建环境，安装 gcc、musl-dev、sqlite-dev 等 CGO 依赖
  - **运行时阶段**：使用 alpine:latest 最小化镜像，只包含必要的运行时依赖 (ca-certificates、tzdata、sqlite)
  - **安全实践**：创建非 root 用户 (app:1000)，设置适当的文件权限，暴露标准端口 12345
  - **构建优化**：使用 CGO_ENABLED=1 支持 SQLite，通过 -ldflags="-w -s" 减小二进制文件大小

此项目强调模块化、可扩展性和高性能，在修改时请保持现有的架构模式和编码约定。