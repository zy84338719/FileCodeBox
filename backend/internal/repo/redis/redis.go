package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zy84338719/fileCodeBox/backend/internal/conf"
)

var Client *redis.Client

func Init(cfg *conf.RedisConfig) error {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect redis: %w", err)
	}

	return nil
}

func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

func GetClient() *redis.Client {
	return Client
}

func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return Client.Set(ctx, key, value, expiration).Err()
}

func Get(ctx context.Context, key string) (string, error) {
	return Client.Get(ctx, key).Result()
}

func Del(ctx context.Context, keys ...string) error {
	return Client.Del(ctx, keys...).Err()
}

func Exists(ctx context.Context, keys ...string) (int64, error) {
	return Client.Exists(ctx, keys...).Result()
}

func Expire(ctx context.Context, key string, expiration time.Duration) error {
	return Client.Expire(ctx, key, expiration).Err()
}

func TTL(ctx context.Context, key string) (time.Duration, error) {
	return Client.TTL(ctx, key).Result()
}

func Incr(ctx context.Context, key string) (int64, error) {
	return Client.Incr(ctx, key).Result()
}

func Decr(ctx context.Context, key string) (int64, error) {
	return Client.Decr(ctx, key).Result()
}

func HSet(ctx context.Context, key string, values ...interface{}) error {
	return Client.HSet(ctx, key, values...).Err()
}

func HGet(ctx context.Context, key, field string) (string, error) {
	return Client.HGet(ctx, key, field).Result()
}

func HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return Client.HGetAll(ctx, key).Result()
}

func HDel(ctx context.Context, key string, fields ...string) error {
	return Client.HDel(ctx, key, fields...).Err()
}

func LPush(ctx context.Context, key string, values ...interface{}) error {
	return Client.LPush(ctx, key, values...).Err()
}

func RPush(ctx context.Context, key string, values ...interface{}) error {
	return Client.RPush(ctx, key, values...).Err()
}

func LPop(ctx context.Context, key string) (string, error) {
	return Client.LPop(ctx, key).Result()
}

func RPop(ctx context.Context, key string) (string, error) {
	return Client.RPop(ctx, key).Result()
}

func LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return Client.LRange(ctx, key, start, stop).Result()
}

func SAdd(ctx context.Context, key string, members ...interface{}) error {
	return Client.SAdd(ctx, key, members...).Err()
}

func SMembers(ctx context.Context, key string) ([]string, error) {
	return Client.SMembers(ctx, key).Result()
}

func SRem(ctx context.Context, key string, members ...interface{}) error {
	return Client.SRem(ctx, key, members...).Err()
}

func ZAdd(ctx context.Context, key string, score float64, member string) error {
	return Client.ZAdd(ctx, key, redis.Z{Score: score, Member: member}).Err()
}

func ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return Client.ZRange(ctx, key, start, stop).Result()
}

func ZRem(ctx context.Context, key string, members ...string) error {
	return Client.ZRem(ctx, key, members).Err()
}
