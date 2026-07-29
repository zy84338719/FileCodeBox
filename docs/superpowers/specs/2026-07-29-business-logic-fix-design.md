# FileCodeBox 业务逻辑修复设计

> 日期：2026-07-29
> 状态：已确认，待实现
> 范围：修复取件/分享链路中的安全、数据一致性、业务流程、工程四大类共 13 项问题

---

## 一、背景与问题清单

通读核心业务链路（分享生成 → 取件下载 → 过期清理 → 权限/安全）后，发现以下问题（详见前序对话）：

| # | 类别 | 问题 |
|---|---|---|
| 1 | 🔴安全 | 取件密码明文存储（`"sha256:"+明文`） |
| 2 | 🔴安全 | 密码校验仅检查非空，`RequireAuth` 形同虚设 |
| 3 | 🔴安全 | 分享码用 `math/rand`+时间种子，可预测、无冲突重试 |
| 4 | 🔴一致性 | 过期清理只删 DB，物理文件永久泄漏（孤儿文件） |
| 5 | 🔴一致性 | 次数扣减非原子（读改写），并发取件超卖 |
| 6 | 🟠流程 | `ExpiredCount` 语义混乱（0/-1/>0 三态错位） |
| 7 | 🟠流程 | 匿名取件 `ShareCode` 用 `file_name` 占位，与 `file_codes` 表脱节 |
| 8 | 🟠流程 | 匿名取件硬编码 24h 过期、密码/限流参数丢失 |
| 9 | 🟠流程 | DB 与 Redis 双重计数互不同步 |
| 10 | 🟠流程 | `RecordViewerAndNotify` 用空 IP 跑通知 |
| 11 | 🟡工程 | 过期清理无定时任务，只能 admin 手动触发 |
| 12 | 🟡工程 | 全局单例 + 延迟初始化 |
| 13 | 🟡工程 | `FileCode` 模型职责过多 |

## 二、已确认的关键决策

1. **打通匿名取件与 share 体系**：取件码只做 `pickup_code → share_code` 映射，真实状态走 `file_codes`
2. **DB 为唯一计数真相源**：Redis 不计数，扣减用原子 SQL
3. **过期清理：定时 ticker + 懒清理双保险**
4. **密码哈希：bcrypt**（cost=10）

## 三、详细设计

### 3.1 安全修复

#### 3.1.1 密码哈希 → bcrypt

新建 `internal/pkg/utils/password.go`：

```go
func HashPassword(pw string) (string, error)        // bcrypt cost=10
func CheckPassword(hash, pw string) bool             // CompareHashAndPassword
```

- 空密码 → 返回空 hash（不哈希），校验时空 hash 视为"无密码"
- `anonymous.go` 的 `defaultPasswordHash` 删除，改用 `utils.HashPassword`
- `anonymous.Retrieve` 密码校验改用 `utils.CheckPassword`
- share 侧 `GetFileWithUsage` 补齐真实密码校验（替换 TODO）

#### 3.1.2 Code 生成 → crypto/rand + 冲突重试

- `share/service.go:GenerateCode` 改用 `crypto/rand`（复用 anonymous.go 的 `randomCode` 风格，但字符表和长度保持 8 位兼容）
- `CreateShare`/`ShareText` 写库遇 `unique` 冲突（`errors.Is(err, gorm.ErrDuplicatedKey)`）时重新生成 code 重试，最多 5 次；超过则返回错误

### 3.2 数据一致性修复

#### 3.2.1 次数扣减原子化

`dao/filecode.go` 新增：

```go
// DecrementExpiredCount 原子扣减剩余次数，返回扣减后是否成功
// expired_count: -1=无限(只 +used_count), 0=已耗尽(拒绝), >0=剩余(扣减)
func (r *FileCodeRepository) DecrementExpiredCount(ctx, code string) (ok bool, err error)
```

SQL（GORM 表达式）：
```go
db.Model(&FileCode{}).Where("code = ?", code).
  Where("expired_count = -1 OR expired_count > 0").
  UpdateColumns(map[string]interface{}{
    "expired_count": gorm.Expr("CASE WHEN expired_count > 0 THEN expired_count - 1 ELSE expired_count END"),
    "used_count":    gorm.Expr("used_count + 1"),
  })
```
受影响行数=0 → 已耗尽，返回 `ok=false`。

- share service `UpdateFileUsage` 改调此方法
- anonymous `IncrementCount`（Redis）保留方法签名但改为调 share 的 DB 扣减（通过接口注入），保证调用方不变
- **统一 `ExpiredCount` 语义**：`-1=无限, 0=已耗尽, >0=剩余`，在 model 注释中固化

#### 3.2.2 过期文件清理（删物理文件 + 定时任务）

- `admin/service.go:CleanExpiredFiles` 接收 storage 注入；遍历过期文件时对有 `FilePath` 的记录调 `storage.DeleteFile`
- 物理删除失败 → 记日志（warn），不阻断 DB 删除；DB 删除失败 → 记日志，不阻断（下轮重试）
- `bootstrap` 启动一个 goroutine + `time.Ticker`（默认 1h，配置项 `storage.cleanup_interval`），调 `CleanExpiredFiles`
- **懒清理**：`share.GetFileByCode` 发现过期时，启动一个 goroutine 异步删除该单条（DB+物理），不阻塞当前请求（当前请求返回 expired 错误）

### 3.3 业务流程修复（打通匿名取件）

#### 3.3.1 CodeMeta 精简

`anonymous.CodeMeta` 字段调整：

| 字段 | 保留？ | 说明 |
|---|---|---|
| `ShareCode` | ✅ 必填 | 真实 file_code（不再用 file_name 占位） |
| `FileName` | ✅ | 展示用 |
| `FileSize` | ✅ | 展示用 |
| `ContentType` | ✅ | 展示用 |
| `RequireAuth` | ✅ | 展示用（是否需要密码） |
| `PasswordHash` | ❌ 删 | 以 DB 为准 |
| `ExpireAt` | ❌ 删 | 以 DB expired_at 为准 |
| `MaxPickupCount` | ❌ 删 | 以 DB expired_count 为准 |

Redis meta 格式简化为 `share_code|file_name|file_size|content_type|require_auth`。

#### 3.3.2 Retrieve 查 DB

`anonymous.Retrieve`：
1. 从 Redis 取 `share_code`（仅映射）
2. 用 `share_code` 查 `file_codes`：校验存在、未软删、未过期（时间+次数）
3. 密码校验：用 DB 记录的密码（需新增 `PasswordHash` 字段到 FileCode）做 bcrypt 校验
4. 扣减次数：调 `dao.DecrementExpiredCount`
5. 返回 meta（文件名等展示信息）

> **FileCode 新增 `PasswordHash string` 字段**（`size:255`，默认空）。`RequireAuth` 保留为"是否需要密码"的开关，`PasswordHash` 存实际哈希。

#### 3.3.3 上传链路透传

`gen/http/handler/share_anonymous/share_anonymous_service.go:GenerateCode`：
- 去掉硬编码 `ExpireAt = 24h`、空 `PasswordHash`
- 上传请求必须携带已创建的 share 信息（share_code、expire、password）；handler 负责调 `share.CreateShare` 拿到真实 code 后再调 `anonymous.GenerateCode`
- 密码在上传时即 bcrypt 哈希后存入 DB

> 说明：当前 `GenerateCode` handler 是 IDL 生成壳。实际"上传+生成取件码"的完整调用方在上传 handler（需在实现阶段定位）。若上传 handler 当前未走匿名取件码，则补一个端到端路径：上传 → share.CreateShare → anonymous.GenerateCode。

#### 3.3.4 viewer IP 修复

- 取件 handler 从 `c.ClientIP()` 取 IP
- `GetFileWithUsage`/`RecordViewerAndNotify` 接收真实 IP 参数，删除空串占位
- 通知去重：`RecordViewerAndNotify` 用 `LastNotifiedAt`，若距上次通知 <5 分钟则跳过

### 3.4 工程改进

- **#11 定时任务**：见 3.2.2
- **#12 单例**：维持现状，补注释说明是为避免循环依赖的权衡（不做大重构）
- **#13 模型职责**：维持现状，仅在注释中标注字段分组（上传/分片/viewer/通知），不做拆分（超出本次范围）

## 四、文件改动清单

| 文件 | 改动类型 | 内容 |
|---|---|---|
| `internal/repo/db/model/filecode.go` | 改 | 新增 `PasswordHash` 字段；统一 ExpiredCount 语义注释 |
| `internal/pkg/utils/password.go` | 新建 | bcrypt 封装 |
| `internal/pkg/utils/password_test.go` | 新建 | 单测 |
| `internal/repo/db/dao/filecode.go` | 改 | 新增 `DecrementExpiredCount` |
| `internal/app/share/service.go` | 改 | GenerateCode→crypto/rand+重试；UpdateFileUsage→原子扣减；密码校验；viewer IP；通知去重 |
| `internal/app/share/service_test.go` | 改 | 更新测试 |
| `internal/app/anonymous/anonymous.go` | 改 | 移除自实现 hash，改 bcrypt；CodeMeta 精简；Retrieve 查 DB |
| `internal/app/anonymous/anonymous_test.go` | 改 | 更新测试 |
| `internal/app/admin/service.go` | 改 | CleanExpiredFiles 删物理文件（注入 storage） |
| `cmd/server/bootstrap/bootstrap.go` | 改 | 启动定时清理 ticker；注入 storage 到 admin；anonymous 注入 share repo |
| `gen/http/handler/share_anonymous/share_anonymous_service.go` | 改 | 透传真实 share_code/expire/password；viewer IP |

## 五、测试策略

- `password_test.go`：哈希非空、校验正确/错误密码、空密码行为
- share service：Code 生成唯一性（并发）、原子扣减至耗尽返回 ok=false、密码校验正/误、通知去重
- anonymous：Retrieve 走 DB、密码错误、次数耗尽、过期
- 过期清理：mock storage 验证物理文件被删、DB 记录被删
- viewer：真实 IP 被记录

## 六、风险与回退

- **DB schema 变更**：新增 `PasswordHash` 字段，靠 AutoMigrate 自动加列（已有迁移机制），无破坏性
- **Code 长度**：保持 8 位不变，不影响存量数据
- **匿名取件 Redis meta 格式变更**：存量 key 仍按旧格式 TTL 过期，新 key 用新格式；过渡期 Retrieve 对旧格式做兼容解析（解析失败视为 code not found）
- **回退**：所有改动集中在上述文件，git revert 即可
