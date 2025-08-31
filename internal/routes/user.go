package routes

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/handlers"
	"github.com/zy84338719/filecodebox/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupUserRoutes 设置用户相关路由
func SetupUserRoutes(
	router *gin.Engine,
	userHandler *handlers.UserHandler,
	cfg *config.Config,
	userService interface {
		ValidateToken(string) (interface{}, error)
	},
) {
	// 用户系统路由
	userGroup := router.Group("/user")
	{
		// 公开路由（不需要认证）
		userGroup.POST("/register", userHandler.Register)
		userGroup.POST("/login", userHandler.Login)
		userGroup.GET("/system-info", userHandler.GetSystemInfo)

		// 需要认证的路由
		authGroup := userGroup.Group("/")
		authGroup.Use(middleware.UserAuth(cfg, userService))
		{
			authGroup.POST("/logout", userHandler.Logout)
			authGroup.GET("/profile", userHandler.GetProfile)
			authGroup.PUT("/profile", userHandler.UpdateProfile)
			authGroup.POST("/change-password", userHandler.ChangePassword)
			authGroup.GET("/files", userHandler.GetUserFiles)
			authGroup.GET("/stats", userHandler.GetUserStats)
			authGroup.GET("/check-auth", userHandler.CheckAuth)
			authGroup.DELETE("/files/:id", userHandler.DeleteFile)
		}
	}

	// 用户页面路由
	userPageGroup := router.Group("/user")
	{
		userPageGroup.GET("/login", func(c *gin.Context) {
			ServeUserPage(c, cfg, "login.html")
		})
		userPageGroup.GET("/register", func(c *gin.Context) {
			ServeUserPage(c, cfg, "register.html")
		})
		userPageGroup.GET("/dashboard", func(c *gin.Context) {
			ServeUserPage(c, cfg, "dashboard.html")
		})
		userPageGroup.GET("/forgot-password", func(c *gin.Context) {
			ServeUserPage(c, cfg, "forgot-password.html")
		})
	}
}

// ServeUserPage 服务用户页面
func ServeUserPage(c *gin.Context, cfg *config.Config, pageName string) {
	userPagePath := filepath.Join(".", cfg.ThemesSelect, pageName)

	content, err := os.ReadFile(userPagePath)
	if err != nil {
		c.String(http.StatusNotFound, "User page not found: "+pageName)
		return
	}

	c.Header("Cache-Control", "no-cache")
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, string(content))
}
