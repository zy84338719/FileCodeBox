# internal/repo/db/model/ - 数据库模型

此目录存放 GORM 数据库模型定义。

## 内容说明

- 数据库表映射结构体
- 字段定义和约束
- 模型关联关系
- 数据转换方法

## 示例

```go
type User struct {
    gorm.Model
    Username string `gorm:"uniqueIndex;size:50;not null"`
    Email    string `gorm:"uniqueIndex;size:100;not null"`
    Password string `gorm:"size:255;not null"`
    Nickname string `gorm:"size:50"`
    Avatar   string `gorm:"size:255"`
    Status   int8   `gorm:"default:1"`
}

func (u *User) TableName() string {
    return "users"
}

func (u *User) ToResp() *UserResp {
    return &UserResp{
        ID:       u.ID,
        Username: u.Username,
        // ...
    }
}
```

## 注意

- 模型应只包含数据结构定义
- 复杂查询逻辑应放在 `dao/` 中
- HTTP 请求/响应模型应放在 `gen/http/model/`
