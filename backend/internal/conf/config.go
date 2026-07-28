package conf

import "fmt"

// 全局配置
var globalConfig *AppConfiguration

// AppConfiguration 完整应用配置
type AppConfiguration struct {
	Server        ServerConfig        `mapstructure:"server"`
	Database      DatabaseConfig      `mapstructure:"database"`
	Redis         RedisConfig         `mapstructure:"redis"`
	Log           LogConfig           `mapstructure:"log"`
	App           AppConfig           `mapstructure:"app"`
	User          UserConfig          `mapstructure:"user"`
	Upload        UploadConfig        `mapstructure:"upload"`
	Download      DownloadConfig      `mapstructure:"download"`
	Storage       StorageConfig       `mapstructure:"storage"`
	UI            UIConfig            `mapstructure:"ui"`
	Observability ObservabilityConfig `mapstructure:"observability"`
}

// SetGlobalConfig 设置全局配置
func SetGlobalConfig(cfg *AppConfiguration) {
	globalConfig = cfg
}

// GetGlobalConfig 获取全局配置
func GetGlobalConfig() *AppConfiguration {
	return globalConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Mode         string `mapstructure:"mode"`     // debug, release, test
	BaseURL      string `mapstructure:"base_url"` // 对外可访问的基础 URL（用于生成分享链接等）
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"` // sqlite, mysql, postgres
	DBName   string `mapstructure:"db_name"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// Addr 返回 Redis 地址
func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Description string `mapstructure:"description"`
	DataPath    string `mapstructure:"datapath"`
	Production  bool   `mapstructure:"production"`
}

// UserConfig 用户配置
type UserConfig struct {
	AllowUserRegistration bool   `mapstructure:"allow_user_registration"`
	RequireEmailVerify    bool   `mapstructure:"require_email_verify"`
	UserUploadSize        int64  `mapstructure:"user_upload_size"`
	UserStorageQuota      int64  `mapstructure:"user_storage_quota"`
	SessionExpiryHours    int    `mapstructure:"session_expiry_hours"`
	MaxSessionsPerUser    int    `mapstructure:"max_sessions_per_user"`
	JWTSecret             string `mapstructure:"jwt_secret"`
}

// UploadConfig 上传配置
type UploadConfig struct {
	OpenUpload     bool  `mapstructure:"open_upload"`
	UploadSize     int64 `mapstructure:"upload_size"`
	EnableChunk    bool  `mapstructure:"enable_chunk"`
	ChunkSize      int64 `mapstructure:"chunk_size"`
	MaxSaveSeconds int   `mapstructure:"max_save_seconds"`
	RequireLogin   bool  `mapstructure:"require_login"`
}

// DownloadConfig 下载配置
type DownloadConfig struct {
	EnableConcurrentDownload bool `mapstructure:"enable_concurrent_download"`
	MaxConcurrentDownloads   int  `mapstructure:"max_concurrent_downloads"`
	DownloadTimeout          int  `mapstructure:"download_timeout"`
	RequireLogin             bool `mapstructure:"require_login"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type        string       `mapstructure:"type"` // local, s3, webdav, onedrive, nfs
	StoragePath string       `mapstructure:"storage_path"`
	S3          *S3Config    `mapstructure:"s3"`
	WebDAV      *WebDAVConfig `mapstructure:"webdav"`
}

// S3Config S3 兼容对象存储配置（AWS S3 / 阿里云 OSS / 腾讯云 COS 等）
type S3Config struct {
	Endpoint        string `mapstructure:"endpoint"`
	Region          string `mapstructure:"region"`
	Bucket          string `mapstructure:"bucket"`
	AccessKey       string `mapstructure:"access_key"`
	SecretKey       string `mapstructure:"secret_key"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	PathStyle       bool   `mapstructure:"path_style"`
}

// WebDAVConfig WebDAV 存储配置
type WebDAVConfig struct {
	Endpoint string `mapstructure:"endpoint"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// UIConfig 前端 UI 相关配置（透传给前端展示）
type UIConfig struct {
	Theme        string  `mapstructure:"theme"`
	Background   string  `mapstructure:"background"`
	PageExplain  string  `mapstructure:"page_explain"`
	RobotsText   string  `mapstructure:"robots_text"`
	ShowAdminAddr bool   `mapstructure:"show_admin_addr"`
	Opacity      float64 `mapstructure:"opacity"`
	NotifyTitle  string  `mapstructure:"notify_title"`
	NotifyContent string `mapstructure:"notify_content"`
}

// ObservabilityConfig 可观测性配置（metrics / tracing）
type ObservabilityConfig struct {
	Metrics MetricsConfig `mapstructure:"metrics"`
	Tracing TracingConfig `mapstructure:"tracing"`
}

// MetricsConfig Prometheus 指标配置
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"` // 指标暴露路径，默认 /metrics
}

// TracingConfig 分布式追踪配置（OpenTelemetry）
type TracingConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Exporter string `mapstructure:"exporter"` // otlp / stdout / 空表示禁用
	Endpoint string `mapstructure:"endpoint"` // OTLP collector 地址
	ServiceName string `mapstructure:"service_name"`
}

// IsProduction 是否生产模式（综合 server.mode 与 app.production 判断）
func (c *AppConfiguration) IsProduction() bool {
	return c.App.Production || c.Server.Mode == "release"
}
