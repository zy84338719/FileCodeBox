package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

// AdminAuth 管理员 JWT 认证中间件
// 验证 JWT token 并确保用户具有管理员角色
func AdminAuth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		tokenString := extractBearerToken(c)
		if tokenString == "" {
			respondUnauthorized(c, "Authorization header is required")
			return
		}

		if err := parseAndSetClaims(c, tokenString); err != nil {
			respondUnauthorized(c, "Invalid or expired token")
			return
		}

		// 检查管理员权限
		if !IsAdmin(c) {
			respondForbidden(c, "Admin access required")
			return
		}

		c.Next(ctx)
	}
}

// AdminAuthStrict 严格的管理员认证中间件
// 与 AdminAuth 功能相同，用于强调权限检查的严格性
// 可以用于系统关键操作的保护
func AdminAuthStrict() app.HandlerFunc {
	return AdminAuth()
}

// SuperAdminAuth 超级管理员认证中间件
// 预留接口，可用于未来多级权限系统
// 目前与 AdminAuth 行为一致
func SuperAdminAuth() app.HandlerFunc {
	return AdminAuth()
}

// AdminOrSystemAdmin 管理员或系统管理员认证中间件
// 预留接口，用于未来多级权限系统
func AdminOrSystemAdmin() app.HandlerFunc {
	return AdminAuth()
}

// AdminOnly 仅允许管理员访问（不检查 JWT，从上下文获取角色）
// 用于在已认证的基础上二次验证管理员权限
func AdminOnly() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if !IsAuthenticated(c) {
			respondUnauthorized(c, "Authentication required")
			return
		}

		if !IsAdmin(c) {
			respondForbidden(c, "Admin access required")
			return
		}

		c.Next(ctx)
	}
}
