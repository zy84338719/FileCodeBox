# internal/repo/external/ - 外部服务调用

此目录存放外部服务（第三方 API、微服务等）的调用封装。

## 适用场景

- 第三方 API 调用（支付、短信、邮件等）
- 内部微服务调用
- 外部数据源访问

## 目录结构建议

```
external/
├── payment/        # 支付服务
│   ├── client.go
│   └── types.go
├── sms/            # 短信服务
│   └── client.go
└── email/          # 邮件服务
    └── client.go
```

## 示例

```go
package payment

type Client struct {
    baseURL string
    apiKey  string
}

func NewClient(baseURL, apiKey string) *Client {
    return &Client{baseURL: baseURL, apiKey: apiKey}
}

func (c *Client) CreateOrder(ctx context.Context, req *CreateOrderReq) (*Order, error) {
    // HTTP 调用实现
}
```

## 最佳实践

- 统一错误处理
- 超时控制
- 重试机制
- 熔断降级
- 日志记录
