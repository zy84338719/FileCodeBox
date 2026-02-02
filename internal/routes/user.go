package routes

import (
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/handlers"
	"github.com/zy84338719/filecodebox/internal/middleware"
	"github.com/zy84338719/filecodebox/internal/services"
	"github.com/zy84338719/filecodebox/internal/static"

	"github.com/gin-gonic/gin"
)

// SetupUserRoutes 设置用户相关路由
func SetupUserRoutes(
	router *gin.Engine,
	userHandler *handlers.UserHandler,
	cfg *config.ConfigManager,
	userService *services.UserService,
) {
	// 注册完整的用户路由（API + 页面）
	SetupUserAPIRoutes(router, userHandler, cfg, userService)

	// 用户页面路由
	userPageGroup := router.Group("/user")
	{
		userPageGroup.GET("/login", func(c *gin.Context) {
			static.ServeUserPage(c, cfg, "login.html")
		})
		// 只有允许注册时才提供注册页面
		if cfg.User.IsRegistrationAllowed() {
			userPageGroup.GET("/register", func(c *gin.Context) {
				static.ServeUserPage(c, cfg, "register.html")
			})
		}
		userPageGroup.GET("/dashboard", func(c *gin.Context) {
			static.ServeUserPage(c, cfg, "dashboard.html")
		})
		userPageGroup.GET("/forgot-password", func(c *gin.Context) {
			static.ServeUserPage(c, cfg, "forgot-password.html")
		})
	}
}

// SetupUserAPIRoutes 仅注册用户相关的 API 路由（供动态注册时使用，避免重复注册页面路由）
func SetupUserAPIRoutes(
	router *gin.Engine,
	userHandler *handlers.UserHandler,
	cfg *config.ConfigManager,
	userService *services.UserService,
) {
	// 用户系统路由
	userGroup := router.Group("/user")
	{
		// 公开路由（不需要认证）
		// 只有允许注册时才注册这个路由
		if cfg.User.IsRegistrationAllowed() {
			userGroup.POST("/register", userHandler.Register)
		}
		userGroup.POST("/login", userHandler.Login)
		// `/user/system-info` 由 `SetupBaseRoutes` 全局注册并在有 `userHandler` 时委托处理，
		// 因此在此处不要重复注册以避免路由冲突（Gin 在重复注册同一路径时会 panic）
		userGroup.GET("/check-initialization", userHandler.CheckSystemInitialization)

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
			authGroup.DELETE("/files/:code", userHandler.DeleteFile)
			authGroup.GET("/api-keys", userHandler.ListAPIKeys)
			authGroup.POST("/api-keys", userHandler.CreateAPIKey)
			authGroup.DELETE("/api-keys/:id", userHandler.DeleteAPIKey)
		}
	}
}

// ServeUserPage 服务用户页面
// ServeUserPage has been moved to internal/static package (static.ServeUserPage)
