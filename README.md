# FileCodeBox v2

> 一个专为自托管场景打造的高性能文件/文本分享平台，采用前后端分离架构。

---

## 📌 项目概览

FileCodeBox 是一个使用 Go + Vue 3 实现的轻量级分享服务，采用前后端分离架构：

- **Backend**: Go + CloudWeGo Hertz + GORM
- **Frontend**: Vue 3 + TypeScript + Vite + Element Plus

### 你可以用它做什么？

- 📁 **拖拽上传**，生成短链接后分享文件或文本片段
- 🔐 **后台管理**，集中查看、搜索、统计、审核、删除分享内容
- 🪶 **多存储后端**，根据需求切换本地/对象存储方案
- ⚙️ **自定义配置**，调整限速、过期策略等

---

## 🌟 关键特性

| 分类 | 能力速览 |
| --- | --- |
| 性能 | Go 原生并发、分片上传、断点续传、秒传校验 |
| 分享体验 | 文本/文件双通道、链接有效期控制、密码和访问次数限制 |
| 管理后台 | 仪表板、文件列表、用户管理、存储面板、系统配置 |
| 安全 | JWT 认证、API Key 支持、限流中间件 |
| 存储 | 本地磁盘、S3 兼容对象存储（可扩展） |
| 部署 | Docker / Docker Compose、单二进制部署 |
| 前端 | Vue 3 + TypeScript、自适应布局、现代化 UI |

---

## 🧩 项目结构

```
FileCodeBox/
├── backend/           # Go 后端 (Hertz + GORM)
│   ├── cmd/server/    # 入口
│   ├── internal/      # 内部包
│   │   ├── app/       # 业务逻辑
│   │   ├── repo/      # 数据访问
│   │   ├── conf/      # 配置
│   │   └── pkg/       # 工具库
│   ├── biz/           # Hertz 生成代码
│   │   ├── handler/   # HTTP handlers
│   │   ├── model/     # Proto 生成的模型
│   │   └── router/    # 路由注册
│   ├── idl/           # Proto API 定义
│   └── configs/       # 配置文件
│
├── frontend/          # Vue 3 前端
│   ├── src/
│   │   ├── views/     # 页面组件
│   │   ├── api/       # API 调用
│   │   ├── stores/    # 状态管理
│   │   └── router/    # 路由
│   └── dist/          # 构建产物
│
├── Makefile           # 构建脚本
├── docker-compose.yml # Docker Compose 配置
└── README.md
```

---

## 🚀 快速起步

### 1. 环境要求

- **Go** 1.21+
- **Node.js** 18+
- **SQLite / MySQL / PostgreSQL**（默认 SQLite）
- Docker 20+（可选）

### 2. 本地开发

```bash
# 安装依赖
make deps

# 开发模式（前后端同时启动）
make dev

# 或分别启动
make dev-backend   # 后端 :12346
make dev-frontend  # 前端 :5173
```

### 3. 生产构建

```bash
# 完整构建
make build

# 或分步构建
make build-frontend  # 构建前端
make build-backend   # 构建后端
make copy-frontend   # 复制前端到 backend/static/
```

### 4. Docker 部署

```bash
# 使用 docker-compose
docker-compose up -d

# 或手动构建
docker build -t filecodebox ./backend
docker run -d -p 12346:12346 -v $(pwd)/data:/data filecodebox
```

服务默认监听 `http://127.0.0.1:12346`。

---

## 📚 API 概览

### 公开 API

| 接口 | 说明 |
| --- | --- |
| `POST /share/text/` | 分享文本 |
| `POST /share/file/` | 分享文件 |
| `GET /share/select/?code=...` | 获取分享内容 |
| `GET /share/download` | 下载文件 |
| `POST /user/register` | 用户注册 |
| `POST /user/login` | 用户登录 |
| `GET /health` | 健康检查 |

### 认证 API (需要 JWT Token)

| 接口 | 说明 |
| --- | --- |
| `GET /user/info` | 用户信息 |
| `GET /user/files` | 用户文件列表 |
| `GET /user/api-keys` | API Key 列表 |
| `POST /user/api-keys` | 创建 API Key |

### 管理 API (需要 Admin Token)

| 接口 | 说明 |
| --- | --- |
| `POST /admin/login` | 管理员登录 |
| `GET /admin/stats` | 系统统计 |
| `GET /admin/files` | 文件列表 |
| `GET /admin/users` | 用户列表 |
| `GET /admin/storage` | 存储信息 |

### 分片上传

| 接口 | 说明 |
| --- | --- |
| `POST /chunk/upload/init/` | 初始化上传 |
| `POST /chunk/upload/chunk/:id/:idx` | 上传分片 |
| `POST /chunk/upload/complete/:id` | 完成上传 |
| `GET /chunk/upload/status/:id` | 上传状态 |
| `DELETE /chunk/upload/cancel/:id` | 取消上传 |

---

## 🔐 认证方式

1. **JWT Token**: 用户/管理员登录后获取，放在 `Authorization: Bearer <token>` 头中
2. **API Key**: 格式 `fcb_sk_xxx`，放在 `X-API-Key` 头或 `api_key` 查询参数中

---

## 🛠️ 配置说明

配置文件位于 `backend/configs/config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 12346

database:
  driver: "sqlite"
  db_name: "./data/filecodebox.db"

storage:
  type: "local"
  path: "./data/uploads"

user:
  allow_user_registration: true
```

---

## 🧑‍💻 开发指南

### 添加新 API

1. 在 `backend/idl/http/` 添加 `.proto` 文件
2. 运行 `make gen-http-update IDL=http/your_api.proto`
3. 实现 `backend/biz/handler/` 中的 handler
4. 在 `backend/internal/app/` 中添加业务逻辑

### 修改前端

1. 修改 `frontend/src/` 下的代码
2. 运行 `make build-frontend` 构建
3. 运行 `make copy-frontend` 复制到后端

---

## 📄 许可证

MIT License © FileCodeBox Contributors
