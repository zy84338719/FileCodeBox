# FileCodeBox · Go Edition

> 一个专为自托管场景打造的高性能文件 / 文本分享平台，提供安全、可扩展、可插拔的“随手分享”体验。

---

## 📌 项目概览

FileCodeBox 是一个使用 Go 实现的轻量级分享服务，核心目标是“部署简单、使用顺手、运营安心”。无论你想在团队内搭建一个文件投递站，还是希望为个人项目提供临时分享通道，都可以通过它快速上线并稳定运行。

### 你可以用它做什么？

- 📁 **拖拽上传**，生成短链接后分享文件或文本片段
- 🔐 **后台管理**，集中查看、搜索、统计、审核、删除分享内容
- 🪶 **多存储后端**，根据需求切换本地/对象存储/WebDAV/OneDrive 等方案
- ⚙️ **自定义配置**，调整限速、过期策略、主题皮肤，让系统和业务完美贴合

---

## 🌟 关键特性

| 分类 | 能力速览 |
| --- | --- |
| 性能 | Go 原生并发、分片上传、断点续传、秒传校验 |
| 分享体验 | 文本 / 文件双通道、链接有效期控制、密码和访问次数限制 |
| 管理后台 | 仪表板、文件列表、用户管理、存储面板、系统配置、维护工具 |
| 安全 | 初始化向导、动态路由热载、JWT 管理登录、限流中间件、可选用户体系 |
| 存储 | 本地磁盘、S3 兼容对象存储、WebDAV、OneDrive（均可扩展） |
| 部署 | Docker / Docker Compose、systemd、Nginx 反代、跨平台二进制 |
| 前端 | 主题系统（`themes/2025`）、自适应布局、可自定义静态资源 |

---

## 🧩 架构速写

```
┌───────────────────────────────────────────────────────────┐
│                           Client                           │
│  Web UI (themes) · 管理后台 · RESTful API · CLI · WebDAV   │
└───────────────▲───────────────────────────────▲───────────┘
                │                               │
        Admin Console                     Public Sharing
                │                               │
┌──────────────┴───────────────────────────────┴─────────────┐
│                       FileCodeBox Server                    │
│                                                             │
│  Routes  ─ internal/routes           Middleware             │
│  Handlers ─ internal/handlers        RateLimit · Auth       │
│  Services ─ internal/services        Chunk · Share · User   │
│  Repos    ─ internal/repository      GORM Data Access       │
│  Config   ─ internal/config          动态配置管理           │
└──────────────┬───────────────────────────────┬─────────────┘
               │                               │
       Storage Manager                 Database (SQLite/MySQL/PG)
          └─ local / S3 / WebDAV / OneDrive · 可插拔
```

初始化流程已经内置了“未初始化只允许访问 `/setup`”的安全护栏：

1. 首次运行 → 自动跳转至 Setup 向导创建数据库 + 管理员
2. 一旦初始化成功 → 所有用户请求 `/setup` 会被重定向至首页，同时拒绝重复初始化

---

## 🚀 快速起步

### 1. 环境要求

- **Go** 1.21 及以上（开发环境推荐 1.24+）
- **SQLite / MySQL / PostgreSQL**（三选一，默认 SQLite）
- 可选：Docker 20+、docker-compose v2、Make、Git

### 2. 本地开发

```bash
# 拉取依赖
go mod tidy

# 运行服务
go run ./...
# 或编译后运行
make build && ./bin/filecodebox
```

服务默认监听 `http://127.0.0.1:12345`。首次访问会被引导到 `/setup` 完成初始化。

### 3. Docker / Compose

```bash
# Docker
docker build -t filecodebox .
docker run -d \
  --name filecodebox \
  -p 12345:12345 \
  -v $(pwd)/data:/data \
  filecodebox

# docker-compose
cp docker-compose.yml docker-compose.override.yml   # 如需自定义
docker-compose up -d
```

**生产环境建议**：

- 使用 `docker-compose.prod.yml` + `.env` 管理数据库凭证
- 通过 Nginx/Traefik 等反向代理启用 HTTPS 与缓存策略
- 将 `data/`、数据库与对象存储做持久化与定期备份

### 4. CLI 管理

```bash
./filecodebox admin user list
./filecodebox admin stats
```

所有 CLI 子命令定义在 `internal/cli`，适合在自动化运维或脚本中使用。

---

## 🛠️ 配置指南

所有配置均由 `config.yml` + 数据库动态配置组合而成：

| 配置域 | 说明 |
| --- | --- |
| `base` | 服务端口、站点名称、DataPath |
| `storage` | 存储类型、凭证、路径 |
| `transfer` | 上传限额、断点续传、分片策略 |
| `user` | 用户系统开关、配额、注册策略 |
| `mcp` | 消息通道 / WebSocket 服务配置 |
| `ui` | 主题、背景、页面说明文案 |

初始化完成后，配置会同步写入数据库并可在后台在线修改。每次操作都走事务保证一致性。

> **提示**：未初始化时，仅开放 `/setup`、部分静态资源与 `GET /user/system-info`，避免系统在部署初期被误用。

---

## 🧑‍💻 管理后台一览

- **仪表盘**：吞吐趋势、存储占用、系统告警
- **文件管理**：模糊搜索、批量操作、访问日志
- **用户管理**：启用用户系统、分配配额、状态冻结
- **存储配置**：即时切换存储后端，并对接健康检查
- **系统配置**：修改站点基础信息、主题、分享策略
- **维护工具**：清理过期数据、生成导出、查看审计日志

访问入口：`/admin/`，登录采用 JWT + Bearer Token。自 2025 版起，所有 `admin` 静态资源均由后台鉴权动态下发，避免公共缓存泄露。

---

## 📦 存储与上传

| 类型 | 说明 |
| --- | --- |
| `local` | 默认方案，数据持久化在 `data/uploads` |
| `s3` | 兼容 S3 的对象存储（如 MinIO、阿里云 OSS） |
| `webdav` | 适合挂载 NAS / Nextcloud |
| `onedrive` | 利用 Microsoft Graph 的云端存储 |

上传采用“分片 + 秒传 + 断点续传”的三段式策略：

1. `POST /chunk/upload/init/` 初始化会返回 upload_id
2. 并行调用 `POST /chunk/upload/chunk/:id/:idx`
3. 最后 `POST /chunk/upload/complete/:id` 合并并校验

上传状态可通过 `GET /chunk/upload/status/:id` 观察，也可主动 `DELETE /chunk/upload/cancel/:id` 终止。

---

## 📚 API 与 SDK

虽然 FileCodeBox 主要针对 Web 场景，但服务本身围绕 REST API 架构，便于集成：

| 模块 | 典型接口 |
| --- | --- |
| 分享 | `POST /share/text/` · `POST /share/file/` · `GET /share/select/?code=...` |
| 分片 | `POST /chunk/upload/init/` · `POST /chunk/upload/complete/:id` |
| 用户 | `POST /user/login` · `POST /user/register`（启用用户系统时） |
| 管理 | `GET /admin/stats` · `POST /admin/files/delete` 等 |
| 健康检查 | `GET /health` |

API 文档位于 `docs/swagger-enhanced.yaml`，可通过 `go install github.com/swaggo/swag/cmd/swag@latest` 生成最新文档。

---

## 🧑‍🔬 本地开发与贡献

1. Fork & clone 仓库
2. `make dev`（或参考 `Makefile`）
3. 运行测试：`go test ./...`
4. 提交前确保通过 `golangci-lint`/`go fmt`（在 CI 中亦会执行）

项目保持模块化、接口清晰，欢迎贡献以下方向：

- 新的存储适配器 / 用户登录方式
- 管理后台的交互优化与国际化支持
- 自动化部署脚本（Helm / Terraform）
- 更丰富的 API 客户端（Python/Node/Swift）

提交 PR 时请附上：变更说明、测试方式、潜在影响。如涉及迁移，请编写相应文档放在 `docs/`。

---

## 🗺️ 路线图（节选）

- [ ] Webhook / EventHook（上传完成、分享到期、超额告警）
- [ ] 更细颗粒度的访问控制（到期提醒、下载密码、白名单）
- [ ] 多节点部署指南（对象存储 + Redis + MySQL）
- [ ] 管理后台模块化主题系统 & 深色主题
- [ ] CLI 支持导入/导出配置模板

欢迎在 Issues 区讨论新需求或报告缺陷。

---

## 📄 许可证

MIT License  © FileCodeBox Contributors

---

> 💬 有任何问题、部署疑问或定制需求，欢迎通过 Issue / Discussions / 邮件联系。我们乐于协助每一个想把 FileCodeBox 搭建成“团队内部效率神器”的你。
