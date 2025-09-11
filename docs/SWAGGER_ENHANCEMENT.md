# Swagger API 文档增强说明

## 概述

本文档记录了对 FileCodeBox 项目 Swagger API 文档的全面增强工作。通过这次升级，API 文档更加完整、专业，覆盖了所有主要功能模块。

## 增强内容

### 1. 文档结构完善

- **完整的 API 信息**: 添加了详细的 API 描述、联系信息、许可证信息
- **安全定义**: 定义了 `ApiKeyAuth`、`BearerAuth`、`BasicAuth` 三种认证方式
- **标签分类**: 将 API 按功能分为 6 个主要标签：
  - 系统 (System)
  - 分享 (Share)
  - 分片上传 (Chunk Upload)
  - 用户 (User)
  - 管理员 (Admin)
  - 存储 (Storage)

### 2. API 端点扩展

#### 系统接口
- `GET /health` - 健康检查 (增强版)
- `GET /api/config` - 获取系统配置

#### 分享接口
- `POST /share/text/` - 分享文本内容 (增强)
- `POST /share/file/` - 分享文件 (增强)
- `GET /share/select/` - 获取分享信息 (增强)
- `POST /share/select/` - 获取分享信息 POST 方式 (增强)
- `GET /share/download` - 下载分享文件 (增强)

#### 分片上传接口
- `POST /chunk/upload/init/` - 初始化分片上传 (增强)
- `POST /chunk/upload/chunk/{upload_id}/{chunk_index}` - 上传文件分片 (增强)
- `POST /chunk/upload/complete/{upload_id}` - 完成分片上传 (增强)
- `GET /chunk/upload/status/{upload_id}` - 获取上传状态 (新增)
- `DELETE /chunk/upload/cancel/{upload_id}` - 取消分片上传 (新增)

#### 用户接口 (新增)
- `POST /user/register` - 用户注册
- `POST /user/login` - 用户登录
- `POST /user/logout` - 用户退出
- `GET /user/profile` - 获取用户信息
- `PUT /user/profile` - 更新用户信息
- `GET /user/files` - 获取用户文件列表

#### 管理员接口 (新增)
- `GET /admin/stats` - 获取系统统计
- `GET /admin/files` - 获取所有文件列表
- `DELETE /admin/files/{id}` - 删除文件
- `GET /admin/config` - 获取系统配置
- `PUT /admin/config` - 更新系统配置

#### 存储管理接口 (新增)
- `POST /admin/storage/test` - 测试存储连接
- `POST /admin/storage/switch` - 切换存储方式

### 3. 数据模型完善

#### 通用响应模型
- `SuccessResponse` - 成功响应
- `ErrorResponse` - 错误响应
- `Pagination` - 分页信息

#### 系统相关模型
- `SystemConfig` - 系统配置
- `HealthResponse` - 健康检查响应

#### 分享相关模型
- `ShareResponse` - 分享响应
- `ShareInfo` - 分享信息

#### 分片上传模型
- `InitUploadRequest` / `InitUploadResponse` - 初始化上传
- `CompleteUploadRequest` - 完成上传请求
- `UploadChunkResponse` - 分片上传响应
- `ChunkStatusResponse` - 分片状态响应

#### 用户相关模型
- `RegisterRequest` - 注册请求
- `LoginRequest` / `LoginResponse` - 登录请求/响应
- `UserResponse` - 用户响应
- `UpdateUserRequest` - 更新用户请求
- `FileListResponse` / `FileItem` - 文件列表

#### 管理员相关模型
- `AdminStatsResponse` - 管理员统计响应
- `AdminFileListResponse` / `AdminFileItem` - 管理员文件列表
- `AdminConfigRequest` / `AdminConfigResponse` - 配置管理

#### 存储相关模型
- `StorageTestRequest` / `StorageTestResponse` - 存储测试
- `StorageSwitchRequest` - 存储切换

### 4. 安全性增强

- **认证机制**: 为需要认证的端点添加了安全要求
- **权限控制**: 区分了用户级别和管理员级别的 API
- **参数验证**: 为所有参数添加了详细的验证规则和示例

### 5. 错误处理

- **标准化错误码**: 定义了标准的 HTTP 状态码
- **详细错误信息**: 为每个可能的错误情况提供了说明
- **错误响应模型**: 统一的错误响应格式

## 文件变更

### 新增文件
- `docs/swagger-enhanced.yaml` - 增强版 Swagger 文档
- `docs/SWAGGER_ENHANCEMENT.md` - 本文档

### 修改文件
- `docs/swagger.yaml` - 更新为增强版本
- `docs/docs.go` - 更新生成的 Swagger 代码
- `internal/handlers/api.go` - 增强 API 处理器
- `internal/routes/base.go` - 添加新的路由
- `internal/models/models.go` - 添加 API 响应结构体

## 使用说明

### 访问 Swagger UI
启动应用后，访问 `http://localhost:12345/swagger/index.html` 查看完整的 API 文档。

### API 测试
1. 系统接口无需认证，可直接测试
2. 用户接口需要先注册/登录获取 JWT Token
3. 管理员接口需要管理员 API Key
4. 分享和分片上传接口支持匿名和认证两种模式

### 认证方式
- **BearerAuth**: 在 Authorization header 中使用 `Bearer <token>` 格式
- **ApiKeyAuth**: 在 X-API-Key header 中传递管理员密钥
- **BasicAuth**: 基础认证 (部分接口支持)

## 技术特点

1. **标准化**: 严格遵循 OpenAPI 2.0 规范
2. **完整性**: 覆盖了所有主要功能的 API
3. **可测试性**: 提供了详细的示例和参数说明
4. **可维护性**: 结构清晰，易于扩展和维护
5. **国际化**: 支持中文标签和描述，便于国内开发者使用

## 后续计划

1. **API 实现**: 根据文档实现缺失的 API 端点
2. **自动化测试**: 基于 Swagger 文档生成自动化测试用例
3. **客户端生成**: 利用 Swagger Codegen 生成多语言客户端 SDK
4. **持续更新**: 随着功能迭代持续更新 API 文档

## 总结

通过这次 Swagger API 文档的全面增强，FileCodeBox 项目的 API 文档达到了企业级标准。文档不仅提供了完整的 API 参考，还包含了详细的示例和说明，大大提升了开发者体验和项目的专业性。
