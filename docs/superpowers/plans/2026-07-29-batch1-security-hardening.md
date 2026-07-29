# 第一批 P0 安全加固 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修复 6 项 P0 安全漏洞——前端 XSS、限流未挂载、JWT 弱密钥、metrics 暴露、CORS 过宽、上传无校验。

**Architecture:** 后端改动集中在 bootstrap（限流挂载/默认值/CORS/metrics）+ auth（JWT 加锁）+ utils（文件校验）；前端仅 FilePreview 消毒。每项独立可测。

**Tech Stack:** Go 1.26 + Hertz + golang.org/x/time/rate；Vue3 + marked + dompurify

## Global Constraints

- 工作根目录：`/Users/zhangyi/my_project/FileCodeBox`，后端代码在 `backend/`，前端在 `frontend/`
- go module path：`github.com/zy84338719/fileCodeBox/backend`
- 限流实例获取：`middleware.GetDefaultRateLimiter()`（已存在，`backend/internal/pkg/middleware/ratelimit.go:249`）
- 限流方法签名：`(*RateLimiter).LoginMiddleware() / UploadMiddleware() / DownloadMiddleware() app.HandlerFunc`（:235,225,230）
- 提交前缀：`fix(security):`
- 每个任务结束前 `cd backend && go build ./...` 通过
- 不改 `gen/` 下 IDL 生成的路由注册文件（限流通过 bootstrap 自定义路由层挂载）

---

### Task 1: JWT 并发安全 + 全环境 fail-fast

**Files:**
- Modify: `backend/internal/pkg/auth/jwt.go`
- Modify: `backend/cmd/server/bootstrap/bootstrap.go`（validateSecrets 段）

**Interfaces:**
- Produces: `jwtSecret` 改为 RWMutex 保护；`validateSecrets` 全环境校验

- [ ] **Step 1: jwt.go 加锁保护 jwtSecret**

替换 `backend/internal/pkg/auth/jwt.go` 的 `var` 块和所有读写 `jwtSecret` 处。完整重写文件：

```go
package auth

import (
	"errors"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")

	jwtSecretMu sync.RWMutex
	jwtSecret   = []byte("FileCodeBox2025SecretKey")
)

// getSecret 读锁获取当前密钥
func getSecret() []byte {
	jwtSecretMu.RLock()
	defer jwtSecretMu.RUnlock()
	return jwtSecret
}

// Claims JWT claims
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT token
func GenerateToken(userID uint, username, role string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * 7)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "FileCodeBox",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getSecret())
}

// ParseToken 解析 JWT token
func ParseToken(tokenString string) (*Claims, error) {
	secret := getSecret()
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrInvalidToken
}

// RefreshToken 刷新 token
func RefreshToken(tokenString string) (string, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}
	return GenerateToken(claims.UserID, claims.Username, claims.Role)
}

// GenerateAdminToken 生成管理员 token
func GenerateAdminToken(userID uint, username string) (string, error) {
	return GenerateToken(userID, username, "admin")
}

// SetJWTSecret 设置 JWT secret（写锁，从配置注入）
func SetJWTSecret(secret string) {
	jwtSecretMu.Lock()
	defer jwtSecretMu.Unlock()
	jwtSecret = []byte(secret)
}
```

- [ ] **Step 2: 构建 + 现有测试回归**

Run: `cd backend && go build ./... && go test ./internal/app/... ./internal/pkg/... 2>&1 | grep -E "^(ok|FAIL)"`
Expected: 全 ok

- [ ] **Step 3: bootstrap validateSecrets 全环境 fail-fast**

先读 `backend/cmd/server/bootstrap/bootstrap.go` 的 `validateSecrets` 函数（grep 定位 `func validateSecrets`）。当前逻辑含 `if bootstrap.IsProduction()` 前置守卫。移除该守卫，改为无条件校验。具体改动：

找到 `validateSecrets` 函数体，将所有 `if config.IsProduction()` 或类似的生产模式判断**删除**，使校验逻辑对所有环境生效。校验内容保持：JWT secret 为空或命中黑名单则 `logger.Fatal`。

黑名单数组扩充，加入：`"FileCodeBox2025SecretKey"`、`"FileCodeBox2025JWT"`、`"filecodebox-dev-signing-key-change-me"`、`"please-change-me"`、`"dev-only-change-me"`。

- [ ] **Step 4: 构建 + 提交**

Run: `cd backend && go build ./...`
Expected: 通过（注意：本地若有默认弱密钥配置，构建不受影响——校验在运行时 bootstrap 阶段）

```bash
git add backend/internal/pkg/auth/jwt.go backend/cmd/server/bootstrap/bootstrap.go
git commit -m "fix(security): JWT 密钥并发安全(加锁) + 全环境 fail-fast 校验"
```

---

### Task 2: 上传文件类型/大小校验工具

**Files:**
- Create: `backend/internal/pkg/utils/filecheck.go`
- Test: `backend/internal/pkg/utils/filecheck_test.go`

**Interfaces:**
- Produces: `IsBlockedExtension(filename string, blacklist []string) bool`、`DefaultBlockedExtensions() []string`、`CheckUploadSize(fileSize, maxSize int64) error`

- [ ] **Step 1: 写失败测试**

```go
// backend/internal/pkg/utils/filecheck_test.go
package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsBlockedExtension_Blocked(t *testing.T) {
	bl := DefaultBlockedExtensions()
	assert.True(t, IsBlockedExtension("malware.exe", bl))
	assert.True(t, IsBlockedExtension("script.SH", bl))   // 大小写不敏感
	assert.True(t, IsBlockedExtension("a.b.bat", bl))
	assert.True(t, IsBlockedExtension("scr.scr", bl))
}

func TestIsBlockedExtension_Allowed(t *testing.T) {
	bl := DefaultBlockedExtensions()
	assert.False(t, IsBlockedExtension("photo.jpg", bl))
	assert.False(t, IsBlockedExtension("doc.pdf", bl))
	assert.False(t, IsBlockedExtension("noext", bl))
	assert.False(t, IsBlockedExtension("archive.tar.gz", bl))
}

func TestCheckUploadSize_OK(t *testing.T) {
	require.NoError(t, CheckUploadSize(1024, 10*1024*1024))
}

func TestCheckUploadSize_TooLarge(t *testing.T) {
	err := CheckUploadSize(11*1024*1024, 10*1024*1024)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrFileTooLarge))
}
```

- [ ] **Step 2: 运行确认失败**

Run: `cd backend && go test ./internal/pkg/utils/ -run TestIsBlocked -v`
Expected: FAIL（undefined）

- [ ] **Step 3: 实现**

```go
// backend/internal/pkg/utils/filecheck.go
package utils

import (
	"errors"
	"path/filepath"
	"strings"
)

// ErrFileTooLarge 文件超过允许大小
var ErrFileTooLarge = errors.New("file size exceeds limit")

// defaultBlockedExtensions 默认拒绝的可执行文件扩展名
var defaultBlockedExtensions = []string{
	".exe", ".bat", ".cmd", ".com", ".scr", ".msi",
	".sh", ".bash", ".ps1", ".vbs", ".js", ".jar",
	".dll", ".so", ".dylib", ".app",
}

// DefaultBlockedExtensions 返回默认黑名单扩展名（调用方可追加）
func DefaultBlockedExtensions() []string {
	cp := make([]string, len(defaultBlockedExtensions))
	copy(cp, defaultBlockedExtensions)
	return cp
}

// IsBlockedExtension 判断文件扩展名是否在黑名单（大小写不敏感）
func IsBlockedExtension(filename string, blacklist []string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return false
	}
	for _, b := range blacklist {
		if strings.ToLower(b) == ext {
			return true
		}
	}
	return false
}

// CheckUploadSize 校验文件大小是否超限
func CheckUploadSize(fileSize, maxSize int64) error {
	if maxSize > 0 && fileSize > maxSize {
		return ErrFileTooLarge
	}
	return nil
}
```

- [ ] **Step 4: 运行确认通过**

Run: `cd backend && go test ./internal/pkg/utils/ -run 'TestIsBlocked|TestCheckUploadSize' -v`
Expected: PASS（4 个全过）

- [ ] **Step 5: 提交**

```bash
git add backend/internal/pkg/utils/filecheck.go backend/internal/pkg/utils/filecheck_test.go
git commit -m "fix(security): 上传文件类型(黑名单)+大小校验工具"
```

---

### Task 3: 上传入口接入校验（chunk/presign/anonymous）

**Files:**
- Modify: `backend/internal/app/chunk/service.go`
- Modify: `backend/gen/http/handler/share_anonymous/share_anonymous_service.go`（GenerateCode handler）

**Interfaces:**
- Consumes: `utils.CheckUploadSize`、`utils.IsBlockedExtension`、`utils.DefaultBlockedExtensions`（Task 2）
- 需要 config 访问 UploadSize：通过 `conf.GetGlobalConfig().Upload.UploadSize`（全局 config 单例，已存在）

- [ ] **Step 1: chunk service 加校验**

读 `backend/internal/app/chunk/service.go`，找到接收 `FileSize` 的入口方法（如 `InitUpload` 或类似，grep `FileSize`）。在方法开头加：

```go
import (
	"github.com/zy84338719/fileCodeBox/backend/internal/conf"
	"github.com/zy84338719/fileCodeBox/backend/internal/pkg/utils"
)

// 在接收 FileSize 的入口处加：
maxSize := conf.GetGlobalConfig().Upload.UploadSize
if err := utils.CheckUploadSize(req.FileSize, maxSize); err != nil {
	return fmt.Errorf("文件过大: 最大允许 %d 字节", maxSize)
}
if utils.IsBlockedExtension(req.FileName, utils.DefaultBlockedExtensions()) {
	return fmt.Errorf("该文件类型禁止上传")
}
```

注意：`req.FileName` 字段名按实际确认（grep chunk service 的 struct）。若无 FileName 字段（纯分片无文件名），则跳过扩展名校验，仅校验大小。

- [ ] **Step 2: anonymous GenerateCode handler 加校验**

读 `backend/gen/http/handler/share_anonymous/share_anonymous_service.go` 的 `GenerateCode` 函数。在 `CreateAnonymousShare` 调用前加：

```go
import (
	"github.com/zy84338719/fileCodeBox/backend/internal/conf"
	fcutils "github.com/zy84338719/fileCodeBox/backend/internal/pkg/utils"
)

// GenerateCode 函数内，BindAndValidate 后：
maxSize := conf.GetGlobalConfig().Upload.UploadSize
if err := fcutils.CheckUploadSize(req.FileSize, maxSize); err != nil {
	resp.NewErrorWithMessage(c, errcode.CodeInvalidParam, "文件过大")
	return
}
if fcutils.IsBlockedExtension(req.FileName, fcutils.DefaultBlockedExtensions()) {
	resp.NewErrorWithMessage(c, errcode.CodeInvalidParam, "该文件类型禁止上传")
	return
}
```

（注意 import 别名 `fcutils` 避免与已有的 `utils` 冲突——该 handler 已 import `utils` 用于 expire 计算）

- [ ] **Step 3: presign 加校验**

读 `backend/internal/app/presign/presign.go`，找到 Init/Complete 接收 size 处，加同样的 `CheckUploadSize` 校验。

- [ ] **Step 4: 构建 + 提交**

Run: `cd backend && go build ./...`
Expected: 通过

```bash
git add backend/internal/app/chunk/service.go backend/gen/http/handler/share_anonymous/share_anonymous_service.go backend/internal/app/presign/presign.go
git commit -m "fix(security): 上传入口接入大小+类型校验(chunk/presign/anonymous)"
```

---

### Task 4: bootstrap 设 body 上限 + 限流挂载

**Files:**
- Modify: `backend/cmd/server/bootstrap/bootstrap.go`

**Interfaces:**
- Consumes: `middleware.GetDefaultRateLimiter()`（:249）、`(*RateLimiter).LoginMiddleware/UploadMiddleware/DownloadMiddleware`、`config.Upload.UploadSize`、`server.WithMaxRequestBodySize`

- [ ] **Step 1: 设 WithMaxRequestBodySize**

读 `bootstrap.go` 的 server 初始化段（grep `server.WithHostPorts`，约 :428）。在 `server.Default` 或 `server.New` 的 options 里加：

```go
uploadSize := int(config.Upload.UploadSize)
if uploadSize <= 0 {
	uploadSize = 10 * 1024 * 1024 // 默认 10MB
}
h := server.Default(
	server.WithHostPorts(fmt.Sprintf("%s:%d", config.Server.Host, port)),
	server.WithMaxRequestBodySize(uploadSize),
)
```

注意：`config.Upload.UploadSize` 类型是 int64，`WithMaxRequestBodySize` 接收 int，需转换。确认当前 `h := server.Default(...)` 的写法（可能是 `server.Default(opts...)`），在 opts 里追加。

- [ ] **Step 2: 限流挂载到路由组**

读 `bootstrap.go` 的自定义路由注册段（grep `apiV1 := r.Group`，约 :528）。在敏感路由组挂载限流中间件：

```go
rl := middleware.GetDefaultRateLimiter()

// 登录接口限流
r.Group("/admin", rl.LoginMiddleware())  // /admin/login 等
r.Group("/api/v1/user/login", rl.LoginMiddleware())

// 上传接口限流
apiV1.Use(rl.UploadMiddleware())  // presign/chunk 在 /api/v1 下

// 匿名取件/上传限流
r.Group("/anonymous", rl.UploadMiddleware())  // generate/retrieve

// 下载限流 —— 对 /share/download 和 /anonymous/download 单独挂
```

注意：限流挂载需精确。由于 gen router 已注册了具体路径，bootstrap 在 `r.Group(prefix)` 上 `Use` 中间件会作用于该前缀下所有路由。**实现时需核实**：gen router 注册的路径前缀（如 `/anonymous/generate`），确保限流 group 前缀匹配且不重复挂载。

若 gen router 与自定义 group 冲突（同前缀重复注册），改为：在全局中间件链（:454-461）按路径前缀判断挂载，或用 hertz 的 `server.Hertz.Use` 配合路径判断中间件。**优先方案**：写一个轻量 `pathBasedRateLimit` 中间件，根据 `c.Request.URI().Path()` 前缀选择 scope：

```go
// 在全局链加一个路径感知限流
h.Use(func(ctx context.Context, c *app.RequestContext) {
	path := string(c.Request.URI().Path())
	switch {
	case strings.HasPrefix(path, "/admin/login") || strings.HasPrefix(path, "/api/v1/user/login"):
		rl.LoginMiddleware()(ctx, c)
	case strings.HasPrefix(path, "/anonymous/generate") || strings.HasPrefix(path, "/presign") || strings.HasPrefix(path, "/api/v1/chunk"):
		rl.UploadMiddleware()(ctx, c)
	case strings.Contains(path, "/download"):
		rl.DownloadMiddleware()(ctx, c)
	default:
		c.Next(ctx)
	}
})
```

此方案最稳妥，不受 gen router 注册顺序影响。

- [ ] **Step 3: 构建 + 手动验证限流生效**

Run: `cd backend && go build ./...`
Expected: 通过

启动服务后用 curl 快速发 10 次 `/admin/login`，预期第 6 次起返回 429。

- [ ] **Step 4: 提交**

```bash
git add backend/cmd/server/bootstrap/bootstrap.go
git commit -m "fix(security): 设 body 上限 + 限流中间件挂载(登录/上传/下载)"
```

---

### Task 5: CORS 收紧（默认 localhost 跨域）

**Files:**
- Modify: `backend/cmd/server/bootstrap/bootstrap.go`（CORS 函数，约 :60-115）

- [ ] **Step 1: 改 CORS 默认行为**

读 `bootstrap.go` 的 `CORS()` 函数（:60 起）。当前逻辑：无白名单时 `allowCredentials=true` + 反射任意 Origin。改为：

```go
allowCredentials := config.Security.CORS.AllowCredentials
if len(allowOrigins) == 0 {
	// 无白名单：默认 localhost 跨域（开发友好），凭证仅对 localhost 生效
	localhostOrigins := []string{
		"http://localhost", "http://127.0.0.1",
	}
	for _, lo := range localhostOrigins {
		allowOrigins[lo] = true
	}
	// 同时匹配带端口的 localhost（运行时动态判断）
}
```

然后在 origin 匹配段（:84-92），改为：白名单精确匹配 **或** localhost 前缀匹配：

```go
allowedOrigin := ""
if origin != "" {
	if allowOrigins[origin] {
		allowedOrigin = origin
	} else if isLocalhostOrigin(origin) {
		allowedOrigin = origin
	}
}
```

新增辅助函数（CORS 函数外或内）：

```go
// isLocalhostOrigin 判断是否 localhost/127.0.0.1 的任意端口
func isLocalhostOrigin(origin string) bool {
	return strings.HasPrefix(origin, "http://localhost:") ||
		strings.HasPrefix(origin, "http://127.0.0.1:") ||
		origin == "http://localhost" || origin == "http://127.0.0.1"
}
```

删除原"宽松模式：反射任意 Origin"分支（:89-92）。

- [ ] **Step 2: 构建确认**

Run: `cd backend && go build ./...`
Expected: 通过

- [ ] **Step 3: 提交**

```bash
git add backend/cmd/server/bootstrap/bootstrap.go
git commit -m "fix(security): CORS 收紧-默认仅 localhost 跨域,禁止反射任意 Origin"
```

---

### Task 6: metrics 默认关闭 + 内网绑定

**Files:**
- Modify: `backend/cmd/server/bootstrap/bootstrap.go`
- Modify: `docker-compose.prod.yml`

- [ ] **Step 1: setDefaults 改 metrics 默认关闭**

读 `bootstrap.go` 的 `setDefaults` 函数（:175）。找 metrics enabled 的默认设置，改为 false：

```go
// 在 setDefaults 内
v.SetDefault("observability.metrics.enabled", false)
```

若当前是 `true` 或通过 env 默认开启，改为 `false`。

- [ ] **Step 2: metrics 开启时内网绑定**

读 metrics 注册段（:514-518）。当前是主 server 上 `r.GET(metricsPath, ...)`。改为：metrics 开启时，起独立 hertz server 监听 127.0.0.1：

```go
if config.Observability.Metrics.Enabled {
	metricsAddr := os.Getenv("FCB_METRICS_ADDR")
	if metricsAddr == "" {
		metricsAddr = "127.0.0.1:9090"
	}
	metricsServer := server.Default(server.WithHostPorts(metricsAddr))
	metricsServer.GET(metricsPath, metricsHandler)
	go func() {
		if err := metricsServer.Run(); err != nil {
			logger.Error("metrics server failed", zap.Error(err))
		}
	}()
	// 主 server 不再注册 /metrics
} else {
	// 关闭时不注册任何 metrics 路由
}
```

删除原 `r.GET(metricsPath, metricsHandler)`（主 server 上的注册）。

- [ ] **Step 3: docker-compose.prod 去掉默认开启**

读 `docker-compose.prod.yml`，找 `FCB_METRICS_ENABLED: "true"`，改为注释说明：

```yaml
# FCB_METRICS_ENABLED: "true"  # 按需开启，默认关闭（开启时绑定 127.0.0.1:9090）
```

- [ ] **Step 4: 构建 + 提交**

Run: `cd backend && go build ./...`
Expected: 通过

```bash
git add backend/cmd/server/bootstrap/bootstrap.go docker-compose.prod.yml
git commit -m "fix(security): /metrics 默认关闭+开启时内网绑定(127.0.0.1:9090)"
```

---

### Task 7: 前端 XSS 消毒（DOMPurify）

**Files:**
- Modify: `frontend/package.json`（加 dompurify）
- Create: `frontend/src/utils/sanitize.ts`
- Modify: `frontend/src/components/FilePreview.vue`

**Interfaces:**
- Produces: `sanitizeHtml(html: string): string`（DOMPurify 封装）

- [ ] **Step 1: 安装 dompurify**

Run: `cd frontend && npm install dompurify && npm install -D @types/dompurify`
（若用 pnpm 则 `pnpm add dompurify && pnpm add -D @types/dompurify`，按现有 lockfile 判断）

确认 `package.json` 出现 `dompurify` 依赖。

- [ ] **Step 2: 新建 sanitize 封装**

```ts
// frontend/src/utils/sanitize.ts
import DOMPurify from 'dompurify'

/**
 * 消毒 HTML，移除 script/事件处理器等危险内容。
 * 用于 v-html 渲染用户内容（markdown/代码高亮输出）前。
 */
export function sanitizeHtml(html: string): string {
  return DOMPurify.sanitize(html, {
    FORBID_TAGS: ['script', 'style', 'iframe', 'object', 'embed', 'form'],
    FORBID_ATTR: ['onerror', 'onload', 'onclick', 'onmouseover'],
  })
}
```

- [ ] **Step 3: FilePreview.vue 接入消毒**

读 `frontend/src/components/FilePreview.vue`，找 `renderedMarkdown`（:350 附近）和 `highlightedCode`（:360）的计算逻辑。在 `marked.parse()` 输出后包一层 `sanitizeHtml`：

```ts
import { sanitizeHtml } from '@/utils/sanitize'

// renderedMarkdown 计算属性
const renderedMarkdown = computed(() => {
  if (!props.textContent) return ''
  const raw = marked.parse(props.textContent, { breaks: true }) as string
  return sanitizeHtml(raw)
})

// highlightedCode 计算属性
const highlightedCode = computed(() => {
  // ... 原高亮逻辑 ...
  return sanitizeHtml(highlighted)
})
```

注意：确认 `marked.parse` 返回 string（marked v5+ 默认返回 string，旧版可能需 `as string`）。`v-html` 绑定保持不变（:105, :122），现在内容已消毒。

- [ ] **Step 4: 前端构建验证**

Run: `cd frontend && npm run build`
Expected: 构建通过，无 TS 报错

- [ ] **Step 5: 提交**

```bash
git add frontend/package.json frontend/package-lock.json frontend/src/utils/sanitize.ts frontend/src/components/FilePreview.vue
git commit -m "fix(security): 前端 XSS 消毒-DOMPurify 净化 markdown/高亮输出"
```

---

### Task 8: 全量回归 + .env.example 补充

**Files:**
- Modify: `backend/.env.example`（若有）或 `backend/configs/config.yaml`

- [ ] **Step 1: 后端全量构建+测试**

Run: `cd backend && go build ./... && go test ./... 2>&1 | grep -E "^(ok|FAIL)" | tail -15`
Expected: 全 ok

- [ ] **Step 2: .env.example 补 JWT 开发占位**

读 `.env.example`（项目根或 backend/）。补充：

```
# JWT 密钥（必填，全环境 fail-fast；开发可用此占位，生产必须强随机）
FCB_JWT_SECRET=dev-only-change-me-to-random-32chars
# metrics 默认关闭，开启时绑定内网
FCB_METRICS_ENABLED=false
FCB_METRICS_ADDR=127.0.0.1:9090
```

- [ ] **Step 3: 提交**

```bash
git add backend/.env.example
git commit -m "docs(security): .env.example 补充 JWT/metrics 安全配置说明"
```

---

## Self-Review 结果

**1. Spec coverage：**
- XSS → Task 7 ✅
- 限流挂载 → Task 4 ✅
- JWT fail-fast → Task 1 ✅
- metrics 内网 → Task 6 ✅
- CORS 收紧 → Task 5 ✅
- 上传校验 → Task 2（工具）+ Task 3（接入）+ Task 4 Step1（body上限）✅

**2. Placeholder scan：** Task 3/4 中"按实际字段名确认"已说明用 grep 确认方式，非占位。Task 4 Step2 提供了两种挂载方案（group.Use 与路径感知），主推路径感知方案有完整代码。

**3. Type consistency：** `CheckUploadSize(fileSize, maxSize int64) error`、`IsBlockedExtension(filename string, blacklist []string) bool`、`DefaultBlockedExtensions() []string` 在 Task 2 定义、Task 3 使用一致；`sanitizeHtml(html: string): string` Task 7 内一致。
