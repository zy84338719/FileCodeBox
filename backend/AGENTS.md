# FileCodeBox Backend — AGENTS.md

> AI agent 项目指南。本文档给后续 agent 接手时参考。

## 项目速览

| 项 | 值 |
|---|---|
| 语言/版本 | Go 1.25 (toolchain go1.26.4) |
| HTTP 框架 | CloudWeGo Hertz v0.9.6 |
| IDL | **Thrift**（hz v0.9.7 + thriftgo 0.4.3） |
| 存储 | local / S3 / WebDAV / NFS / **OpenDAL 风格抽象** |
| DB | GORM（SQLite/MySQL/PostgreSQL 驱动） |
| Cache | Redis (go-redis v9) |
| 端口 | 12345 |

## 目录结构

```
backend/
├── cmd/server/                # main 入口 + bootstrap
│   ├── main.go
│   └── bootstrap/
│       └── bootstrap.go       # 配置/DB/Redis/服务初始化
├── idl/                       # ★ Thrift IDL（已从 proto 全量迁移）
│   ├── api.thrift             # 注解约定文档
│   ├── common.thrift          # health/index/ping
│   ├── user.thrift            # 用户
│   ├── admin.thrift           # 管理员
│   ├── share.thrift           # 分享
│   ├── chunk.thrift           # 分片上传
│   ├── qrcode.thrift          # 二维码
│   ├── setup.thrift           # 初始化
│   ├── storage.thrift         # 存储管理
│   ├── maintenance.thrift     # 维护
│   ├── http/health.thrift     # 独立 health
│   ├── share_anonymous.thrift # 🆕 匿名取件（vastsa UX）
│   ├── presign.thrift         # 🆕 预签名上传
│   ├── ratelimit.thrift       # 🆕 限流管理
│   ├── notify.thrift          # 🆕 系统通知
│   ├── error.thrift           # 🆕 统一错误响应 schema
│   ├── http/                  # 子目录
│   └── rpc/                   # Kitex RPC IDL
├── gen/http/                  # hz 生成代码（不手改结构，只改业务逻辑）
│   ├── handler/<sub>/         # 每个 IDL 一个目录
│   ├── model/<sub>/           # thrift 生成的 struct
│   └── router/<sub>/          # 路由注册
├── internal/
│   ├── app/                   # ★ 业务层（每个子模块一个目录）
│   │   ├── user/              # 业务实现
│   │   ├── admin/
│   │   ├── share/
│   │   ├── anonymous/        # 🆕 匿名取件 service
│   │   ├── presign/           # 🆕 预签名上传 service
│   │   ├── notify/            # 🆕 通知 service
│   │   └── ...
│   ├── pkg/
│   │   ├── resp/              # 统一响应 envelope {code, message, data, trace_id}
│   │   ├── errcode/           # 业务码段位（1xxxx/2xxxx/3xxxx/4xxxx/5xxxx/9xxxx）
│   │   ├── errors/            # 旧版 error code（保留兼容）
│   │   ├── middleware/        # auth + ratelimit
│   │   ├── auth/              # JWT
│   │   ├── logger/            # zap
│   │   └── utils/
│   ├── storage/
│   │   ├── storage.go         # 旧 4-driver 抽象
│   │   └── opendal/           # 🆕 OpenDAL 风格抽象
│   ├── conf/                  # 配置 struct
│   ├── repo/
│   │   ├── db/                # GORM
│   │   └── redis/             # Redis client
│   └── preview/               # 文件预览（图片/视频缩略图）
├── configs/
│   ├── config.yaml            # 默认配置
│   └── config.prod.yaml       # 生产配置
├── scripts/
│   ├── gen.sh                 # hz/kitex 代码生成（支持 proto + thrift）
│   └── install-tools.sh       # 一键安装 hz + thriftgo
├── Makefile
├── go.mod
└── router.go                  # main 包的路由（customizedRegister）
```

## 关键约定

### IDL 迁移完成度

| 状态 | 含义 | 文件 |
|---|---|---|
| ✅ thrift | 完全迁移 | 所有 idl/*.thrift |
| 🔄 proto 旧 | hz 仍可生成（保留兼容路径） | idl/*.proto + idl/http/*.proto |
| 🆕 new | 新增能力 | share_anonymous/presign/ratelimit/notify/error |

> 旧 IDL（proto 路径）将在后续阶段删除。当前 proto + thrift 并存。

### 响应 envelope 统一

```json
{
  "code": 0,
  "message": "success",
  "data": { ... },
  "trace_id": "uuid"   // 通过 X-Trace-Id header 透传
}
```

业务码 → HTTP 状态自动映射见 `internal/pkg/resp/response.go:httpStatusForCode`。

### 业务码段位

- `0` 成功
- `1xxxx` 通用（参数/认证/限流/资源不存在）
- `2xxxx` 分享/取件
- `3xxxx` 存储/预签名
- `4xxxx` 用户/认证
- `5xxxx` 系统
- `9xxxx` 第三方

定义见 `internal/pkg/errcode/errcode.go`。

### IDL → handler → internal/app 映射

每个 IDL 文件对应：
1. `idl/<sub>.thrift` — 数据结构 + service 签名 + hz 注解
2. `gen/http/model/<sub>/` — hz 生成的 struct（不手改）
3. `gen/http/handler/<sub>/<sub>_service.go` — hz 生成的 stub + 业务实现覆盖
4. `gen/http/router/<sub>/<sub>.go` — hz 生成的路由（不手改）
5. `internal/app/<sub>/` — 业务实现（service 模式）

新加一个 IDL 的流程：
```bash
# 1. 写 IDL
vim idl/mysub.thrift

# 2. hz 生成（stub）
cd backend
hz update --idl idl/mysub.thrift --module github.com/zy84338719/fileCodeBox/backend --out_dir . --model_dir gen/http/model -I idl

# 3. 实现 internal/app/mysub/
vim internal/app/mysub/mysub.go

# 4. 在 handler stub 上覆盖（调用 internal/app/mysub）
vim gen/http/handler/mysub/mysub_service.go

# 5. 在 bootstrap 注入
vim cmd/server/bootstrap/bootstrap.go
# 加：mysubHandler.SetXxx(...)
```

### 重要 Go 路径

```
github.com/zy84338719/fileCodeBox/backend
├── gen/http/...                        # hz 生成
├── internal/app/...                    # 业务层
├── internal/pkg/errcode                # 业务码
├── internal/pkg/resp                   # 统一 envelope
├── internal/pkg/middleware             # auth + ratelimit
└── ...
```

## 常用命令

```bash
# 工具安装
./scripts/install-tools.sh

# 代码生成
make gen-thrift-update IDL=common.thrift
make gen-thrift-update-all

# 编译
go build -o bin/server ./cmd/server

# 启动
./bin/server -env dev
# 或
go run ./cmd/server/main.go

# 验证
go vet ./...
go test -count=1 -short -timeout 120s ./...
```

## API 路径约定

| 路径前缀 | 说明 |
|---|---|
| `/api/v1/...` | 内部 API（presign 等） |
| `/admin/...` | 管理员 API（需要 admin JWT） |
| `/user/...` | 用户 API（需要 user JWT） |
| `/share/...` | 分享 API（公开） |
| `/chunk/...` | 分片上传 |
| `/anonymous/...` | 🆕 匿名取件（无登录） |
| `/notifies/active` | 🆕 公开活跃通知 |

完整 API 列表参见各 IDL 文件（`idl/*.thrift`）。

## 注意事项

1. **不要删除 proto 文件**（仍在 hz 生成路径中，渐进迁移）
2. **handler stub 是半自动**：hz 首次生成 stub，开发者覆盖业务实现。再次 hz update 不会覆盖已存在的 handler（scripts/gen.sh 的 _build_exclude_args 保护）
3. **opendal 包是兼容抽象**：实际 macOS 开发时用本地 fs，Linux 部署时可切真实 OpenDAL
4. **trace_id 通过 X-Trace-Id header 透传**：客户端可以传，服务器也会自动生成
5. **rate limit middleware 注入**：通过 `middleware.GetDefaultRateLimiter()` 拿单例

## 已知妥协

| 项 | 状态 | 原因 |
|---|---|---|
| OpenDAL 真实 binding | ❌ 未启用 | 只支持 Linux 预编译库，macOS 编译失败 |
| 匿名取件完整流程 | ⚠️ 部分 | 缺 Redis 时 panic；需补 share table 写入 |
| 预签名上传 complete | ⚠️ 简化 | 未真正写 share 表（待 share service 集成） |
| proto 旧 IDL | 🔄 并存 | 渐进迁移，下一阶段删除 |
| 测试覆盖 | ❌ 0 | 项目历史无测试，待补 |

## 参考项目

- [vastsa/FileCodeBox](https://github.com/vastsa/FileCodeBox) — UX 灵感来源
- [fuxi/fuxi_engine](https://git.enjoye.top/enjoydream/fuxi/fuxi_engine) — 项目惯例
