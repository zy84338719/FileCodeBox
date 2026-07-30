package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

// SecurityHeaders 设置常见安全响应头，缓解 XSS / 点击劫持 / MIME 嗅探等风险。
//
//   - X-Content-Type-Options: nosniff  → 禁止浏览器嗅探 MIME 类型
//   - X-Frame-Options: SAMEORIGIN      → 防止点击劫持（页面只能被同源 iframe 嵌入）
//   - X-XSS-Protection: 1; mode=block  → 启用浏览器 XSS 过滤器（旧浏览器）
//   - Referrer-Policy: strict-origin-when-cross-origin → 限制 Referrer 泄露
//   - Strict-Transport-Security        → 仅在生产(HTTPS)下启用 HSTS
//
// 可通过 SetSecurityHeadersConfig 调整。
type SecurityHeadersConfig struct {
	// EnableHSTS 是否启用 Strict-Transport-Security（仅 HTTPS 部署时启用，
	// 否则可能把用户锁在无法访问的 https 状态）。
	EnableHSTS bool
}

var securityHeadersConfig = SecurityHeadersConfig{
	EnableHSTS: false, // 默认关闭，由生产配置显式开启
}

// SetSecurityHeadersConfig 设置安全头全局配置（bootstrap 调用）。
func SetSecurityHeadersConfig(cfg SecurityHeadersConfig) {
	securityHeadersConfig = cfg
}

// SecurityHeaders 安全响应头中间件。
func SecurityHeaders() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		if securityHeadersConfig.EnableHSTS {
			// max-age=31536000（1年），含子域名，预加载
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}
		c.Next(ctx)
	}
}
