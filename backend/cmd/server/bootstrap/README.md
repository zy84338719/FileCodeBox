# cmd/server/bootstrap/ - 服务初始化

此目录存放服务启动时的初始化代码。

## 文件说明

- `bootstrap.go` - 初始化入口，协调各组件的初始化顺序

## 初始化顺序

1. **配置** - 加载配置文件
2. **日志** - 初始化日志系统
3. **数据库** - 建立数据库连接
4. **缓存** - 建立 Redis 连接
5. **服务器** - 创建 HTTP 服务器并注册路由

## 使用方式

```go
// main.go
func main() {
    h, err := bootstrap.Bootstrap()
    if err != nil {
        log.Fatal(err)
    }
    defer bootstrap.Cleanup()

    h.Spin()
}
```

## 添加新组件初始化

```go
// bootstrap.go
func Bootstrap() (*server.Hertz, error) {
    // ... 现有初始化

    // 添加新组件
    if err := initNewComponent(); err != nil {
        return nil, err
    }

    // ...
}
```

## 清理资源

```go
func Cleanup() {
    logger.Sync()
    _ = db.Close()
    _ = redis.Close()
    // 添加新组件的清理
}
```
