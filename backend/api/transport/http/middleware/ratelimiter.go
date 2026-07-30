package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"golang.org/x/time/rate"
)

// RateLimiterConfig 限流器配置
type RateLimiterConfig struct {
	// RequestsPerSecond 每秒允许的请求数
	RequestsPerSecond float64
	// BurstSize 突发流量大小
	BurstSize int
	// KeyGenerator 用于生成限流键的函数
	KeyGenerator func(ctx context.Context, c *app.RequestContext) string
	// Expiration 限流记录的过期时间
	Expiration time.Duration
}

// RateLimiter 限流器
type RateLimiter struct {
	limiterMap sync.Map // map[string]*rateLimiterEntry
	config     *RateLimiterConfig
}

type rateLimiterEntry struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// NewRateLimiter 创建新的限流器
func NewRateLimiter(config *RateLimiterConfig) *RateLimiter {
	if config == nil {
		config = &RateLimiterConfig{
			RequestsPerSecond: 10,
			BurstSize:         20,
			KeyGenerator:      IPKeyGenerator,
			Expiration:        time.Hour,
		}
	}

	// 设置默认的键生成器
	if config.KeyGenerator == nil {
		config.KeyGenerator = IPKeyGenerator
	}

	// 设置默认的过期时间
	if config.Expiration == 0 {
		config.Expiration = time.Hour
	}

	return &RateLimiter{
		config: config,
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(ctx context.Context, c *app.RequestContext) bool {
	key := rl.config.KeyGenerator(ctx, c)
	limiter, exists := rl.limiterMap.Load(key)

	if !exists {
		// 创建新的限流器
		newLimiter := &rateLimiterEntry{
			limiter:    rate.NewLimiter(rate.Limit(rl.config.RequestsPerSecond), rl.config.BurstSize),
			lastAccess: time.Now(),
		}
		rl.limiterMap.Store(key, newLimiter)
		limiter = newLimiter
	}

	entry := limiter.(*rateLimiterEntry)
	entry.lastAccess = time.Now()

	return entry.limiter.Allow()
}

// Cleanup 清理过期的限流记录
func (rl *RateLimiter) Cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			now := time.Now()
			rl.limiterMap.Range(func(key, value interface{}) bool {
				entry := value.(*rateLimiterEntry)
				if now.Sub(entry.lastAccess) > rl.config.Expiration {
					rl.limiterMap.Delete(key)
				}
				return true
			})
		}
	}()
}

// Middleware 返回限流中间件
func (rl *RateLimiter) Middleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if !rl.Allow(ctx, c) {
			c.Abort()
			c.JSON(http.StatusTooManyRequests, map[string]interface{}{
				"code":    http.StatusTooManyRequests,
				"message": "Too many requests, please try again later",
			})
			return
		}

		c.Next(ctx)
	}
}

// IPKeyGenerator 基于 IP 地址生成限流键
func IPKeyGenerator(ctx context.Context, c *app.RequestContext) string {
	return c.ClientIP()
}

// UserKeyGenerator 基于用户 ID 生成限流键
// 如果用户未认证，则使用 IP
func UserKeyGenerator(ctx context.Context, c *app.RequestContext) string {
	if userID := GetUserID(c); userID > 0 {
		return "user:" + string(rune(userID))
	}
	return "ip:" + c.ClientIP()
}

// PathKeyGenerator 基于 IP 和路径生成限流键
func PathKeyGenerator(ctx context.Context, c *app.RequestContext) string {
	path := string(c.URI().Path())
	return c.ClientIP() + ":" + path
}

// GlobalRateLimit 全局限流中间件（基于 IP）
func GlobalRateLimit(requestsPerSecond float64, burstSize int) app.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), burstSize)

	return func(ctx context.Context, c *app.RequestContext) {
		if !limiter.Allow() {
			c.Abort()
			c.JSON(http.StatusTooManyRequests, map[string]interface{}{
				"code":    http.StatusTooManyRequests,
				"message": "Too many requests, please try again later",
			})
			return
		}

		c.Next(ctx)
	}
}

// PerIPRateLimit 基于 IP 的限流中间件
func PerIPRateLimit(requestsPerSecond float64, burstSize int) app.HandlerFunc {
	config := &RateLimiterConfig{
		RequestsPerSecond: requestsPerSecond,
		BurstSize:         burstSize,
		KeyGenerator:      IPKeyGenerator,
		Expiration:        time.Hour,
	}
	limiter := NewRateLimiter(config)
	limiter.Cleanup()

	return limiter.Middleware()
}

// PerUserRateLimit 基于用户的限流中间件
func PerUserRateLimit(requestsPerSecond float64, burstSize int) app.HandlerFunc {
	config := &RateLimiterConfig{
		RequestsPerSecond: requestsPerSecond,
		BurstSize:         burstSize,
		KeyGenerator:      UserKeyGenerator,
		Expiration:        time.Hour,
	}
	limiter := NewRateLimiter(config)
	limiter.Cleanup()

	return limiter.Middleware()
}

// PerPathRateLimit 基于路径的限流中间件
func PerPathRateLimit(requestsPerSecond float64, burstSize int) app.HandlerFunc {
	config := &RateLimiterConfig{
		RequestsPerSecond: requestsPerSecond,
		BurstSize:         burstSize,
		KeyGenerator:      PathKeyGenerator,
		Expiration:        time.Hour,
	}
	limiter := NewRateLimiter(config)
	limiter.Cleanup()

	return limiter.Middleware()
}

// ConfigurableRateLimit 可配置的限流中间件
func ConfigurableRateLimit(config *RateLimiterConfig) app.HandlerFunc {
	limiter := NewRateLimiter(config)
	limiter.Cleanup()

	return limiter.Middleware()
}

// DefaultRateLimit 默认的限流中间件
// 每个 IP 每秒最多 10 个请求，突发 20 个
func DefaultRateLimit() app.HandlerFunc {
	return PerIPRateLimit(10, 20)
}

// StrictRateLimit 严格的限流中间件
// 每个 IP 每秒最多 5 个请求，突发 10 个
func StrictRateLimit() app.HandlerFunc {
	return PerIPRateLimit(5, 10)
}

// LooseRateLimit 宽松的限流中间件
// 每个 IP 每秒最多 100 个请求，突发 200 个
func LooseRateLimit() app.HandlerFunc {
	return PerIPRateLimit(100, 200)
}
