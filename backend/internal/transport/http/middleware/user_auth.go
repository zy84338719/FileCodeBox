package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

const (
	// UserRoleAdmin 管理员角色
	UserRoleAdmin = "admin"
	// UserRoleUser 普通用户角色
	UserRoleUser = "user"
)

// UserAuth 用户 JWT 认证中间件
// 验证 Bearer JWT token 并将用户信息注入到上下文
func UserAuth() app.HandlerFunc {
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

		c.Next(ctx)
	}
}

// OptionalUserAuth 可选用户认证中间件
// 不强制要求登录，但如果提供了有效的 token 则解析用户信息
func OptionalUserAuth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		tokenString := extractBearerToken(c)
		if tokenString == "" {
			// 没有提供 token，继续执行（匿名用户）
			c.Next(ctx)
			return
		}

		// 尝试解析 token，如果失败则继续执行（匿名用户）
		_ = parseAndSetClaims(c, tokenString)

		c.Next(ctx)
	}
}

// RequireAdmin 要求管理员权限的中间件
// 必须配合 UserAuth 或其他认证中间件使用
func RequireAdmin() app.HandlerFunc {
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

// RequireRole 要求特定角色的中间件
// 必须配合 UserAuth 或其他认证中间件使用
func RequireRole(role string) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if !IsAuthenticated(c) {
			respondUnauthorized(c, "Authentication required")
			return
		}

		if GetUserRole(c) != role {
			respondForbidden(c, "Required role: "+role)
			return
		}

		c.Next(ctx)
	}
}

// RequireAnyRole 要求具备任一指定角色的中间件
// 必须配合 UserAuth 或其他认证中间件使用
func RequireAnyRole(roles ...string) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if !IsAuthenticated(c) {
			respondUnauthorized(c, "Authentication required")
			return
		}

		currentRole := GetUserRole(c)
		for _, role := range roles {
			if currentRole == role {
				c.Next(ctx)
				return
			}
		}

		respondForbidden(c, "Access denied: insufficient permissions")
	}
}

// UserAuthWithAdmin 用户认证 + 管理员权限检查的组合中间件
func UserAuthWithAdmin() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 先执行用户认证
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
