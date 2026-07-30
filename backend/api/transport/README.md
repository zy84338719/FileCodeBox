# internal/transport/ - 传输层

此目录存放协议适配代码，处理不同传输协议的请求/响应转换。

## 目录结构

```
transport/
├── http/           # HTTP 协议适配
│   ├── handler/    # HTTP 请求处理器实现
│   └── middleware/ # HTTP 中间件
└── rpc/            # RPC 协议适配（如需要）
    └── handler/    # RPC 请求处理器实现
```

## 职责

- 协议相关的参数解析和验证
- 调用 `internal/app/` 中的服务层
- 响应数据的组装和格式化
- 中间件实现（认证、日志、限流等）

## 与 gen/http/handler 的区别

- `gen/http/handler/`: Hz 生成的骨架代码，定义接口入口
- `internal/transport/http/handler/`: 手写的 handler 实现（复杂场景）
- 简单场景可直接在 gen/http/handler 中调用 app 层

## 依赖规则

- 可以依赖：`internal/app/`、`internal/pkg/`
- 不应依赖：`internal/repo/`（通过 app 层间接访问）
