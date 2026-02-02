package config

import (
	"errors"
	"fmt"
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
	var configPath string

	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		configPath = envPath
		sources = append(sources, YAMLFileSource{Path: configPath})
	} else {
		configPath = "./config.yaml"
		// 检查配置文件是否存在
		if _, err := os.Stat(configPath); err == nil {
			sources = append(sources, YAMLFileSource{Path: configPath})
		} else {
			// 配置文件不存在，自动生成默认配置
			if err := generateDefaultConfigFile(configPath, cm); err != nil {
				fmt.Fprintf(os.Stderr, "警告: 无法生成默认配置文件 %s: %v\n", configPath, err)
			} else {
				fmt.Printf("已生成默认配置文件: %s\n", configPath)
				// 生成后加载配置
				sources = append(sources, YAMLFileSource{Path: configPath})
			}
		}
	}

	sources = append(sources, NewDefaultEnvSource())

	_ = cm.ApplySources(sources...)
	return cm
}

// generateDefaultConfigFile 生成默认配置文件
func generateDefaultConfigFile(path string, cm *ConfigManager) error {
	// 使用当前 ConfigManager 的默认值生成配置
	return writeConfigToPath(path, cm)
}

func (cm *ConfigManager) SetDB(db *gorm.DB) { cm.db = db }
func (cm *ConfigManager) GetDB() *gorm.DB   { return cm.db }

// mergeUserConfig 合并用户配置，但允许覆盖为 0 的开关值
func (cm *ConfigManager) mergeUserConfig(fileUser *UserSystemConfig) {
	if fileUser == nil {
		return
	}

	// 对于开关型配置（0/1），需要始终应用，包括 0 值
	// 注意：这里不能检查 != 0，因为 0 是有效的配置值（禁用）
	cm.User.AllowUserRegistration = fileUser.AllowUserRegistration
	cm.User.RequireEmailVerify = fileUser.RequireEmailVerify

	// 对于数值配置，只在非 0 时应用，0 表示使用默认值
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
	return writeConfigToPath(cm.configFilePath(), cm)
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

// UpdateTransaction 在配置上执行事务式更新：先克隆、应用变更、落盘、重载；若任一步失败则回滚到原状态并恢复文件
func (cm *ConfigManager) UpdateTransaction(apply func(draft *ConfigManager) error) error {
	if apply == nil {
		return errors.New("更新操作不能为空")
	}

	original := cm.Clone()
	draft := cm.Clone()

	if err := apply(draft); err != nil {
		return err
	}

	if err := draft.Validate(); err != nil {
		return err
	}

	cm.applyFrom(draft)
	path := cm.configFilePath()

	if err := writeConfigToPath(path, cm); err != nil {
		cm.applyFrom(original)
		_ = writeConfigToPath(path, original)
		return err
	}

	reloaded := NewConfigManager()
	if err := reloaded.LoadFromYAML(path); err != nil {
		cm.applyFrom(original)
		_ = writeConfigToPath(path, original)
		return err
	}
	reloaded.applyEnvironmentOverrides()
	reloaded.db = cm.db

	cm.applyFrom(reloaded)

	if err := cm.ReloadConfig(); err != nil {
		cm.applyFrom(original)
		if rollbackErr := writeConfigToPath(path, original); rollbackErr != nil {
			return errors.Join(err, fmt.Errorf("rollback failed: %w", rollbackErr))
		}
		return err
	}

	return nil
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
	newManager.UploadMinute = cm.UploadMinute
	newManager.UploadCount = cm.UploadCount
	newManager.ErrorMinute = cm.ErrorMinute
	newManager.ErrorCount = cm.ErrorCount
	if len(cm.ExpireStyle) > 0 {
		newManager.ExpireStyle = append([]string(nil), cm.ExpireStyle...)
	}

	// 克隆 UI 配置
	if cm.UI != nil {
		ui := *cm.UI
		newManager.UI = &ui
	}

	return newManager
}

// applyFrom 用于将另外一个配置实例的内容拷贝到当前实例，保留数据库连接句柄
func (cm *ConfigManager) applyFrom(src *ConfigManager) {
	if src == nil {
		return
	}

	db := cm.db

	if src.Base != nil {
		cm.Base = src.Base.Clone()
	} else {
		cm.Base = nil
	}

	if src.Database != nil {
		cm.Database = src.Database.Clone()
	} else {
		cm.Database = nil
	}

	if src.Transfer != nil {
		cm.Transfer = src.Transfer.Clone()
	} else {
		cm.Transfer = nil
	}

	if src.Storage != nil {
		cm.Storage = src.Storage.Clone()
	} else {
		cm.Storage = nil
	}

	if src.User != nil {
		cm.User = src.User.Clone()
	} else {
		cm.User = nil
	}

	if src.MCP != nil {
		cm.MCP = src.MCP.Clone()
	} else {
		cm.MCP = nil
	}

	if src.UI != nil {
		ui := *src.UI
		cm.UI = &ui
	} else {
		cm.UI = nil
	}

	cm.NotifyTitle = src.NotifyTitle
	cm.NotifyContent = src.NotifyContent
	cm.SysStart = src.SysStart
	cm.UploadMinute = src.UploadMinute
	cm.UploadCount = src.UploadCount
	cm.ErrorMinute = src.ErrorMinute
	cm.ErrorCount = src.ErrorCount
	cm.ExpireStyle = append([]string(nil), src.ExpireStyle...)

	cm.db = db
}

func (cm *ConfigManager) configFilePath() string {
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}
	return "./config.yaml"
}

func writeConfigToPath(path string, cfg *ConfigManager) error {
	if cfg == nil {
		return errors.New("配置不能为空")
	}

	clone := cfg.Clone()
	if clone.UI == nil {
		clone.UI = &UIConfig{}
	}

	data, err := yaml.Marshal(clone)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
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
