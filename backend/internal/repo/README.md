# internal/repo/ - 数据访问层（Repository）

此目录存放数据访问相关代码，实现 Repository 模式。

## 目录结构

```
repo/
├── db/             # 关系型数据库访问
│   ├── database.go # 数据库初始化
│   ├── model/      # 数据库模型（GORM）
│   └── dao/        # 数据访问对象
├── redis/          # Redis 缓存访问
└── external/       # 外部服务调用
```

## 职责

- 封装所有数据访问逻辑
- 提供统一的数据操作接口
- 隐藏底层存储实现细节
- 实现缓存策略

## 依赖规则

- 可以依赖：`internal/pkg/`、`internal/conf/`
- 不应依赖：`internal/app/`、`internal/transport/`
- 应被依赖于：`internal/app/`

## 设计原则

- 每个聚合根对应一个 Repository
- Repository 返回领域模型，不返回数据库模型
- 事务控制在 app 层或 repo 层，保持一致性
