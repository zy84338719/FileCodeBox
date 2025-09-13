// Package config 分层配置管理器
package config

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
)

// LayeredConfigManager 分层配置管理器
type LayeredConfigManager struct {
	storageStrategy *ConfigStorageStrategy
	cache           map[string]interface{}
	cacheMutex      sync.RWMutex
	cacheExpiry     map[string]time.Time
	db              *gorm.DB

	// 配置模块
	Base     *BaseConfig       `json:"base"`
	Database *DatabaseConfig   `json:"database"`
	Transfer *TransferConfig   `json:"transfer"`
	Storage  *StorageConfig    `json:"storage"`
	User     *UserSystemConfig `json:"user"`
	MCP      *MCPConfig        `json:"mcp"`

	// 业务配置
	Notification *NotificationConfig `json:"notification"`
	RateLimit    *RateLimitConfig    `json:"rate_limit"`
	Theme        *ThemeConfig        `json:"theme"`
}

// NotificationConfig 通知配置
type NotificationConfig struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Explain string `json:"explain"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	UploadMinute int `json:"upload_minute"`
	UploadCount  int `json:"upload_count"`
	ErrorMinute  int `json:"error_minute"`
	ErrorCount   int `json:"error_count"`
}

// ThemeConfig 主题配置
type ThemeConfig struct {
	Select     string  `json:"select"`
	Choices    []Theme `json:"choices"`
	Opacity    float64 `json:"opacity"`
	Background string  `json:"background"`
}

// NewLayeredConfigManager 创建分层配置管理器
func NewLayeredConfigManager(db *gorm.DB) *LayeredConfigManager {
	manager := &LayeredConfigManager{
		storageStrategy: NewConfigStorageStrategy(db),
		cache:           make(map[string]interface{}),
		cacheExpiry:     make(map[string]time.Time),
		db:              db,
	}

	// 初始化配置模块
	manager.initConfigModules()

	return manager
}

// initConfigModules 初始化配置模块
func (m *LayeredConfigManager) initConfigModules() {
	m.Base = &BaseConfig{}
	m.Database = &DatabaseConfig{}
	m.Transfer = &TransferConfig{}
	m.Storage = &StorageConfig{}
	m.User = &UserSystemConfig{}
	m.MCP = &MCPConfig{}
	m.Notification = &NotificationConfig{}
	m.RateLimit = &RateLimitConfig{}
	m.Theme = &ThemeConfig{}
}

// InitTables 初始化数据库表
func (m *LayeredConfigManager) InitTables() error {
	return m.storageStrategy.InitTables()
}

// LoadAllConfigs 加载所有配置
func (m *LayeredConfigManager) LoadAllConfigs() error {
	// 1. 加载静态配置（基础设置）
	if err := m.loadStaticConfigs(); err != nil {
		return fmt.Errorf("加载静态配置失败: %v", err)
	}

	// 2. 加载系统配置（存储、传输、数据库）
	if err := m.loadSystemConfigs(); err != nil {
		return fmt.Errorf("加载系统配置失败: %v", err)
	}

	// 3. 加载运行时配置（用户、MCP）
	if err := m.loadRuntimeConfigs(); err != nil {
		return fmt.Errorf("加载运行时配置失败: %v", err)
	}

	// 4. 加载业务配置（通知、限流）
	if err := m.loadBusinessConfigs(); err != nil {
		return fmt.Errorf("加载业务配置失败: %v", err)
	}

	return nil
}

// loadStaticConfigs 加载静态配置
func (m *LayeredConfigManager) loadStaticConfigs() error {
	// 加载基础配置
	if err := m.getConfig("base.name", &m.Base.Name); err != nil {
		m.Base.Name = "文件快递柜 - FileCodeBox" // 默认值
	}
	if err := m.getConfig("base.description", &m.Base.Description); err != nil {
		m.Base.Description = "开箱即用的文件快传系统"
	}

	// 加载主题配置
	if err := m.getConfig("theme.select", &m.Theme.Select); err != nil {
		m.Theme.Select = "themes/2025"
	}

	return nil
}

// loadSystemConfigs 加载系统配置
func (m *LayeredConfigManager) loadSystemConfigs() error {
	// 加载存储配置
	if err := m.getConfig("storage.config", m.Storage); err != nil {
		// 设置默认存储配置
		m.Storage.Type = "local"
		m.Storage.StoragePath = ""
	}

	// 加载传输配置
	if err := m.getConfig("transfer.config", m.Transfer); err != nil {
		// 设置默认传输配置
		m.Transfer.Upload = &UploadConfig{
			OpenUpload:     1,
			UploadSize:     10 * 1024 * 1024,
			EnableChunk:    0,
			ChunkSize:      2 * 1024 * 1024,
			MaxSaveSeconds: 0,
		}
		m.Transfer.Download = &DownloadConfig{
			EnableConcurrentDownload: 1,
			MaxConcurrentDownloads:   10,
			DownloadTimeout:          300,
		}
	}

	// 加载数据库配置
	if err := m.getConfig("database.config", m.Database); err != nil {
		// 设置默认数据库配置
		m.Database.Type = "sqlite"
		m.Database.Host = "localhost"
		m.Database.Port = 3306
		m.Database.Name = "filecodebox"
		m.Database.User = "root"
		m.Database.Pass = ""
		m.Database.SSL = "disable"
	}

	return nil
}

// loadRuntimeConfigs 加载运行时配置
func (m *LayeredConfigManager) loadRuntimeConfigs() error {
	// 用户系统始终启用，无需加载配置

	// 加载MCP配置
	var mcpEnabled int
	if err := m.getConfig("mcp.enabled", &mcpEnabled); err != nil {
		mcpEnabled = 0 // 默认禁用
	}
	m.MCP.EnableMCPServer = mcpEnabled

	return nil
}

// loadBusinessConfigs 加载业务配置
func (m *LayeredConfigManager) loadBusinessConfigs() error {
	// 加载通知配置
	if err := m.getConfig("notification.config", m.Notification); err != nil {
		// 设置默认通知配置
		m.Notification.Title = "系统通知"
		m.Notification.Content = "欢迎使用 FileCodeBox"
		m.Notification.Explain = "请勿上传或分享违法内容"
	}

	// 加载限流配置
	if err := m.getConfig("ratelimit.config", m.RateLimit); err != nil {
		// 设置默认限流配置
		m.RateLimit.UploadMinute = 1
		m.RateLimit.UploadCount = 10
		m.RateLimit.ErrorMinute = 1
		m.RateLimit.ErrorCount = 1
	}

	return nil
}

// getConfig 获取配置（带缓存）
func (m *LayeredConfigManager) getConfig(key string, result interface{}) error {
	// 首先检查缓存
	if cached, ok := m.getCachedConfig(key); ok {
		// 复制缓存值到result
		return m.copyValue(cached, result)
	}

	// 从存储策略获取
	if err := m.storageStrategy.GetConfig(key, result); err != nil {
		return err
	}

	// 更新缓存
	m.setCachedConfig(key, result)

	return nil
}

// setConfig 设置配置
func (m *LayeredConfigManager) setConfig(key string, value interface{}) error {
	// 设置到存储策略
	if err := m.storageStrategy.SetConfig(key, value); err != nil {
		return err
	}

	// 更新缓存
	m.setCachedConfig(key, value)

	return nil
}

// getCachedConfig 获取缓存配置
func (m *LayeredConfigManager) getCachedConfig(key string) (interface{}, bool) {
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()

	// 检查是否过期
	if expiry, exists := m.cacheExpiry[key]; exists && time.Now().After(expiry) {
		delete(m.cache, key)
		delete(m.cacheExpiry, key)
		return nil, false
	}

	value, exists := m.cache[key]
	return value, exists
}

// setCachedConfig 设置缓存配置
func (m *LayeredConfigManager) setCachedConfig(key string, value interface{}) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	m.cache[key] = value

	// 设置缓存过期时间（5分钟）
	m.cacheExpiry[key] = time.Now().Add(5 * time.Minute)
}

// copyValue 复制值
func (m *LayeredConfigManager) copyValue(src, dst interface{}) error {
	srcBytes, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(srcBytes, dst)
}

// SaveAllConfigs 保存所有配置
func (m *LayeredConfigManager) SaveAllConfigs() error {
	// 保存静态配置
	if err := m.saveStaticConfigs(); err != nil {
		return fmt.Errorf("保存静态配置失败: %v", err)
	}

	// 保存系统配置
	if err := m.saveSystemConfigs(); err != nil {
		return fmt.Errorf("保存系统配置失败: %v", err)
	}

	// 保存运行时配置
	if err := m.saveRuntimeConfigs(); err != nil {
		return fmt.Errorf("保存运行时配置失败: %v", err)
	}

	// 保存业务配置
	if err := m.saveBusinessConfigs(); err != nil {
		return fmt.Errorf("保存业务配置失败: %v", err)
	}

	return nil
}

// saveStaticConfigs 保存静态配置
func (m *LayeredConfigManager) saveStaticConfigs() error {
	if err := m.setConfig("base.name", m.Base.Name); err != nil {
		return err
	}
	if err := m.setConfig("base.description", m.Base.Description); err != nil {
		return err
	}
	if err := m.setConfig("theme.select", m.Theme.Select); err != nil {
		return err
	}
	return nil
}

// saveSystemConfigs 保存系统配置
func (m *LayeredConfigManager) saveSystemConfigs() error {
	if err := m.setConfig("storage.config", m.Storage); err != nil {
		return err
	}
	if err := m.setConfig("transfer.config", m.Transfer); err != nil {
		return err
	}
	if err := m.setConfig("database.config", m.Database); err != nil {
		return err
	}
	return nil
}

// saveRuntimeConfigs 保存运行时配置
func (m *LayeredConfigManager) saveRuntimeConfigs() error {
	// 用户系统始终启用，无需保存配置
	if err := m.setConfig("mcp.enabled", m.MCP.EnableMCPServer); err != nil {
		return err
	}
	return nil
}

// saveBusinessConfigs 保存业务配置
func (m *LayeredConfigManager) saveBusinessConfigs() error {
	if err := m.setConfig("notification.config", m.Notification); err != nil {
		return err
	}
	if err := m.setConfig("ratelimit.config", m.RateLimit); err != nil {
		return err
	}
	return nil
}

// GetConfigByCategory 按分类获取配置
func (m *LayeredConfigManager) GetConfigByCategory(category ConfigCategory) (map[string]interface{}, error) {
	configs, err := m.storageStrategy.ListConfigsByCategory(category)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, config := range configs {
		var value interface{}
		if err := m.getConfig(config.Key, &value); err == nil {
			result[config.Key] = value
		}
	}

	return result, nil
}

// ValidateAllConfigs 验证所有配置
func (m *LayeredConfigManager) ValidateAllConfigs() error {
	// 验证各个模块
	if err := m.Base.Validate(); err != nil {
		return fmt.Errorf("基础配置验证失败: %v", err)
	}
	if err := m.Database.Validate(); err != nil {
		return fmt.Errorf("数据库配置验证失败: %v", err)
	}
	if err := m.Transfer.Validate(); err != nil {
		return fmt.Errorf("传输配置验证失败: %v", err)
	}
	if err := m.Storage.Validate(); err != nil {
		return fmt.Errorf("存储配置验证失败: %v", err)
	}
	if err := m.User.Validate(); err != nil {
		return fmt.Errorf("用户配置验证失败: %v", err)
	}
	if err := m.MCP.Validate(); err != nil {
		return fmt.Errorf("MCP配置验证失败: %v", err)
	}

	return nil
}

// ClearCache 清除缓存
func (m *LayeredConfigManager) ClearCache() {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	m.cache = make(map[string]interface{})
	m.cacheExpiry = make(map[string]time.Time)
}

// GetCacheStats 获取缓存统计
func (m *LayeredConfigManager) GetCacheStats() map[string]interface{} {
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()

	stats := map[string]interface{}{
		"cache_size":    len(m.cache),
		"cache_entries": make([]string, 0, len(m.cache)),
	}

	for key := range m.cache {
		stats["cache_entries"] = append(stats["cache_entries"].([]string), key)
	}

	return stats
}
