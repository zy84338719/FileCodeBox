package bootstrap

import (
	"fmt"
	"log"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/spf13/viper"
	"github.com/zy84338719/fileCodeBox/biz/router"
	"github.com/zy84338719/fileCodeBox/internal/conf"
	"github.com/zy84338719/fileCodeBox/internal/pkg/logger"
	previewPkg "github.com/zy84338719/fileCodeBox/internal/preview"
	"github.com/zy84338719/fileCodeBox/internal/repo/db"
	"github.com/zy84338719/fileCodeBox/internal/repo/db/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 使用 internal/conf 包中的统一配置类型
type Config = conf.AppConfiguration

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
	config, err = InitConfig("./conf/config.yaml")
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

	// 5. 创建 HTTP 服务器
	port := config.Server.Port
	if port == 0 {
		port = 12345
	}
	h := server.New(
		server.WithHostPorts(fmt.Sprintf("%s:%d", config.Server.Host, port)),
	)

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

// customizedRegister registers customize routers.
func customizedRegister(r *server.Hertz) {
	// 这里可以添加自定义路由，现在为空，所有的路由都通过 GeneratedRegister 注册
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
