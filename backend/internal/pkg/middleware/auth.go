package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/fileCodeBox/internal/pkg/auth"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 获取Authorization头
		authHeader := string(c.GetHeader("Authorization"))
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    http.StatusUnauthorized,
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// 验证Bearer token格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    http.StatusUnauthorized,
				"message": "Authorization header format must be Bearer {token}",
			})
			c.Abort()
			return
		}

		// 解析JWT token
		claims, err := auth.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    http.StatusUnauthorized,
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		
		// 同时设置 Header，方便 handler 读取
		c.Header("X-User-ID", fmt.Sprintf("%d", claims.UserID))
		c.Header("X-Username", claims.Username)
		c.Header("X-Role", claims.Role)

		c.Next(ctx)
	}
}

// AdminMiddleware 管理员权限中间件
func AdminMiddleware() app.HandlerFunc {
	// 不需要认证的路径白名单
	skipPaths := map[string]bool{
		"/admin/login": true,
	}

	return func(ctx context.Context, c *app.RequestContext) {
		// 检查是否在白名单中
		path := string(c.URI().Path())
		if skipPaths[path] {
			c.Next(ctx)
			return
		}

		// 先进行身份认证
		AuthMiddleware()(ctx, c)
		if c.IsAborted() {
			return
		}

		// 检查是否为管理员
		role, _ := c.Get("role")
		if role != "admin" {
			c.JSON(http.StatusForbidden, map[string]interface{}{
				"code":    http.StatusForbidden,
				"message": "Admin access required",
			})
			c.Abort()
			return
		}

		c.Next(ctx)
	}
}

// OptionalAuthMiddleware 可选认证中间件（不强制要求登录）
func OptionalAuthMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 获取Authorization头
		authHeader := string(c.GetHeader("Authorization"))
		if authHeader == "" {
			c.Next(ctx)
			return
		}

		// 验证Bearer token格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.Next(ctx)
			return
		}

		// 解析JWT token
		claims, err := auth.ParseToken(parts[1])
		if err != nil {
			c.Next(ctx)
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		
		// 同时设置 Header
		c.Header("X-User-ID", fmt.Sprintf("%d", claims.UserID))
		c.Header("X-Username", claims.Username)
		c.Header("X-Role", claims.Role)

		c.Next(ctx)
	}
}
