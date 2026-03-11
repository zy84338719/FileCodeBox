# internal/conf/ - 配置代码

此目录存放配置相关的 Go 代码，包括配置结构体定义和加载逻辑。

## 文件说明

- `config.go` - 配置结构体定义和加载函数

## 与 configs/ 的区别

| 目录 | 内容 | 说明 |
|------|------|------|
| `internal/conf/` | Go 代码 | 配置结构体、加载逻辑 |
| `configs/` | YAML/TOML/JSON | 配置文件 |

## 配置结构

```go
type AppConfig struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    Log      LogConfig
    App      AppInfo
}

type ServerConfig struct {
    Host string
    Port int
}
// ...
```

## 使用方式

```go
// 初始化配置
conf.Init("configs/config.yaml")

// 使用全局配置
cfg := conf.GlobalConfig
addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
```

## 注意

- 配置文件放在 `configs/` 目录
- 敏感信息（密码等）应通过环境变量覆盖
- 支持默认配置 InitWithDefault()
