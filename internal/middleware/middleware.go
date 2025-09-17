package middleware

import (
	"strings"
	"sync"
	"time"

	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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

// Cleanup 清理过期的限流器
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
func RateLimit(manager *config.ConfigManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		// 根据路径选择不同的限流策略
		if c.Request.URL.Path == "/share/file/" || c.Request.URL.Path == "/share/text/" {
			// 上传限流
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
			// 只对GET请求的select进行错误限流，POST请求更宽松
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

// AdminAuth 管理员认证中间件（基于用户权限）
func AdminAuth(manager *config.ConfigManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文中获取用户角色（由UserAuth中间件设置）
		role, exists := c.Get("role")
		if !exists {
			common.UnauthorizedResponse(c, "用户权限信息不存在")
			c.Abort()
			return
		}

		// 检查用户角色是否为管理员
		roleStr, ok := role.(string)
		if !ok || roleStr != "admin" {
			common.ForbiddenResponse(c, "需要管理员权限")
			c.Abort()
			return
		}

		c.Next()
	}
}

// ShareAuth 分享认证中间件
func ShareAuth(manager *config.ConfigManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if manager.Transfer.Upload.OpenUpload == 0 {
			common.ForbiddenResponse(c, "上传功能已关闭")
			c.Abort()
			return
		}
		c.Next()
	}
}

// UserAuth 用户认证中间件
func UserAuth(manager *config.ConfigManager, userService interface {
	ValidateToken(string) (interface{}, error)
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 用户系统始终启用，直接进行认证验证

		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			common.UnauthorizedResponse(c, "缺少认证信息")
			c.Abort()
			return
		}

		// 检查Bearer前缀
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			common.UnauthorizedResponse(c, "认证格式错误")
			c.Abort()
			return
		}

		// 验证token
		claimsInterface, err := userService.ValidateToken(tokenParts[1])
		if err != nil {
			common.UnauthorizedResponse(c, "认证失败: "+err.Error())
			c.Abort()
			return
		}

		// 类型断言获取claims
		if claims, ok := claimsInterface.(*services.AuthClaims); ok {
			// 将用户信息设置到上下文
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("role", claims.Role)
			c.Set("session_id", claims.SessionID)
		} else {
			common.UnauthorizedResponse(c, "token格式错误")
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserClaims JWT claims 结构体定义
// OptionalUserAuth 可选用户认证中间件（支持匿名和登录用户）
func OptionalUserAuth(manager *config.ConfigManager, userService interface {
	ValidateToken(string) (interface{}, error)
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 用户系统始终启用，直接进行认证验证

		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 没有认证信息，允许匿名访问
			c.Set("is_anonymous", true)
			c.Next()
			return
		}

		// 检查Bearer前缀
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			// 认证格式错误，但仍允许匿名访问
			c.Set("is_anonymous", true)
			c.Next()
			return
		}

		// 尝试验证token
		claimsInterface, err := userService.ValidateToken(tokenParts[1])
		if err != nil {
			// token验证失败，但仍允许匿名访问
			c.Set("is_anonymous", true)
			c.Next()
			return
		}

		// 类型断言获取claims
		if claims, ok := claimsInterface.(*services.AuthClaims); ok {
			// 将用户信息设置到上下文
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("role", claims.Role)
			c.Set("session_id", claims.SessionID)
			c.Set("is_anonymous", false)
		} else {
			// claims格式错误，但仍允许匿名访问
			c.Set("is_anonymous", true)
		}

		c.Next()
	}
}
