package bootstrap

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/spf13/viper"
	"github.com/zy84338719/fileCodeBox/backend/gen/http/router"
	"github.com/zy84338719/fileCodeBox/backend/internal/conf"
	"github.com/zy84338719/fileCodeBox/backend/internal/pkg/auth"
	"github.com/zy84338719/fileCodeBox/backend/internal/pkg/logger"
	"github.com/zy84338719/fileCodeBox/backend/internal/pkg/middleware"
	"github.com/zy84338719/fileCodeBox/backend/internal/pkg/resp"
	previewPkg "github.com/zy84338719/fileCodeBox/backend/internal/preview"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db/model"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/redis"
	"github.com/zy84338719/fileCodeBox/backend/internal/storage"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	notifyHandler "github.com/zy84338719/fileCodeBox/backend/gen/http/handler/notify"
	presignHandler "github.com/zy84338719/fileCodeBox/backend/gen/http/handler/presign"
	ratelimitHandler "github.com/zy84338719/fileCodeBox/backend/gen/http/handler/ratelimit"
	adminHandler "github.com/zy84338719/fileCodeBox/backend/gen/http/handler/admin"
	anonHandler "github.com/zy84338719/fileCodeBox/backend/gen/http/handler/share_anonymous"
	notifyAppService "github.com/zy84338719/fileCodeBox/backend/internal/app/notify"
	shareService "github.com/zy84338719/fileCodeBox/backend/internal/app/share"
	adminApp "github.com/zy84338719/fileCodeBox/backend/internal/app/admin"
	customHandler "github.com/zy84338719/fileCodeBox/backend/internal/transport/http/handler"
	customMw "github.com/zy84338719/fileCodeBox/backend/internal/transport/http/middleware"
)

// 使用 internal/conf 包中的统一配置类型
type Config = conf.AppConfiguration

// CORS 跨域中间件（配置化）。
//
// 安全策略：
//   - 配置了 allow_origins 白名单时，仅放行白名单内的 Origin（生产推荐）
//   - 未配置白名单时，退化为反射 Origin（便于本地开发，等同于宽松模式）
//   - allow_credentials=true 时，绝不返回 "*"，而是精确匹配的 Origin
//
// 同时允许 X-Trace-Id / X-API-Key 等自定义请求头跨域。
// allow_origins 来源：yaml 的 security.cors.allow_origins（数组）或
// 环境变量 FCB_CORS_ALLOW_ORIGINS（逗号分隔，如 "https://a.com,https://b.com"）。
func CORS() app.HandlerFunc {
	allowOrigins := map[string]bool{}
	// 优先从环境变量读取（逗号分隔），兼容 slice 字段在 env 下的传递
	if envOrigins := os.Getenv("FCB_CORS_ALLOW_ORIGINS"); envOrigins != "" {
		for _, o := range strings.Split(envOrigins, ",") {
			if o = strings.TrimSpace(o); o != "" {
				allowOrigins[o] = true
			}
		}
	}
	// 再合并配置文件中的白名单
	for _, o := range config.Security.CORS.AllowOrigins {
		allowOrigins[o] = true
	}
	allowCredentials := config.Security.CORS.AllowCredentials
	if len(allowOrigins) == 0 {
		// 无白名单默认允许凭证（开发友好）；生产应显式配置白名单
		allowCredentials = true
	}

	return func(ctx context.Context, c *app.RequestContext) {
		origin := string(c.GetHeader("Origin"))

		allowedOrigin := ""
		if origin != "" {
			if len(allowOrigins) > 0 {
				// 白名单模式：精确匹配
				if allowOrigins[origin] {
					allowedOrigin = origin
				}
			} else {
				// 宽松模式：反射任意 Origin（开发用）
				allowedOrigin = origin
			}
		}

		if allowedOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
			c.Header("Vary", "Origin")
			if allowCredentials {
				// 凭证模式下不能返回 "*"，必须是具体 origin（上面已保证）
				c.Header("Access-Control-Allow-Credentials", "true")
			}
		} else if origin == "" {
			// 无 Origin（同源请求），放行且不设 ACAO
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-Trace-Id, X-API-Key")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type, X-Trace-Id")
		c.Header("Access-Control-Max-Age", "86400")

		// 处理预检请求
		if string(c.Method()) == "OPTIONS" {
			c.AbortWithStatus(consts.StatusNoContent)
			return
		}

		c.Next(ctx)
	}
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	return config
}

// InitConfig 初始化配置。
//
// 配置来源优先级（高 → 低）：
//  1. 环境变量（FCB_ 前缀完整名 / 文档化的短名，见 bindEnvironment）
//  2. 配置文件（yaml，路径由 configPath 或 CONFIG_PATH 决定）
//  3. 代码内默认值（SetDefault）
//
// 这样容器化部署（K8s/Docker）可通过 env 注入敏感配置（jwt_secret、db 密码等），
// 而无需修改镜像内的配置文件，符合 12-factor。
func InitConfig(configPath string) (*Config, error) {
	// 解析最终配置文件路径：参数 > CONFIG_PATH env > 默认
	if configPath == "" {
		configPath = os.Getenv("CONFIG_PATH")
	}
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 代码内默认值（仅当文件与 env 均未设置时生效）
	setDefaults(v)

	// 绑定环境变量（优先级最高）
	bindEnvironment(v)

	// 读取配置文件（缺失时降级为仅默认值 + env，便于无文件启动）
	if err := v.ReadInConfig(); err != nil {
		log.Printf("Warning: Failed to read config file %s: %v, using defaults + env", configPath, err)
	} else {
		log.Printf("Loaded config from: %s", v.ConfigFileUsed())
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 生产环境敏感配置 fail-fast 校验
	if err := validateSecrets(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// setDefaults 设置代码内默认值。
func setDefaults(v *viper.Viper) {
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 12345)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.base_url", "")
	v.SetDefault("database.driver", "sqlite")
	v.SetDefault("database.db_name", "./data/filecodebox.db")
	v.SetDefault("user.allow_user_registration", true)
	v.SetDefault("user.require_email_verify", false)
	v.SetDefault("observability.metrics.enabled", true)
	v.SetDefault("observability.metrics.path", "/metrics")
	v.SetDefault("observability.tracing.enabled", false)
}

// envBindings 环境变量 → 配置 key 的映射。
// 同时支持两套命名：
//   - 文档化的短扁平名（PORT / DATABASE_HOST 等，便于运维记忆）
//   - FCB_ 前缀 + 下划线的完整名（FCB_SERVER_PORT，与 mapstructure key 对齐）
var envBindings = map[string][]string{
	// server
	"server.host":          {"FCB_SERVER_HOST", "HOST"},
	"server.port":          {"FCB_SERVER_PORT", "PORT"},
	"server.mode":          {"FCB_SERVER_MODE"},
	"server.base_url":      {"FCB_SERVER_BASE_URL", "BASE_URL"},
	"server.read_timeout":  {"FCB_SERVER_READ_TIMEOUT"},
	"server.write_timeout": {"FCB_SERVER_WRITE_TIMEOUT"},
	// database
	"database.driver":   {"FCB_DATABASE_DRIVER", "DATABASE_TYPE", "DB_TYPE"},
	"database.db_name":  {"FCB_DATABASE_DB_NAME", "DATABASE_NAME", "DB_NAME"},
	"database.host":     {"FCB_DATABASE_HOST", "DATABASE_HOST", "DB_HOST"},
	"database.port":     {"FCB_DATABASE_PORT", "DATABASE_PORT", "DB_PORT"},
	"database.user":     {"FCB_DATABASE_USER", "DATABASE_USER", "DB_USER"},
	"database.password": {"FCB_DATABASE_PASSWORD", "DATABASE_PASS", "DB_PASS"},
	// redis
	"redis.host":     {"FCB_REDIS_HOST", "REDIS_HOST"},
	"redis.port":     {"FCB_REDIS_PORT", "REDIS_PORT"},
	"redis.password": {"FCB_REDIS_PASSWORD", "REDIS_PASSWORD"},
	"redis.db":       {"FCB_REDIS_DB", "REDIS_DB"},
	// app
	"app.datapath":   {"FCB_DATA_PATH", "DATA_PATH"},
	"app.production": {"FCB_PRODUCTION", "PRODUCTION"},
	// user
	"user.jwt_secret":             {"FCB_JWT_SECRET", "JWT_SECRET"},
	"user.allow_user_registration": {"FCB_USER_ALLOW_REGISTRATION"},
	// upload
	"upload.open_upload": {"FCB_OPEN_UPLOAD", "OPEN_UPLOAD"},
	"upload.upload_size": {"FCB_UPLOAD_SIZE", "UPLOAD_SIZE"},
	// storage
	"storage.type":         {"FCB_STORAGE_TYPE"},
	"storage.storage_path": {"FCB_STORAGE_PATH"},
	// observability
	"observability.metrics.enabled": {"FCB_METRICS_ENABLED"},
	"observability.metrics.path":    {"FCB_METRICS_PATH"},
	"observability.tracing.enabled": {"FCB_TRACING_ENABLED"},
	// security
	"security.cors.allow_origins":     {"FCB_CORS_ALLOW_ORIGINS"},
	"security.cors.enable_hsts":       {"FCB_ENABLE_HSTS"},
}

// bindEnvironment 把环境变量绑定到 viper 配置 key。
// 列表中靠前的 env 名优先（viper BindEnv 只绑定第一个非空）。
func bindEnvironment(v *viper.Viper) {
	for key, envs := range envBindings {
		// 绑定所有候选 env 名；viper 会取最后一个 BindEnv 的值，
		// 因此我们逐个检查并显式设置，确保优先级正确。
		for _, env := range envs {
			if val, ok := os.LookupEnv(env); ok {
				v.Set(key, val)
				break
			}
		}
	}
}

// insecureDefaultSecrets 已知的不安全默认密钥（禁止在生产环境使用）。
var insecureDefaultSecrets = map[string]string{
	"FileCodeBox2025JWT":                   "user.jwt_secret",
	"filecodebox-dev-signing-key-change-me": "presign signing key",
}

// validateSecrets 在生产环境校验敏感配置，避免使用默认/弱密钥启动（fail-fast）。
func validateSecrets(cfg *Config) error {
	if !cfg.IsProduction() {
		return nil
	}
	// jwt_secret
	if sec := cfg.User.JWTSecret; sec == "" || insecureDefaultSecrets[sec] != "" {
		return fmt.Errorf("production mode requires a secure user.jwt_secret: current value is empty or a known default; set FCB_JWT_SECRET env to a strong random string")
	}
	return nil
}

// InitDatabase 初始化数据库。
//
// 迁移策略（按配置）：
//   - database.migrate=true：先执行版本化迁移(migrations/*.sql)，适合生产/需要版本控制
//   - database.auto_migrate（默认 true）：GORM AutoMigrate，开发友好、自动补表/列
//   - 两者可共存：版本化迁移建表后，AutoMigrate 兜底补充新字段
func InitDatabase(config *conf.DatabaseConfig) (*gorm.DB, error) {
	// 创建数据目录
	if config.Driver == "sqlite" {
		dbPath := config.DBName
		if dbPath != ":memory:" {
			log.Printf("SQLite database path: %s", dbPath)
		}
	}

	// 初始化数据库连接（db.Init 内部会执行 AutoMigrate 兜底）
	err := db.Init(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	database := db.GetDB()

	// 版本化迁移（企业级，可选）
	if config.Migrate {
		log.Println("Running versioned database migrations...")
		migrator, err := db.NewMigrator(database, config.Driver)
		if err != nil {
			return nil, fmt.Errorf("failed to create migrator: %w", err)
		}
		applied, err := migrator.Up()
		if err != nil {
			return nil, fmt.Errorf("versioned migration failed: %w", err)
		}
		if len(applied) > 0 {
			log.Printf("Applied %d migration(s): %v", len(applied), applied)
		} else {
			log.Println("No new migrations to apply (already up to date)")
		}
		// AutoMigrate 兜底：补充 baseline 未覆盖的表（如 file_previews/notifies）
		if err := database.AutoMigrate(
			&model.FilePreview{},
		); err != nil {
			logger.Error("AutoMigrate fallback for previews failed", zap.Error(err))
		}
	}

	log.Println("Database initialized successfully")
	return database, nil
}

// CreateDefaultAdmin 创建默认管理员。
//
// 密码来源（优先级）：FCB_ADMIN_PASSWORD 环境变量 > 默认 admin123。
// 密码用 bcrypt 现场哈希（此前硬编码的哈希与 admin123 不匹配，导致管理员无法登录）。
// 生产环境务必通过 FCB_ADMIN_PASSWORD 注入强密码并在首次登录后修改。
func CreateDefaultAdmin(database *gorm.DB) error {
	var count int64
	database.Model(&model.User{}).Where("role = ?", "admin").Count(&count)

	if count > 0 {
		log.Println("Admin user already exists")
		return nil
	}

	// 密码：env 注入优先，否则默认 admin123
	password := os.Getenv("FCB_ADMIN_PASSWORD")
	if password == "" {
		password = "admin123"
	}

	// 现场生成 bcrypt 哈希（避免硬编码哈希与明文不一致）
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	admin := &model.User{
		Username:     "admin",
		Email:        "admin@filecodebox.local",
		PasswordHash: string(hashed),
		Nickname:     "Administrator",
		Role:         "admin",
		Status:       "active",
	}

	if err := database.Create(admin).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	if os.Getenv("FCB_ADMIN_PASSWORD") == "" {
		logger.Warn("Default admin created with default password 'admin123' — change it immediately in production (set FCB_ADMIN_PASSWORD for a custom one)")
	} else {
		logger.Info("Default admin created with password from FCB_ADMIN_PASSWORD")
	}
	return nil
}

var (
	database *gorm.DB
	config   *Config
)

// Bootstrap 应用程序启动入口。
// configPath 为空时依次回退到 CONFIG_PATH 环境变量、默认 configs/config.yaml。
func Bootstrap(configPath string) (*server.Hertz, error) {
	// 1. 初始化配置
	var err error
	config, err = InitConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to init config: %w", err)
	}

	// 设置全局配置（供其他包访问）
	conf.SetGlobalConfig(config)

	// 1.1 注入 JWT secret 到 auth 包（覆盖硬编码默认值）。
	// 此前 jwt.go 使用硬编码 "FileCodeBox2025SecretKey"，且与 config 的 jwt_secret
	// 不一致，导致配置中的 secret 从未生效。此处统一从配置/env 读取。
	auth.SetJWTSecret(config.User.JWTSecret)

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
	logger.Info("JWT secret loaded from configuration",
		zap.Bool("production", config.IsProduction()))

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

	// 可观测性：初始化 Prometheus 指标（在注册中间件前完成）
	if config.Observability.Metrics.Enabled {
		middleware.NewMetrics()
		logger.Info("Prometheus metrics enabled",
			zap.String("path", config.Observability.Metrics.Path))
	}

	// 安全：配置安全响应头（HSTS 仅在显式启用时开启，避免非 HTTPS 部署锁死）
	middleware.SetSecurityHeadersConfig(middleware.SecurityHeadersConfig{
		EnableHSTS: config.Security.CORS.EnableHSTS,
	})
	if config.Security.CORS.EnableHSTS {
		logger.Info("HSTS enabled (ensure HTTPS deployment)")
	}

	// 全局中间件链（顺序敏感）：
	//   Recovery → RequestID → AccessLog → Metrics → SecurityHeaders → CORS → handler
	//   - Recovery 最外层，捕获任意 panic 转 500，避免进程崩溃
	//   - RequestID 生成/透传 trace_id（X-Trace-Id），写入 ctx 供下游使用
	//   - AccessLog 结构化访问日志（带 trace_id），在 handler 执行后记录 status
	//   - Metrics 采集 RED 指标（延迟/计数/在途），在 handler 执行后记录
	//   - SecurityHeaders 设置 X-Content-Type-Options / X-Frame-Options / HSTS 等
	//   - CORS 跨域，最贴近 handler
	h.Use(middleware.Recovery())
	h.Use(middleware.RequestID())
	h.Use(middleware.AccessLog())
	if config.Observability.Metrics.Enabled {
		h.Use(middleware.MetricsMiddleware())
	}
	h.Use(middleware.SecurityHeaders())
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

	// ===== 公开配置端点（前端 configStore 启动时拉取）=====
	// 前端 publicApi.getConfig() 请求 /api/config 获取站点配置（名称、上传限制等），
	// 此前端点缺失导致前端启动报 "获取配置失败: Network Error"。此处补齐。
	r.GET("/api/config", publicConfigHandler)

	// ===== 前端构建产物静态资源 =====
	// Vite 输出的 index.html 用根级绝对路径引用资源（/assets/xxx、/vite.svg），
	// 故把 ./static/assets 挂到 /assets。用 StaticFS 正确处理 Range 请求、MIME、
	// 缓存（NoRoute 里手写的 c.File 对大文件 ES module 的 Range/缓冲处理不够稳定，
	// 会导致浏览器 "Failed to fetch dynamically imported module"）。
	r.StaticFS("/assets", &app.FS{
		Root:          "./static/assets",
		PathRewrite:   app.NewPathSlashesStripper(1),
		CacheDuration: 7 * 24 * time.Hour,
	})

	// ===== Prometheus 指标端点 =====
	if config.Observability.Metrics.Enabled {
		metricsPath := config.Observability.Metrics.Path
		if metricsPath == "" {
			metricsPath = "/metrics"
		}
		r.GET(metricsPath, metricsHandler)
	}

	// ===== 深度就绪检查 =====
	// /readyz 检查 DB 等依赖连通性，供 K8s readinessProbe 使用；
	// 依赖不可用时返回 503，避免流量打到未就绪实例。
	// （IDL 生成的 /ready 为轻量 stub，此处用 /readyz 做深度检查以避免路由冲突）
	r.GET("/readyz", readinessHandler)

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
	// Vite 构建的 index.html 使用根级绝对路径引用资源（/assets/xxx.js、/vite.svg），
	// 因此不能把资源挂在 /static 前缀下。这里采用"文件优先 + SPA 回退"策略：
	//   1. 静态资源请求（/assets/*、/vite.svg 等带扩展名路径）→ 从 ./static 读取
	//   2. 非 API 的其他路径 → 回退 index.html（交给前端 hash 路由）
	//   3. API 路径未命中 → 404 JSON
	r.NoRoute(func(ctx context.Context, c *app.RequestContext) {
		path := string(c.Request.URI().Path())
		if isAPIPath(path) {
			c.JSON(consts.StatusNotFound, map[string]interface{}{
				"code":    404,
				"message": "API endpoint not found",
			})
			return
		}
		// 静态资源：尝试从 ./static 下读取（去掉前导 /）
		if tryServeStatic(c, path) {
			return
		}
		// 其余路径回退到 SPA index.html
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
	"/health", "/live", "/ready", "/readyz", "/ping", "/version",
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

// staticFileExtensions 视为静态资源（而非 SPA 路由）的文件扩展名。
// Vite 构建产物（js/css/图片/字体等）走这里直接返回文件。
var staticFileExtensions = map[string]bool{
	".js": true, ".mjs": true, ".css": true,
	".html": true, ".svg": true, ".png": true, ".jpg": true, ".jpeg": true,
	".gif": true, ".ico": true, ".webp": true, ".woff": true, ".woff2": true,
	".ttf": true, ".eot": true, ".map": true, ".json": true, ".txt": true,
}

// tryServeStatic 尝试从 ./static 目录服务静态资源。
// 命中（文件存在且是静态资源扩展名）时写入响应并返回 true，否则返回 false。
// 用于在 NoRoute 中优先服务 Vite 构建产物（/assets/xxx.js 等），再回退 SPA。
func tryServeStatic(c *app.RequestContext, path string) bool {
	// 仅对带静态资源扩展名的路径尝试（避免目录穿越和无谓的文件系统查找）
	ext := strings.ToLower(filepath.Ext(path))
	if !staticFileExtensions[ext] {
		return false
	}
	// 去掉前导 /，拼接到 static 根目录；filepath.Join 会清理 ../ 等穿越
	rel := strings.TrimPrefix(path, "/")
	fullPath := filepath.Join("./static", rel)
	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		return false
	}
	c.File(fullPath)
	return true
}

// publicConfigHandler 返回公开配置（前端 configStore 启动时拉取）。
// 返回结构与前端 PublicConfig 接口对齐：
// name / description / uploadSize / enableChunk / openUpload / expireStyle。
func publicConfigHandler(ctx context.Context, c *app.RequestContext) {
	resp.Success(c, map[string]interface{}{
		"name":        config.App.Name,
		"description": config.App.Description,
		"uploadSize":  config.Upload.UploadSize,
		"enableChunk": config.Upload.EnableChunk,
		"openUpload":  config.Upload.OpenUpload,
		// 前端 expireStyle 下拉选项（与 utils.CalculateExpireTime 支持的风格对齐）
		"expireStyle": []string{"minute", "hour", "day", "week", "month", "year", "forever"},
		"initialized": true,
	})
}

// metricsHandler 暴露 Prometheus 指标（/metrics）。
// Hertz 与标准 net/http 接口不同，不能直接用 promhttp.Handler()，
// 这里手动 gather 指标并用 expfmt 文本格式输出到 buffer 再写入响应。
func metricsHandler(ctx context.Context, c *app.RequestContext) {
	mfs, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]interface{}{
			"code":    500,
			"message": "failed to gather metrics: " + err.Error(),
		})
		return
	}
	var buf bytes.Buffer
	enc := expfmt.NewEncoder(&buf, expfmt.NewFormat(expfmt.TypeTextPlain))
	for _, mf := range mfs {
		if err := enc.Encode(mf); err != nil {
			logger.Error("failed to encode metric", zap.String("name", mf.GetName()), zap.Error(err))
		}
	}
	c.SetStatusCode(consts.StatusOK)
	c.Response.Header.SetContentType(string(expfmt.NewFormat(expfmt.TypeTextPlain)))
	_, _ = c.Write(buf.Bytes())
}

// readinessHandler 深度就绪检查：校验 DB 连通性，失败返回 503。
// 供 K8s readinessProbe 使用——依赖未就绪时不接流量。
func readinessHandler(ctx context.Context, c *app.RequestContext) {
	checks := map[string]bool{}
	allOK := true

	// DB ping
	if database != nil {
		if sqlDB, err := database.DB(); err == nil {
			if err := sqlDB.Ping(); err == nil {
				checks["database"] = true
			} else {
				checks["database"] = false
				allOK = false
			}
		} else {
			checks["database"] = false
			allOK = false
		}
	} else {
		checks["database"] = false
		allOK = false
	}

	status := "ready"
	httpStatus := consts.StatusOK
	if !allOK {
		status = "not ready"
		httpStatus = consts.StatusServiceUnavailable
	}
	c.JSON(httpStatus, map[string]interface{}{
		"status": status,
		"checks": checks,
	})
}


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
	// baseURL 优先用配置的对外地址（server.base_url），否则用 host:port
	baseURL := config.Server.BaseURL
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://%s:%d", config.Server.Host, config.Server.Port)
	}
	// presign 签名密钥：优先专用 FCB_PRESIGN_SIGNING_KEY，否则复用 jwt_secret
	signingKey := os.Getenv("FCB_PRESIGN_SIGNING_KEY")
	if signingKey == "" {
		signingKey = config.User.JWTSecret
	}
	presignHandler.SetService(redis.GetClient(),
		baseURL,
		signingKey)
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

	// 6. 注入 storage 到 admin handler 的 service（过期清理删物理文件）
	bootstrapStorage := getBootstrapStorageService()
	adminHandler.SetStorage(bootstrapStorage)

	// 7. 启动过期文件定时清理（默认每小时，删 DB 记录 + 物理文件）
	//    独立 admin service 实例（避免与 handler 实例竞争），注入 storage
	cleanupSvc := adminApp.NewService()
	cleanupSvc.SetStorage(bootstrapStorage)
	go startExpiredFileCleanup(cleanupSvc)
}

// startExpiredFileCleanup 定时清理过期文件（DB 记录 + 物理文件）。
// 默认每 1 小时执行一次；懒清理由取件路径覆盖（GetFileByCode 发现过期即返回错误）。
func startExpiredFileCleanup(svc *adminApp.Service) {
	interval := time.Hour
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	ctx := context.Background()
	for range ticker.C {
		if n, freed, err := svc.CleanExpiredFiles(ctx); err != nil {
			logger.Error("expired file cleanup failed", zap.Error(err))
		} else if n > 0 {
			logger.Info("expired files cleaned",
				zap.Int64("count", n),
				zap.Int64("freed_bytes", freed))
		}
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
