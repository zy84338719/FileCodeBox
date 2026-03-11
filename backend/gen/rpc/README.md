# gen/rpc/ - Kitex RPC 生成代码

此目录存放 Kitex 生成的 RPC 相关代码。

## 目录结构

```
rpc/
└── health/             # 健康检查服务
    ├── health.pb.go            # Protobuf 消息定义
    ├── healthservice/          # 服务端代码
    │   └── healthservice.go
    └── client.go               # 客户端代码
```

## 生成命令

```bash
# 生成 RPC 代码
make gen-rpc IDL=rpc/health.proto

# 或使用脚本
./scripts/gen.sh kitex idl/rpc/health.proto
```

## 使用示例

### 服务端

```go
import (
    "github.com/zy84338719/fileCodeBox/gen/rpc/health/healthservice"
)

// 实现服务接口
type HealthServiceImpl struct{}

func (s *HealthServiceImpl) Ping(ctx context.Context, req *health.PingReq) (*health.PingResp, error) {
    return &health.PingResp{Message: "pong"}, nil
}

func (s *HealthServiceImpl) Check(ctx context.Context, req *health.HealthCheckReq) (*health.HealthCheckResp, error) {
    return &health.HealthCheckResp{
        Healthy:   true,
        Status:    "serving",
        Timestamp: time.Now().Format(time.RFC3339),
    }, nil
}

// 启动服务
svr := healthservice.NewServer(new(HealthServiceImpl))
svr.Run()
```

### 客户端

```go
import (
    "github.com/zy84338719/fileCodeBox/gen/rpc/health"
)

// 创建客户端
cli, err := health.NewClient("health-service")

// 调用 RPC
resp, err := cli.Ping(ctx, &health.PingReq{})
// resp.Message = "pong"
```

## 注意

- **禁止手动修改**此目录下的文件
- 接口定义应在 `idl/rpc/` 目录中维护
- 重新生成时会覆盖现有文件
- 服务实现放在 `internal/transport/rpc/handler/`
