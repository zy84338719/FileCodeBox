# 第一批：P0 安全加固设计

> 日期：2026-07-29
> 状态：已确认，待实现
> 范围：修复 6 项 P0 安全漏洞（XSS/限流/JWT/metrics/CORS/上传校验）
> 这是"项目不足修复"三批计划的第一批，独立可交付、可回退

## 已确认决策

1. **JWT**：全环境 fail-fast（含开发），未设 FCB_JWT_SECRET 启动即报错
2. **CORS**：未配时默认允许 localhost 跨域，AllowCredentials 仅对白名单生效
3. **metrics**：默认关闭，开启时内网绑定（127.0.0.1）
4. **限流**：挂载到登录/anonymous/presign+chunk/upload/download 全部接口

---

## 一、前端存储型 XSS 修复（token 窃取链）

**问题**：`frontend/src/components/FilePreview.vue:105` 用 `v-html="renderedMarkdown"`，而 `renderedMarkdown` 来自 `marked.parse()` 无消毒。恶意分享文本可执行任意 JS，配合 localStorage 存 token → 窃取登录态。

**修复**：
- 引入 `dompurify` 依赖
- `FilePreview.vue` 中 `marked.parse()` 输出后经 `DOMPurify.sanitize()` 消毒
- `highlightedCode` 路径（`:122`）一并消毒（hljs 默认转义但防御性消毒）
- 新增 `utils/sanitize.ts` 统一封装 `sanitizeHtml(html: string): string`

**验收**：分享含 `<script>alert(1)</script>` 或 `<img onerror=...>` 的文本，渲染后不执行脚本。

---

## 二、限流中间件挂载

**问题**：`internal/pkg/middleware/ratelimit.go` 定义了完整限流，但 `bootstrap.go:454-461` 主链未挂载，gen router 的 middleware.go 全 `return nil`。登录可暴力破解、取件码可枚举、上传下载无防护。

**修复**：
- 限流挂载到自定义路由层（`customMw` 已有模式，不走 gen router）
- 登录接口（`/admin/login`、`/user/login`）→ `LoginMiddleware`（5 QPS）
- anonymous（generate/retrieve/download）→ `UploadMiddleware`/`DownloadMiddleware`
- presign/chunk → `UploadMiddleware`（10 QPS）
- share download、anonymous download → `DownloadMiddleware`（50 QPS）
- 挂载方式：在 bootstrap 注册路由组时，对敏感前缀 `Use` 对应中间件

**挑战与方案**：限流中间件当前签名需确认与 hertz `server.Hertz` 的 `Use` 兼容。若 ratelimit.go 的 `XxxMiddleware()` 返回 `app.HandlerFunc`，可直接在路由组 `group.Use()`。需在实现时核实并适配。

---

## 三、JWT 全环境 fail-fast + 并发安全

**问题**：`auth/jwt.go:13` 硬编码默认密钥 `"FileCodeBox2025SecretKey"`，`bootstrap.go:256` 的 `validateSecrets` 仅 `IsProduction()` 时校验。非生产环境用公开默认密钥可伪造 token。全局 `jwtSecret` 变量并发读写无保护。

**修复**：
- `validateSecrets` 移除 `IsProduction()` 前置条件，所有环境校验
- 未设 `FCB_JWT_SECRET` 或为空或命中黑名单 → `log.Fatal` 退出
- `.env.example` 提供 `FCB_JWT_SECRET=dev-only-change-me-<random>` 占位（明确标注仅开发）
- `auth/jwt.go` 的 `jwtSecret` 改为 `sync.RWMutex` 保护：`SetJWTSecret` 写锁，`GenerateToken`/`ParseToken` 读锁
- 黑名单扩充：加入 `.env.example` 的占位串、`admin123` 派生串等

---

## 四、/metrics 默认关闭 + 内网绑定

**问题**：`bootstrap.go:184` `FCB_METRICS_ENABLED` 默认 true，`:518` 注册 `/metrics` 无认证，docker-compose.prod.yml:39 默认开启。公网泄露请求量/延迟/内部状态。

**修复**：
- `FCB_METRICS_ENABLED` 默认改为 `false`（setDefaults 处）
- metrics 路由保持注册，但开启时绑定独立内网监听：新增 `FCB_METRICS_ADDR`（默认 `127.0.0.1:9090`），metrics 起独立 hertz server 监听该地址，主 server 不暴露 `/metrics`
- docker-compose.prod.yml 去掉 `FCB_METRICS_ENABLED: "true"`，改为注释说明按需开启
- 文档说明：生产用 Prometheus 抓 `127.0.0.1:9090/metrics`（sidecar/同节点）

---

## 五、CORS 收紧（默认 localhost 跨域）

**问题**：`bootstrap.go:74-92` 未配 allow_origins 时反射任意 Origin + AllowCredentials=true。生产默认即"任意源+携带凭证"，CSRF/凭证泄露风险。

**修复**：
- 未配 `allow_origins` 时，默认白名单：`http://localhost:*`、`http://127.0.0.1:*`（开发友好）
- `AllowCredentials=true` 仅当请求 Origin 命中白名单时设置；非白名单 Origin 不反射、不带 `Access-Control-Allow-Credentials`
- 显式配了 `allow_origins` 时，严格按白名单（不反射）

---

## 六、上传应用层校验

**问题**：`chunk/service.go:16` 等直接信任客户端 `FileSize` 入库不校验；`bootstrap.go:427` 未设 `WithMaxRequestBodySize`；无文件类型黑白名单。

**修复**：
- bootstrap 设 `server.WithMaxRequestBodySize(config.Upload.UploadSize)`（Hertz 选项）
- chunk/presign/anonymous 上传入口校验 `FileSize <= config.Upload.UploadSize`，超限返回 413 + 错误码
- 加文件扩展名黑名单：`.exe .bat .cmd .sh .com .scr .msi` 等可执行文件拒绝（黑名单默认值可配）
- 新增 `internal/pkg/utils/filecheck.go`：`IsAllowedExtension(filename string, blacklist []string) bool`

---

## 文件改动清单

| 文件 | 改动 |
|---|---|
| `frontend/package.json` | 加 `dompurify` 依赖 |
| `frontend/src/utils/sanitize.ts` | 新建：DOMPurify 封装 |
| `frontend/src/components/FilePreview.vue` | marked 输出消毒 |
| `backend/internal/pkg/middleware/ratelimit.go` | 确认/适配路由组挂载签名 |
| `backend/cmd/server/bootstrap/bootstrap.go` | 限流挂载 + JWT fail-fast + metrics 内网 + CORS 收紧 + body 上限 |
| `backend/internal/pkg/auth/jwt.go` | jwtSecret 加锁 |
| `backend/.env.example`（若有） | JWT 开发占位 |
| `backend/internal/pkg/utils/filecheck.go` | 新建：扩展名校验 |
| `backend/internal/app/chunk/service.go` | 上传大小+类型校验 |
| `backend/gen/http/handler/share_anonymous/share_anonymous_service.go` | 上传校验 |
| `backend/internal/app/presign/presign.go` | 上传校验 |
| `docker-compose.prod.yml` | metrics 默认关闭 |

## 测试

- 前端：XSS 消毒单测（marked+DOMPurify 输出不含 script 标签）
- 后端：filecheck 单测（黑名单拒绝/白名单通过）
- JWT：fail-fast 触发测试（空密钥/黑名单密钥 fatal）
- 限流：集成测试（连续请求超限返回 429）—— 若现有测试框架难测中间件，至少手动验证

## 风险

- **JWT fail-fast 全环境**：本地开发必须设 env。缓解：.env.example 提供开发占位
- **metrics 独立监听**：改动 bootstrap 启动流程。缓解：独立 server 失败不阻断主 server
- **限流挂载**：可能误伤现有测试/集成。缓解：限流阈值已宽松（登录5/上传10/下载50 QPS）
