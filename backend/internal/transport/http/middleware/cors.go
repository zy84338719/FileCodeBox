package middleware

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
)

// CORSConfig CORS 配置选项
type CORSConfig struct {
	// AllowOrigins 允许的源，* 表示允许所有源
	AllowOrigins []string
	// AllowMethods 允许的 HTTP 方法
	AllowMethods []string
	// AllowHeaders 允许的请求头
	AllowHeaders []string
	// ExposeHeaders 暴露给客户端的响应头
	ExposeHeaders []string
	// AllowCredentials 是否允许携带凭证
	AllowCredentials bool
	// MaxAge 预检请求的缓存时间（秒）
	MaxAge int
	// OptionsPassthrough 是否在 OPTIONS 请求后继续执行后续中间件
	OptionsPassthrough bool
}

// DefaultCORSConfig 返回默认的 CORS 配置
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Requested-With", "X-API-Key"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           86400,
		OptionsPassthrough: false,
	}
}

// CORS 返回默认配置的 CORS 中间件
func CORS() app.HandlerFunc {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig 使用自定义配置返回 CORS 中间件
func CORSWithConfig(config *CORSConfig) app.HandlerFunc {
	if config == nil {
		config = DefaultCORSConfig()
	}

	return func(ctx context.Context, c *app.RequestContext) {
		origin := string(c.GetHeader("Origin"))
		method := string(c.Method())

		// 设置 Allow-Origin
		if len(config.AllowOrigins) == 1 && config.AllowOrigins[0] == "*" {
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			for _, allowedOrigin := range config.AllowOrigins {
				if allowedOrigin == origin || allowedOrigin == "*" {
					c.Header("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		// 设置 Allow-Methods
		if len(config.AllowMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
		}

		// 设置 Allow-Headers
		if len(config.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
		}

		// 设置 Expose-Headers
		if len(config.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
		}

		// 设置 Allow-Credentials
		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 设置 Max-Age
		if config.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", string(rune(config.MaxAge)))
		}

		// 处理 OPTIONS 预检请求
		if method == "OPTIONS" {
			if !config.OptionsPassthrough {
				c.AbortWithStatus(204)
				return
			}
		}

		c.Next(ctx)
	}
}

// CORSWithAllowOrigins 设置允许的源
func CORSWithAllowOrigins(origins []string) app.HandlerFunc {
	config := DefaultCORSConfig()
	config.AllowOrigins = origins
	return CORSWithConfig(config)
}

// CORSWithAllowCredentials 允许携带凭证的 CORS 中间件
func CORSWithAllowCredentials() app.HandlerFunc {
	config := DefaultCORSConfig()
	config.AllowCredentials = true
	config.AllowOrigins = []string{} // 不能使用 * 当 AllowCredentials 为 true 时
	return CORSWithConfig(config)
}
