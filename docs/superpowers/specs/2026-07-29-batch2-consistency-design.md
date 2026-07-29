# 第二批：一致性 + 分层设计

> 日期：2026-07-29
> 状态：已确认，待实现
> 范围：核心写操作加事务、notify 抽 DAO、全量补结构化日志、presign 死路由修复、全局单例文档化
> 这是"项目不足修复"三批计划的第二批，独立可交付、可回退

## 已确认决策

1. **notify**：完整抽 DAO（迁移 11 处 gorm 直连为 repo 方法）
2. **事务**：仅核心写操作（CreateShare/ShareText/DeleteFileByCode），物理删除放事务外
3. **日志**：全量补 zap 结构化日志，覆盖 15+ 处静默吞错
4. **presign**：修复死路由（注册 upload-direct 并实现 handler）
5. **全局单例**：维持现状，补注释文档化（不做 DI 重构）

---

## 一、核心写操作加事务

**问题**：`share/service.go` 的 CreateShare（:215）、ShareText（:145）、DeleteFileByCode（:299）多步 DB 写无事务，统计漂移、孤儿文件。

**修复**：
- `dao` 新增事务辅助：`WithTransaction(ctx, fn func(txRepo) error) error`，或在 service 层用 `db.GetDB().Transaction()`
- **CreateShare/ShareText**：用事务包裹"create fileCode + UpdateUserStats"。统计更新失败则回滚记录创建
- **DeleteFileByCode**：事务内"验证所有权 + 删 DB 记录 + 扣 storage 统计"；事务提交成功后才删物理文件（不可回滚，放事务外，失败记 warn 日志不阻断）

**事务实现方式**：由于 DAO 通过全局 `db.GetDB()` 获取连接，service 层直接用 `db.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {...})`，在闭包内用 `tx` 操作。需给 repo 方法增加接受 `*gorm.DB` 的变体，或在事务闭包内直接用 `tx.Model(&FileCode{})` 操作（统计更新逻辑简单，可直接内联）。

**简化方案（采用）**：统计更新（UpdateUserStats）走 userService 接口，无法注入 tx。因此事务仅包裹"fileCode 创建"单步（目前本就原子），统计更新保持 best-effort 但加日志。真正需要事务的是 **DeleteFileByCode** 的"删 DB + 扣统计"两步——这两步都用 fileCodeRepo，可用事务。

最终：
- CreateShare/ShareText：保持现状（单步 create 已原子），但 UpdateUserStats 失败补 logger.Warn
- DeleteFileByCode：用事务包裹"删 fileCode 记录 + 扣 storage 统计"，物理删除放事务外

---

## 二、notify 抽 DAO

**问题**：`notify/notify.go` 直接持有 `*gorm.DB`，11 处 gorm 调用绕过 DAO 层。

**修复**：
- 新建 `dao.NotifyRepository`，方法覆盖现有 11 处调用：Create/GetByID/GetByUserID/ListUnread/MarkRead/MarkAllRead/Delete/Count 等
- notify Service 改持有 `*dao.NotifyRepository`，`NewService(db)` 改为 `NewService()`（与其他 service 一致，内部 `dao.NewNotifyRepository()`）
- 更新 notify_test.go 适配（用 `db.SetDatabaseInstance` 注入测试库，与 share 测试模式一致）
- 更新 bootstrap 的 notify 初始化调用

---

## 三、全量补结构化日志

**问题**：15+ 处静默吞错（`_ = err`、注释"记日志"但未实现），核心业务零结构化日志。

**修复**：
- share/admin/anonymous/chunk/presign service 补 zap logger import
- 静默吞错处全部改为 `logger.Warn/Error`：
  - `share.UpdateUserStats` 失败（:168,243,290,335）
  - `share.DeleteFile` 物理删除失败（:321）
  - `share.RecordViewerAndNotify` 各失败点（:389,407）
  - `admin.CleanExpiredFiles` 物理删除失败（:331，已有 TODO）
  - `anonymous.CreateAnonymousShare` 回滚失败（:281）
- 删除 `cmd/server/main.go:23` 的 `fmt.Printf` 改 `logger.Fatal`

---

## 四、presign 死路由修复

**问题**：`presign.go:125` 生成 `/api/v1/presign/upload-direct/:uploadID` 的 UploadURL，但 router 未注册该路由，客户端 PUT 直传 404。

**修复**：
- 在 presign handler（`gen/http/handler/presign/`）注册 `PUT /upload-direct/:uploadID`
- 实现 handler：校验 presign token（签名验证）→ 从 body 读取文件 → 写入 storage → 更新 meta 为 complete
- 复用 `CheckUploadSize`/`IsBlockedExtension` 校验
- 由于 gen handler 不易新增路由，在 bootstrap 自定义路由层注册该端点（已有 customHandler 模式）

---

## 五、全局单例文档化

**问题**：DB/Config/Preview/Metrics 全局单例，对象生命周期不清。

**修复**（仅文档化，不重构）：
- `database.go` 的 `var DB` 补注释说明全局单例设计权衡
- `config.go` 的 `globalConfig` 补注释
- `preview/service.go` 的 `var svc` 补注释
- 注释说明：为避免构造期依赖循环采用全局单例，未来可迁移到 DI 容器

---

## 文件改动清单

| 文件 | 改动 |
|---|---|
| `backend/internal/repo/db/dao/notify.go` | 新建：NotifyRepository |
| `backend/internal/app/notify/notify.go` | 改：持有 repo 替代 db |
| `backend/internal/app/notify/notify_test.go` | 改：适配 repo 注入 |
| `backend/internal/app/share/service.go` | 改：DeleteFileByCode 加事务 + 补日志 |
| `backend/internal/app/admin/service.go` | 改：补日志 |
| `backend/internal/app/anonymous/anonymous.go` | 改：补日志 |
| `backend/internal/app/chunk/service.go` | 改：补日志 |
| `backend/internal/app/presign/presign.go` | 改：补日志 + upload-direct 支持 |
| `backend/cmd/server/bootstrap/bootstrap.go` | 改：notify 初始化 + presign 路由注册 |
| `backend/cmd/server/main.go` | 改：fmt.Printf → logger.Fatal |
| `backend/internal/repo/db/database.go` | 改：补注释 |
| `backend/internal/conf/config.go` | 改：补注释 |
| `backend/internal/preview/service.go` | 改：补注释 |

## 测试

- notify：repo 单测（Create/MarkRead/List 等用 sqlite）
- share：DeleteFileByCode 事务测试（模拟统计更新失败验证回滚）
- presign：upload-direct handler 集成测试（token 校验 + 文件写入）

## 风险

- **notify 抽 DAO**：改动 notify.go 全文件 + bootstrap + test，需仔细保持行为一致
- **事务**：DeleteFileByCode 事务需确保 repo 方法支持 tx 传入（或事务内直接用 gorm）
- **presign 新路由**：需在 bootstrap 注册，避免与 gen router 冲突
