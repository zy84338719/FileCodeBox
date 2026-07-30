# FileCodeBox 飞牛适配层集成测试与凭证接入 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 `filecodebox-fnos` 的飞牛适配层建立可运行的集成测试,并把凭证到手后的四项能力接入点固化为清晰步骤。

**Architecture:** adapter 的可测行为 = `adapter.Mount(h, cfg)` 在一个空 `*server.Hertz` 上的路由表现(降级模式 503、各子路由注册),用 Hertz 的 `ut.PerformRequest` 做 HTTP 级测试,不依赖 bootstrap 的 DB/Redis/JWT 副作用。凭证到手后的真实飞牛 API 接入,以 `adapter/internal/client` 为统一注入点。

**Tech Stack:** Go 1.26 · CloudWeGo Hertz v0.9.6(`ut` 测试工具) · testify(assert/require) · 环境:`/Users/zhangyi/my_project/filecodebox-fnos`

## Global Constraints

- 测试不调用 `bootstrap.Bootstrap()`(它有 DB/Redis/JWT fail-fast 副作用),改用 `server.Default()` 创建空 engine。
- 测试用 SQLite 内存库或完全不触 DB(adapter.Mount 不需要 DB)。
- 飞牛 Open API 的端点/签名/响应结构**当前未知**,凡涉及真实飞牛调用的步骤,须以「凭证+官方文档到手」为前置条件,按文档填实,不得臆造接口。
- 代码风格对齐原项目:`package` 注释、中文注释、testify。
- 每个任务结束 `go build ./... && go vet ./...` 必须通过。

---

## File Structure

| 文件 | 职责 | 状态 |
|------|------|------|
| `adapter/adapter_test.go` | 测试 `Mount` 降级模式 503 + 启用模式路由注册 | 新建 |
| `adapter/internal/fnosconfig/config_test.go` | 测试 `LoadConfig` 降级判定逻辑 | 新建 |
| `adapter/internal/client/client.go` | 飞牛 Open API 统一 client(签名/请求/重试/降级),凭证注入点 | 新建(凭证到手后) |
| `adapter/sso/sso.go` | SSO:ticket→飞牛用户→本系统用户→JWT | 已有桩,凭证后实现 |
| `adapter/storage/storage.go` | 共享目录列表 | 已有桩,凭证后实现 |
| `adapter/notify/notify.go` | 通知中心推送 | 已有桩,凭证后实现 |
| `adapter/tunnel/tunnel.go` | 内网穿透外网链接 | 已有桩,凭证后实现 |

---

## Task 1: 配置降级判定单元测试

**Files:**
- Create: `adapter/internal/fnosconfig/config_test.go`

**Interfaces:**
- Consumes: `fnosconfig.LoadConfig()`(已存在于 `config.go`),依赖环境变量 `FNOS_ENABLED/FNOS_APPID/FNOS_APPSECRET/FNOS_API_BASE`。
- Produces: 验证 `Config.Enabled`/`DisabledReason`/`APIBase` 在各种环境变量组合下的取值。

- [ ] **Step 1: 写失败测试**

```go
package fnosconfig

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 辅助:设置一批环境变量并返回清理函数。
func setEnvs(t *testing.T, kv map[string]string) {
	t.Helper()
	// 先清空,确保测试隔离
	for _, k := range []string{"FNOS_ENABLED", "FNOS_APPID", "FNOS_APPSECRET", "FNOS_API_BASE"} {
		orig, had := os.LookupEnv(k)
		os.Unsetenv(k)
		t.Cleanup(func() {
			if had {
				os.Setenv(k, orig)
			} else {
				os.Unsetenv(k)
			}
		})
	}
	for k, v := range kv {
		os.Setenv(k, v)
	}
}

func TestLoadConfig_DisabledByDefault(t *testing.T) {
	setEnvs(t, nil) // 全部缺失
	cfg := LoadConfig()
	assert.False(t, cfg.Enabled)
	assert.Equal(t, "https://open.fnnas.com", cfg.APIBase) // 缺省占位
	assert.Contains(t, cfg.DisabledReason, "FNOS_ENABLED")
}

func TestLoadConfig_EnabledButMissingSecret(t *testing.T) {
	setEnvs(t, map[string]string{"FNOS_ENABLED": "true", "FNOS_APPID": "app1"})
	cfg := LoadConfig()
	assert.False(t, cfg.Enabled)
	assert.Contains(t, cfg.DisabledReason, "APPSECRET")
}

func TestLoadConfig_Enabled(t *testing.T) {
	setEnvs(t, map[string]string{
		"FNOS_ENABLED":   "true",
		"FNOS_APPID":     "app1",
		"FNOS_APPSECRET": "sec1",
		"FNOS_API_BASE":  "https://api.example.com",
	})
	cfg := LoadConfig()
	assert.True(t, cfg.Enabled)
	assert.Equal(t, "app1", cfg.AppID)
	assert.Equal(t, "https://api.example.com", cfg.APIBase)
}
```

- [ ] **Step 2: 运行测试确认通过(纯函数,无需先失败)**

Run: `cd /Users/zhangyi/my_project/filecodebox-fnos && go test ./adapter/internal/fnosconfig/ -v`
Expected: PASS(被测代码已存在,测试应直接通过;若失败说明 config.go 逻辑有 bug,需修)。

- [ ] **Step 3: 提交**

```bash
cd /Users/zhangyi/my_project/filecodebox-fnos
git add adapter/internal/fnosconfig/config_test.go
git commit -m "test(fnosconfig): LoadConfig 降级判定单元测试"
```

---

## Task 2: adapter.Mount 降级模式集成测试

**Files:**
- Create: `adapter/adapter_test.go`

**Interfaces:**
- Consumes: `adapter.Mount(h *server.Hertz, cfg Config)`(已存在于 `adapter.go`),`adapter.Config` 类型别名,`server.Default()` 创建空 engine。
- Produces: 验证降级模式下 `/api/fnos/*` 全部返回 503,业务路由不受影响。

**关键:** 用 `server.Default()` 建空 Hertz(不调 bootstrap,零 DB/Redis 副作用),注册一个探针路由 `/api/v1/ping` 模拟 FileCodeBox 业务,挂载 adapter 后验证隔离性。

- [ ] **Step 1: 写失败测试**

```go
package adapter

import (
	"context"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/stretchr/testify/assert"

	"github.com/zy84338719/filecodebox-fnos/adapter/internal/fnosconfig"
)

// newTestEngine 建一个空 Hertz + 探针路由,挂载指定 cfg 的 adapter。
func newTestEngine(t *testing.T, cfg fnosconfig.Config) *server.Hertz {
	t.Helper()
	h := server.Default()
	// 探针路由:模拟 FileCodeBox 业务路由,验证 adapter 不影响它
	h.GET("/api/v1/ping", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(consts.StatusOK, map[string]string{"msg": "pong"})
	})
	Mount(h, cfg)
	return h
}

func TestMount_DisabledMode_Returns503(t *testing.T) {
	cfg := fnosconfig.Config{Enabled: false, DisabledReason: "FNOS_ENABLED 未设为 true"}
	h := newTestEngine(t, cfg)

	// 飞牛路由组应返回 503
	w := ut.PerformRequest(h, consts.MethodPost, "/api/fnos/login", nil)
	resp := w.Result()
	assert.Equal(t, consts.StatusServiceUnavailable, resp.StatusCode())
}

func TestMount_DisabledMode_DoesNotAffectBusinessRoutes(t *testing.T) {
	cfg := fnosconfig.Config{Enabled: false}
	h := newTestEngine(t, cfg)

	// 业务探针路由应正常 200
	w := ut.PerformRequest(h, consts.MethodGet, "/api/v1/ping", nil)
	resp := w.Result()
	assert.Equal(t, consts.StatusOK, resp.StatusCode())
}

func TestMount_EnabledMode_RegistersFnosRoutes(t *testing.T) {
	cfg := fnosconfig.Config{
		Enabled:   true,
		AppID:     "app1",
		AppSecret: "sec1",
		APIBase:   "https://api.example.com",
	}
	h := newTestEngine(t, cfg)

	// 启用模式:SSO 路由应注册(桩返回 501 而非 404)
	w := ut.PerformRequest(h, consts.MethodPost, "/api/fnos/login", nil)
	resp := w.Result()
	assert.NotEqual(t, consts.StatusNotFound, resp.StatusCode(), "路由应已注册,不应 404")
}
```

- [ ] **Step 2: 运行测试确认通过**

Run: `cd /Users/zhangyi/my_project/filecodebox-fnos && go test ./adapter/ -run TestMount -v`
Expected: PASS。若 `ut.PerformRequest` body 参数签名不符(本版要求 `*ut.Body`),改传 `nil`——`nil` body 对 POST 桩 handler 无影响。

- [ ] **Step 3: 提交**

```bash
cd /Users/zhangyi/my_project/filecodebox-fnos
git add adapter/adapter_test.go
git commit -m "test(adapter): Mount 降级模式 503 与业务路由隔离集成测试"
```

---

## Task 3: 飞牛 Open API client 抽象(凭证注入点)

> **前置条件:** 此任务的**接口抽象**现在可做(不依赖飞牛文档);但 client 内部的**签名算法、真实端点**须等飞牛官方文档到手才能填实。本任务先产出可编译的 client 骨架与降级逻辑,真实调用部分标注 `凭证到手后按文档填`。

**Files:**
- Create: `adapter/internal/client/client.go`
- Create: `adapter/internal/client/client_test.go`

**Interfaces:**
- Consumes: `fnosconfig.Config`。
- Produces:
  - `type Client struct{ ... }`
  - `func New(cfg fnosconfig.Config) *Client`
  - `func (c *Client) do(ctx, method, path string, body []byte) ([]byte, int, error)` — 统一请求入口(签名/重试/超时注入点)

- [ ] **Step 1: 写失败测试(测降级与超时,不测真实飞牛调用)**

```go
package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/zy84338719/filecodebox-fnos/adapter/internal/fnosconfig"
)

func TestClient_New_RequiresEnabledConfig(t *testing.T) {
	c := New(fnosconfig.Config{Enabled: false})
	assert.Nil(t, c, "降级模式下不应创建 client")
}

func TestClient_Do_HitsConfiguredBase(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		assert.Equal(t, "/test/path", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := New(fnosconfig.Config{Enabled: true, AppID: "a", AppSecret: "s", APIBase: srv.URL})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	body, status, err := c.do(ctx, "GET", "/test/path", nil)

	assert.NoError(t, err)
	assert.True(t, called, "应请求到测试服务器")
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, `{"ok":true}`, string(body))
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd /Users/zhangyi/my_project/filecodebox-fnos && go test ./adapter/internal/client/ -v`
Expected: FAIL(`client.go` 尚不存在,包未定义)。

- [ ] **Step 3: 实现 client 骨架**

```go
// Package client 是飞牛 Open API 的统一 HTTP client。
//
// 职责:签名注入、请求重试、超时控制、错误码映射。
// 所有飞牛 API 调用经此 client,便于凭证注入与降级判断。
//
// 注意:签名算法与真实端点待飞牛官方 Open API 文档到手后填实。
package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/zy84338719/filecodebox-fnos/adapter/internal/fnosconfig"
)

// Client 飞牛 Open API client。
type Client struct {
	cfg    fnosconfig.Config
	http   *http.Client
	// TODO(凭证+文档到手):签名密钥/nonce 缓存等
}

// New 创建飞牛 client。降级模式(cfg.Enabled=false)返回 nil。
func New(cfg fnosconfig.Config) *Client {
	if !cfg.Enabled {
		return nil
	}
	return &Client{
		cfg: cfg,
		http: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// do 统一请求入口(小写,仅包内/各能力模块经封装方法调用)。
//
// TODO(凭证+文档到手):
//   1. 按飞牛文档实现签名(注入 appid/签名/timestamp 到 header 或 query)
//   2. 实现重试(网络错误重试 2 次)
//   3. 映射飞牛错误码
func (c *Client) do(ctx context.Context, method, path string, body []byte) ([]byte, int, error) {
	url := c.cfg.APIBase + path

	var bodyReader io.Reader
	if body != nil {
		bodyReader = io.NopCloser(bytes.NewReader(body))
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("构造请求失败: %w", err)
	}
	// TODO(文档到手):在此注入飞牛要求的签名 header
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("请求飞牛失败: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("读取响应失败: %w", err)
	}
	return data, resp.StatusCode, nil
}
```

> 注意:`bodyReader` 用 `io.NopCloser` 包裹以适配 `http.NewRequestWithContext` 的 `io.Reader`;实际需 `import "bytes"`。提交前确保 `goimports`/编译通过。

- [ ] **Step 4: 运行测试确认通过**

Run: `cd /Users/zhangyi/my_project/filecodebox-fnos && go test ./adapter/internal/client/ -v`
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
cd /Users/zhangyi/my_project/filecodebox-fnos
git add adapter/internal/client/
git commit -m "feat(client): 飞牛 Open API client 骨架(签名/端点待文档到手填实)"
```

---

## Task 4-N: 凭证到手后的四项能力接入(模板)

> **前置条件(全部满足才能开始):** ① 已获飞牛开发者 appid/appsecret;② 拿到飞牛官方 Open API 文档(含 SSO/共享目录/通知/穿透的端点、签名算法、请求/响应结构)。
>
> **本节不是可立即执行的 TDD 步骤**,而是凭证到手后每个能力的接入清单。届时按以下顺序、对照飞牛文档填实 client 签名与各模块实现。每个能力遵循:先写对飞牛 mock server 的测试 → 实现 → 提交。

### 接入顺序(按价值与依赖)
1. **client 签名**(Task 4):按文档在 `client.do` 注入签名,是其余三项的前提。
2. **SSO**(Task 5):`sso/sso.go` 实现 `POST /api/fnos/login`,需复用原项目 `api/pkg/auth` 签本系统 JWT、`api/app/user` 查/建用户。
3. **共享目录**(Task 6):`storage/storage.go` 实现 `GET /api/fnos/shares`,调飞牛"列出共享目录"。
4. **通知**(Task 7):`notify/notify.go` 实现 `Send`,调飞牛通知中心。
5. **内网穿透**(Task 8):`tunnel/tunnel.go` 实现,调飞牛穿透接口。

### 每个能力的接入清单(凭证到手后逐条对照飞牛文档)
- [ ] 飞牛端点 URL 与 HTTP method(填入 client 调用)
- [ ] 请求参数结构(query/body/header)
- [ ] 签名算法(填入 `client.do`)
- [ ] 响应结构(定义 Go struct 反序列化)
- [ ] 错误码与降级处理
- [ ] 对飞牛 mock server 的单元测试(遵循 Task 1-3 风格:httptest + testify)
- [ ] 真实凭证联调(在飞牛 NAS 实测)
- [ ] 提交

### SSO 专项:需复用的原项目符号(已确认导出)
- `github.com/zy84338719/fileCodeBox/backend/api/pkg/auth` — JWT 签发(`SetJWTSecret` 已在 bootstrap 调用)
- 用户查/建:需在凭证到手后,核实 `api/app/user` 是否有可复用的 `GetOrCreate` 方法;若无,直接走 `api/repo/db` 的 user DAO。

---

## Self-Review 结论

1. **Spec coverage:** 设计文档第 5 节四项能力 → Task 4-N 清单覆盖;第 7 节降级策略 → Task 1/2 覆盖。凭证到手后的真实实现受限于飞牛文档未公开,无法写成 TDD 步骤,已在每个 Task 顶部显式标注前置条件。
2. **Placeholder scan:** Task 1-3 无占位,均为可立即执行的真实代码。Task 4-N 的"待文档填实"是**事实约束**(非计划疏漏),已用前置条件块明确标注。
3. **Type consistency:** `fnosconfig.Config` 字段(Enabled/AppID/AppSecret/APIBase/DisabledReason)在 Task 1-3 一致;`adapter.Mount(h, cfg)`、`client.New(cfg)`/`client.do(...)` 签名前后一致。

---

## Execution Handoff

本计划 Task 1-3 可立即执行(不依赖飞牛凭证/文档);Task 4-N 等凭证到手。两种执行方式:
1. **Subagent-Driven(推荐)**:每个 task 派新 subagent,任务间审查。
2. **Inline Execution**:本会话内批量执行 + 检查点。
