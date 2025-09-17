集中管理静态资源（internal/static）
=================================

概述
----
`internal/static` 提供集中注册静态资源的帮助函数，避免在各个路由文件中重复调用 `router.Static(...)` 或 `group.Static(...)`。

提供的 API
-----------
- `RegisterStaticRoutes(router *gin.Engine, cfg *config.ConfigManager)`
  - 在应用根上注册公共静态资源（`/assets`, `/css`, `/js`, `/components`）。

- `RegisterAdminStaticRoutes(adminGroup *gin.RouterGroup, cfg *config.ConfigManager)`
  - 在管理路由组上注册管理后台需要的静态资源（例如 `/admin/css`, `/admin/js`, `/admin/templates` 等）。

使用示例
--------
在 `internal/routes/base.go` 中：

```go
import "github.com/zy84338719/filecodebox/internal/static"

// ...
static.RegisterStaticRoutes(router, cfg)
```

在 `internal/routes/admin.go` 中：

```go
import "github.com/zy84338719/filecodebox/internal/static"

// ...
static.RegisterAdminStaticRoutes(adminGroup, cfg)
```

扩展性
----
- 如果需要添加缓存 header、CDN 前缀或将资源改为嵌入（`embed`），可以在此包中统一实现并将选项通过参数传入 `Register*` 函数。

注意
---
- 本模块使用 `cfg.ThemesSelect` 与相对路径 `./<theme>` 来定位资源，行为与历史实现一致。
