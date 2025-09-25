# FileCodeBox API 文档

FileCodeBox 已集成 Swagger/OpenAPI 文档系统，提供完整的 API 接口文档。

## Swagger 集成特性

### 🚀 功能特点
- **完整的 API 文档**: 自动生成的 OpenAPI 3.0 规范文档
- **交互式界面**: 在线测试 API 接口
- **实时更新**: 代码注释自动同步到文档
- **多种格式**: 支持 JSON 和 YAML 格式

### 📖 访问文档

启动应用后，可以通过以下方式访问 API 文档：

```bash
# 启动应用
./filecodebox

# 访问 Swagger UI (推荐)
http://localhost:8080/swagger/index.html

# 访问传统 API 文档
http://localhost:8080/api/doc

# 健康检查接口
http://localhost:8080/health

# 获取 OpenAPI JSON 规范
http://localhost:8080/swagger/doc.json
```

### 🔧 开发者使用

#### 重新生成文档
```bash
# 安装 swag 工具 (如果未安装)
go install github.com/swaggo/swag/cmd/swag@latest

# 生成/更新 Swagger 文档
swag init
```

#### 添加 API 注释
在处理器方法上添加 Swagger 注释：

```go
// ShareText 分享文本
// @Summary 分享文本内容
// @Description 分享文本内容并生成分享代码
// @Tags 分享
// @Accept multipart/form-data
// @Produce json
// @Param text formData string true "文本内容"
// @Param expire_value formData int false "过期值" default(1)
// @Success 200 {object} map[string]interface{} "分享成功"
// @Router /share/text/ [post]
func (h *ShareHandler) ShareText(c *gin.Context) {
    // 实现代码...
}
```

### 📋 已集成的 API 分组

| 分组 | 描述 | 端点数量 |
|------|------|----------|
| **系统** | 健康检查、系统信息 | 2 |
| **分享** | 文本分享、文件分享、下载 | 4 |
| **分片上传** | 大文件分片上传管理 | 6 |
| **用户管理** | 用户注册、登录、个人信息 | 8 |
| **管理员** | 后台管理、用户管理、存储管理 | 15+ |
| **API文档** | 文档接口和规范 | 2 |

### 🔐 认证方式

API 支持多种认证方式：

1. **API Key 认证**: 在请求头中添加 `X-API-Key`
2. **Basic 认证**: 用户名密码认证
3. **JWT Token**: Bearer token 认证
4. **可选认证**: 部分接口支持匿名访问

### � 用户 API Key 管理

登录后的用户可以在 `/user/api-keys` 接口管理个人 API Key，用于从命令行或第三方应用直接上传/下载：

- `GET /user/api-keys`：列出当前用户的全部 API Key（需要 Bearer Token）
- `POST /user/api-keys`：创建新的 API Key，可选字段 `name`、`expires_in_days` 或 `expires_at`
- `DELETE /user/api-keys/{id}`：撤销指定的 API Key

创建成功后，响应会包含一次性返回的明文 API Key。后续请求需在 `Authorization: ApiKey <key>` 或 `X-API-Key` 头中携带，系统会自动识别并注入用户身份，可用于 `/share/*` 和 `/chunk/*` 等上传/下载接口。

### �📊 响应格式

所有 API 响应都遵循统一格式：

```json
{
    "code": 200,
    "message": "success",
    "detail": {
        // 具体数据
    }
}
```

### 🔗 相关文件

- `main.go`: Swagger 配置和路由设置
- `docs/`: 自动生成的文档文件
  - `docs.go`: Go 文档包
  - `swagger.json`: OpenAPI JSON 规范
  - `swagger.yaml`: OpenAPI YAML 规范
- `internal/handlers/`: 各种处理器及其 Swagger 注释
- `internal/routes/`: 路由配置

### 🛠️ 技术栈

- **Swagger/OpenAPI**: API 文档标准
- **gin-swagger**: Gin 框架的 Swagger 中间件
- **swaggo/swag**: Go 语言的 Swagger 文档生成工具
- **swaggo/files**: Swagger UI 静态文件服务

### 🎯 使用示例

#### 分享文本示例
```bash
curl -X POST "http://localhost:8080/share/text/" \
  -H "Content-Type: multipart/form-data" \
  -F "text=Hello World" \
  -F "expire_value=1" \
  -F "expire_style=day"
```

#### 分片上传示例
```bash
# 1. 初始化上传
curl -X POST "http://localhost:8080/chunk/upload/init/" \
  -H "Content-Type: application/json" \
  -d '{
    "file_name": "large_file.zip",
    "file_size": 1048576,
    "chunk_size": 1024,
    "file_hash": "abc123"
  }'

# 2. 上传分片
curl -X POST "http://localhost:8080/chunk/upload/chunk/{upload_id}/0" \
  -F "chunk=@chunk_0.bin"

# 3. 完成上传
curl -X POST "http://localhost:8080/chunk/upload/complete/{upload_id}" \
  -H "Content-Type: application/json" \
  -d '{
    "expire_value": 7,
    "expire_style": "day"
  }'
```

---

> 📝 **注意**: 文档会随着代码的更新自动同步，确保始终是最新的 API 接口信息。
