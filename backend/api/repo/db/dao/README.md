# internal/repo/db/dao/ - 数据访问对象

此目录存放数据访问对象（DAO），封装数据库 CRUD 操作。

## 命名规范

- 文件名：`{entity}.go`（如 `user.go`）
- 结构体：`{Entity}Repository`（如 `UserRepository`）

## 职责

- 封装单表 CRUD 操作
- 实现复杂查询
- 提供分页、排序等通用功能

## 示例

```go
type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository() *UserRepository {
    return &UserRepository{db: db.GetDB()}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
    var user model.User
    err := r.db.WithContext(ctx).First(&user, id).Error
    return &user, err
}

func (r *UserRepository) List(ctx context.Context, page, pageSize int) ([]*model.User, int64, error) {
    // 分页查询实现
}
```

## 最佳实践

- 使用 Context 传递超时和取消信号
- 返回业务友好的错误信息
- 复杂事务在 app 层处理
