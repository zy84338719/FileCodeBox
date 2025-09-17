package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/config"
	"golang.org/x/time/rate"
)

// RateLimiter 限流器
type RateLimiter struct {
	limiters map[string]*rate.Limiter
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{limiters: make(map[string]*rate.Limiter)}
}

func (rl *RateLimiter) GetLimiter(key string, r rate.Limit, b int) *rate.Limiter {
	// Note: for simplicity we avoid mutex here; callers should ensure safe usage
	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(r, b)
		rl.limiters[key] = limiter
	}
	return limiter
}

func (rl *RateLimiter) Cleanup() {
	ticker := time.NewTicker(time.Hour)
	go func() {
		for range ticker.C {
			for key, limiter := range rl.limiters {
				if limiter.Allow() {
					delete(rl.limiters, key)
				}
			}
		}
	}()
}

var (
	uploadLimiter = NewRateLimiter()
	errorLimiter  = NewRateLimiter()
)

func init() {
	uploadLimiter.Cleanup()
	errorLimiter.Cleanup()
}

// RateLimit 限流中间件
func RateLimit(manager *config.ConfigManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		if c.Request.URL.Path == "/share/file/" || c.Request.URL.Path == "/share/text/" {
			limiter := uploadLimiter.GetLimiter(
				ip,
				rate.Every(time.Duration(manager.UploadMinute)*time.Minute/time.Duration(manager.UploadCount)),
				manager.UploadCount,
			)
			if !limiter.Allow() {
				common.TooManyRequestsResponse(c, "上传频率过快，请稍后再试")
				c.Abort()
				return
			}
		} else if c.Request.URL.Path == "/share/select/" && c.Request.Method == "GET" {
			limiter := errorLimiter.GetLimiter(
				ip,
				rate.Every(time.Duration(manager.ErrorMinute)*time.Minute/time.Duration(manager.ErrorCount)),
				manager.ErrorCount,
			)
			if !limiter.Allow() {
				common.TooManyRequestsResponse(c, "请求频率过快，请稍后再试")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
