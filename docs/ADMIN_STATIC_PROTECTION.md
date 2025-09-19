# 管理后台静态资源保护（Admin Static Protection）

目的
----
确保管理后台 (`/admin/`) 及其专用静态资源仅由通过管理员认证的用户访问，防止普通用户或未授权者直接加载管理前端代码或模板。

背景
----
历史实现中，静态资源由 `internal/static.RegisterStaticRoutes` 和 `internal/static.RegisterAdminStaticRoutes` 注册，后者会把 `themes/<theme>/admin/*` 目录暴露为 `group.Static`，存在被误用或误配置的风险，从而导致未授权访问。

变更
----
1. 移除或弃用 `RegisterAdminStaticRoutes`（在 `internal/static/assets.go` 中删除了此函数的注册行为，并在 README 中说明）。
2. 在 `internal/routes/admin.go` 中，将 admin 前端入口和 admin 静态资源注册到受保护的 `authGroup` 中，并为每个静态子路径提供显式的 `GET` 和 `HEAD` 处理器：
   - `/admin/`（前端入口）
   - `/admin/js/*filepath`
   - `/admin/css/*filepath`
   - `/admin/templates/*filepath`
   - `/admin/assets/*filepath`
   - `/admin/components/*filepath`

实现细节
--------
- 受保护的处理器使用 `CombinedAdminAuth` 中间件（在 `internal/middleware/combined_auth.go`），该中间件通过 `userService.ValidateToken` 验证传入 JWT，并确保 `claims.Role == "admin"`。
- 静态处理器会先 `os.Stat` 检查文件是否存在，再通过 `c.File` 返回文件内容，避免使用 `group.Static` 在路由层意外公开目录。
- 同时为 `HEAD` 注册处理器，确保 HEAD 与 GET 行为一致（采用相同的中间件和文件存在检查），避免代理或客户端通过 HEAD 绕过认证。

验证步骤
--------
1. 启动服务（或在已运行服务上重启以加载最新代码）。
2. 未认证时访问应返回 401：
   ```bash
   curl -i http://127.0.0.1:12346/admin/
   curl -i http://127.0.0.1:12346/admin/js/main.js
   curl -I http://127.0.0.1:12346/admin/css/base.css
   ```
3. 公共资源仍可访问：
   ```bash
   curl -i http://127.0.0.1:12346/js/main.js
   curl -i http://127.0.0.1:12346/user/login
   ```
4. 认证后（使用管理员 JWT）应能访问 admin 页面与 admin 静态资源。

部署注意
--------
- 在代理（Nginx/Cloudflare 等）层，禁止缓存 `/admin/*` 路径或设置 `Cache-Control: no-store`，以防代理缓存导致绕过认证。
- 如果需要让某些少量前端文件在未认证时可用（例如 favicon 或登录页面需要的最小脚本），请将这些放入 `themes/<theme>/assets/` 并通过 `/assets/*` 提供，而不是放在 `themes/<theme>/admin/`。

回滚策略
--------
如果发现兼容性或前端构建问题，回滚步骤：
1. 在版本控制中回退相关提交（`internal/routes/admin.go` 与 `internal/static/assets.go` 的变更）。
2. 恢复 `RegisterAdminStaticRoutes` 并在部署层追加访问控制（不推荐，存在安全风险）。

作者与审计
----------
变更由开发者在 `remove-initwithdb` 分支上实现与测试，建议在合并到主分支前进行代码审计与部署验证。

*** End of Document ***
