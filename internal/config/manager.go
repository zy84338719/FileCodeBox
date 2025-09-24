package config

import (
	"errors"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"gorm.io/gorm"
)

// ConfigManager 配置管理器，实现 环境变量 > YAML > 默认值 的优先级语义
type ConfigManager struct {
	// === 核心配置模块 ===
	Base     *BaseConfig       `json:"base" yaml:"base"`
	Database *DatabaseConfig   `json:"database" yaml:"database"`
	Transfer *TransferConfig   `json:"transfer" yaml:"transfer"`
	Storage  *StorageConfig    `json:"storage" yaml:"storage"`
	User     *UserSystemConfig `json:"user" yaml:"user"`
	MCP      *MCPConfig        `json:"mcp" yaml:"mcp"`
	UI       *UIConfig         `json:"ui" yaml:"ui"`

	// === 通知配置 ===
	NotifyTitle   string `json:"notify_title" yaml:"notify_title"`
	NotifyContent string `json:"notify_content" yaml:"notify_content"`

	// === 运行时元数据 ===
	SysStart string `json:"sys_start" yaml:"sys_start"`

	// === 业务配置（保持顶层以兼容性） ===
	UploadMinute int      `json:"upload_minute" yaml:"upload_minute"`
	UploadCount  int      `json:"upload_count" yaml:"upload_count"`
	ErrorMinute  int      `json:"error_minute" yaml:"error_minute"`
	ErrorCount   int      `json:"error_count" yaml:"error_count"`
	ExpireStyle  []string `json:"expire_style" yaml:"expire_style"`

	// === 内部状态 ===
	db *gorm.DB
}

func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		Base:     NewBaseConfig(),
		Database: NewDatabaseConfig(),
		Transfer: NewTransferConfig(),
		Storage:  NewStorageConfig(),
		User:     NewUserSystemConfig(),
		MCP:      NewMCPConfig(),
		UI:       &UIConfig{},
	}
}

// InitManager 初始化配置管理器，加载 YAML 配置并应用环境变量覆盖
func InitManager() *ConfigManager {
	cm := NewConfigManager()

	var sources []ConfigSource

	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		sources = append(sources, YAMLFileSource{Path: configPath})
	} else if _, err := os.Stat("./config.yaml"); err == nil {
		sources = append(sources, YAMLFileSource{Path: "./config.yaml"})
	}

	sources = append(sources, NewDefaultEnvSource())

	_ = cm.ApplySources(sources...)
	return cm
}

func (cm *ConfigManager) SetDB(db *gorm.DB) { cm.db = db }
func (cm *ConfigManager) GetDB() *gorm.DB   { return cm.db }

// mergeUserConfig 合并用户配置，避免覆盖默认值
func (cm *ConfigManager) mergeUserConfig(fileUser *UserSystemConfig) {
	if fileUser == nil {
		return
	}

	if fileUser.AllowUserRegistration != 0 {
		cm.User.AllowUserRegistration = fileUser.AllowUserRegistration
	}
	if fileUser.RequireEmailVerify != 0 {
		cm.User.RequireEmailVerify = fileUser.RequireEmailVerify
	}
	if fileUser.UserUploadSize != 0 {
		cm.User.UserUploadSize = fileUser.UserUploadSize
	}
	if fileUser.UserStorageQuota != 0 {
		cm.User.UserStorageQuota = fileUser.UserStorageQuota
	}
	if fileUser.SessionExpiryHours != 0 {
		cm.User.SessionExpiryHours = fileUser.SessionExpiryHours
	}
	if fileUser.MaxSessionsPerUser != 0 {
		cm.User.MaxSessionsPerUser = fileUser.MaxSessionsPerUser
	}
	if strings.TrimSpace(fileUser.JWTSecret) != "" {
		cm.User.JWTSecret = fileUser.JWTSecret
	}
}

// mergeConfigModules 合并配置模块
func (cm *ConfigManager) mergeConfigModules(fileCfg *ConfigManager) {
	if fileCfg.Base != nil {
		cm.Base = fileCfg.Base
	}
	if fileCfg.Database != nil {
		cm.Database = fileCfg.Database
	}
	if fileCfg.Transfer != nil {
		cm.Transfer = fileCfg.Transfer
	}
	if fileCfg.Storage != nil {
		cm.Storage = fileCfg.Storage
	}
	if fileCfg.MCP != nil {
		cm.MCP = fileCfg.MCP
	}
	if fileCfg.UI != nil {
		cm.UI = fileCfg.UI
	}
}

// mergeSimpleFields 合并简单字段
func (cm *ConfigManager) mergeSimpleFields(fileCfg *ConfigManager) {
	if fileCfg.NotifyTitle != "" {
		cm.NotifyTitle = fileCfg.NotifyTitle
	}
	if fileCfg.NotifyContent != "" {
		cm.NotifyContent = fileCfg.NotifyContent
	}
	if fileCfg.SysStart != "" {
		cm.SysStart = fileCfg.SysStart
	}
}

// ApplySources processes a group of configuration sources and collects errors.
func (cm *ConfigManager) ApplySources(sources ...ConfigSource) error {
	var errs []error
	for _, source := range sources {
		if source == nil {
			continue
		}
		if err := source.Apply(cm); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// LoadFromYAML 从 YAML 文件加载配置
func (cm *ConfigManager) LoadFromYAML(path string) error {
	return cm.ApplySources(YAMLFileSource{Path: path})
}

// ReloadConfig 重新加载配置（仅支持环境变量，保持端口不变）
func (cm *ConfigManager) ReloadConfig() error {
	// 保存当前端口设置
	currentPort := cm.Base.Port

	// 重新应用环境变量覆盖
	cm.applyEnvironmentOverrides()

	// 恢复端口设置（端口在运行时不可变）
	cm.Base.Port = currentPort
	return nil
}

// PersistYAML 将当前配置保存到 YAML 文件
func (cm *ConfigManager) PersistYAML() error {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config.yaml"
	}

	// 确保 UI 配置存在
	if cm.UI == nil {
		cm.UI = &UIConfig{}
	}

	data, err := yaml.Marshal(cm)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// applyEnvironmentOverrides 应用环境变量覆盖配置
func (cm *ConfigManager) applyEnvironmentOverrides() {
	// 收集错误以便在调用者中统一处理，保持现有签名
	_ = NewDefaultEnvSource().Apply(cm)
}

// Save 保存配置（已废弃，请使用 config.yaml 和环境变量）
func (cm *ConfigManager) Save() error {
	return errors.New("数据库配置保存已不支持，请使用 config.yaml 和环境变量")
}

// === 配置访问助手方法 ===

func (cm *ConfigManager) GetAddress() string              { return cm.Base.GetAddress() }
func (cm *ConfigManager) GetDatabaseDSN() (string, error) { return cm.Database.GetDSN() }
func (cm *ConfigManager) IsUserSystemEnabled() bool       { return cm.User.IsUserSystemEnabled() }
func (cm *ConfigManager) IsMCPEnabled() bool              { return cm.MCP.IsMCPEnabled() }

// Clone 创建配置管理器的深拷贝
func (cm *ConfigManager) Clone() *ConfigManager {
	newManager := NewConfigManager()

	// 克隆配置模块
	newManager.Base = cm.Base.Clone()
	newManager.Database = cm.Database.Clone()
	newManager.Transfer = cm.Transfer.Clone()
	newManager.Storage = cm.Storage.Clone()
	newManager.User = cm.User.Clone()
	newManager.MCP = cm.MCP.Clone()

	// 克隆简单字段
	newManager.NotifyTitle = cm.NotifyTitle
	newManager.NotifyContent = cm.NotifyContent
	newManager.SysStart = cm.SysStart

	// 克隆 UI 配置
	if cm.UI != nil {
		ui := *cm.UI
		newManager.UI = &ui
	}

	return newManager
}

// Validate 验证所有配置模块
func (cm *ConfigManager) Validate() error {
	// 检查配置模块是否初始化
	if cm.Base == nil || cm.Database == nil || cm.Transfer == nil ||
		cm.Storage == nil || cm.User == nil || cm.MCP == nil {
		return errors.New("配置模块未完全初始化")
	}

	// 验证关键配置模块
	if err := cm.Base.Validate(); err != nil {
		return err
	}
	if err := cm.Database.Validate(); err != nil {
		return err
	}

	return nil
}
