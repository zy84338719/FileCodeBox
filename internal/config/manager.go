// Package config 配置管理器
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/zy84338719/filecodebox/internal/models"
	"gorm.io/gorm"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	Base     *BaseConfig       `json:"base"`
	Database *DatabaseConfig   `json:"database"`
	Transfer *TransferConfig   `json:"transfer"`
	Storage  *StorageConfig    `json:"storage"`
	User     *UserSystemConfig `json:"user"`
	MCP      *MCPConfig        `json:"mcp"`

	// 其他配置字段
	NotifyTitle   string   `json:"notify_title"`
	NotifyContent string   `json:"notify_content"`
	PageExplain   string   `json:"page_explain"`
	ExpireStyle   []string `json:"expire_style"`

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

	// 管理配置
	AdminToken    string `json:"admin_token"`
	ShowAdminAddr int    `json:"show_admin_address"`
	RobotsText    string `json:"robots_text"`

	// 数据库连接（内部使用）
	db *gorm.DB `json:"-"`
}

// NewConfigManager 创建配置管理器
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		Base:     NewBaseConfig(),
		Database: NewDatabaseConfig(),
		Transfer: NewTransferConfig(),
		Storage:  NewStorageConfig(),
		User:     NewUserSystemConfig(),
		MCP:      NewMCPConfig(),

		NotifyTitle:   "系统通知",
		NotifyContent: `欢迎使用 FileCodeBox，本程序开源于 <a href="https://github.com/vastsa/FileCodeBox" target="_blank">Github</a> ，欢迎Star和Fork。`,
		PageExplain:   "请勿上传或分享违法内容。根据《中华人民共和国网络安全法》、《中华人民共和国刑法》、《中华人民共和国治安管理处罚法》等相关规定。 传播或存储违法、违规内容，会受到相关处罚，严重者将承担刑事责任。本站坚决配合相关部门，确保网络内容的安全，和谐，打造绿色网络环境。",
		ExpireStyle:   []string{"day", "hour", "minute", "forever", "count"},

		UploadMinute: 1,
		UploadCount:  10,
		ErrorMinute:  1,
		ErrorCount:   1,

		ThemesSelect: "themes/2025",
		ThemesChoices: []Theme{
			{Name: "2025", Key: "themes/2025", Author: "Yi", Version: "1.0"},
		},
		Opacity:    0.9,
		Background: "",

		AdminToken:    "FileCodeBox2025",
		ShowAdminAddr: 0,
		RobotsText:    "User-agent: *\nDisallow: /",
	}
}

// InitManager 初始化配置管理器
func InitManager() *ConfigManager {
	cm := NewConfigManager()

	// 从环境变量读取配置
	cm.applyEnvironmentOverrides()

	// 创建数据目录（仅SQLite需要）
	if cm.Database.IsSQLite() {
		if err := os.MkdirAll(cm.Base.DataPath, 0750); err != nil {
			panic("创建数据目录失败: " + err.Error())
		}
	}

	return cm
}

// InitWithDB 使用数据库初始化配置管理器
func (cm *ConfigManager) InitWithDB(db *gorm.DB) error {
	cm.db = db

	// 保存环境变量和默认配置中的端口和管理员密码
	envPort := cm.Base.Port
	envAdminToken := cm.AdminToken

	// 尝试从数据库加载配置
	if err := cm.LoadFromDatabase(); err != nil {
		// 数据库中没有配置，初始化基础数据
		if err := cm.InitDefaultDataInDB(); err != nil {
			return fmt.Errorf("初始化数据库默认配置失败: %w", err)
		}
		// 重新加载配置
		if err := cm.LoadFromDatabase(); err != nil {
			return fmt.Errorf("加载初始化后的配置失败: %w", err)
		}
	}

	// 启动时优先级：端口和管理员密码优先使用环境变量或默认配置
	cm.Base.Port = envPort
	cm.AdminToken = envAdminToken

	// 再次应用环境变量覆盖以确保优先级正确
	cm.applyEnvironmentOverrides()

	return nil
}

// Validate 验证所有配置
func (cm *ConfigManager) Validate() error {
	if err := cm.Base.Validate(); err != nil {
		return fmt.Errorf("基础配置验证失败: %w", err)
	}

	if err := cm.Database.Validate(); err != nil {
		return fmt.Errorf("数据库配置验证失败: %w", err)
	}

	if err := cm.Transfer.Validate(); err != nil {
		return fmt.Errorf("传输配置验证失败: %w", err)
	}

	if err := cm.Storage.Validate(); err != nil {
		return fmt.Errorf("存储配置验证失败: %w", err)
	}

	if err := cm.User.Validate(); err != nil {
		return fmt.Errorf("用户系统配置验证失败: %w", err)
	}

	if err := cm.MCP.Validate(); err != nil {
		return fmt.Errorf("MCP配置验证失败: %w", err)
	}

	return nil
}

// buildConfigMap 构建配置映射表
func (cm *ConfigManager) buildConfigMap() map[string]string {
	result := make(map[string]string)

	// 合并各个配置模块的映射
	for k, v := range cm.Base.ToMap() {
		result[k] = v
	}

	for k, v := range cm.Database.ToMap() {
		result[k] = v
	}

	for k, v := range cm.Transfer.ToMap() {
		result[k] = v
	}

	for k, v := range cm.Storage.ToMap() {
		result[k] = v
	}

	for k, v := range cm.User.ToMap() {
		result[k] = v
	}

	for k, v := range cm.MCP.ToMap() {
		result[k] = v
	}

	// 添加其他配置
	result["notify_title"] = cm.NotifyTitle
	result["notify_content"] = cm.NotifyContent
	result["page_explain"] = cm.PageExplain
	result["upload_minute"] = fmt.Sprintf("%d", cm.UploadMinute)
	result["upload_count"] = fmt.Sprintf("%d", cm.UploadCount)
	result["error_minute"] = fmt.Sprintf("%d", cm.ErrorMinute)
	result["error_count"] = fmt.Sprintf("%d", cm.ErrorCount)
	result["themes_select"] = cm.ThemesSelect
	result["opacity"] = fmt.Sprintf("%f", cm.Opacity)
	result["background"] = cm.Background
	result["admin_token"] = cm.AdminToken
	result["show_admin_address"] = fmt.Sprintf("%d", cm.ShowAdminAddr)
	result["robots_text"] = cm.RobotsText

	return result
}

// InitDefaultDataInDB 在数据库中初始化默认配置数据
func (cm *ConfigManager) InitDefaultDataInDB() error {
	if cm.db == nil {
		return errors.New("数据库连接未设置")
	}

	// 检查是否已经有配置数据
	var count int64
	if err := cm.db.Model(&models.KeyValue{}).Count(&count).Error; err != nil {
		return fmt.Errorf("检查配置数据失败: %w", err)
	}

	// 如果已有数据，不进行初始化
	if count > 0 {
		return nil
	}

	// 使用公共方法获取配置映射
	defaultConfigs := cm.buildConfigMap()

	// 批量插入默认配置
	var keyValues []models.KeyValue
	for key, value := range defaultConfigs {
		keyValues = append(keyValues, models.KeyValue{
			Key:   key,
			Value: value,
		})
	}

	if err := cm.db.CreateInBatches(keyValues, 50).Error; err != nil {
		return fmt.Errorf("插入默认配置失败: %w", err)
	}

	return nil
}

// SetDB 设置数据库连接
func (cm *ConfigManager) SetDB(db *gorm.DB) {
	cm.db = db
}

// Save 保存配置
func (cm *ConfigManager) Save() error {
	// 验证配置
	if err := cm.Validate(); err != nil {
		return err
	}

	// 只保存到数据库
	if cm.db != nil {
		return cm.saveToDatabase()
	}

	return errors.New("数据库连接未设置，无法保存配置")
}

// saveToDatabase 保存配置到数据库
func (cm *ConfigManager) saveToDatabase() error {
	if cm.db == nil {
		return errors.New("数据库连接未设置")
	}

	// 使用公共方法获取配置映射
	configMap := cm.buildConfigMap()

	for key, value := range configMap {
		kv := models.KeyValue{
			Key:   key,
			Value: value,
		}

		// 使用 UPSERT 操作
		if err := cm.db.Where("key = ?", key).Assign(models.KeyValue{Value: value}).FirstOrCreate(&kv).Error; err != nil {
			return fmt.Errorf("保存配置项 %s 失败: %w", key, err)
		}
	}

	return nil
}

// LoadFromDatabase 从数据库加载配置
func (cm *ConfigManager) LoadFromDatabase() error {
	if cm.db == nil {
		return errors.New("数据库连接未设置")
	}

	var kvPairs []models.KeyValue
	if err := cm.db.Find(&kvPairs).Error; err != nil {
		return fmt.Errorf("查询配置失败: %w", err)
	}

	// 如果数据库中没有配置，返回错误以触发初始化
	if len(kvPairs) == 0 {
		return fmt.Errorf("数据库中没有配置数据")
	}

	// 构建数据映射
	data := make(map[string]string)
	for _, kv := range kvPairs {
		// 支持嵌套格式的键，转换为平面格式
		if strings.Contains(kv.Key, ".") {
			// 将嵌套格式转换为平面格式，例如 "user.allow_user_registration" -> "allow_user_registration"
			parts := strings.SplitN(kv.Key, ".", 2)
			if len(parts) == 2 {
				// 如果是已知的嵌套格式，去掉前缀
				switch parts[0] {
				case "user", "base", "transfer", "storage", "database", "mcp":
					data[parts[1]] = kv.Value
				default:
					// 保持原样
					data[kv.Key] = kv.Value
				}
			} else {
				data[kv.Key] = kv.Value
			}
		} else {
			data[kv.Key] = kv.Value
		}
	}

	// 加载各个配置模块
	if err := cm.Base.FromMap(data); err != nil {
		return fmt.Errorf("加载基础配置失败: %w", err)
	}

	if err := cm.Database.FromMap(data); err != nil {
		return fmt.Errorf("加载数据库配置失败: %w", err)
	}

	if err := cm.Transfer.FromMap(data); err != nil {
		return fmt.Errorf("加载传输配置失败: %w", err)
	}

	if err := cm.Storage.FromMap(data); err != nil {
		return fmt.Errorf("加载存储配置失败: %w", err)
	}

	if err := cm.User.FromMap(data); err != nil {
		return fmt.Errorf("加载用户系统配置失败: %w", err)
	}

	if err := cm.MCP.FromMap(data); err != nil {
		return fmt.Errorf("加载MCP配置失败: %w", err)
	}

	// 加载其他配置
	if val, ok := data["notify_title"]; ok {
		cm.NotifyTitle = val
	}
	if val, ok := data["notify_content"]; ok {
		cm.NotifyContent = val
	}
	if val, ok := data["page_explain"]; ok {
		cm.PageExplain = val
	}
	if val, ok := data["upload_minute"]; ok {
		if v, err := strconv.Atoi(val); err == nil {
			cm.UploadMinute = v
		}
	}
	if val, ok := data["upload_count"]; ok {
		if v, err := strconv.Atoi(val); err == nil {
			cm.UploadCount = v
		}
	}
	if val, ok := data["error_minute"]; ok {
		if v, err := strconv.Atoi(val); err == nil {
			cm.ErrorMinute = v
		}
	}
	if val, ok := data["error_count"]; ok {
		if v, err := strconv.Atoi(val); err == nil {
			cm.ErrorCount = v
		}
	}
	if val, ok := data["themes_select"]; ok {
		cm.ThemesSelect = val
	}
	if val, ok := data["opacity"]; ok {
		if v, err := strconv.ParseFloat(val, 64); err == nil {
			cm.Opacity = v
		}
	}
	if val, ok := data["background"]; ok {
		cm.Background = val
	}
	if val, ok := data["admin_token"]; ok {
		cm.AdminToken = val
	}
	if val, ok := data["show_admin_address"]; ok {
		if v, err := strconv.Atoi(val); err == nil {
			cm.ShowAdminAddr = v
		}
	}
	if val, ok := data["robots_text"]; ok {
		cm.RobotsText = val
	}

	return nil
}

// ReloadConfig 重新加载配置（热重载）
func (cm *ConfigManager) ReloadConfig() error {
	if cm.db == nil {
		return errors.New("数据库连接未设置")
	}

	// 保存当前的端口和管理员密码
	originalPort := cm.Base.Port
	originalAdminToken := cm.AdminToken

	// 从数据库重新加载配置
	if err := cm.LoadFromDatabase(); err != nil {
		return fmt.Errorf("重新加载配置失败: %w", err)
	}

	// 恢复原始的端口和管理员密码设置
	cm.Base.Port = originalPort
	cm.AdminToken = originalAdminToken

	// 重新应用环境变量中的优先级设置
	cm.applyEnvironmentOverrides()

	return nil
}

// applyEnvironmentOverrides 应用环境变量覆盖
func (cm *ConfigManager) applyEnvironmentOverrides() {
	// 端口配置 - 总是优先环境变量
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cm.Base.Port = p
		}
	}

	// 管理员密码配置 - 总是优先环境变量
	if adminToken := os.Getenv("ADMIN_TOKEN"); adminToken != "" {
		cm.AdminToken = adminToken
	}

	// 主机绑定配置
	if host := os.Getenv("HOST"); host != "" {
		cm.Base.Host = host
	}

	// 数据路径配置
	if dataPath := os.Getenv("DATA_PATH"); dataPath != "" {
		cm.Base.DataPath = dataPath
	}

	// 生产模式配置
	if production := os.Getenv("PRODUCTION"); production != "" {
		if prod, err := strconv.Atoi(production); err == nil {
			cm.Base.Production = prod == 1
		}
	}

	// 上传配置
	if openUpload := os.Getenv("OPEN_UPLOAD"); openUpload != "" {
		if upload, err := strconv.Atoi(openUpload); err == nil {
			cm.Transfer.Upload.OpenUpload = upload
		}
	}

	if uploadSize := os.Getenv("UPLOAD_SIZE"); uploadSize != "" {
		if size, err := strconv.ParseInt(uploadSize, 10, 64); err == nil {
			cm.Transfer.Upload.UploadSize = size
		}
	}

	// 数据库配置
	if dbType := os.Getenv("DATABASE_TYPE"); dbType != "" {
		cm.Database.Type = dbType
	}

	if dbHost := os.Getenv("DATABASE_HOST"); dbHost != "" {
		cm.Database.Host = dbHost
	}

	if dbPort := os.Getenv("DATABASE_PORT"); dbPort != "" {
		if port, err := strconv.Atoi(dbPort); err == nil {
			cm.Database.Port = port
		}
	}

	if dbName := os.Getenv("DATABASE_NAME"); dbName != "" {
		cm.Database.Name = dbName
	}

	if dbUser := os.Getenv("DATABASE_USER"); dbUser != "" {
		cm.Database.User = dbUser
	}

	if dbPass := os.Getenv("DATABASE_PASS"); dbPass != "" {
		cm.Database.Pass = dbPass
	}

	if dbSSL := os.Getenv("DATABASE_SSL"); dbSSL != "" {
		cm.Database.SSL = dbSSL
	}
}

// GetAddress 获取服务器完整地址
func (cm *ConfigManager) GetAddress() string {
	return cm.Base.GetAddress()
}

// GetDatabaseDSN 获取数据库连接字符串
func (cm *ConfigManager) GetDatabaseDSN() (string, error) {
	return cm.Database.GetDSN()
}

// IsUserSystemEnabled 判断是否启用用户系统
func (cm *ConfigManager) IsUserSystemEnabled() bool {
	return cm.User.IsUserSystemEnabled()
}

// IsMCPEnabled 判断是否启用MCP服务器
func (cm *ConfigManager) IsMCPEnabled() bool {
	return cm.MCP.IsMCPEnabled()
}

// Clone 克隆整个配置管理器
func (cm *ConfigManager) Clone() *ConfigManager {
	clone := &ConfigManager{
		Base:     cm.Base.Clone(),
		Database: cm.Database.Clone(),
		Transfer: cm.Transfer.Clone(),
		Storage:  cm.Storage.Clone(),
		User:     cm.User.Clone(),
		MCP:      cm.MCP.Clone(),

		NotifyTitle:   cm.NotifyTitle,
		NotifyContent: cm.NotifyContent,
		PageExplain:   cm.PageExplain,
		ExpireStyle:   make([]string, len(cm.ExpireStyle)),

		UploadMinute: cm.UploadMinute,
		UploadCount:  cm.UploadCount,
		ErrorMinute:  cm.ErrorMinute,
		ErrorCount:   cm.ErrorCount,

		ThemesSelect:  cm.ThemesSelect,
		ThemesChoices: make([]Theme, len(cm.ThemesChoices)),
		Opacity:       cm.Opacity,
		Background:    cm.Background,

		AdminToken:    cm.AdminToken,
		ShowAdminAddr: cm.ShowAdminAddr,
		RobotsText:    cm.RobotsText,
	}

	copy(clone.ExpireStyle, cm.ExpireStyle)
	copy(clone.ThemesChoices, cm.ThemesChoices)

	return clone
}
