# 路由结构说明

本项目已将路由按照功能模块拆分到不同的文件中，提高代码的可维护性和可读性。

## 文件结构

```
internal/routes/
├── routes.go       # 主路由文件（保持向后兼容）
├── setup.go        # 路由整合设置
├── base.go         # 基础路由（首页、健康检查、配置获取）
├── share.go        # 分享相关路由 (/share/*)
├── user.go         # 用户相关路由 (/user/*)
├── chunk.go        # 分片上传路由 (/chunk/*)
├── admin.go        # 管理员路由 (/admin/*)
└── README.md       # 本说明文档
```

## 路由分类

### 1. 基础路由 (base.go)
- `GET /` - 首页
- `POST /` - 获取系统配置
- `GET /health` - 健康检查
- `GET /robots.txt` - robots.txt

### 2. 分享路由 (share.go) 
**路径前缀**: `/share`
- `POST /share/text/` - 分享文本
- `POST /share/file/` - 分享文件
- `GET|POST /share/select/` - 获取分享内容
- `GET /share/download` - 下载文件

### 3. 用户路由 (user.go)
**路径前缀**: `/user`

**API 路由**:
- `POST /user/register` - 用户注册
- `POST /user/login` - 用户登录
- `GET /user/system-info` - 获取系统信息
- `POST /user/logout` - 用户登出（需认证）
- `GET /user/profile` - 获取用户资料（需认证）
- `PUT /user/profile` - 更新用户资料（需认证）
- `POST /user/change-password` - 修改密码（需认证）
- `GET /user/files` - 获取用户文件（需认证）
- `GET /user/stats` - 获取用户统计（需认证）
- `GET /user/check-auth` - 检查认证状态（需认证）
- `DELETE /user/files/:id` - 删除用户文件（需认证）

**页面路由**:
- `GET /user/login` - 登录页面
- `GET /user/register` - 注册页面
- `GET /user/dashboard` - 用户仪表板
- `GET /user/forgot-password` - 忘记密码页面

### 4. 分片上传路由 (chunk.go)
**路径前缀**: `/chunk`
- `POST /chunk/upload/init/` - 初始化分片上传
- `POST /chunk/upload/chunk/:upload_id/:chunk_index` - 上传分片
- `POST /chunk/upload/complete/:upload_id` - 完成上传
- `GET /chunk/upload/status/:upload_id` - 获取上传状态
- `POST /chunk/upload/verify/:upload_id/:chunk_index` - 验证分片
- `DELETE /chunk/upload/cancel/:upload_id` - 取消上传

### 5. 管理员路由 (admin.go)
**路径前缀**: `/admin`

**页面路由**:
- `GET /admin/` - 管理页面

**API 路由**:
- `POST /admin/login` - 管理员登录

**需要认证的路由**:
- `GET /admin/dashboard` - 仪表板
- `GET /admin/stats` - 统计信息
- `GET /admin/files` - 文件列表
- `GET /admin/files/:code` - 获取文件信息
- `DELETE /admin/files/:code` - 删除文件
- `PUT /admin/files/:code` - 更新文件
- `GET /admin/files/download` - 下载文件
- `GET /admin/config` - 获取配置
- `PUT /admin/config` - 更新配置
- `POST /admin/clean` - 清理过期文件

**用户管理**:
- `GET /admin/users` - 用户列表
- `GET /admin/users/:id` - 获取用户信息
- `POST /admin/users` - 创建用户
- `PUT /admin/users/:id` - 更新用户
- `DELETE /admin/users/:id` - 删除用户
- `PUT /admin/users/:id/status` - 更新用户状态
- `GET /admin/users/:id/files` - 获取用户文件

**存储管理**:
- `GET /admin/storage` - 获取存储信息
- `POST /admin/storage/switch` - 切换存储
- `GET /admin/storage/test/:type` - 测试存储连接
- `PUT /admin/storage/config` - 更新存储配置

## 使用方式

### 在 main.go 中使用
```go
// 方式1：使用原有接口（向后兼容）
routes.SetupRoutes(router, shareHandler, chunkHandler, adminHandler, storageHandler, userHandler, cfg, userService)

// 方式2：使用新的整合接口
routes.SetupAllRoutes(router, shareHandler, chunkHandler, adminHandler, storageHandler, userHandler, cfg, userService)

// 方式3：按需设置特定模块路由
routes.SetupBaseRoutes(router, cfg)
routes.SetupShareRoutes(router, shareHandler, cfg, userService)
routes.SetupUserRoutes(router, userHandler, cfg, userService)
routes.SetupChunkRoutes(router, chunkHandler, cfg)
routes.SetupAdminRoutes(router, adminHandler, storageHandler, cfg)
```

## 优势

1. **模块化**: 每个功能模块的路由独立管理
2. **可维护性**: 修改特定功能的路由时不影响其他模块
3. **可读性**: 路由定义更加清晰，易于理解
4. **可扩展性**: 新增功能模块时只需添加对应的路由文件
5. **向后兼容**: 保留原有的 SetupRoutes 函数，不影响现有代码

## 注意事项

- 所有函数名已改为公开（首字母大写），便于跨包调用
- 保持了原有的中间件和认证逻辑
- 路由的逻辑和功能完全保持不变，仅做了结构上的拆分
