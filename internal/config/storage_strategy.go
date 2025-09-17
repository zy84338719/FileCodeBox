// Package config 配置存储策略
package config

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ConfigStorageType 配置存储类型
type ConfigStorageType string

const (
	StorageTypeFile     ConfigStorageType = "file"     // 配置文件
	StorageTypeDatabase ConfigStorageType = "database" // 数据库表
	StorageTypeJSON     ConfigStorageType = "json"     // JSON配置
)

// ConfigCategory 配置分类
type ConfigCategory string

const (
	CategoryStatic   ConfigCategory = "static"   // 静态配置（文件）
	CategoryRuntime  ConfigCategory = "runtime"  // 运行时配置（数据库）
	CategorySystem   ConfigCategory = "system"   // 系统配置（专用表）
	CategoryBusiness ConfigCategory = "business" // 业务配置（业务表）
)

// ConfigMetadata 配置元数据
type ConfigMetadata struct {
	Key          string            `json:"key"`
	Category     ConfigCategory    `json:"category"`
	StorageType  ConfigStorageType `json:"storage_type"`
	TableName    string            `json:"table_name,omitempty"`
	FileName     string            `json:"file_name,omitempty"`
	Description  string            `json:"description"`
	Volatile     bool              `json:"volatile"`  // 是否易变
	Cacheable    bool              `json:"cacheable"` // 是否可缓存
	Version      int               `json:"version"`   // 配置版本
	LastModified time.Time         `json:"last_modified"`
}

// StaticConfig 静态配置（存储在文件中）
type StaticConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ConfigKey   string    `gorm:"uniqueIndex;size:100" json:"config_key"`
	ConfigValue string    `gorm:"type:text" json:"config_value"`
	Category    string    `gorm:"size:50;index" json:"category"`
	Description string    `gorm:"size:255" json:"description"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RuntimeConfig 运行时配置（存储在数据库中）
type RuntimeConfig struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	ConfigKey   string     `gorm:"uniqueIndex;size:100" json:"config_key"`
	ConfigValue string     `gorm:"type:text" json:"config_value"`
	DataType    string     `gorm:"size:20" json:"data_type"` // string, int, bool, json
	Category    string     `gorm:"size:50;index" json:"category"`
	UserID      *uint      `gorm:"index" json:"user_id,omitempty"` // 用户专属配置
	IsGlobal    bool       `gorm:"default:true" json:"is_global"`
	Priority    int        `gorm:"default:0" json:"priority"` // 优先级
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`      // 过期时间
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// SystemConfig 系统配置（存储配置、传输配置等）
type SystemConfig struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ConfigKey    string    `gorm:"uniqueIndex;size:100" json:"config_key"`
	ConfigValue  string    `gorm:"type:json" json:"config_value"`
	ConfigSchema string    `gorm:"type:text" json:"config_schema"` // JSON Schema
	Version      int       `gorm:"default:1" json:"version"`
	Environment  string    `gorm:"size:20;default:'production'" json:"environment"`
	IsEncrypted  bool      `gorm:"default:false" json:"is_encrypted"`
	Checksum     string    `gorm:"size:64" json:"checksum"` // 配置校验和
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// BusinessConfig 业务配置（通知、限流等）
type BusinessConfig struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	ConfigKey   string     `gorm:"uniqueIndex;size:100" json:"config_key"`
	ConfigValue string     `gorm:"type:json" json:"config_value"`
	Module      string     `gorm:"size:50;index" json:"module"` // 模块名称
	SubModule   string     `gorm:"size:50;index" json:"sub_module,omitempty"`
	IsEnabled   bool       `gorm:"default:true" json:"is_enabled"`
	ValidFrom   time.Time  `json:"valid_from"`
	ValidTo     *time.Time `json:"valid_to,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ConfigStorageStrategy 配置存储策略
type ConfigStorageStrategy struct {
	db            *gorm.DB
	configMapping map[string]ConfigMetadata
}

// NewConfigStorageStrategy 创建配置存储策略
func NewConfigStorageStrategy(db *gorm.DB) *ConfigStorageStrategy {
	strategy := &ConfigStorageStrategy{
		db:            db,
		configMapping: make(map[string]ConfigMetadata),
	}

	strategy.initConfigMapping()
	return strategy
}

// initConfigMapping 初始化配置映射
func (s *ConfigStorageStrategy) initConfigMapping() {
	// 静态配置（基础设置、主题）
	s.configMapping["base.name"] = ConfigMetadata{
		Key:         "base.name",
		Category:    CategoryStatic,
		StorageType: StorageTypeFile,
		FileName:    "static_config.json",
		Description: "应用名称",
		Cacheable:   true,
	}

	s.configMapping["base.description"] = ConfigMetadata{
		Key:         "base.description",
		Category:    CategoryStatic,
		StorageType: StorageTypeFile,
		FileName:    "static_config.json",
		Description: "应用描述",
		Cacheable:   true,
	}

	s.configMapping["theme.select"] = ConfigMetadata{
		Key:         "theme.select",
		Category:    CategoryStatic,
		StorageType: StorageTypeFile,
		FileName:    "theme_config.json",
		Description: "主题选择",
		Cacheable:   true,
	}

	// 运行时配置（用户设置）
	s.configMapping["user.system_enabled"] = ConfigMetadata{
		Key:         "user.system_enabled",
		Category:    CategoryRuntime,
		StorageType: StorageTypeDatabase,
		TableName:   "runtime_configs",
		Description: "用户系统启用状态",
		Volatile:    true,
		Cacheable:   true,
	}

	s.configMapping["mcp.enabled"] = ConfigMetadata{
		Key:         "mcp.enabled",
		Category:    CategoryRuntime,
		StorageType: StorageTypeDatabase,
		TableName:   "runtime_configs",
		Description: "MCP服务器启用状态",
		Volatile:    true,
		Cacheable:   false, // MCP配置变化需要立即生效
	}

	// 系统配置（存储、传输）
	s.configMapping["storage.config"] = ConfigMetadata{
		Key:         "storage.config",
		Category:    CategorySystem,
		StorageType: StorageTypeJSON,
		TableName:   "system_configs",
		Description: "存储配置",
		Volatile:    false,
		Cacheable:   true,
	}

	s.configMapping["transfer.config"] = ConfigMetadata{
		Key:         "transfer.config",
		Category:    CategorySystem,
		StorageType: StorageTypeJSON,
		TableName:   "system_configs",
		Description: "传输配置",
		Volatile:    false,
		Cacheable:   true,
	}

	s.configMapping["database.config"] = ConfigMetadata{
		Key:         "database.config",
		Category:    CategorySystem,
		StorageType: StorageTypeJSON,
		TableName:   "system_configs",
		Description: "数据库配置",
		Volatile:    false,
		Cacheable:   true,
	}

	// 业务配置（通知、限流）
	s.configMapping["notification.config"] = ConfigMetadata{
		Key:         "notification.config",
		Category:    CategoryBusiness,
		StorageType: StorageTypeJSON,
		TableName:   "business_configs",
		Description: "通知配置",
		Volatile:    true,
		Cacheable:   true,
	}

	s.configMapping["ratelimit.config"] = ConfigMetadata{
		Key:         "ratelimit.config",
		Category:    CategoryBusiness,
		StorageType: StorageTypeJSON,
		TableName:   "business_configs",
		Description: "限流配置",
		Volatile:    true,
		Cacheable:   true,
	}
}

// GetConfig 获取配置
func (s *ConfigStorageStrategy) GetConfig(key string, result interface{}) error {
	metadata, exists := s.configMapping[key]
	if !exists {
		return fmt.Errorf("配置键 %s 不存在", key)
	}

	switch metadata.StorageType {
	case StorageTypeFile:
		return s.getFileConfig(key, metadata, result)
	case StorageTypeDatabase:
		return s.getDatabaseConfig(key, metadata, result)
	case StorageTypeJSON:
		return s.getJSONConfig(key, metadata, result)
	default:
		return fmt.Errorf("不支持的存储类型: %s", metadata.StorageType)
	}
}

// SetConfig 设置配置
func (s *ConfigStorageStrategy) SetConfig(key string, value interface{}) error {
	metadata, exists := s.configMapping[key]
	if !exists {
		return fmt.Errorf("配置键 %s 不存在", key)
	}

	switch metadata.StorageType {
	case StorageTypeFile:
		return s.setFileConfig(key, metadata, value)
	case StorageTypeDatabase:
		return s.setDatabaseConfig(key, metadata, value)
	case StorageTypeJSON:
		return s.setJSONConfig(key, metadata, value)
	default:
		return fmt.Errorf("不支持的存储类型: %s", metadata.StorageType)
	}
}

// 文件配置操作
func (s *ConfigStorageStrategy) getFileConfig(key string, metadata ConfigMetadata, result interface{}) error {
	// 从静态配置表获取
	var staticConfig StaticConfig
	err := s.db.Where("config_key = ? AND is_active = ?", key, true).First(&staticConfig).Error
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(staticConfig.ConfigValue), result)
}

func (s *ConfigStorageStrategy) setFileConfig(key string, metadata ConfigMetadata, value interface{}) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	staticConfig := StaticConfig{
		ConfigKey:   key,
		ConfigValue: string(valueBytes),
		Category:    string(metadata.Category),
		Description: metadata.Description,
		IsActive:    true,
	}

	return s.db.Save(&staticConfig).Error
}

// 数据库配置操作
func (s *ConfigStorageStrategy) getDatabaseConfig(key string, metadata ConfigMetadata, result interface{}) error {
	var runtimeConfig RuntimeConfig
	err := s.db.Where("config_key = ? AND is_global = ?", key, true).First(&runtimeConfig).Error
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(runtimeConfig.ConfigValue), result)
}

func (s *ConfigStorageStrategy) setDatabaseConfig(key string, metadata ConfigMetadata, value interface{}) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	runtimeConfig := RuntimeConfig{
		ConfigKey:   key,
		ConfigValue: string(valueBytes),
		DataType:    "json",
		Category:    string(metadata.Category),
		IsGlobal:    true,
		Priority:    0,
	}

	return s.db.Save(&runtimeConfig).Error
}

// JSON配置操作
func (s *ConfigStorageStrategy) getJSONConfig(key string, metadata ConfigMetadata, result interface{}) error {
	var systemConfig SystemConfig
	err := s.db.Where("config_key = ?", key).First(&systemConfig).Error
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(systemConfig.ConfigValue), result)
}

func (s *ConfigStorageStrategy) setJSONConfig(key string, metadata ConfigMetadata, value interface{}) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	systemConfig := SystemConfig{
		ConfigKey:   key,
		ConfigValue: string(valueBytes),
		Version:     1,
		Environment: "production",
	}

	return s.db.Save(&systemConfig).Error
}

// InitTables 初始化配置相关表
func (s *ConfigStorageStrategy) InitTables() error {
	// 自动迁移所有配置表
	return s.db.AutoMigrate(
		&StaticConfig{},
		&RuntimeConfig{},
		&SystemConfig{},
		&BusinessConfig{},
	)
}

// GetConfigMetadata 获取配置元数据
func (s *ConfigStorageStrategy) GetConfigMetadata(key string) (ConfigMetadata, error) {
	metadata, exists := s.configMapping[key]
	if !exists {
		return ConfigMetadata{}, fmt.Errorf("配置键 %s 不存在", key)
	}
	return metadata, nil
}

// ListConfigsByCategory 按分类列出配置
func (s *ConfigStorageStrategy) ListConfigsByCategory(category ConfigCategory) ([]ConfigMetadata, error) {
	var result []ConfigMetadata
	for _, metadata := range s.configMapping {
		if metadata.Category == category {
			result = append(result, metadata)
		}
	}
	return result, nil
}

// ValidateConfig 验证配置
func (s *ConfigStorageStrategy) ValidateConfig(key string, value interface{}) error {
	_, exists := s.configMapping[key]
	if !exists {
		return fmt.Errorf("配置键 %s 不存在", key)
	}

	// 这里可以添加具体的验证逻辑
	// 比如检查JSON Schema、数据类型等

	return nil
}
