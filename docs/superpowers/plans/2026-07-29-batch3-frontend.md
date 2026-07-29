# 第三批 前端工程 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修复前端 token 刷新机制、死拦截器+错误码 i18n、关键类型、冗余依赖清理。

**Architecture:** 后端补 refresh 路由；前端 request.ts 接入 api-interceptor 拦截器 + 401 自动刷新；补 errcode locale；补 admin config 类型 + 删影子 shims；清理未用依赖。

**Tech Stack:** Vue3 + TS + axios + vue-i18n + Pinia

## Global Constraints

- 工作根目录：`/Users/zhangyi/my_project/FileCodeBox`，前端在 `frontend/`，后端在 `backend/`
- 前端包管理器：npm（本批统一为 npm，删 pnpm lockfile）
- 后端 errcode 约 30 条，需全部补到 locale
- api-interceptor 已导出 `onFulfilled`/`onRejected`（:59,:80）但 request.ts 未接入
- 提交前缀：`fix(frontend):` 或 `feat(frontend):`
- 每个前端任务结束前 `cd frontend && npx vue-tsc --noEmit` 通过

---

### Task 1: 后端 refresh token 路由

**Files:**
- Modify: `backend/cmd/server/bootstrap/bootstrap.go`

**Interfaces:**
- Produces: `POST /api/v1/user/refresh` 端点

- [ ] **Step 1: 在 customizedRegister 注册 refresh 路由**

读 `backend/cmd/server/bootstrap/bootstrap.go` 的 `customizedRegister` 函数。在 presign upload-direct 路由后加：

```go
// ===== token 刷新端点（前端 401 拦截器调用）=====
r.POST("/api/v1/user/refresh", func(ctx context.Context, c *app.RequestContext) {
	authHeader := string(c.GetHeader("Authorization"))
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(consts.StatusUnauthorized, map[string]interface{}{"code": 401, "message": "missing token"})
		return
	}
	oldToken := strings.TrimPrefix(authHeader, "Bearer ")
	newToken, err := auth.RefreshToken(oldToken)
	if err != nil {
		c.JSON(consts.StatusUnauthorized, map[string]interface{}{"code": 401, "message": "token invalid or expired"})
		return
	}
	c.JSON(consts.StatusOK, map[string]interface{}{
		"code": 200, "message": "ok",
		"data": map[string]string{"token": newToken},
	})
})
```

需确认 import 含 `auth`（已有，bootstrap 用了 auth 包）和 `strings`（已有）。

- [ ] **Step 2: 构建**

Run: `cd backend && go build ./...`
Expected: 通过

- [ ] **Step 3: 提交**

```bash
git add backend/cmd/server/bootstrap/bootstrap.go
git commit -m "feat(frontend): 后端补 /user/refresh token 刷新路由"
```

---

### Task 2: 错误码 i18n 补全（locale 文件）

**Files:**
- Modify: `frontend/src/i18n/locales/zh-CN.ts`
- Modify: `frontend/src/i18n/locales/en-US.ts`

**Interfaces:**
- Produces: `errcode.*` key 覆盖后端全部 ~30 个错误码

- [ ] **Step 1: 读 errcode 清单**

Run: `grep -E "Code[A-Z].*= [0-9]" backend/internal/pkg/errcode/errcode.go`
记录所有错误码常量名和数值。

- [ ] **Step 2: zh-CN.ts 补 errcode 块**

读 `frontend/src/i18n/locales/zh-CN.ts`，在合适层级（与其他顶级 key 平级）加 `errcode` 对象。完整清单（按后端 errcode.go）：

```ts
errcode: {
  '10000': '未知错误',
  '10001': '参数错误',
  '10002': '未登录或登录已过期',
  '10003': '无权限',
  '10004': '资源不存在',
  '10005': '请求过于频繁，请稍后再试',
  '10006': '服务不可用',
  '10007': '请求超时',
  '10008': '服务器内部错误',
  '10009': '请求方法不允许',
  '10010': '请求体过大',
  '20001': '分享不存在',
  '20002': '分享已过期',
  '20003': '密码错误',
  '20004': '超过取件次数',
  '20005': '取件码不存在',
  '20006': '取件码已过期',
  '20007': '取件码已用完',
  '20008': '文件不存在',
  '20009': '分片无效',
  '30001': '存储初始化失败',
  '30002': '存储配额超限',
  '30003': '上传失败',
  '30004': '下载失败',
  '30005': '删除失败',
  '30006': '存储方案不支持',
  '30007': '存储 IO 错误',
  '30008': '预签名生成失败',
  '30009': '预签名 token 无效',
  '40001': '用户不存在',
  '40002': '用户已存在',
  '40004': '用户已禁用',
},
```

- [ ] **Step 3: en-US.ts 补对应英文 errcode**

读 `frontend/src/i18n/locales/en-US.ts`，加同样结构的 `errcode` 对象（英文文案）：

```ts
errcode: {
  '10000': 'Unknown error',
  '10001': 'Invalid parameter',
  '10002': 'Not logged in or session expired',
  '10003': 'Permission denied',
  '10004': 'Resource not found',
  '10005': 'Too many requests, please try later',
  '10006': 'Service unavailable',
  '10007': 'Request timeout',
  '10008': 'Internal server error',
  '10009': 'Method not allowed',
  '10010': 'Request body too large',
  '20001': 'Share not found',
  '20002': 'Share expired',
  '20003': 'Wrong password',
  '20004': 'Pickup limit reached',
  '20005': 'Pickup code not found',
  '20006': 'Pickup code expired',
  '20007': 'Pickup code exhausted',
  '20008': 'File not found',
  '20009': 'Invalid chunk',
  '30001': 'Storage init failed',
  '30002': 'Storage quota exceeded',
  '30003': 'Upload failed',
  '30004': 'Download failed',
  '30005': 'Delete failed',
  '30006': 'Storage scheme not supported',
  '30007': 'Storage IO error',
  '30008': 'Presign generation failed',
  '30009': 'Presign token invalid',
  '40001': 'User not found',
  '40002': 'User already exists',
  '40004': 'User disabled',
},
```

- [ ] **Step 4: tsc 验证**

Run: `cd frontend && npx vue-tsc --noEmit 2>&1 | tail -5`
Expected: 通过（纯数据补充）

- [ ] **Step 5: 提交**

```bash
git add frontend/src/i18n/locales/zh-CN.ts frontend/src/i18n/locales/en-US.ts
git commit -m "fix(frontend): 补全 errcode i18n(30+错误码中英文)"
```

---

### Task 3: request.ts 接入 api-interceptor 拦截器 + 401 自动刷新

**Files:**
- Modify: `frontend/src/utils/request.ts`
- Modify: `frontend/src/stores/user.ts`

**Interfaces:**
- Consumes: `onFulfilled`/`onRejected`（api-interceptor.ts）、`/api/v1/user/refresh`（Task 1）
- Produces: request.ts 响应拦截器接入 BizError 包装 + 401 刷新

- [ ] **Step 1: 读 request.ts 和 api-interceptor.ts 全文**

Run: `cat frontend/src/utils/request.ts` 和 `cat frontend/src/utils/api-interceptor.ts`
理解两者职责：request.ts 提取 trace_id + 归一化错误；api-interceptor 的 onFulfilled 把响应包成 BizResponse、onRejected 包成 BizError。

- [ ] **Step 2: user store 加 refresh 方法**

读 `frontend/src/stores/user.ts`，加 token 刷新方法：

```ts
async refreshToken(): Promise<string | null> {
  const oldToken = this.token
  if (!oldToken) return null
  try {
    const res = await axios.post('/api/v1/user/refresh', null, {
      baseURL: import.meta.env.VITE_API_BASE_URL || '',
      headers: { Authorization: `Bearer ${oldToken}` },
    })
    const newToken = res.data?.data?.token
    if (newToken) {
      this.token = newToken
      localStorage.setItem('token', newToken)
      return newToken
    }
  } catch {
    // refresh 失败
  }
  return null
}
```

需 import axios（store 文件顶部）。

- [ ] **Step 3: request.ts 改造响应拦截器**

替换 `frontend/src/utils/request.ts` 的响应拦截器（:37-69）。合并 api-interceptor 的 onFulfilled（BizResponse 包装）+ 保留 trace_id 提取 + 加 401 刷新：

```ts
import { onRejected } from './api-interceptor'

// 防并发刷新：多个 401 共享同一次 refresh
let refreshPromise: Promise<string | null> | null = null

instance.interceptors.response.use(
  (response) => {
    // 提取 X-Trace-Id
    const traceId = response.headers?.['x-trace-id'] as string | undefined
    if (traceId && response.data && typeof response.data === 'object') {
      ;(response.data as Record<string, unknown>).trace_id = traceId
    }
    return response.data
  },
  async (error) => {
    const originalRequest = error.config
    // 401 且未在刷新中且非 refresh 接口本身 → 尝试刷新
    if (
      error.response?.status === 401 &&
      !originalRequest._retry &&
      !originalRequest.url?.includes('/user/refresh') &&
      !originalRequest.url?.includes('/user/login') &&
      !originalRequest.url?.includes('/admin/login')
    ) {
      originalRequest._retry = true
      if (!refreshPromise) {
        refreshPromise = useUserStore().refreshToken().finally(() => {
          refreshPromise = null
        })
      }
      const newToken = await refreshPromise
      if (newToken) {
        originalRequest.headers.Authorization = `Bearer ${newToken}`
        return instance(originalRequest)
      }
      // refresh 失败 → 跳登录
      useUserStore().logout()
      window.location.href = '/login'
      return Promise.reject(error)
    }
    // 其他错误：归一化（保留原逻辑）
    return onRejected(error)
  }
)
```

注意：`useUserStore` 需 import（从 stores/user）。`onRejected` 来自 api-interceptor，负责 BizError 包装。需确认 api-interceptor 的 onRejected 签名兼容（它接收 AxiosError 返回 Promise<never>）。

- [ ] **Step 4: tsc 验证**

Run: `cd frontend && npx vue-tsc --noEmit 2>&1 | tail -10`
Expected: 通过（若有类型错误，按提示修正——常见：_retry 不在 AxiosRequestConfig 类型上，需 `as any` 或扩展接口）

- [ ] **Step 5: 提交**

```bash
git add frontend/src/utils/request.ts frontend/src/stores/user.ts
git commit -m "fix(frontend): 接入 api-interceptor 拦截器 + 401 token 自动刷新"
```

---

### Task 4: admin config 类型 + 删影子 shims

**Files:**
- Modify: `frontend/src/api/admin.ts`
- Modify: `frontend/src/types/shims.d.ts`

- [ ] **Step 1: 读 admin.ts 的 config 接口签名**

Run: `grep -n "updateConfig\|updateBasicConfig\|updateSecurityConfig\|updateEmailConfig\|any" frontend/src/api/admin.ts`
记录 4 个 any 参数的位置。

- [ ] **Step 2: 定义 config 类型替换 any**

读 `frontend/src/api/admin.ts`，为 4 个 config 接口定义参数类型。在文件顶部（接口定义区）加：

```ts
export interface BasicConfigUpdate {
  site_name?: string
  site_description?: string
  [key: string]: unknown
}

export interface SecurityConfigUpdate {
  allow_registration?: boolean
  [key: string]: unknown
}

export interface EmailConfigUpdate {
  smtp_host?: string
  smtp_port?: number
  [key: string]: unknown
}

export type AdminConfigUpdate = BasicConfigUpdate | SecurityConfigUpdate | EmailConfigUpdate | Record<string, unknown>
```

然后替换：
- `updateConfig(config: any)` → `updateConfig(config: AdminConfigUpdate)`
- `updateBasicConfig(data: any)` → `updateBasicConfig(data: BasicConfigUpdate)`
- `updateSecurityConfig(data: any)` → `updateSecurityConfig(data: SecurityConfigUpdate)`
- `updateEmailConfig(data: any)` → `updateEmailConfig(data: EmailConfigUpdate)`
- `PaginatedResponse<any>` → `PaginatedResponse<Record<string, unknown>>`（或具体行类型）

- [ ] **Step 3: 删 shims.d.ts 覆盖真实模块的声明**

读 `frontend/src/types/shims.d.ts`，删除以下覆盖真实模块的声明（保留 CSS/IMG/swagger 等环境声明）：
```ts
declare module '@/api/user' { ... }   // 删
declare module '@/api/share' { ... }  // 删
declare module '@/api/admin' { ... }  // 删
```

- [ ] **Step 4: tsc 验证 + 修复暴露的错误**

Run: `cd frontend && npx vue-tsc --noEmit 2>&1 | head -20`
Expected: 删 shims 后可能暴露之前被 any 掩盖的类型错误。逐个修复（通常是补 import 路径或具体类型）。若错误来自调用方误用 any，补正确类型。

- [ ] **Step 5: 构建验证**

Run: `cd frontend && npm run build 2>&1 | tail -5`
Expected: 通过

- [ ] **Step 6: 提交**

```bash
git add frontend/src/api/admin.ts frontend/src/types/shims.d.ts
git commit -m "fix(frontend): admin config 补类型 + 删影子 shims(暴露真实类型检查)"
```

---

### Task 5: 清理冗余依赖 + 统一 lockfile

**Files:**
- Modify: `frontend/package.json`
- Modify: `frontend/src/main.ts`
- Delete: `frontend/pnpm-lock.yaml`、`frontend/pnpm-workspace.yaml`

- [ ] **Step 1: 确认 vue-query 确实未用**

Run: `grep -rn "useQuery\|useMutation\|useQueryClient" frontend/src/`
Expected: 无结果（确认仅 main.ts 注册插件，无实际使用）

- [ ] **Step 2: 删 vue-query 注册**

读 `frontend/src/main.ts`，删除：
- `import { VueQueryPlugin } from '@tanstack/vue-query'`（:7）
- `app.use(VueQueryPlugin)`（:26）

- [ ] **Step 3: 卸载未用依赖**

Run: `cd frontend && npm uninstall @vueuse/core @tanstack/vue-query 2>&1 | tail -3`

- [ ] **Step 4: 删 pnpm lockfile**

Run: `rm frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml`

- [ ] **Step 5: tsc + 构建验证**

Run: `cd frontend && npx vue-tsc --noEmit 2>&1 | tail -3 && npm run build 2>&1 | tail -3`
Expected: 通过

- [ ] **Step 6: 提交**

```bash
git add frontend/package.json frontend/package-lock.json frontend/src/main.ts
git rm frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml
git commit -m "fix(frontend): 清理未用依赖(vue-query/vueuse) + 统一 npm lockfile"
```

---

### Task 6: admin 登录移除 atob（用 /user/info）+ 路由守卫注释

**Files:**
- Modify: `frontend/src/views/admin/Login.vue`
- Modify: `frontend/router/index.ts`

- [ ] **Step 1: 读 admin Login.vue 的 atob 逻辑**

Run: `grep -n "atob\|userRole\|localStorage\|fetchUserInfo\|/user/info" frontend/src/views/admin/Login.vue`
记录 :124-138 的 atob 解析逻辑。

- [ ] **Step 2: 移除 atob，登录后调 user store**

读 `frontend/src/views/admin/Login.vue` 的登录成功处理（:124-149）。移除 atob 解析 JWT 的代码，改为：

```ts
// 登录成功后，用 token 拉取用户信息获取 role
import { useUserStore } from '@/stores/user'
const userStore = useUserStore()

// 登录成功后：
await userStore.fetchUserInfo()  // 拉取 /user/info，store 内设置 role
const role = userStore.userInfo?.role
if (role !== 'admin') {
  ElMessage.error('非管理员账号')
  return
}
localStorage.setItem('userRole', 'admin')  // 仅 UX 缓存
```

注意：`fetchUserInfo` 是否存在于 user store 需确认（grep）。若不存在，用 axios 直接调 `/api/v1/user/info`。

- [ ] **Step 3: 路由守卫补注释**

读 `frontend/router/index.ts:159-165`，在 `localStorage.getItem('userRole')` 判断处补注释：

```ts
// NOTE: userRole 仅作客户端 UX 路由控制，非安全鉴权。
// 真正的 admin 权限由后端中间件（admin_auth）在每个 /admin API 强制校验。
const role = localStorage.getItem('userRole')
```

- [ ] **Step 4: tsc + 构建**

Run: `cd frontend && npx vue-tsc --noEmit 2>&1 | tail -5 && npm run build 2>&1 | tail -3`
Expected: 通过

- [ ] **Step 5: 提交**

```bash
git add frontend/src/views/admin/Login.vue frontend/router/index.ts
git commit -m "fix(frontend): admin 登录移除 atob 解析(改用/user/info) + 路由守卫补鉴权注释"
```

---

### Task 7: 全量回归

- [ ] **Step 1: 前端 tsc + build**

Run: `cd frontend && npx vue-tsc --noEmit && npm run build`
Expected: 通过

- [ ] **Step 2: 后端构建**

Run: `cd backend && go build ./...`
Expected: 通过

- [ ] **Step 3: 后端测试回归**

Run: `cd backend && go test ./... 2>&1 | grep -E "^(ok|FAIL)" | tail`
Expected: 全 ok

---

## Self-Review 结果

**1. Spec coverage：**
- token 刷新 → Task 1（后端路由）+ Task 3（前端拦截器）✅
- 死拦截器 + 错误码 i18n → Task 2（locale）+ Task 3（接入拦截器）✅
- 关键类型 + 删 shims → Task 4 ✅
- 冗余依赖 + lockfile → Task 5 ✅
- admin atob → Task 6 ✅

**2. Placeholder scan：** Task 3 Step 3 的 `_retry` 类型扩展说明"需 `as any` 或扩展接口"——这是必要的实现时适配，非占位。Task 6 Step 2 的 fetchUserInfo 存在性说明"若不存在用 axios"——给了明确 fallback。

**3. Type consistency：** `BasicConfigUpdate/SecurityConfigUpdate/EmailConfigUpdate/AdminConfigUpdate` 在 Task 4 定义并使用一致；`refreshToken()` 在 Task 3 Step2 定义、Step3 调用一致。
