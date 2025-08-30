// Package config 处理应用程序配置管理
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/zy84338719/filecodebox/internal/models"

	"gorm.io/gorm"
)

type Config struct {
	// 基础配置
	Name        string `json:"name"`
	Description string `json:"description"`
	Keywords    string `json:"keywords"`
	Port        int    `json:"port"`
	Host        string `json:"host"` // 新增：绑定IP地址
	DataPath    string `json:"data_path"`
	Production  bool   `json:"production"`

	// 通知配置
	NotifyTitle   string `json:"notify_title"`
	NotifyContent string `json:"notify_content"`
	PageExplain   string `json:"page_explain"`

	// 上传配置
	OpenUpload     int   `json:"open_upload"`
	UploadSize     int64 `json:"upload_size"`
	EnableChunk    int   `json:"enable_chunk"`
	ChunkSize      int64 `json:"chunk_size"` // 新增：分片大小
	MaxSaveSeconds int   `json:"max_save_seconds"`

	// 下载配置 (新增)
	EnableConcurrentDownload int `json:"enable_concurrent_download"` // 是否启用并发下载
	MaxConcurrentDownloads   int `json:"max_concurrent_downloads"`   // 最大并发下载数
	DownloadTimeout          int `json:"download_timeout"`           // 下载超时时间(秒)

	// 过期样式
	ExpireStyle []string `json:"expire_style"`

	// 限流配置
	UploadMinute int `json:"upload_minute"`
	UploadCount  int `json:"upload_count"`
	ErrorMinute  int `json:"error_minute"`
	ErrorCount   int `json:"error_count"`

	// 主题配置
	ThemesSelect  string  `json:"themes_select"`
	ThemesChoices []Theme `json:"themes_choices"`
	Opacity       float64 `json:"opacity"`
	Background    string  `json:"background"`

	// 存储配置
	FileStorage string `json:"file_storage"`
	StoragePath string `json:"storage_path"`

	// S3配置
	S3AccessKeyID      string `json:"s3_access_key_id"`
	S3SecretAccessKey  string `json:"s3_secret_access_key"`
	S3BucketName       string `json:"s3_bucket_name"`
	S3EndpointURL      string `json:"s3_endpoint_url"`
	S3RegionName       string `json:"s3_region_name"`
	S3SignatureVersion string `json:"s3_signature_version"`
	S3Hostname         string `json:"s3_hostname"`
	S3Proxy            int    `json:"s3_proxy"`
	AWSSessionToken    string `json:"aws_session_token"`

	// WebDAV配置
	WebDAVHostname string `json:"webdav_hostname"`
	WebDAVRootPath string `json:"webdav_root_path"`
	WebDAVProxy    int    `json:"webdav_proxy"`
	WebDAVURL      string `json:"webdav_url"`
	WebDAVPassword string `json:"webdav_password"`
	WebDAVUsername string `json:"webdav_username"`

	// OneDrive配置
	OneDriveDomain   string `json:"onedrive_domain"`
	OneDriveClientID string `json:"onedrive_client_id"`
	OneDriveUsername string `json:"onedrive_username"`
	OneDrivePassword string `json:"onedrive_password"`
	OneDriveRootPath string `json:"onedrive_root_path"`
	OneDriveProxy    int    `json:"onedrive_proxy"`

	// 管理配置
	AdminToken    string `json:"admin_token"`
	ShowAdminAddr int    `json:"show_admin_address"`
	RobotsText    string `json:"robots_text"`

	// 用户系统配置 (新增)
	EnableUserSystem      int   `json:"enable_user_system"`      // 是否启用用户系统 0-禁用 1-启用
	AllowUserRegistration int   `json:"allow_user_registration"` // 是否允许用户注册 0-禁用 1-启用
	RequireEmailVerify    int   `json:"require_email_verify"`    // 是否需要邮箱验证 0-不需要 1-需要
	UserUploadSize        int64 `json:"user_upload_size"`        // 用户上传文件大小限制(字节)
	UserStorageQuota      int64 `json:"user_storage_quota"`      // 用户存储配额(字节) 0-无限制
	SessionExpiryHours    int   `json:"session_expiry_hours"`    // 用户会话过期时间(小时)
	MaxSessionsPerUser    int   `json:"max_sessions_per_user"`   // 每个用户最大会话数

	// JWT 密钥配置
	JWTSecret string `json:"jwt_secret"` // JWT签名密钥

	// 数据库连接（内部使用，不保存到JSON）
	db *gorm.DB `json:"-"`
}

type Theme struct {
	Name    string `json:"name"`
	Key     string `json:"key"`
	Author  string `json:"author"`
	Version string `json:"version"`
}

var defaultConfig = Config{
	Name:        "文件快递柜 - FileCodeBox",
	Description: "开箱即用的文件快传系统",
	Keywords:    "FileCodeBox, 文件快递柜, 口令传送箱, 匿名口令分享文本, 文件",
	Port:        12345,
	Host:        "0.0.0.0", // 默认绑定所有IP
	DataPath:    "./data",
	Production:  false,

	NotifyTitle:   "系统通知",
	NotifyContent: `欢迎使用 FileCodeBox，本程序开源于 <a href="https://github.com/vastsa/FileCodeBox" target="_blank">Github</a> ，欢迎Star和Fork。`,
	PageExplain:   "请勿上传或分享违法内容。根据《中华人民共和国网络安全法》、《中华人民共和国刑法》、《中华人民共和国治安管理处罚法》等相关规定。 传播或存储违法、违规内容，会受到相关处罚，严重者将承担刑事责任。本站坚决配合相关部门，确保网络内容的安全，和谐，打造绿色网络环境。",

	OpenUpload:     1,
	UploadSize:     10 * 1024 * 1024, // 10MB
	EnableChunk:    0,
	ChunkSize:      2 * 1024 * 1024, // 2MB 默认分片大小
	MaxSaveSeconds: 0,

	// 并发下载配置
	EnableConcurrentDownload: 1,   // 默认启用
	MaxConcurrentDownloads:   10,  // 最大10个并发下载
	DownloadTimeout:          300, // 5分钟超时

	ExpireStyle: []string{"day", "hour", "minute", "forever", "count"},

	UploadMinute: 1,
	UploadCount:  10,
	ErrorMinute:  1,
	ErrorCount:   1,

	ThemesSelect: "themes/2024",
	ThemesChoices: []Theme{
		// {Name: "2023", Key: "themes/2023", Author: "Lan", Version: "1.0"},
		{Name: "2024", Key: "themes/2024", Author: "Lan", Version: "1.0"},
	},
	Opacity:    0.9,
	Background: "",

	FileStorage: "local",
	StoragePath: "",

	S3RegionName:       "auto",
	S3SignatureVersion: "s3v2",

	OneDriveRootPath: "filebox_storage",
	WebDAVRootPath:   "filebox_storage",

	AdminToken:    "FileCodeBox2025",
	ShowAdminAddr: 0,
	RobotsText:    "User-agent: *\nDisallow: /",

	// 用户系统默认配置
	EnableUserSystem:      0,                    // 默认禁用用户系统
	AllowUserRegistration: 1,                    // 允许用户注册
	RequireEmailVerify:    0,                    // 不要求邮箱验证
	UserUploadSize:        50 * 1024 * 1024,     // 用户上传限制50MB
	UserStorageQuota:      1024 * 1024 * 1024,   // 用户存储配额1GB
	SessionExpiryHours:    24 * 7,               // 会话7天过期
	MaxSessionsPerUser:    5,                    // 每用户最多5个会话
	JWTSecret:             "FileCodeBox2025JWT", // JWT密钥
}

func Init() *Config {
	config := defaultConfig

	// 创建数据目录
	if err := os.MkdirAll(config.DataPath, 0750); err != nil {
		panic("创建数据目录失败: " + err.Error())
	}

	return &config
}

// InitWithDB 使用数据库初始化配置
func (c *Config) InitWithDB(db *gorm.DB) error {
	c.db = db

	// 尝试从数据库加载配置
	if err := c.LoadFromDatabase(); err != nil {
		// 数据库中没有配置，初始化基础数据
		if err := c.InitDefaultDataInDB(); err != nil {
			return fmt.Errorf("初始化数据库默认配置失败: %w", err)
		}
		// 重新加载配置
		if err := c.LoadFromDatabase(); err != nil {
			return fmt.Errorf("加载初始化后的配置失败: %w", err)
		}
	}

	return nil
}

// buildConfigMap 构建配置映射表
func (c *Config) buildConfigMap() map[string]string {
	return map[string]string{
		"name":                       c.Name,
		"description":                c.Description,
		"host":                       c.Host,
		"upload_size":                fmt.Sprintf("%d", c.UploadSize),
		"admin_token":                c.AdminToken,
		"storage_path":               c.StoragePath,
		"open_upload":                fmt.Sprintf("%d", c.OpenUpload),
		"enable_chunk":               fmt.Sprintf("%d", c.EnableChunk),
		"chunk_size":                 fmt.Sprintf("%d", c.ChunkSize),
		"enable_concurrent_download": fmt.Sprintf("%d", c.EnableConcurrentDownload),
		"max_concurrent_downloads":   fmt.Sprintf("%d", c.MaxConcurrentDownloads),
		"download_timeout":           fmt.Sprintf("%d", c.DownloadTimeout),
		"notify_title":               c.NotifyTitle,
		"notify_content":             c.NotifyContent,
		"page_explain":               c.PageExplain,
		"themes_select":              c.ThemesSelect,
		"file_storage":               c.FileStorage,
		"webdav_hostname":            c.WebDAVHostname,
		"webdav_username":            c.WebDAVUsername,
		"webdav_password":            c.WebDAVPassword,
		"webdav_root_path":           c.WebDAVRootPath,
		"webdav_url":                 c.WebDAVURL,
		"s3_access_key_id":           c.S3AccessKeyID,
		"s3_secret_access_key":       c.S3SecretAccessKey,
		"s3_bucket_name":             c.S3BucketName,
		"s3_endpoint_url":            c.S3EndpointURL,
		"s3_region_name":             c.S3RegionName,
		"s3_hostname":                c.S3Hostname,
		"s3_signature_version":       c.S3SignatureVersion,
		"s3_proxy":                   fmt.Sprintf("%d", c.S3Proxy),

		// 用户系统配置
		"enable_user_system":      fmt.Sprintf("%d", c.EnableUserSystem),
		"allow_user_registration": fmt.Sprintf("%d", c.AllowUserRegistration),
		"require_email_verify":    fmt.Sprintf("%d", c.RequireEmailVerify),
		"user_upload_size":        fmt.Sprintf("%d", c.UserUploadSize),
		"user_storage_quota":      fmt.Sprintf("%d", c.UserStorageQuota),
		"session_expiry_hours":    fmt.Sprintf("%d", c.SessionExpiryHours),
		"max_sessions_per_user":   fmt.Sprintf("%d", c.MaxSessionsPerUser),
		"jwt_secret":              c.JWTSecret,
	}
}

// InitDefaultDataInDB 在数据库中初始化默认配置数据
func (c *Config) InitDefaultDataInDB() error {
	if c.db == nil {
		return errors.New("数据库连接未设置")
	}

	// 检查是否已经有配置数据
	var count int64
	if err := c.db.Model(&models.KeyValue{}).Count(&count).Error; err != nil {
		return fmt.Errorf("检查配置数据失败: %w", err)
	}

	// 如果已有数据，不进行初始化
	if count > 0 {
		return nil
	}

	// 使用公共方法获取配置映射
	defaultConfigs := c.buildConfigMap()

	// 批量插入默认配置
	var keyValues []models.KeyValue
	for key, value := range defaultConfigs {
		keyValues = append(keyValues, models.KeyValue{
			Key:   key,
			Value: value,
		})
	}

	if err := c.db.CreateInBatches(keyValues, 50).Error; err != nil {
		return fmt.Errorf("插入默认配置失败: %w", err)
	}

	return nil
}

// SetDB 设置数据库连接
func (c *Config) SetDB(db *gorm.DB) {
	c.db = db
}

func (c *Config) Save() error {
	// 只保存到数据库，不再保存到文件
	if c.db != nil {
		return c.saveToDatabase()
	}

	return errors.New("数据库连接未设置，无法保存配置")
}

// saveToDatabase 保存配置到数据库
func (c *Config) saveToDatabase() error {
	if c.db == nil {
		return errors.New("数据库连接未设置")
	}

	// 使用公共方法获取配置映射
	configMap := c.buildConfigMap()

	for key, value := range configMap {
		kv := models.KeyValue{
			Key:   key,
			Value: value,
		}

		// 使用 UPSERT 操作
		if err := c.db.Where("key = ?", key).Assign(models.KeyValue{Value: value}).FirstOrCreate(&kv).Error; err != nil {
			return fmt.Errorf("保存配置项 %s 失败: %w", key, err)
		}
	}

	return nil
} // LoadFromDatabase 从数据库加载配置
func (c *Config) LoadFromDatabase() error {
	if c.db == nil {
		return errors.New("数据库连接未设置")
	}

	var kvPairs []models.KeyValue
	if err := c.db.Find(&kvPairs).Error; err != nil {
		return fmt.Errorf("查询配置失败: %w", err)
	}

	// 如果数据库中没有配置，返回错误以触发初始化
	if len(kvPairs) == 0 {
		return fmt.Errorf("数据库中没有配置数据")
	}

	for _, kv := range kvPairs {
		switch kv.Key {
		case "name":
			c.Name = kv.Value
		case "description":
			c.Description = kv.Value
		case "upload_size":
			if size, err := strconv.ParseInt(kv.Value, 10, 64); err == nil {
				c.UploadSize = size
			}
		case "admin_token":
			c.AdminToken = kv.Value
		case "storage_path":
			c.StoragePath = kv.Value
		case "open_upload":
			if val, err := strconv.Atoi(kv.Value); err == nil {
				c.OpenUpload = val
			}
		case "enable_chunk":
			if val, err := strconv.Atoi(kv.Value); err == nil {
				c.EnableChunk = val
			}
		case "notify_title":
			c.NotifyTitle = kv.Value
		case "notify_content":
			c.NotifyContent = kv.Value
		case "page_explain":
			c.PageExplain = kv.Value
		case "themes_select":
			c.ThemesSelect = kv.Value
		case "file_storage":
			c.FileStorage = kv.Value
		case "webdav_hostname":
			c.WebDAVHostname = kv.Value
		case "webdav_username":
			c.WebDAVUsername = kv.Value
		case "webdav_password":
			c.WebDAVPassword = kv.Value
		case "webdav_root_path":
			c.WebDAVRootPath = kv.Value
		case "webdav_url":
			c.WebDAVURL = kv.Value
		case "s3_access_key_id":
			c.S3AccessKeyID = kv.Value
		case "s3_secret_access_key":
			c.S3SecretAccessKey = kv.Value
		case "s3_bucket_name":
			c.S3BucketName = kv.Value
		case "s3_endpoint_url":
			c.S3EndpointURL = kv.Value
		case "s3_region_name":
			c.S3RegionName = kv.Value
		case "s3_hostname":
			c.S3Hostname = kv.Value
		case "host":
			c.Host = kv.Value
		case "chunk_size":
			if size, err := strconv.ParseInt(kv.Value, 10, 64); err == nil {
				c.ChunkSize = size
			}
		case "enable_concurrent_download":
			if val, err := strconv.Atoi(kv.Value); err == nil {
				c.EnableConcurrentDownload = val
			}
		case "max_concurrent_downloads":
			if val, err := strconv.Atoi(kv.Value); err == nil {
				c.MaxConcurrentDownloads = val
			}
		case "download_timeout":
			if val, err := strconv.Atoi(kv.Value); err == nil {
				c.DownloadTimeout = val
			}
		// 用户系统配置加载
		case "enable_user_system":
			if val, err := strconv.Atoi(kv.Value); err == nil {
				c.EnableUserSystem = val
			}
		case "allow_user_registration":
			if val, err := strconv.Atoi(kv.Value); err == nil {
				c.AllowUserRegistration = val
			}
		case "require_email_verify":
			if val, err := strconv.Atoi(kv.Value); err == nil {
				c.RequireEmailVerify = val
			}
		case "user_upload_size":
			if size, err := strconv.ParseInt(kv.Value, 10, 64); err == nil {
				c.UserUploadSize = size
			}
		case "user_storage_quota":
			if size, err := strconv.ParseInt(kv.Value, 10, 64); err == nil {
				c.UserStorageQuota = size
			}
		case "session_expiry_hours":
			if val, err := strconv.Atoi(kv.Value); err == nil {
				c.SessionExpiryHours = val
			}
		case "max_sessions_per_user":
			if val, err := strconv.Atoi(kv.Value); err == nil {
				c.MaxSessionsPerUser = val
			}
		case "jwt_secret":
			c.JWTSecret = kv.Value
		}
	}

	return nil
}
