# FileCodeBox 分层配置系统设计文档

## 概述

FileCodeBox 的分层配置系统将原来的单一配置结构拆分为多个独立的配置层，每层采用最适合的存储策略，实现了更好的性能、可维护性和扩展性。

## 架构设计

### 配置分层

```
┌─────────────────────────────────────────────────────────────┐
│                   FileCodeBox 分层配置系统                    │
├─────────────────────────────────────────────────────────────┤
│ 📁 静态配置层 (Static Config)                               │
│   存储: static_configs 表                                   │
│   内容: 基础设置、主题配置                                   │
│   特点: 启动时加载，很少变更                                 │
├─────────────────────────────────────────────────────────────┤
│ 💾 系统配置层 (System Config)                               │
│   存储: system_configs 表 (JSON格式)                        │
│   内容: 存储配置、传输配置、数据库配置                       │
│   特点: 版本控制、校验和验证                                 │
├─────────────────────────────────────────────────────────────┤
│ 🔄 运行时配置层 (Runtime Config)                            │
│   存储: runtime_configs 表                                  │
│   内容: 用户系统、MCP服务配置                                │
│   特点: 支持用户级配置、优先级管理、过期时间                 │
├─────────────────────────────────────────────────────────────┤
│ 📋 业务配置层 (Business Config)                             │
│   存储: business_configs 表                                 │
│   内容: 通知配置、限流配置、模块配置                         │
│   特点: 按模块组织、支持时效性                               │
└─────────────────────────────────────────────────────────────┘
```

## 数据库表设计

### 1. static_configs 表 - 静态配置
```sql
CREATE TABLE static_configs (
    id          INTEGER PRIMARY KEY,
    config_key  VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT NOT NULL,
    category    VARCHAR(50),
    description VARCHAR(255),
    is_active   BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMP,
    updated_at  TIMESTAMP
);
```

**用途**: 基础设置、主题配置等很少变更的配置
**特点**: 文件级缓存，启动时全量加载

### 2. system_configs 表 - 系统配置
```sql
CREATE TABLE system_configs (
    id            INTEGER PRIMARY KEY,
    config_key    VARCHAR(100) UNIQUE NOT NULL,
    config_value  JSON NOT NULL,
    config_schema TEXT,
    version       INTEGER DEFAULT 1,
    environment   VARCHAR(20) DEFAULT 'production',
    is_encrypted  BOOLEAN DEFAULT FALSE,
    checksum      VARCHAR(64),
    created_at    TIMESTAMP,
    updated_at    TIMESTAMP
);
```

**用途**: 存储配置、传输配置、数据库配置
**特点**: JSON格式存储、版本控制、数据完整性校验

### 3. runtime_configs 表 - 运行时配置
```sql
CREATE TABLE runtime_configs (
    id           INTEGER PRIMARY KEY,
    config_key   VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT NOT NULL,
    data_type    VARCHAR(20),
    category     VARCHAR(50),
    user_id      INTEGER,
    is_global    BOOLEAN DEFAULT TRUE,
    priority     INTEGER DEFAULT 0,
    expires_at   TIMESTAMP,
    created_at   TIMESTAMP,
    updated_at   TIMESTAMP
);
```

**用途**: 用户系统、MCP服务等运行时可变配置
**特点**: 支持用户级配置、优先级管理、过期时间

### 4. business_configs 表 - 业务配置
```sql
CREATE TABLE business_configs (
    id           INTEGER PRIMARY KEY,
    config_key   VARCHAR(100) UNIQUE NOT NULL,
    config_value JSON NOT NULL,
    module       VARCHAR(50),
    sub_module   VARCHAR(50),
    is_enabled   BOOLEAN DEFAULT TRUE,
    valid_from   TIMESTAMP NOT NULL,
    valid_to     TIMESTAMP,
    created_at   TIMESTAMP,
    updated_at   TIMESTAMP
);
```

**用途**: 通知配置、限流配置、模块配置
**特点**: 按模块组织、支持时效性、功能开关

## 配置映射策略

### 配置键命名规范
```
<层级>.<模块>.<配置项>

示例:
- base.name              // 基础配置 - 应用名称
- theme.select           // 主题配置 - 选择的主题
- storage.config         // 存储配置 - 完整配置对象
- transfer.config        // 传输配置 - 完整配置对象
- user.system_enabled    // 用户配置 - 系统启用状态
- mcp.enabled           // MCP配置 - 服务启用状态
- notification.config    // 通知配置 - 完整配置对象
- ratelimit.config      // 限流配置 - 完整配置对象
```

### 存储类型分配
| 配置类型 | 存储表 | 存储格式 | 缓存策略 |
|---------|--------|----------|----------|
| 静态配置 | static_configs | 键值对 | 启动时加载，长期缓存 |
| 系统配置 | system_configs | JSON | 按需加载，中期缓存 |
| 运行时配置 | runtime_configs | 键值对/JSON | 实时加载，短期缓存 |
| 业务配置 | business_configs | JSON | 按模块加载，智能缓存 |

## 核心组件

### 1. ConfigStorageStrategy - 配置存储策略
负责不同类型配置的存储和检索操作：
- `GetConfig(key, result)` - 获取配置
- `SetConfig(key, value)` - 设置配置
- `ValidateConfig(key, value)` - 验证配置
- `ListConfigsByCategory(category)` - 按分类列出配置

### 2. LayeredConfigManager - 分层配置管理器
统一的配置管理接口：
- `LoadAllConfigs()` - 加载所有配置
- `SaveAllConfigs()` - 保存所有配置
- `ValidateAllConfigs()` - 验证所有配置
- `GetConfigByCategory(category)` - 按分类获取配置

### 3. 缓存系统
智能缓存策略：
- **分级缓存**: 不同类型配置采用不同缓存策略
- **过期管理**: 自动过期清理，减少内存占用
- **缓存统计**: 提供缓存使用情况监控

## 使用示例

### 创建分层配置管理器
```go
// 创建管理器
manager := config.NewLayeredConfigManager(db)

// 初始化配置表
err := manager.InitTables()

// 加载所有配置
err = manager.LoadAllConfigs()
```

### 访问配置
```go
// 基础配置
fmt.Println("应用名称:", manager.Base.Name)
fmt.Println("服务器地址:", manager.Base.GetAddress())

// 系统配置
fmt.Println("数据库类型:", manager.Database.Type)
fmt.Println("存储类型:", manager.Storage.Type)

// 运行时配置
fmt.Println("用户系统启用:", manager.User.IsUserSystemEnabled())
fmt.Println("MCP服务启用:", manager.MCP.IsMCPEnabled())

// 业务配置
fmt.Println("通知标题:", manager.Notification.Title)
fmt.Println("限流设置:", manager.RateLimit.UploadCount)
```

### 按分类管理配置
```go
// 获取静态配置
staticConfigs, err := manager.GetConfigByCategory(config.CategoryStatic)

// 获取运行时配置
runtimeConfigs, err := manager.GetConfigByCategory(config.CategoryRuntime)
```

### 缓存管理
```go
// 获取缓存统计
stats := manager.GetCacheStats()
fmt.Printf("缓存大小: %d\n", stats["cache_size"])

// 清除缓存
manager.ClearCache()
```

## 性能优化

### 1. 查询优化
- **精确查询**: 避免大JSON解析，直接查询需要的配置
- **索引优化**: 配置键、分类、模块等字段建立索引
- **批量操作**: 支持批量加载和保存操作

### 2. 缓存优化
- **分层缓存**: 静态配置长期缓存，运行时配置短期缓存
- **智能过期**: 根据配置变更频率自动调整过期时间
- **内存控制**: 限制缓存大小，防止内存泄漏

### 3. 存储优化
- **数据分离**: 不同类型配置分表存储
- **压缩存储**: 大配置对象支持压缩存储
- **版本控制**: 系统配置支持版本管理和回滚

## 迁移指南

### 从传统配置迁移
```go
// 1. 创建分层管理器
layeredManager := config.NewLayeredConfigManager(db)

// 2. 迁移现有配置
err := migrateFromTraditionalConfig(oldConfig, layeredManager)

// 3. 验证迁移结果
err = layeredManager.ValidateAllConfigs()

// 4. 保存新配置
err = layeredManager.SaveAllConfigs()
```

### 配置兼容性
- **向后兼容**: 保留原有Config结构作为适配器
- **平滑迁移**: 支持逐步迁移，新旧系统可并存
- **数据验证**: 迁移过程中验证数据完整性

## 扩展性

### 添加新配置类型
1. 在配置映射中定义新的配置键
2. 选择合适的存储策略
3. 实现配置验证逻辑
4. 添加到分层管理器中

### 自定义存储策略
```go
// 实现自定义存储策略
type CustomStorageStrategy struct {
    // 自定义字段
}

func (s *CustomStorageStrategy) GetConfig(key string, result interface{}) error {
    // 自定义获取逻辑
}

func (s *CustomStorageStrategy) SetConfig(key string, value interface{}) error {
    // 自定义设置逻辑
}
```

## 监控和调试

### 配置监控
- **访问统计**: 记录配置访问频率
- **性能监控**: 监控配置加载时间
- **错误追踪**: 记录配置错误和异常

### 调试工具
- **配置导出**: 支持导出所有配置用于调试
- **配置比较**: 比较不同环境的配置差异
- **配置历史**: 查看配置变更历史

## 总结

分层配置系统相比传统单一配置的优势：

### ✅ 性能提升
- **查询效率**: 避免大JSON解析，提升查询速度
- **内存优化**: 分层缓存，减少内存占用
- **扩展性**: 易于添加新配置类型

### ✅ 可维护性
- **代码组织**: 配置按功能分层，代码更清晰
- **数据隔离**: 不同类型配置独立存储，减少冲突
- **版本控制**: 系统配置支持版本管理

### ✅ 安全性
- **权限控制**: 支持用户级配置和权限管理
- **数据加密**: 敏感配置支持加密存储
- **审计追踪**: 配置变更可追踪

这个分层配置系统为FileCodeBox提供了更强大、更灵活的配置管理能力，支持未来的功能扩展和性能优化需求。
