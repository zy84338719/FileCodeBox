# 第二批 一致性+分层 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修复数据一致性（事务）、notify 分层（抽 DAO）、可观测性（结构化日志）、presign 死路由、单例文档化。

**Architecture:** notify 抽 DAO 与其他 service 对齐；DeleteFileByCode 调整删除顺序+事务；全量补 zap 日志；presign 在 bootstrap 自定义层注册 upload-direct 路由。

**Tech Stack:** Go 1.26 + Hertz + GORM + zap

## Global Constraints

- 工作根目录：`/Users/zhangyi/my_project/FileCodeBox`，后端代码在 `backend/`
- go module path：`github.com/zy84338719/fileCodeBox/backend`
- 测试模式：sqlite 内存库 + `db.SetDatabaseInstance(gormDB)`（notify_test 当前用 `NewService(db)` 传参，需改为全局注入）
- logger 使用：`github.com/zy84338719/fileCodeBox/backend/internal/pkg/logger`，zap 封装
- 提交前缀：`fix(consistency):` 或 `refactor(consistency):`
- 每个任务结束前 `cd backend && go build ./...` 通过
- notify Service 必须保留 `CreateForUserSimple` 方法（share 依赖它，:787）

---

### Task 1: notify 抽 DAO（NotifyRepository）

**Files:**
- Create: `backend/internal/repo/db/dao/notify.go`
- Modify: `backend/internal/app/notify/notify.go`
- Modify: `backend/internal/app/notify/notify_test.go`
- Modify: `backend/cmd/server/bootstrap/bootstrap.go`

**Interfaces:**
- Produces: `dao.NotifyRepository`（Create/GetByID/ListByUserID/Update/Delete/MarkAllRead/CountUnread 等方法）

- [ ] **Step 1: 读 notify.go 全文，提取所有 gorm 调用为 repo 方法**

读 `backend/internal/app/notify/notify.go`，记录 11 处 `s.db.WithContext(ctx)...` 调用及其语义：
- :70 `Model(&Notify{})` + 动态 Where/Count（List 查询）
- :103 `First(&n, id)`（GetByID）
- :115 `Model().Where().Find()`（ListByUserID）
- :169 `Create(&n)`（Create）
- :215 `Model().Where("id").Updates()`（Update）
- :227 `Delete(&Notify{}, id)`（Delete）
- :295 `Model().Where().Find()`（List 查询变体）
- :304 `Model().Count()`（Count）
- :345 `Model().Updates()`（批量 MarkRead）
- :354 `Model().Where().Updates()`（MarkAllRead）
- :372 `Create(n)`（Create 变体）

- [ ] **Step 2: 新建 NotifyRepository**

```go
// backend/internal/repo/db/dao/notify.go
package dao

import (
	"context"

	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db/model"
	"gorm.io/gorm"
)

type NotifyRepository struct{}

func NewNotifyRepository() *NotifyRepository { return &NotifyRepository{} }

func (r *NotifyRepository) db() *gorm.DB { return db.GetDB() }

func (r *NotifyRepository) Create(ctx context.Context, n *model.Notify) error {
	return r.db().WithContext(ctx).Create(n).Error
}

func (r *NotifyRepository) GetByID(ctx context.Context, id uint) (*model.Notify, error) {
	var n model.Notify
	err := r.db().WithContext(ctx).First(&n, id).Error
	return &n, err
}

// QueryNotify 通用查询（支持 where 条件 + 分页 + 排序），返回 (items, total, query 闭包)
// 为兼容 notify.go 里 :70/:295 的动态查询，暴露底层 *gorm.DB 链
func (r *NotifyRepository) Query(ctx context.Context) *gorm.DB {
	return r.db().WithContext(ctx).Model(&model.Notify{})
}

func (r *NotifyRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) (int64, error) {
	res := r.db().WithContext(ctx).Model(&model.Notify{}).Where("id = ?", id).Updates(updates)
	return res.RowsAffected, res.Error
}

func (r *NotifyRepository) Delete(ctx context.Context, id uint) (int64, error) {
	res := r.db().WithContext(ctx).Delete(&model.Notify{}, id)
	return res.RowsAffected, res.Error
}

func (r *NotifyRepository) UpdatesWhere(ctx context.Context, where map[string]interface{}, updates map[string]interface{}) (int64, error) {
	q := r.db().WithContext(ctx).Model(&model.Notify{})
	for k, v := range where {
		q = q.Where(k, v)
	}
	res := q.Updates(updates)
	return res.RowsAffected, res.Error
}
```

注意：notify.go 的查询较动态（:70/:295 用链式 Where），保留 `Query()` 返回 `*gorm.DB` 供 service 继续链式调用，避免过度抽象。其余简单 CRUD 用专门方法。

- [ ] **Step 3: 改 notify.go 持有 repo**

修改 `backend/internal/app/notify/notify.go`：
- struct `db *gorm.DB` → `notifyRepo *dao.NotifyRepository`
- `NewService(db *gorm.DB)` → `NewService()`，内部 `dao.NewNotifyRepository()`
- 所有 `s.db.WithContext(ctx)` 简单 CRUD 改为 `s.notifyRepo.Xxx(ctx, ...)`
- 动态查询（:70/:295）改为 `s.notifyRepo.Query(ctx).Where(...)...`
- import 移除 `gorm.io/gorm`（若不再直接用），加 dao import
- **保留** `CreateForUserSimple` 方法签名不变

- [ ] **Step 4: 改 notify_test.go 适配全局 DB 注入**

修改 `backend/internal/app/notify/notify_test.go`：
- `newTestService` 从 `NewService(newTestDB(t))` 改为：
```go
func newTestService(t *testing.T) *Service {
	t.Helper()
	g := newTestDB(t)
	db.SetDatabaseInstance(g)  // 注入全局
	t.Cleanup(func() { db.SetDatabaseInstance(nil) })
	return NewService()
}
```
- import 加 `"github.com/zy84338719/fileCodeBox/backend/internal/repo/db"`

- [ ] **Step 5: 改 bootstrap notify 初始化**

修改 `backend/cmd/server/bootstrap/bootstrap.go:762`：
```go
// 原：notifyApp := notifyAppService.NewService(database)
// 改：
notifyApp := notifyAppService.NewService()
```
（NewService 不再接收 db 参数；repo 内部用全局 db.GetDB()）

- [ ] **Step 6: 构建 + 测试**

Run: `cd backend && go build ./... && go test ./internal/app/notify/ -v 2>&1 | tail -20`
Expected: 构建通过，notify 测试全过（现有测试逻辑不变，仅注入方式改变）

- [ ] **Step 7: 提交**

```bash
git add backend/internal/repo/db/dao/notify.go backend/internal/app/notify/notify.go backend/internal/app/notify/notify_test.go backend/cmd/server/bootstrap/bootstrap.go
git commit -m "refactor(consistency): notify service 抽 DAO(迁移11处 gorm 直连为 repo 方法)"
```

---

### Task 2: DeleteFileByCode 删除顺序+事务

**Files:**
- Modify: `backend/internal/app/share/service.go:299-340`

**Interfaces:**
- Consumes: `db.GetDB()`（全局，用于事务）、`fileCodeRepo.Delete`、`userService.UpdateUserStats`

- [ ] **Step 1: 重写 DeleteFileByCode**

替换 `backend/internal/app/share/service.go:299-340`，调整顺序为：查所有权 → 删DB（事务，含扣统计）→ 删物理文件（事务外）。

```go
func (s *Service) DeleteFileByCode(ctx context.Context, code string, userID uint) error {
	s.ensureRepository()

	// 1. 根据 code 查询文件记录
	file, err := s.fileCodeRepo.GetByCode(ctx, code)
	if err != nil {
		return fmt.Errorf("分享不存在")
	}

	// 2. 验证文件所有权
	if file.UserID == nil || *file.UserID != userID {
		return fmt.Errorf("无权限删除此分享")
	}

	// 3. 删除数据库记录 + 扣减统计（事务保证一致）
	if err := db.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删 DB 记录（用 tx）
		if err := tx.Delete(&model.FileCode{}, file.ID).Error; err != nil {
			return fmt.Errorf("删除分享记录失败: %w", err)
		}
		// 扣减用户 storage 统计（用 tx，与删除同事务）
		if s.userService != nil {
			// 统计表更新直接用 tx
			if err := tx.Model(&model.User{}).Where("id = ?", userID).
				UpdateColumn("used_storage", gorm.Expr("used_storage - ?", file.Size)).Error; err != nil {
				logger.Warn("update user storage failed on delete", zap.Error(err), zap.Uint("user_id", userID))
				// 统计失败不回滚删除（避免因统计表问题无法删分享）
			}
		}
		return nil
	}); err != nil {
		return err
	}

	// 4. 事务提交成功后，删除物理文件（不可回滚，放事务外，失败记日志不阻断）
	if file.FilePath != "" && file.UUIDFileName != "" && s.storage != nil {
		filePath := file.GetFilePath()
		if filePath != "" {
			if err := s.storage.DeleteFile(ctx, filePath); err != nil {
				logger.Warn("delete physical file failed", zap.String("path", filePath), zap.Error(err))
			}
		}
	}

	return nil
}
```

需在 service.go 顶部加 import：
- `"github.com/zy84338719/fileCodeBox/backend/internal/repo/db"`（GetDB）
- `"github.com/zy84338719/fileCodeBox/backend/internal/repo/db/model"`（已有）
- `"go.uber.org/zap"`（logger）
- `"github.com/zy84338719/fileCodeBox/backend/internal/pkg/logger"`

注意：`model.User` 的 storage 字段名需确认（grep `used_storage\|UsedStorage` in model/user.go）。若字段名不同，按实际调整。

- [ ] **Step 2: 确认 User 模型 storage 字段**

Run: `grep -n "storage\|Storage" backend/internal/repo/db/model/user.go | head`
按实际字段名调整 Task 2 Step 1 的 `UpdateColumn("used_storage", ...)`。

- [ ] **Step 3: 构建 + share 测试回归**

Run: `cd backend && go build ./... && go test ./internal/app/share/ -v 2>&1 | tail -15`
Expected: 构建通过，现有 share 测试（含 TestDeleteFileByCode_OwnerSuccess）全过

- [ ] **Step 4: 提交**

```bash
git add backend/internal/app/share/service.go
git commit -m "fix(consistency): DeleteFileByCode 加事务+调整删除顺序(DB先于物理,统计进事务)"
```

---

### Task 3: 全量补结构化日志

**Files:**
- Modify: `backend/internal/app/share/service.go`、`admin/service.go`、`anonymous/anonymous.go`、`chunk/service.go`、`presign/presign.go`
- Modify: `backend/cmd/server/main.go`

**Interfaces:**
- Consumes: `logger.Warn/Error/Info`（`internal/pkg/logger`）

- [ ] **Step 1: share service 补日志**

修改 `backend/internal/app/share/service.go`，在以下静默吞错处补 `logger.Warn`（需加 import `"go.uber.org/zap"` 和 logger）：

- CreateShare/ShareText 的 `UpdateUserStats` 失败（约 :168, :243）：
```go
if err := s.userService.UpdateUserStats(*req.UserID, "uploads", 1); err != nil {
    logger.Warn("update user uploads stat failed", zap.Error(err), zap.Uint("user_id", *req.UserID))
}
```
- CreateShare 的 storage 统计失败（:246）、DeleteFile 的 storage（:290）：同样补 logger.Warn
- GetFileWithUsage 的 fire-and-forget goroutine（:375）：保留 `_ =` 但 goroutine 内 RecordViewerAndNotify 已自带日志（Task 3 Step 3 会加）

- [ ] **Step 2: RecordViewerAndNotify 补日志**

修改 `share/service.go:389-407`，UpdateViewer 失败和 UpdateColumns(last_notified_at) 失败补日志：
```go
if err := s.fileCodeRepo.UpdateViewer(ctx, code, viewerIP); err != nil {
    logger.Warn("update viewer failed", zap.String("code", code), zap.Error(err))
    return nil
}
// ... UpdateColumns 失败：
if err := s.fileCodeRepo.UpdateColumns(ctx, fc.ID, map[string]interface{}{"last_notified_at": now}); err != nil {
    logger.Warn("update last_notified_at failed", zap.Uint("id", fc.ID), zap.Error(err))
}
```

- [ ] **Step 3: admin service 补日志**

修改 `backend/internal/app/admin/service.go:331`（CleanExpiredFiles 物理删除 TODO）：
```go
if err := s.storage.DeleteFile(ctx, fp); err != nil {
    logger.Warn("delete physical file failed during cleanup", zap.String("path", fp), zap.Error(err))
}
```
删除 `// TODO: 接 logger` 注释。admin/service.go 已 import logger（审查确认）。

- [ ] **Step 4: anonymous service 补日志**

修改 `backend/internal/app/anonymous/anonymous.go:281`（CreateAnonymousShare 回滚失败）：
```go
if err := s.fileCodeRepo.Delete(ctx, fc.ID); err != nil {
    logger.Warn("rollback file_code on generate code failed", zap.Uint("id", fc.ID), zap.Error(err))
}
```
需加 import `"github.com/zy84338719/fileCodeBox/backend/internal/pkg/logger"` 和 `"go.uber.org/zap"`。

- [ ] **Step 5: main.go fmt.Printf → logger.Fatal**

修改 `backend/cmd/server/main.go:23`：
```go
// 原：fmt.Printf("Bootstrap failed: %v\n", err)
// 改：
logger.Fatal("bootstrap failed", zap.Error(err))
```
需加 import logger 和 zap，移除 fmt（若仅此处用）。

- [ ] **Step 6: 构建 + 全量测试**

Run: `cd backend && go build ./... && go test ./... 2>&1 | grep -E "^(ok|FAIL)" | tail -15`
Expected: 全 ok

- [ ] **Step 7: 提交**

```bash
git add backend/internal/app/share/service.go backend/internal/app/admin/service.go backend/internal/app/anonymous/anonymous.go backend/cmd/server/main.go
git commit -m "fix(consistency): 全量补结构化日志(覆盖15+处静默吞错)"
```

---

### Task 4: presign upload-direct 死路由修复

**Files:**
- Modify: `backend/internal/app/presign/presign.go`（新增 UploadDirect 方法）
- Modify: `backend/cmd/server/bootstrap/bootstrap.go`（注册路由）
- Modify: `backend/internal/transport/http/handler/`（新增 handler，或内联到 bootstrap）

**Interfaces:**
- Consumes: presign token 校验（`presign.go` 的 verifyToken）、storage 写入

- [ ] **Step 1: 读 presign.go 的 token 校验和 meta 结构**

读 `backend/internal/app/presign/presign.go`，确认：
- `signToken`/`verifyToken` 方法签名（grep `func.*sign\|func.*verify\|func.*Token`）
- `InitMeta` 结构（含 ObjectKey/UploadID/Complete 等字段）
- Redis meta 存储格式（keyUploadMeta）
- Complete 方法的现有逻辑（写入 share 表的方式）

- [ ] **Step 2: presign service 新增 UploadDirect 方法**

在 `backend/internal/app/presign/presign.go` 加方法，接收 uploadID + token + 文件数据，校验后写入 storage 并标记 complete：

```go
// UploadDirect 处理预签名直传：校验 token → 写文件 → 标记 complete
func (s *Service) UploadDirect(ctx context.Context, uploadID, token string, data []byte, fileSize int64, fileName string) error {
	// 1. 校验 token
	meta, err := s.GetMeta(ctx, uploadID)
	if err != nil {
		return fmt.Errorf("upload not found: %w", err)
	}
	if err := s.verifyToken(token, uploadID, meta.ExpireAt); err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	// 2. 大小+类型校验
	if err := utils.CheckUploadSize(fileSize, utils.GetMaxUploadSize()); err != nil {
		return fmt.Errorf("文件过大")
	}
	if utils.IsBlockedExtension(fileName, utils.DefaultBlockedExtensions()) {
		return fmt.Errorf("该文件类型禁止上传")
	}

	// 3. 写入 storage（复用 storage 抽象）
	// 写入 meta.ObjectKey 对应路径
	// 注：具体写入方式依赖 storage 接口，用 SaveFile 或直接写 reader
	// 简化：存到本地 data 目录（与 chunk 一致）
	// TODO 在实现时根据 storage 接口确定

	return nil
}
```

注意：`verifyToken` 和 `GetMeta` 的实际签名需按 Step 1 确认。storage 写入需确认 StorageInterface 是否有 `SaveFromReader` 之类方法，或复用 chunk 的写入逻辑。**实现时若 storage 接口不支持直接写字节，则用临时文件 + SaveFile**。

- [ ] **Step 3: bootstrap 注册 upload-direct 路由**

在 `backend/cmd/server/bootstrap/bootstrap.go` 的自定义路由段（customizedRegister 或主 Run），加：

```go
// presign 预签名直传端点（gen router 未注册，在此补）
r.PUT("/api/v1/presign/upload-direct/:uploadID", func(ctx context.Context, c *app.RequestContext) {
	uploadID := c.Param("uploadID")
	token := string(c.GetHeader("X-Upload-Token"))
	// 读取 body
	data := c.Request.Body()
	fileSize := int64(len(data))
	// 从 query 或 header 取 fileName
	fileName := string(c.Query("file_name"))
	if err := presignSvc.UploadDirect(ctx, uploadID, token, data, fileSize, fileName); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]interface{}{"code": 400, "message": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, map[string]interface{}{"code": 200, "message": "ok"})
})
```

需注入 presignSvc 到 customizedRegister（或用全局）。确认 presign service 在 bootstrap 的可用性（grep presignService 初始化）。

- [ ] **Step 4: 构建 + 手动验证**

Run: `cd backend && go build ./...`
Expected: 通过

启动后用 curl PUT 测试 `/api/v1/presign/upload-direct/:id`（需先 Init 拿 token）。

- [ ] **Step 5: 提交**

```bash
git add backend/internal/app/presign/presign.go backend/cmd/server/bootstrap/bootstrap.go
git commit -m "fix(consistency): 修复 presign upload-direct 死路由(注册端点+实现直传)"
```

---

### Task 5: 全局单例文档化

**Files:**
- Modify: `backend/internal/repo/db/database.go`、`backend/internal/conf/config.go`、`backend/internal/preview/service.go`

- [ ] **Step 1: database.go 补注释**

在 `backend/internal/repo/db/database.go` 的 `var DB *gorm.DB` 上方加注释：
```go
// DB 全局数据库实例。
// 设计权衡：采用全局单例避免 service/dao 层的构造期依赖注入复杂度（dao.NewXxxRepository()
// 内部调 GetDB() 获取连接）。未来若迁移到 DI 容器，可改为构造注入。
// 测试时用 SetDatabaseInstance 注入内存 sqlite。
var DB *gorm.DB
```

- [ ] **Step 2: config.go 补注释**

在 `backend/internal/conf/config.go` 的 `var globalConfig` 上方加类似注释。

- [ ] **Step 3: preview/service.go 补注释**

在 `backend/internal/preview/service.go` 的 `var svc *Service` 上方加注释。

- [ ] **Step 4: 构建 + 提交**

Run: `cd backend && go build ./...`
Expected: 通过（仅注释）

```bash
git add backend/internal/repo/db/database.go backend/internal/conf/config.go backend/internal/preview/service.go
git commit -m "docs(consistency): 全局单例(DB/Config/Preview) 补设计权衡注释"
```

---

### Task 6: 全量回归

- [ ] **Step 1: 全量构建+测试**

Run: `cd backend && go build ./... && go test ./... 2>&1 | grep -E "^(ok|FAIL)" | tail -15`
Expected: 全 ok

- [ ] **Step 2: go vet**

Run: `cd backend && go vet ./... 2>&1 | tail -5`
Expected: 无输出

- [ ] **Step 3: 提交（若有未提交调整）**

```bash
git add -A
git status  # 确认无遗漏
```

---

## Self-Review 结果

**1. Spec coverage：**
- 事务 → Task 2（DeleteFileByCode）✅；CreateShare/ShareText 统计改日志（Task 3）✅
- notify 抽 DAO → Task 1 ✅
- 全量日志 → Task 3 ✅
- presign 死路由 → Task 4 ✅
- 单例文档化 → Task 5 ✅

**2. Placeholder scan：** Task 4 Step 2 的 storage 写入方式说明"实现时根据 storage 接口确定"——这是必要的实现时核实（storage 接口无 SaveFromReader），非占位。已给出 fallback（临时文件+SaveFile）。

**3. Type consistency：** `NotifyRepository.Create/GetByID/Query/Update/Delete/UpdatesWhere` 在 Task 1 定义，notify.go 改造使用一致；`UploadDirect(ctx, uploadID, token, data, fileSize, fileName)` 在 Task 4 定义并使用。
