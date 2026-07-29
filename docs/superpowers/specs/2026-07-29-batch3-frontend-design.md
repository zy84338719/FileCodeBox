# 第三批：前端工程设计

> 日期：2026-07-29
> 状态：已确认，待实现
> 范围：token 刷新机制、死拦截器修复+错误码 i18n、关键类型、冗余依赖清理
> 这是"项目不足修复"三批计划的第三批，独立可交付、可回退

## 已确认决策

1. **token**：完整刷新机制（后端加 refresh 路由 + 前端 401 拦截自动刷新）
2. **i18n**：修复死拦截器 + 补错误码 key（硬编码 ElMessage 暂不动）
3. **类型**：补关键类型（admin config/PaginatedResponse）+ 删影子 shims
4. **依赖**：清理 @vueuse/core + 评估 vue-query + 统一 lockfile

---

## 一、token 完整刷新机制（前后端）

**问题**：JWT 7天固定过期，无 refresh，前端 token 过期静默登出无提示。

**后端修复**：
- 新增 `POST /api/v1/user/refresh` 路由（bootstrap 自定义层注册）
- handler：从 Authorization header 取旧 token → 调 `auth.RefreshToken` → 返回新 token
- 复用现有 `auth.RefreshToken`（已实现，仅缺路由暴露）

**前端修复**：
- `request.ts` 加 401 响应拦截器：
  - 捕获 401 → 调 `/api/v1/user/refresh`（用当前 token）换新 → 更新 store 的 token → 重放原请求
  - refresh 也失败 → 清 store + 跳登录页 + ElMessage 提示"登录已过期，请重新登录"
- 防并发刷新：用 promise 缓存（多个 401 请求共享同一次 refresh）

---

## 二、修复死拦截器 + 错误码 i18n

**问题**：`api-interceptor.ts` 的 `onFulfilled/onRejected` 从未接入 axios；`errcode.*` key 在 locale 文件不存在，错误码翻译永远回退原始后端消息。

**修复**：
- `request.ts` 接入 `api-interceptor.ts` 的 `onFulfilled/onRejected`（替换或合并现有内联拦截器）
- 补 `errcode.*` key 到 `zh-CN.ts` / `en-US.ts`，覆盖后端 errcode 包的所有错误码（约 20+ 条）
- `useErrorHandler` 的 `translateError` 现在能真正翻译

---

## 三、关键类型 + 删影子 shims

**问题**：admin config 接口全 `any`；`PaginatedResponse<any>`；`shims.d.ts` 声明 `@/api/*` 为 `any` 模块，覆盖真实有类型模块。

**修复**：
- `api/admin.ts`：`updateConfig/updateBasicConfig/updateSecurityConfig/updateEmailConfig` 补具体参数类型（定义 `AdminConfigUpdate` 等接口）
- `PaginatedResponse<T>` 替换裸 `<any>`
- 删 `types/shims.d.ts` 中 `declare module '@/api/user'` 等覆盖真实模块的声明（保留 swagger 等 CSS/IMG 模块声明）

---

## 四、清理冗余依赖 + 统一 lockfile

**问题**：`@vueuse/core` 未用；`@tanstack/vue-query` 仅注册插件未用；`package-lock.json` + `pnpm-lock.yaml` 双存。

**修复**：
- 删 `@vueuse/core`
- `@tanstack/vue-query`：grep 确认无 `useQuery/useMutation` 使用 → 删（含 main.ts 注册）
- 删 `pnpm-lock.yaml` + `pnpm-workspace.yaml`，统一 npm（package-lock.json）

---

## 五、客户端 admin 鉴权加固（轻量）

**问题**：admin 登录用浏览器 `atob` 解 JWT 拿 role，脆弱；路由守卫信 localStorage.userRole。

**修复**：
- admin `Login.vue` 移除 `atob` 解析，登录后调 `/user/info` 获取 role
- 路由守卫 `localStorage.userRole` 判断保留（UX 层），补注释说明真鉴权在后端

---

## 文件改动清单

| 文件 | 改动 |
|---|---|
| `backend/cmd/server/bootstrap/bootstrap.go` | 注册 refresh 路由 |
| `frontend/src/utils/request.ts` | 接入 api-interceptor + 401 refresh 拦截 |
| `frontend/src/utils/api-interceptor.ts` | 确认/修正拦截器导出 |
| `frontend/src/stores/user.ts` | refresh token 方法 |
| `frontend/src/i18n/locales/zh-CN.ts` | 补 errcode.* key |
| `frontend/src/i18n/locales/en-US.ts` | 补 errcode.* key |
| `frontend/src/api/admin.ts` | 补 config 类型 |
| `frontend/src/api/share.ts` | PaginatedResponse 泛型 |
| `frontend/src/types/shims.d.ts` | 删覆盖真实模块的声明 |
| `frontend/src/main.ts` | 移除 vue-query 注册（若删） |
| `frontend/src/views/admin/Login.vue` | 移除 atob，用 /user/info |
| `frontend/package.json` | 删 @vueuse/core、vue-query |
| `frontend/router/index.ts` | 补鉴权注释 |
| 删除 | `pnpm-lock.yaml`、`pnpm-workspace.yaml` |

## 测试

- 前端 `npm run build` + `vue-tsc --noEmit` 通过
- 后端 refresh 路由构建通过
- 手动验证：token 过期场景刷新链路

## 风险

- **401 拦截器**：错误的拦截逻辑可能阻断正常 401（如密码错误也返回 401）。需精确匹配"token 过期"而非所有 401
- **删 shims.d.ts**：若某处真依赖 any 声明会暴露类型错误——这正是目的（修复后补真实类型）
