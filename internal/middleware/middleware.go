package middleware

import (
	"github.com/zy84338719/filecodebox/internal/config"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/time/rate"
)

// CORS 中间件
func CORS() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowCredentials = true
	config.AllowHeaders = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	return cors.New(config)
}

// RateLimiter 限流器
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
	}
}

func (rl *RateLimiter) GetLimiter(key string, r rate.Limit, b int) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(r, b)
		rl.limiters[key] = limiter
	}

	return limiter
}

// 清理过期的限流器
func (rl *RateLimiter) Cleanup() {
	ticker := time.NewTicker(time.Hour)
	go func() {
		for range ticker.C {
			rl.mu.Lock()
			for key, limiter := range rl.limiters {
				if limiter.Allow() {
					delete(rl.limiters, key)
				}
			}
			rl.mu.Unlock()
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
func RateLimit(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		// 根据路径选择不同的限流策略
		if c.Request.URL.Path == "/share/file/" || c.Request.URL.Path == "/share/text/" {
			// 上传限流
			limiter := uploadLimiter.GetLimiter(
				ip,
				rate.Every(time.Duration(cfg.UploadMinute)*time.Minute/time.Duration(cfg.UploadCount)),
				cfg.UploadCount,
			)
			if !limiter.Allow() {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"code":    429,
					"message": "上传频率过快，请稍后再试",
				})
				c.Abort()
				return
			}
		} else if c.Request.URL.Path == "/share/select/" && c.Request.Method == "GET" {
			// 只对GET请求的select进行错误限流，POST请求更宽松
			limiter := errorLimiter.GetLimiter(
				ip,
				rate.Every(time.Duration(cfg.ErrorMinute)*time.Minute/time.Duration(cfg.ErrorCount)),
				cfg.ErrorCount,
			)
			if !limiter.Allow() {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"code":    429,
					"message": "请求频率过快，请稍后再试",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// AdminAuth 管理员认证中间件
func AdminAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "缺少认证信息",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证格式错误",
			})
			c.Abort()
			return
		}

		// 验证JWT token
		tokenString := tokenParts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.AdminToken), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证失败",
			})
			c.Abort()
			return
		}

		// 检查是否是管理员
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if isAdmin, exists := claims["is_admin"]; !exists || !isAdmin.(bool) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    401,
					"message": "权限不足",
				})
				c.Abort()
				return
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "token格式错误",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ShareAuth 分享认证中间件
func ShareAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.OpenUpload == 0 {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "上传功能已关闭",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
