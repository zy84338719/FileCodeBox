# idl/ - 接口定义语言

此目录存放 IDL（Interface Definition Language）文件。

## 目录结构

```
idl/
├── api/                # Hz HTTP 注解定义
│   └── api.proto       # HTTP 方法注解（get/post/put/delete）
├── http/               # HTTP API 定义
│   └── health.proto    # HTTP 健康检查服务
└── rpc/                # Kitex RPC 服务定义
    └── health.proto    # RPC 探活服务
```

## HTTP IDL (Hz)

用于定义 HTTP API 接口，支持 RESTful 风格。

```protobuf
// idl/http/health.proto
service HealthService {
    rpc Health(HealthCheckReq) returns (HealthCheckResp) {
        option (api.get) = "/health";
    }
}
```

## RPC IDL (Kitex)

用于定义微服务间的 RPC 接口。

```protobuf
// idl/rpc/health.proto
service HealthService {
    rpc Ping(PingReq) returns (PingResp);
    rpc Check(HealthCheckReq) returns (HealthCheckResp);
}
```

## 代码生成

```bash
# 生成 HTTP 代码
make gen-http-new IDL=http/health.proto
make gen-http-update IDL=http/health.proto

# 生成 RPC 代码
make gen-rpc IDL=rpc/health.proto
```

## 注意

- HTTP IDL 生成到 `gen/http/`
- RPC IDL 生成到 `gen/rpc/`
- `api/api.proto` 定义 HTTP 注解，不需要单独生成
