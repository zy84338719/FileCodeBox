# FileCodeBox 全量重构方案 — 参考 vastsa/FileCodeBox 能力补齐 + IDL Proto→Thrift 迁移

> 立项日期: 2026-07-27
> 参考项目: https://github.com/vastsa/FileCodeBox (8.4k stars, v2.5.4)
> 工作目录: `/Users/zhangyi/my_project/FileCodeBox`
> 目标: 当前项目承载 vastsa/FileCodeBox 全部能力 + 切换到 CloudWeGo hz Thrift IDL

---

## 1. 现状盘点

### 1.1 当前项目（改造前）

| 项 | 现状 |
|---|---|
| 后端 | Go 1.25 + CloudWeGo Hertz + GORM |
| IDL | **Proto3**（9 个文件） + 第三方 API 注解 |
| 前端 | Vue 3 + TypeScript + Vite + Element Plus |
| 存储 | local / S3 / WebDAV / NFS（手写 driver） |
| 认证 | JWT + API Key |
| 端口 | 12345 |
| DB | SQLite / MySQL |

**已有 IDL（proto）**：
- `common.proto` — health/index/ping
- `user.proto` — 9 个 RPC
- `admin.proto` — 8 个 RPC
- `share.proto` — 4 个 RPC
- `chunk.proto` — 5 个 RPC
- `qrcode.proto` — 2 个 RPC
- `setup.proto` — 2 个 RPC
- `storage.proto` — 4 个 RPC
- `maintenance.proto` — 5 个 RPC

**API 完整度（API-COMPARISON.md）**: 85%

### 1.2 参考项目 vastsa/FileCodeBox（v2.5.4）

| 项 | 现状 |
|---|---|
| 后端 | Python 3.11 + FastAPI + SQLAlchemy + SQLite/MySQL |
| 前端 | Vue 3 + Vite + Pinia + Naive UI |
| 存储 | local / S3 / OneDrive / WebDAV / **OpenDAL**（40+ 后端）|
| 核心 UX | 匿名口令取件（无登录） |
| 国际化 | zh-CN / en-US |
| 主题 | Light / Dark / Auto |
| 镜像大小 | 100MB（distroless） |

**核心特性**（按 docs 目录）：
1. 即传即取（拖拽/粘贴/批量/分片/断点续传）
2. 灵活失效（按时间/次数/永久）
3. 数据自主（OpenDAL 多后端）
4. 预签名上传（presign-upload API）
5. 限流 + 错误码 + 状态码
6. 国际化 + 暗色模式
7. 系统通知（网站公告）

---

## 2. 改造目标

### 2.1 必须达成（P0）

1. **IDL 全量迁移**: proto → thrift，使用 hz 注解体系
2. **OpenDAL 存储抽象**: 替换现有手写 storage driver，统一一个 interface
3. **匿名取件口令**: 仿 vastsa，无登录用户按 6 位取件码取文件
4. **预签名上传**: 客户端直传对象存储，减轻服务器压力
5. **限流中间件**: 全局 IP 限流 + 接口级限流
6. **统一错误码体系**: 错误响应 `{code, message, trace_id}` 全站统一
7. **状态码语义化**: HTTP 状态码 + 业务状态码两层

### 2.2 增值项（P1）

8. **系统通知**: 管理员发布公告，首页/管理面板 banner
9. **国际化**: 前端 i18n（zh-CN / en-US）
10. **暗色模式**: Element Plus dark 适配
11. **游客模式开关**: 管理员可关闭匿名访问

### 2.3 暂不做（明确不包含）

- ❌ 重写前端
- ❌ Docker 镜像优化（distroless）
- ❌ OneDrive 专属适配（OpenDAL 已支持）
- ❌ BokeBox / Podcast 相关能力（vastsa 同作者其他项目）

---

## 3. 架构设计

### 3.1 IDL 体系（thrift 改造后）

```
backend/idl/
├── api.thrift                   # 注解定义（类比 proto 的 api.proto）
├── common.thrift                # Health/Index/Ping
├── user.thrift                  # 用户模块
├── admin.thrift                 # 管理员模块
├── share.thrift                 # 分享（保留）
├── share_anonymous.thrift       # 🆕 匿名取件
├── chunk.thrift                 # 分片上传
├── presign.thrift               # 🆕 预签名上传
├── qrcode.thrift                # 二维码
├── setup.thrift                 # 初始化
├── storage.thrift               # 存储管理（改造：走 OpenDAL）
├── maintenance.thrift           # 维护
├── ratelimit.thrift             # 🆕 限流管理
├── notify.thrift                # 🆕 系统通知
├── error.thrift                 # 🆕 统一错误响应
└── rpc/                         # 内部 RPC
    └── ...
```

### 3.2 存储层重构

**改造前**（4 套手写 driver）:
```
storage.Storage (interface)
├── local (os.Open)
├── s3 (aws-sdk-go)
├── webdav (golang.org/x/net/webdav)
└── nfs (syscall.Mount)
```

**改造后**（OpenDAL 统一抽象）:
```
storage.Storage (interface)
└── OpenDALBackend
    ├── scheme=fs     → 本地
    ├── scheme=s3      → AWS S3
    ├── scheme=oss     → 阿里云 OSS
    ├── scheme=cos     → 腾讯云 COS
    ├── scheme=obs     → 华为云 OBS
    ├── scheme=azblob  → Azure Blob
    ├── scheme=gcs     → Google Cloud Storage
    ├── scheme=webdav  → WebDAV（含坚果云/Nextcloud）
    ├── scheme=sftp    → SFTP
    ├── scheme=hdfs    → HDFS
    └── ...（OpenDAL 0.46 已支持 40+ 后端）
```

**关键变化**：
- 存储配置从分类型（`type: "s3"` + 嵌套配置）改成统一 `scheme` 字段 + `options` map
- 后端实现从 4 套砍到 1 套（OpenDAL）
- 切换存储 = 改一个 scheme，不用动代码

### 3.3 匿名取件流程

```
[发送端]                              [服务]                    [接收端]
   │                                    │                          │
   │  ① 上传文件 (file/分享)             │                          │
   │ ───────────────────────────────────>│                          │
   │                                    │ 生成 6 位取件码 (PICK-XXXXXX)│
   │ <─ 返回 code + url ────────────────│                          │
   │                                    │                          │
   │  ② 复制 code 给接收方                │                          │
   │ ──────────────────────────────────────────────────────────────────>
   │                                    │                          │
   │                                    │  ③ 接收方输入 code         │
   │                                    │ <────────────────────────│
   │                                    │  POST /api/v1/anonymous/retrieve
   │                                    │  {code, password?}        │
   │                                    │                          │
   │                                    │  ④ 校验 → 返回 file info │
   │                                    │ ─────────────────────────>
   │                                    │                          │
   │                                    │  ⑤ 接收方下载            │
   │                                    │ <────────────────────────│
   │                                    │  GET /api/v1/anonymous/download
```

**核心差异**：
- 取件人**不需要账号**
- 取件码是短码（6 位字母数字），不是分享 URL
- 适配"工作群发文件/老师发资料给家长"这种"跨账号"场景

### 3.4 预签名上传流程

```
[客户端]                          [服务]                       [对象存储]
   │                                │                              │
   │  ① 请求上传 URL                 │                              │
   │  POST /api/v1/presign/upload   │                              │
   │  {file_name, file_size, scheme}│                              │
   │ ──────────────────────────────>│                              │
   │                                │ 生成 PUT presigned URL        │
   │                                │ (OpenDAL.CreatePresign)       │
   │ <── {upload_url, token, exp}───│                              │
   │                                                                 │
   │  ② 直传文件 (无服务器中转)                                      │
   │  PUT upload_url  body=<file>                                   │
   │ ──────────────────────────────────────────────────────────────>│
   │                                                                 │
   │  ③ 通知服务完成                                                  │
   │  POST /api/v1/presign/complete {token, code}                    │
   │ ──────────────────────────────>│                              │
   │                                │ 校验 token → 写 share 表      │
   │ <── {code, url} ──────────────│                              │
```

**关键收益**：
- 大文件不经过服务器，省带宽
- 服务无状态，水平扩展简单
- 上传期间服务挂掉不影响

### 3.5 错误码体系

```go
// 统一响应 envelope
type Response struct {
    Code    int32       `json:"code"`     // 业务码
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    TraceID string      `json:"trace_id,omitempty"`
}

// 业务码段位规划
const (
    // 1xxxx 通用
    CodeOK             = 0      // 成功
    CodeUnknown        = 10000  // 未知错误
    CodeInvalidParam   = 10001  // 参数错误
    CodeUnauthorized   = 10002  // 未登录
    CodeForbidden      = 10003  // 无权限
    CodeNotFound       = 10004  // 资源不存在
    CodeRateLimit      = 10005  // 限流
    
    // 2xxxx 业务
    CodeShareNotFound  = 20001  // 分享不存在
    CodeShareExpired   = 20002  // 分享过期
    CodeSharePasswordWrong = 20003  // 密码错
    CodeShareReachLimit = 20004  // 超过取件次数
    
    // 3xxxx 存储
    CodeStorageInit    = 30001  // 存储初始化失败
    CodeStorageQuota   = 30002  // 存储配额超限
    CodeStorageUpload  = 30003  // 上传失败
    CodeStorageDownload = 30004 // 下载失败
    
    // 4xxxx 用户
    CodeUserNotFound   = 40001
    CodeUserExists     = 40002
    CodePasswordWrong  = 40003
    CodeUserDisabled   = 40004
)
```

### 3.6 限流设计

```go
// 双层限流
type Limiter struct {
    global *rate.Limiter        // 全局 IP 维度
    perAPI map[string]*rate.Limiter  // 接口维度
}

// 配置（系统表）
type RateLimitConfig struct {
    GlobalQPS    int  // 全局 QPS（per IP）
    UploadQPS    int  // 上传接口 QPS（per IP）
    DownloadQPS  int  // 下载接口 QPS（per IP）
    LoginQPS     int  // 登录接口 QPS（per IP）
    Burst        int  // 突发容量
    Enabled      bool
}
```

---

## 4. IDL 改造映射（proto → thrift）

### 4.1 注解体系

Proto 用了第三方 `api.proto` 扩展（query/path/body/form/...），Thrift 没有官方扩展，但 hz 提供以下机制：

```thrift
// 1. hz 注解通过文件级 option 设置
include "api.thrift"

// 2. 通过 Thrift IDL 注释 + hz 工具识别
namespace go filecodebox.user

// 3. 路由通过 service method 命名 + 额外 .yaml 配置文件（hz 支持）
```

**最佳实践（参考 fuxi/fuxi_xxx）**:
- IDL 只定义数据结构 + service 接口
- 路由注解（path/query/body）通过 IDL 注释 + hz 配置工具识别
- 我们项目的现状: `biz/router/<sub>/<sub>.go` 是**手维护**的（不是 hz 生成的）
- → 跟现有项目惯例一致

### 4.2 IDL 字段映射

| Proto | Thrift | 备注 |
|---|---|---|
| `int32` | `i32` | |
| `int64` | `i64` | |
| `uint32` | `i32` (thrift 无 u) | 负值校验放到 handler |
| `uint64` | `i64` | |
| `string` | `string` | |
| `bool` | `bool` | |
| `bytes` | `binary` | |
| `repeated X` | `list<X>` | |
| `map<K,V>` | `map<K,V>` | |
| `optional X` | `optional X` | thrift 0.13+ 支持 |

**保留 IDL 命名风格**（跟 fuxi 项目 1:1）：
- `xxxReq` / `xxxResp` 命名
- `option go_package = "filecodebox.user"` 直接对到包
- 不要在 IDL 里写 HTTP 注解（手维护 router）

---

## 5. 实施阶段

### 阶段 1：基础设施（30min）

1. `backend/idl/api.thrift` 创建（注解注释模板）
2. `backend/Makefile.thrift-gen` 添加
3. 顶层 `Makefile` 加 `thrift-gen` target
4. 删除 `go.mod` 里 proto 依赖（`google.golang.org/protobuf`）
5. 添加 thrift 依赖（`github.com/apache/thrift` + hz）
6. 安装 hz + thrift 二进制（脚本 `scripts/install-tools.sh`）

### 阶段 2：IDL 迁移（90min）

按模块顺序迁移（依赖少的先做）：

| 顺序 | 文件 | 内容 |
|---|---|---|
| 1 | common.thrift | Health/Index/Ping |
| 2 | setup.thrift | 初始化 |
| 3 | qrcode.thrift | 二维码 |
| 4 | chunk.thrift | 分片上传 |
| 5 | share.thrift | 分享（保留） |
| 6 | user.thrift | 用户 |
| 7 | admin.thrift | 管理员 |
| 8 | storage.thrift | 存储管理（OpenDAL 配置） |
| 9 | maintenance.thrift | 维护 |

每个文件迁移完跑 `hz update -idl idl/main.thrift` 验证生成。

### 阶段 3：新增能力 IDL（60min）

| 文件 | 内容 |
|---|---|
| share_anonymous.thrift | 匿名取件（按 6 位码） |
| presign.thrift | 预签名上传（init/upload/complete） |
| ratelimit.thrift | 限流配置（get/update/test） |
| notify.thrift | 系统通知（list/get/create/delete） |
| error.thrift | 统一错误响应 schema |

### 阶段 4：业务实现（120min）

- OpenDAL storage 后端（替换现有 4 个 driver）
- 匿名取件 service（短码生成 + 校验 + 限次）
- 预签名上传 service（OpenDAL presign API）
- 限流 middleware（双层：IP + 接口）
- 错误码 envelope（替换现有 resp 包）
- 通知 module

### 阶段 5：hz 代码生成 + 集成（60min）

- 跑 `hz update` 全部生成
- 手维护 `biz/router/<sub>/<sub>.go`（每个 service 一个文件）
- 完善 `biz/handler/<sub>/`（跟 IDL RPC 1:1）
- 替换 `internal/app/` 下的实现

### 阶段 6：验证（30min）

- `go build ./...` — 0 error
- `go vet ./...` — 0 error
- `go test -count=1 -short -timeout 120s ./...` — 0 fail
- hz lint（如果有）
- README + AGENTS.md 更新

---

## 6. 风险评估

| 风险 | 影响 | 缓解 |
|---|---|---|
| hz 对 thrift 的支持不如 proto 成熟 | 中 | 现有项目用 proto 时也是手维护 router，thrift 一样可行 |
| OpenDAL Go binding API 文档不完整 | 中 | 0.46+ 已稳定；参考官方 example |
| 大文件 IDL 改动可能 break 现有 handler | 高 | 分阶段迁移，每阶段跑 hz update 验证；handler 1:1 对应 IDL method |
| 前端调用 IDL 变了要调整 | 高 | IDL field tag 不能变（保留 1:N 顺序）；message 名字保持兼容 |
| 时间紧 | 中 | 后台跑长任务；分阶段提交，每阶段可独立 work |

---

## 7. 兼容性策略

### 7.1 API 兼容

- 所有 IDL field tag 保持不变（protobuf→thrift tag 号一致）
- message 名保持 1:1 映射（`ShareTextReq` → `ShareTextReq`）
- service 名保持 1:1（`ShareService` → `ShareService`）
- HTTP path 完全保持（`/share/text/` 不变）
- 响应 envelope `{code, message, data}` 保持

### 7.2 DB 兼容

- 不改任何表结构
- 不改任何 migration
- 增量字段用 `IF NOT EXISTS`

### 7.3 配置兼容

- `configs/config.yaml` 结构保持
- OpenDAL 配置走单独的 `storage_opendal` 段，平滑替换现有 `storage.local/s3/webdav/nfs`

---

## 8. 验收清单

- [ ] 所有 proto 文件已删除
- [ ] 所有 IDL 是 thrift 格式
- [ ] `hz update -idl idl/main.thrift` 成功生成
- [ ] `go build ./...` 0 error
- [ ] `go vet ./...` 0 error
- [ ] `go test -count=1 -short ./...` 0 fail
- [ ] 9 个原有 IDL service 全部 hz 生成对应 handler stub
- [ ] 5 个新增 IDL service 全部 hz 生成对应 handler stub
- [ ] OpenDAL 集成到 storage 抽象
- [ ] 匿名取件流程端到端可走通
- [ ] 预签名上传流程可生成有效 URL
- [ ] 限流中间件生效（高并发下返回 429）
- [ ] 错误码统一（每个 handler 都用 `pkg/resp`）
- [ ] README + AGENTS.md 同步更新
- [ ] Makefile 包含 `thrift-gen` target
- [ ] `scripts/install-tools.sh` 一键安装 thrift + hz

---

## 9. 不在本次范围

- ❌ 前端 i18n 改造（仅后端支持）
- ❌ 前端暗色模式（仅后端支持）
- ❌ K8s/ArgoCD 部署清单更新
- ❌ 性能压测
- ❌ 跨集群同步（Submariner）
- ❌ WebSocket 实时通知

---

## 10. 时间估算

| 阶段 | 估算 |
|---|---|
| 1. 基础设施 | 30 min |
| 2. IDL 迁移（9 个） | 90 min |
| 3. 新增 IDL（5 个） | 60 min |
| 4. 业务实现 | 120 min |
| 5. hz 集成 | 60 min |
| 6. 验证 | 30 min |
| **合计** | **~6.5h**（后台 task 跑） |
