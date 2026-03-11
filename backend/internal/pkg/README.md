# internal/pkg/ - 项目内部工具库

此目录存放项目专用的工具函数和通用组件。

## 目录结构

```
pkg/
├── errors/     # 错误码定义和错误处理
├── logger/     # 日志工具封装
└── resp/       # HTTP 响应工具
```

## 与根目录 pkg/ 的区别

| 目录 | 用途 | 可见性 |
|------|------|--------|
| `pkg/` | 跨项目可复用的库 | 外部项目可导入 |
| `internal/pkg/` | 项目专用工具 | 仅本项目可用 |

## 现有组件

### errors/
- 业务错误码定义
- AppError 错误结构
- 错误消息映射

### logger/
- Zap 日志封装
- 日志级别、文件输出配置
- 便捷日志方法

### resp/
- HTTP 响应工具函数
- Success/Error/Page 等标准响应
- 错误码自动映射

## 添加新工具

1. 在 `internal/pkg/` 下创建新目录
2. 实现工具函数
3. 在需要的地方导入使用

```go
import "github.com/zy84338719/fileCodeBox/backend/internal/pkg/yourpkg"
```
