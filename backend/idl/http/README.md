# idl/http/ - HTTP API 定义

此目录存放 Hz HTTP API 的 IDL 定义文件。

## 文件说明

| 文件 | 用途 |
|------|------|
| `health.proto` | 健康检查服务 |

## health.proto - 健康检查服务

提供 K8s 风格的健康检查接口。

### 接口列表

| 接口 | 方法 | 路径 | 用途 |
|------|------|------|------|
| `Health` | GET | `/health` | 基础健康检查 |
| `Readiness` | GET | `/ready` | K8s 就绪探针 |
| `Liveness` | GET | `/live` | K8s 存活探针 |
| `Version` | GET | `/version` | 版本信息 |
| `Ping` | GET | `/ping` | Ping/Pong |

### 响应示例

```json
// GET /health
{
    "status": "healthy",
    "timestamp": "2024-01-01T00:00:00Z"
}

// GET /ready
{
    "status": "ready",
    "database": true,
    "redis": true
}

// GET /version
{
    "name": "my-app",
    "version": "1.0.0",
    "build_time": "2024-01-01T00:00:00Z",
    "git_commit": "abc1234",
    "go_version": "go1.21"
}
```

## 代码生成

```bash
# 生成 HTTP 代码
make gen-http-new IDL=http/health.proto
make gen-http-update IDL=http/health.proto
```

## 添加新服务

1. 在此目录创建新的 `.proto` 文件
2. 引入 `api/api.proto` 使用 HTTP 注解
3. 定义请求/响应消息和服务接口
4. 执行代码生成命令

```protobuf
syntax = "proto3";
package http.example;
option go_package = "github.com/zy84338719/fileCodeBox/gen/http/model/example";

import "api/api.proto";

message ExampleReq {
    string name = 1 [(api.query) = "name"];
}

message ExampleResp {
    string message = 1 [(api.body) = "message"];
}

service ExampleService {
    rpc Hello(ExampleReq) returns (ExampleResp) {
        option (api.get) = "/api/v1/hello";
    }
}
```
