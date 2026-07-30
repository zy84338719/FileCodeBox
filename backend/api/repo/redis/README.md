# internal/repo/redis/ - Redis 缓存访问

此目录存放 Redis 相关代码。

## 文件说明

- `redis.go` - Redis 连接初始化和通用操作封装

## 支持的操作

- 字符串：Set, Get, Del
- 哈希：HSet, HGet, HGetAll, HDel
- 列表：LPush, RPush, LPop, RPop, LRange
- 集合：SAdd, SMembers, SRem
- 有序集合：ZAdd, ZRange, ZRem
- 通用：Exists, Expire, TTL, Incr, Decr

## 使用示例

```go
// 初始化
redis.Init(&conf.RedisConfig)

// 基本操作
redis.Set(ctx, "key", "value", time.Hour)
value, err := redis.Get(ctx, "key")

// 哈希操作
redis.HSet(ctx, "user:1", "name", "Alice", "age", "25")
data, err := redis.HGetAll(ctx, "user:1")
```

## 缓存策略建议

- 热点数据缓存
- 会话存储
- 分布式锁
- 消息队列
- 限流计数
