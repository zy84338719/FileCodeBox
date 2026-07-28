// Package middleware 提供 Hertz 中间件：
//   - RateLimit: IP 维度 + 接口维度双层限流
//
// 基于 golang.org/x/time/rate 实现 token bucket。
// 配置存 runtime_configs 表，运行时可调。
package middleware

import (
	"context"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/zy84338719/fileCodeBox/backend/internal/pkg/errcode"
	"github.com/zy84338719/fileCodeBox/backend/internal/pkg/logger"
	"github.com/zy84338719/fileCodeBox/backend/internal/pkg/resp"
)

// RateLimitConfig 限流配置（运行时可调）
type RateLimitConfig struct {
	GlobalQPS    int  // 全局 IP 维度 QPS（每 IP）
	UploadQPS    int  // 上传接口 QPS
	DownloadQPS  int  // 下载接口 QPS
	LoginQPS     int  // 登录接口 QPS
	Burst        int  // 突发容量
	Enabled      bool // 总开关
	BlockSeconds int  // 阻断后封禁秒数
}

// DefaultRateLimitConfig 默认限流配置
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		GlobalQPS:    100,
		UploadQPS:    10,
		DownloadQPS:  50,
		LoginQPS:     5,
		Burst:        20,
		Enabled:      true,
		BlockSeconds: 60,
	}
}

// scope 限流维度
type scope string

const (
	scopeGlobal   scope = "global"
	scopeUpload   scope = "upload"
	scopeDownload scope = "download"
	scopeLogin    scope = "login"
)

// clientLimiter 单 IP 在某 scope 下的 limiter
type clientLimiter struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// RateLimiter 限流管理器
type RateLimiter struct {
	mu      sync.Mutex
	clients map[scope]map[string]*clientLimiter
	cfg     RateLimitConfig
	stopCh  chan struct{}
}

// NewRateLimiter 创建限流管理器
func NewRateLimiter(cfg RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		clients: map[scope]map[string]*clientLimiter{
			scopeGlobal:   {},
			scopeUpload:   {},
			scopeDownload: {},
			scopeLogin:    {},
		},
		cfg:    cfg,
		stopCh: make(chan struct{}),
	}
	// 后台定期清理过期 limiter（防止 map 无限增长）
	go rl.gcLoop()
	return rl
}

// UpdateConfig 热更新配置
func (rl *RateLimiter) UpdateConfig(cfg RateLimitConfig) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.cfg = cfg
	// 清空所有现有 limiter，让新配置生效
	for s := range rl.clients {
		rl.clients[s] = map[string]*clientLimiter{}
	}
}

// Stop 停止后台 goroutine
func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
}

func (rl *RateLimiter) gcLoop() {
	t := time.NewTicker(5 * time.Minute)
	defer t.Stop()
	for {
		select {
		case <-rl.stopCh:
			return
		case <-t.C:
			rl.gc()
		}
	}
}

func (rl *RateLimiter) gc() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	cutoff := time.Now().Add(-10 * time.Minute)
	for s := range rl.clients {
		for ip, cl := range rl.clients[s] {
			if cl.lastAccess.Before(cutoff) {
				delete(rl.clients[s], ip)
			}
		}
	}
}

// getLimiter 获取或创建某 IP 在某 scope 下的 limiter
func (rl *RateLimiter) getLimiter(s scope, ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	m, ok := rl.clients[s]
	if !ok {
		return nil
	}
	cl, ok := m[ip]
	if !ok {
		var qps int
		switch s {
		case scopeUpload:
			qps = rl.cfg.UploadQPS
		case scopeDownload:
			qps = rl.cfg.DownloadQPS
		case scopeLogin:
			qps = rl.cfg.LoginQPS
		default:
			qps = rl.cfg.GlobalQPS
		}
		if qps <= 0 {
			qps = 1
		}
		burst := rl.cfg.Burst
		if burst <= 0 {
			burst = qps
		}
		cl = &clientLimiter{
			limiter: rate.NewLimiter(rate.Limit(qps), burst),
		}
		m[ip] = cl
	}
	cl.lastAccess = time.Now()
	return cl.limiter
}

// allow 检查 IP 在某 scope 下是否被允许
func (rl *RateLimiter) allow(ctx context.Context, s scope, ip string) bool {
	if !rl.cfg.Enabled {
		return true
	}
	limiter := rl.getLimiter(s, ip)
	if limiter == nil {
		return true
	}
	return limiter.Allow()
}

// getClientIP 取客户端 IP（兼容 X-Forwarded-For / X-Real-IP）
func getClientIP(c *app.RequestContext) string {
	if xff := string(c.GetHeader("X-Forwarded-For")); xff != "" {
		// X-Forwarded-For 格式: client, proxy1, proxy2
		// 取第一个
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	if xri := string(c.GetHeader("X-Real-IP")); xri != "" {
		return xri
	}
	return c.RemoteAddr().String()
}

// Middleware 返回 Hertz 限流中间件
//   - scope: 限流维度（global / upload / download / login）
func (rl *RateLimiter) Middleware(s scope) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		if !rl.cfg.Enabled {
			ctx.Next(c)
			return
		}
		ip := getClientIP(ctx)
		if !rl.allow(c, s, ip) {
			if logger.Logger != nil {
				logger.Logger.Warn("[ratelimit] blocked",
					zap.String("scope", string(s)),
					zap.String("ip", ip))
			}
			resp.NewErrorByCode(ctx, errcode.CodeRateLimit)
			ctx.Abort()
			return
		}
		ctx.Next(c)
	}
}

// GlobalMiddleware 全局限流中间件（用于所有接口）
func (rl *RateLimiter) GlobalMiddleware() app.HandlerFunc {
	return rl.Middleware(scopeGlobal)
}

// UploadMiddleware 上传接口限流
func (rl *RateLimiter) UploadMiddleware() app.HandlerFunc {
	return rl.Middleware(scopeUpload)
}

// DownloadMiddleware 下载接口限流
func (rl *RateLimiter) DownloadMiddleware() app.HandlerFunc {
	return rl.Middleware(scopeDownload)
}

// LoginMiddleware 登录接口限流
func (rl *RateLimiter) LoginMiddleware() app.HandlerFunc {
	return rl.Middleware(scopeLogin)
}

// DefaultRateLimiter 全局默认限流器（单例）
var defaultLimiter *RateLimiter

// InitDefaultRateLimiter 初始化默认限流器
func InitDefaultRateLimiter(cfg RateLimitConfig) *RateLimiter {
	defaultLimiter = NewRateLimiter(cfg)
	return defaultLimiter
}

// GetDefaultRateLimiter 获取默认限流器
func GetDefaultRateLimiter() *RateLimiter {
	if defaultLimiter == nil {
		defaultLimiter = NewRateLimiter(DefaultRateLimitConfig())
	}
	return defaultLimiter
}
