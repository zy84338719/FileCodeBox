package bootstrap

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/spf13/viper"
	"github.com/zy84338719/fileCodeBox/backend/gen/http/router"
	"github.com/zy84338719/fileCodeBox/backend/internal/conf"
	"github.com/zy84338719/fileCodeBox/backend/internal/pkg/logger"
	"github.com/zy84338719/fileCodeBox/backend/internal/pkg/middleware"
	previewPkg "github.com/zy84338719/fileCodeBox/backend/internal/preview"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db/model"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/redis"
	"github.com/zy84338719/fileCodeBox/backend/internal/storage"
	"go.uber.org/zap"
	"gorm.io/gorm"

	notifyHandler "github.com/zy84338719/fileCodeBox/backend/gen/http/handler/notify"
	presignHandler "github.com/zy84338719/fileCodeBox/backend/gen/http/handler/presign"
	ratelimitHandler "github.com/zy84338719/fileCodeBox/backend/gen/http/handler/ratelimit"
	anonHandler "github.com/zy84338719/fileCodeBox/backend/gen/http/handler/share_anonymous"
	notifyAppService "github.com/zy84338719/fileCodeBox/backend/internal/app/notify"
	shareService "github.com/zy84338719/fileCodeBox/backend/internal/app/share"
	customHandler "github.com/zy84338719/fileCodeBox/backend/internal/transport/http/handler"
	customMw "github.com/zy84338719/fileCodeBox/backend/internal/transport/http/middleware"
)

// 使用 internal/conf 包中的统一配置类型
type Config = conf.AppConfiguration

// CORS 跨域中间件
func CORS() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		origin := string(c.GetHeader("Origin"))
		if origin == "" {
			origin = "*"
		}

		// 设置 CORS 头
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24小时

		// 处理预检请求
		if string(c.Method()) == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next(ctx)
	}
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	return config
}

// InitConfig 初始化配置
func InitConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 设置默认值
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 12345)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("database.driver", "sqlite")
	v.SetDefault("database.db_name", "./data/filecodebox.db")
	// 用户配置默认值
	v.SetDefault("user.allow_user_registration", true)
	v.SetDefault("user.require_email_verify", false)
	v.SetDefault("user.jwt_secret", "FileCodeBox2025JWT")

	if err := v.ReadInConfig(); err != nil {
		log.Printf("Warning: Failed to read config file: %v, using defaults", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// InitDatabase 初始化数据库
func InitDatabase(config *conf.DatabaseConfig) (*gorm.DB, error) {
	// 创建数据目录
	if config.Driver == "sqlite" {
		// 确保数据目录存在
		dbPath := config.DBName
		if dbPath != ":memory:" {
			// 创建目录（如果需要）
			// 这里简化处理，GORM 会自动创建数据库文件
			log.Printf("SQLite database path: %s", dbPath)
		}
	}

	// 初始化数据库连接
	err := db.Init(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	database := db.GetDB()

	// 自动迁移表结构
	log.Println("Auto migrating database tables...")
	err = database.AutoMigrate(
		&model.User{},
		&model.FileCode{},
		&model.UploadChunk{},
		&model.TransferLog{},
		&model.AdminOperationLog{},
		&model.UserAPIKey{},
		&model.FilePreview{}, // 添加预览表
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database initialized successfully")
	return database, nil
}

// CreateDefaultAdmin 创建默认管理员
func CreateDefaultAdmin(database *gorm.DB) error {
	var count int64
	database.Model(&model.User{}).Where("role = ?", "admin").Count(&count)

	if count > 0 {
		log.Println("Admin user already exists")
		return nil
	}

	// 创建默认管理员
	admin := &model.User{
		Username:     "admin",
		Email:        "admin@filecodebox.local",
		PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3.rsQ5pPjZ5yVlWK5WAe", // password: admin123
		Nickname:     "Administrator",
		Role:         "admin",
		Status:       "active",
	}

	if err := database.Create(admin).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	log.Println("Default admin user created (username: admin, password: admin123)")
	return nil
}

var (
	database *gorm.DB
	config   *Config
)

// Bootstrap 应用程序启动入口
func Bootstrap() (*server.Hertz, error) {
	// 1. 初始化配置
	var err error
	config, err = InitConfig("configs/config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to init config: %w", err)
	}

	// 设置全局配置（供其他包访问）
	conf.SetGlobalConfig(config)

	// 2. 初始化日志
	loggerConfig := &logger.Config{
		Level:      config.Log.Level,
		Filename:   config.Log.Filename,
		MaxSize:    config.Log.MaxSize,
		MaxBackups: config.Log.MaxBackups,
		MaxAge:     config.Log.MaxAge,
		Compress:   config.Log.Compress,
	}
	if err := logger.Init(loggerConfig); err != nil {
		return nil, fmt.Errorf("failed to init logger: %w", err)
	}

	// 3. 初始化数据库
	database, err = InitDatabase(&config.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	// 4. 创建默认管理员
	if err := CreateDefaultAdmin(database); err != nil {
		logger.Error("Failed to create default admin", zap.Error(err))
	}

	// 4.5 初始化预览服务
	if err := initPreviewService(); err != nil {
		logger.Error("Failed to init preview service", zap.Error(err))
	}

	// 4.6 初始化新服务（thrift IDL 对应：notify/presign/anonymous/ratelimit）
	initThriftIDLServices(database)

	// 5. 创建 HTTP 服务器
	port := config.Server.Port
	if port == 0 {
		port = 12345
	}
	h := server.New(
		server.WithHostPorts(fmt.Sprintf("%s:%d", config.Server.Host, port)),
	)

	// 全局中间件（顺序：Recovery → RequestID → CORS）
	// Recovery 必须最先，确保任何 panic 都被捕获并转为 500，避免进程崩溃
	h.Use(middleware.Recovery())
	// RequestID 为每个请求生成/透传 trace_id（X-Trace-Id），写入 ctx 供日志/resp 使用
	h.Use(middleware.RequestID())
	// CORS 跨域
	h.Use(CORS())

	// 6. 注册路由
	router.GeneratedRegister(h)

	// 7. 注册自定义路由
	customizedRegister(h)

	logger.Info("Application bootstrap completed successfully")
	return h, nil
}

// Cleanup 清理资源
func Cleanup() {
	logger.Info("Cleaning up resources...")

	if database != nil {
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database", zap.Error(err))
		}
	}

	logger.Sync()
}

// customizedRegister 注册自定义路由（不走 thrift IDL 生成）。
//
// 历史背景：该函数此前为空，导致前端 SPA 静态服务、Swagger 文档、
// “我的分享管理”与“用户通知”REST API 全部未生效。本函数把此前残留在
// 根目录 router.go（死代码）中的逻辑正式接线，使单二进制 / Docker 部署
// 即可打开前端页面并使用全部功能。
func customizedRegister(r *server.Hertz) {
	// ===== OpenAPI 文档（Swagger UI）=====
	r.GET("/openapi.json", customHandler.OpenAPISpec)

	// ===== 自定义 REST API（需用户 JWT 认证）=====
	apiV1 := r.Group("/api/v1", customMw.UserAuth())
	{
		// 我的分享管理（批量删除 / 批量延期 / 恢复 / 永久删除）
		userShares := apiV1.Group("/user/shares")
		userShares.GET("", customHandler.ListUserShares)
		userShares.POST("/batch-delete", customHandler.BatchDeleteUserShares)
		userShares.POST("/batch-extend", customHandler.BatchExtendUserShares)
		userShares.POST("/:code/restore", customHandler.RestoreUserShare)
		userShares.DELETE("/:code/hard", customHandler.HardDeleteUserShare)

		// 用户站内通知（列表 / 未读数 / 标记已读）
		apiV1.GET("/notifies/mine", customHandler.ListMyNotifications)
		apiV1.GET("/notifies/unread-count", customHandler.UnreadNotifyCount)
		apiV1.POST("/notifies/mark-read", customHandler.MarkNotifyRead)
	}

	// ===== 前端 SPA 静态资源服务 =====
	// 服务 /static 前缀的构建产物（JS/CSS/图片等）。
	// 注意：必须用 NewPathSlashesStripper(1) 剥离 "/static" 前缀，
	// 否则 handler 会在 root 下重复拼接导致 404。
	r.StaticFS("/static", &app.FS{
		Root:         "./static",
		PathRewrite:  app.NewPathSlashesStripper(1),
		CacheDuration: 7 * 24 * time.Hour,
	})
	// SPA fallback：根路径 "/" 与所有非 API 路径一律回退到 index.html，
	// 交给前端 hash 路由处理（前端使用 createWebHashHistory）。
	r.NoRoute(func(ctx context.Context, c *app.RequestContext) {
		path := string(c.Request.URI().Path())
		if isAPIPath(path) {
			c.JSON(consts.StatusNotFound, map[string]interface{}{
				"code":    404,
				"message": "API endpoint not found",
			})
			return
		}
		c.File("./static/index.html")
	})
}

// apiPathPrefixes 仅含"纯 API"前缀——这些前缀下不存在前端页面，
// 未命中路由时返回 JSON 404，而不是回退到 SPA index.html。
//
// 注意：/user、/admin、/anonymous、/share 同时是后端 API 前缀与前端 hash
// 路由的页面路径（如 /user/login、/admin/dashboard、/share/:code）。前端使用
// createWebHashHistory，浏览器实际只请求 "/"，但若用户直接访问这些 history
// 路径（如刷新、外链），应回退到 SPA 由前端路由处理，而非返回 API 404。
// 因此它们不在本列表中。
var apiPathPrefixes = []string{
	"/api", "/chunk", "/notifies",
	"/health", "/live", "/ready", "/ping", "/version",
	"/openapi", "/metrics", "/preview", "/qrcode", "/setup",
}

// isAPIPath 判断路径是否属于后端 API（而非前端 SPA 路由）。
func isAPIPath(path string) bool {
	for _, p := range apiPathPrefixes {
		if path == p || strings.HasPrefix(path, p+"/") {
			return true
		}
	}
	return false
}

// initPreviewService 初始化预览服务
func initPreviewService() error {
	previewConfig := &previewPkg.Config{
		EnablePreview:    true,
		ThumbnailWidth:   300,
		ThumbnailHeight:  200,
		MaxFileSize:      50 * 1024 * 1024,
		PreviewCachePath: "./data/previews",
		FFmpegPath:       "ffmpeg",
	}

	return previewPkg.InitService(previewConfig)
}

// initThriftIDLServices 初始化 thrift IDL 对应的新服务
// 关联 internal/app/ → gen/http/handler/ 各 SetXxx 入口
func initThriftIDLServices(database *gorm.DB) {
	// 1. notify service（需要 DB）
	notifyApp := notifyAppService.NewService(database)
	notifyHandler.SetDB(database)
	// 1.1 注入定制路由的 notify service
	customHandler.SetNotifyService(notifyApp)

	// 2. presign service（需要 Redis + baseURL + signingKey + share service）
	baseURL := fmt.Sprintf("http://%s:%d", config.Server.Host, config.Server.Port)
	presignHandler.SetService(redis.GetClient(),
		baseURL,
		"filecodebox-dev-signing-key-change-me")
	// 2.1 注入 share service（Complete 时写分享表）
	shareSvc := shareService.NewService(baseURL, getBootstrapStorageService())
	presignHandler.SetShareService(shareSvc)
	// 2.2 注入定制路由的 share service
	customHandler.SetShareService(shareSvc)
	// 2.3 注入 notify service（取件时给 owner 发通知）
	shareSvc.SetNotifyService(notifyApp) // *Service 已实现 CreateForUserSimple

	// 3. anonymous service（需要 Redis）
	anonHandler.SetService(redis.GetClient())

	// 4. ratelimit service（直接用 default limiter）
	ratelimitHandler.SetLimiter(middleware.GetDefaultRateLimiter())

	// 5. 自动迁移 notify 表 + file_codes viewer 字段
	if err := database.AutoMigrate(&model.Notify{}); err != nil {
		logger.Error("Failed to migrate notify table", zap.Error(err))
	} else {
		logger.Info("Notify table migrated")
	}
	if err := database.AutoMigrate(&model.FileCode{}); err != nil {
		logger.Error("Failed to migrate file_codes table", zap.Error(err))
	} else {
		logger.Info("FileCode table migrated (viewer fields added)")
	}
}

// getBootstrapStorageService bootstrap 用的 storage service
// 后续应该从 conf 读 storage 配置（task4 后续）
func getBootstrapStorageService() *storage.StorageService {
	dataPath := "./data"
	if config != nil && config.Storage.StoragePath != "" {
		dataPath = config.Storage.StoragePath
	}
	storageType := storage.StorageTypeLocal
	if config != nil && config.Storage.Type != "" {
		switch config.Storage.Type {
		case "s3":
			storageType = storage.StorageTypeS3
		case "webdav":
			storageType = storage.StorageTypeWebDAV
		}
	}
	return storage.NewStorageService(&storage.StorageConfig{
		Type:     storageType,
		DataPath: dataPath,
		BaseURL:  fmt.Sprintf("http://%s:%d", config.Server.Host, config.Server.Port),
	})
}
