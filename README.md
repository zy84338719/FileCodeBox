# FileCodeBox Go版本

🚀 FileCodeBox的高性能Golang实现 - 开箱即用的文件快传系统

## ✨ 特性

- 🔥 **高性能** - Golang实现，并发处理能力强
- 📁 **文件分享** - 支持各种格式文件的快速分享
- 📝 **文本分享** - 支持文本内容的快速分享
- 🔧 **分片上传** - 支持大文件分片上传，断点续传
- ⚡ **断点续传** - 网络中断后可恢复上传，支持秒传
- ⏰ **灵活过期** - 支持时间和次数两种过期方式
- 🛡️ **安全可靠** - JWT认证、限流保护、权限控制
- 📊 **管理后台** - 完整的文件管理和系统配置
- 🗄️ **多存储** - 支持本地、WebDAV、S3存储
- 🐳 **容器化** - 开箱即用的Docker部署
- 🎨 **主题系统** - 支持多主题切换

## 快速开始

### 直接运行

```bash
# 安装依赖
go mod tidy

# 编译并运行
go build -o filecodebox
./filecodebox
```

### Docker运行

```bash
# 构建镜像
docker build -t filecodebox-go .

# 运行容器
docker run -d -p 12345:12345 -v ./data:/root/data filecodebox-go
```

### Docker Compose

```bash
docker-compose up -d
```

## 配置

配置文件会自动生成在 `data/config.json`，主要配置项：

- `port`: 服务端口（默认12345）
- `name`: 站点名称
- `upload_size`: 最大上传大小
- `file_storage`: 存储类型（local/s3/webdav/onedrive）
- `admin_token`: 管理员访问令牌

## API接口

### 分享文本
```
POST /share/text/
```

### 分享文件
```
POST /share/file/
```

### 获取分享内容
```
GET /share/select/?code=xxx
POST /share/select/
```

### 分片上传和断点续传
```
POST /chunk/upload/init/          # 初始化分片上传
POST /chunk/upload/chunk/:upload_id/:chunk_index # 上传分片
POST /chunk/upload/complete/:upload_id            # 完成上传
GET  /chunk/upload/status/:upload_id              # 获取上传状态
POST /chunk/upload/verify/:upload_id/:chunk_index # 验证分片
DELETE /chunk/upload/cancel/:upload_id            # 取消上传
```

### 管理接口
```
GET /admin/stats         # 统计信息
GET /admin/files         # 文件列表
DELETE /admin/files/:id  # 删除文件
GET /admin/config        # 获取配置
PUT /admin/config        # 更新配置
```

## 项目结构

```
├── main.go                 # 主程序入口
├── internal/
│   ├── config/            # 配置管理
│   ├── database/          # 数据库初始化
│   ├── models/            # 数据模型
│   ├── services/          # 业务逻辑
│   ├── handlers/          # HTTP处理器
│   ├── middleware/        # 中间件
│   ├── storage/           # 存储接口
│   ├── routes/            # 路由设置
│   └── tasks/             # 后台任务
├── data/                  # 数据目录
├── themes/                # 主题文件
└── docker-compose.yml     # Docker编排
```

## 与Python版本的差异

1. **性能提升**: Go版本具有更好的并发性能和更低的内存占用
2. **依赖更少**: 不需要Python运行时环境
3. **部署简单**: 编译后为单一可执行文件
4. **类型安全**: 静态类型系统减少运行时错误

## 开发

### 添加新的存储后端

1. 实现 `storage.StorageInterface` 接口
2. 在 `storage.NewStorageManager` 中注册新存储
3. 在配置中添加相关配置项

### 添加新功能

1. 在 `models/` 中定义数据模型
2. 在 `services/` 中实现业务逻辑
3. 在 `handlers/` 中添加HTTP处理器
4. 在 `routes/` 中注册路由

## 许可证

与原Python版本保持一致的开源许可证。
