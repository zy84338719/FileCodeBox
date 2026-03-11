# internal/repo/db/ - 数据库访问层

此目录存放关系型数据库相关代码。

## 目录结构

```
db/
├── database.go     # 数据库连接初始化
├── model/          # GORM 数据模型
└── dao/            # 数据访问对象（Data Access Object）
```

## 支持的数据库

- MySQL
- PostgreSQL
- SQLite（纯 Go 实现，无 CGO）

## 文件说明

### database.go
- 数据库连接初始化
- 连接池配置
- GetDB() 获取数据库实例

### model/
- GORM 模型定义
- 数据库表映射
- 模型转换方法

### dao/
- 数据访问对象
- CRUD 操作封装
- 复杂查询实现

## 使用示例

```go
// 获取数据库实例
db := db.GetDB()

// 使用 DAO
userRepo := dao.NewUserRepository()
user, err := userRepo.GetByID(ctx, 1)
```
