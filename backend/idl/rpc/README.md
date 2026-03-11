# idl/rpc/ - Kitex RPC 服务定义

此目录存放 Kitex RPC 服务的 IDL 定义文件。

## 文件说明

| 文件 | 用途 |
|------|------|
| `health.proto` | 健康检查/探活服务 |

## health.proto - 健康检查服务

提供 RPC 服务探活接口。

### 接口列表

| 接口 | 用途 |
|------|------|
| `Ping` | Ping/Pong 探活 |
| `Check` | 健康检查 |
| `Info` | 服务信息 |

### 使用示例

```go
// 客户端调用
resp, err := healthClient.Ping(ctx, &health.PingReq{})
// resp.Message = "pong"

resp, err := healthClient.Check(ctx, &health.HealthCheckReq{})
// resp.Healthy = true
// resp.Status = "serving"

resp, err := healthClient.Info(ctx, &health.ServiceInfoReq{})
// resp.Name = "my-service"
// resp.Version = "1.0.0"
```

## 代码生成

```bash
# 生成 RPC 代码
make gen-rpc IDL=rpc/health.proto
```

## 添加新服务

1. 在此目录创建新的 `.proto` 文件
2. 定义请求/响应消息和服务接口
3. 执行代码生成命令

```protobuf
syntax = "proto3";
package rpc.example;
option go_package = "github.com/zy84338719/fileCodeBox/gen/rpc/example";

message ExampleReq {
    string name = 1;
}

message ExampleResp {
    string message = 1;
}

service ExampleService {
    rpc Hello(ExampleReq) returns (ExampleResp);
}
```

## 注意事项

- 生成的代码在 `gen/rpc/` 目录，禁止手动修改
- 新增服务后需要在 `internal/transport/rpc/handler/` 中实现
