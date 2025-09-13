# DAO 层迁移文档

## 概述

我们已经将所有数据库操作逻辑迁移到了 `internal/dao` 文件夹中，实现了数据访问层的统一管理。这种架构提供了更好的代码组织和维护性。

## 创建的 DAO 文件

### 1. `/internal/dao/dao_manager.go`
- **作用**: DAO 管理器，统一管理所有 DAO 实例
- **核心功能**:
  - 统一创建和管理所有 DAO 实例
  - 提供数据库连接和事务管理
  - 作为服务层与数据层的桥梁

### 2. `/internal/dao/file_code.go`
- **作用**: 文件代码相关的数据库操作
- **主要方法**:
  - `Create()` - 创建文件记录
  - `GetByID()`, `GetByCode()` - 查询文件
  - `Update()`, `UpdateColumns()` - 更新文件
  - `Delete()`, `DeleteByFileCode()` - 删除文件
  - `Count()`, `CountToday()`, `CountActive()` - 统计方法
  - `List()` - 分页查询
  - `GetExpiredFiles()` - 获取过期文件
  - `GetByHash()` - 根据哈希查找重复文件

### 3. `/internal/dao/user.go`
- **作用**: 用户相关的数据库操作
- **主要方法**:
  - `Create()` - 创建用户
  - `GetByID()`, `GetByUsername()`, `GetByEmail()` - 查询用户
  - `GetByUsernameOrEmail()` - 登录查询
  - `Update()`, `UpdateColumns()` - 更新用户
  - `UpdatePassword()`, `UpdateStatus()` - 专项更新
  - `CheckExists()`, `CheckEmailExists()` - 存在性检查
  - `Count()`, `CountActive()`, `CountTodayRegistrations()` - 统计
  - `List()` - 分页查询用户列表

### 4. `/internal/dao/user_session.go`
- **作用**: 用户会话相关的数据库操作
- **主要方法**:
  - `Create()` - 创建会话
  - `GetBySessionID()` - 根据会话ID查询
  - `CountActiveSessionsByUserID()` - 统计用户活跃会话
  - `GetOldestSessionByUserID()` - 获取最老会话
  - `UpdateIsActive()` - 更新会话状态
  - `DeactivateUserSessions()` - 停用用户所有会话
  - `CleanExpiredSessions()` - 清理过期会话

### 5. `/internal/dao/chunk.go`
- **作用**: 分片上传相关的数据库操作
- **主要方法**:
  - `Create()` - 创建分片记录
  - `GetByHash()`, `GetByUploadID()` - 查询分片信息
  - `GetChunkByIndex()` - 根据索引获取分片
  - `UpdateUploadProgress()` - 更新上传进度
  - `UpdateChunkCompleted()` - 标记分片完成
  - `CountCompletedChunks()` - 统计完成的分片
  - `DeleteByUploadID()` - 删除上传相关记录

### 6. `/internal/dao/key_value.go`
- **作用**: 配置键值对的数据库操作
- **主要方法**:
  - `Create()`, `Update()` - 基础CRUD
  - `GetByKey()`, `GetByKeys()` - 查询键值对
  - `SetValue()` - 设置键值（存在则更新，不存在则创建）
  - `BatchSet()` - 批量设置
  - `Search()` - 搜索键值对
  - `Count()` - 统计数量

### 7. `/internal/dao/admin.go`
- **作用**: 管理员相关的数据库操作
- **主要方法**:
  - `GetDB()` - 获取数据库连接（兼容现有代码）
  - `BeginTransaction()` - 开始事务

## 使用示例

### 在服务层中使用 DAO

```go
// 服务结构体中添加 DAO 管理器
type AdminService struct {
    db             *gorm.DB
    config         *config.Config
    storageManager *storage.StorageManager
    daoManager     *dao.DAOManager  // 新增
}

// 初始化服务时创建 DAO 管理器
func NewAdminService(db *gorm.DB, config *config.Config, storageManager *storage.StorageManager) *AdminService {
    return &AdminService{
        db:             db,
        config:         config,
        storageManager: storageManager,
        daoManager:     dao.NewDAOManager(db),  // 新增
    }
}

// 使用 DAO 方法替代直接的数据库操作
func (s *AdminService) GetStats() (map[string]interface{}, error) {
    stats := make(map[string]interface{})

    // 旧方式：s.db.Model(&models.FileCode{}).Count(&totalFiles)
    // 新方式：
    totalFiles, err := s.daoManager.FileCode.Count()
    if err != nil {
        return nil, err
    }
    stats["total_files"] = totalFiles

    // ... 其他统计
    return stats, nil
}
```

## 迁移的好处

### 1. **代码组织更清晰**
- 数据访问逻辑集中在 DAO 层
- 服务层专注于业务逻辑
- 更好的分层架构

### 2. **可维护性提升**
- 数据库操作统一管理
- 减少重复代码
- 更容易进行单元测试

### 3. **扩展性更好**
- 可以轻松添加新的数据操作方法
- 支持复杂查询的封装
- 便于缓存层的集成

### 4. **错误处理更统一**
- 统一的错误处理策略
- 更好的错误信息返回
- 便于日志记录和监控

## 后续工作建议

### 1. **完成其他服务的迁移**
- 继续将 `share.go`, `chunk.go`, `user.go` 等服务迁移到使用 DAO
- 逐步移除服务层中的直接数据库操作

### 2. **添加事务支持**
```go
// 在 DAO 中支持事务操作
func (s *UserService) DeleteUserWithFiles(userID uint) error {
    return s.daoManager.GetDB().Transaction(func(tx *gorm.DB) error {
        // 使用事务删除用户相关数据
        if err := s.daoManager.FileCode.DeleteByUserID(tx, userID); err != nil {
            return err
        }
        if err := s.daoManager.UserSession.DeleteByUserID(tx, userID); err != nil {
            return err
        }
        return s.daoManager.User.Delete(tx, user)
    })
}
```

### 3. **添加缓存层**
- 在 DAO 层添加 Redis 缓存支持
- 对热点数据进行缓存优化

### 4. **性能优化**
- 添加数据库查询性能监控
- 优化复杂查询
- 添加索引建议

## 测试验证

当前代码已经通过编译测试，说明基础架构正确。建议进行以下测试：

1. **单元测试**: 为每个 DAO 方法编写单元测试
2. **集成测试**: 测试 DAO 与服务层的集成
3. **性能测试**: 验证迁移后的性能表现

这个 DAO 层的迁移为 FileCodeBox 项目提供了更好的架构基础，有利于后续的维护和扩展。
