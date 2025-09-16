package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"

	"gorm.io/gorm"
)

// ConfigManager implements Option A semantics (Env > YAML > DB > Defaults).
// Keys present in config.yaml are recorded in yamlManagedKeys and are authoritative.
type ConfigManager struct {
	Base     *BaseConfig
	Database *DatabaseConfig
	Transfer *TransferConfig
	Storage  *StorageConfig
	User     *UserSystemConfig
	MCP      *MCPConfig

	NotifyTitle   string
	NotifyContent string
	AdminToken    string

	// UI / Theme / Page fields (kept at top-level for backward compatibility with callers)
	ThemesSelect  string  `yaml:"themes_select" json:"themes_select"`
	RobotsText    string  `yaml:"robots_text" json:"robots_text"`
	PageExplain   string  `yaml:"page_explain" json:"page_explain"`
	ShowAdminAddr int     `yaml:"show_admin_addr" json:"show_admin_addr"`
	Opacity       float64 `yaml:"opacity" json:"opacity"`
	Background    string  `yaml:"background" json:"background"`

	// rate limit / business fields (kept at top-level for backwards compatibility)
	UploadMinute int      `yaml:"upload_minute" json:"upload_minute"`
	UploadCount  int      `yaml:"upload_count" json:"upload_count"`
	ErrorMinute  int      `yaml:"error_minute" json:"error_minute"`
	ErrorCount   int      `yaml:"error_count" json:"error_count"`
	ExpireStyle  []string `yaml:"expire_style" json:"expire_style"`

	db *gorm.DB

	// yamlManagedKeys stores flat keys (module_field or single keys) that are managed by YAML
	// and must not be overwritten by DB nor written back to DB.
	yamlManagedKeys map[string]bool
}

func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		Base:     NewBaseConfig(),
		Database: NewDatabaseConfig(),
		Transfer: NewTransferConfig(),
		Storage:  NewStorageConfig(),
		User:     NewUserSystemConfig(),
		MCP:      NewMCPConfig(),
	}
}

// InitManager loads config.yaml early if present and applies environment overrides.
func InitManager() *ConfigManager {
	cm := NewConfigManager()
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		_ = cm.LoadFromYAML(p)
	} else if _, err := os.Stat("./config.yaml"); err == nil {
		_ = cm.LoadFromYAML("./config.yaml")
	}
	cm.applyEnvironmentOverrides()
	return cm
}

func (cm *ConfigManager) SetDB(db *gorm.DB) { cm.db = db }
func (cm *ConfigManager) GetDB() *gorm.DB   { return cm.db }

// LoadFromYAML loads hierarchical YAML into module structs and records their flat keys.
func (cm *ConfigManager) LoadFromYAML(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var fileCfg ConfigManager
	if err := yaml.Unmarshal(b, &fileCfg); err != nil {
		return err
	}
	if cm.yamlManagedKeys == nil {
		cm.yamlManagedKeys = make(map[string]bool)
	}
	if fileCfg.Base != nil {
		cm.Base = fileCfg.Base
		for k := range cm.Base.ToMap() {
			cm.yamlManagedKeys[k] = true
		}
	}
	if fileCfg.Database != nil {
		cm.Database = fileCfg.Database
		for k := range cm.Database.ToMap() {
			cm.yamlManagedKeys[k] = true
		}
	}
	if fileCfg.Transfer != nil {
		cm.Transfer = fileCfg.Transfer
		for k := range cm.Transfer.ToMap() {
			cm.yamlManagedKeys[k] = true
		}
	}
	if fileCfg.Storage != nil {
		cm.Storage = fileCfg.Storage
		for k := range cm.Storage.ToMap() {
			cm.yamlManagedKeys[k] = true
		}
	}
	if fileCfg.User != nil {
		cm.User = fileCfg.User
		for k := range cm.User.ToMap() {
			cm.yamlManagedKeys[k] = true
		}
	}
	if fileCfg.MCP != nil {
		cm.MCP = fileCfg.MCP
		for k := range cm.MCP.ToMap() {
			cm.yamlManagedKeys[k] = true
		}
	}
	if fileCfg.NotifyTitle != "" {
		cm.NotifyTitle = fileCfg.NotifyTitle
		cm.yamlManagedKeys["notify_title"] = true
	}
	if fileCfg.NotifyContent != "" {
		cm.NotifyContent = fileCfg.NotifyContent
		cm.yamlManagedKeys["notify_content"] = true
	}
	if fileCfg.AdminToken != "" {
		cm.AdminToken = fileCfg.AdminToken
		cm.yamlManagedKeys["admin_token"] = true
	}
	if fileCfg.ThemesSelect != "" {
		cm.ThemesSelect = fileCfg.ThemesSelect
		cm.yamlManagedKeys["themes_select"] = true
	}
	if fileCfg.RobotsText != "" {
		cm.RobotsText = fileCfg.RobotsText
		cm.yamlManagedKeys["robots_text"] = true
	}
	if fileCfg.PageExplain != "" {
		cm.PageExplain = fileCfg.PageExplain
		cm.yamlManagedKeys["page_explain"] = true
	}
	if fileCfg.ShowAdminAddr != 0 {
		cm.ShowAdminAddr = fileCfg.ShowAdminAddr
		cm.yamlManagedKeys["show_admin_addr"] = true
	}
	if fileCfg.Opacity != 0 {
		cm.Opacity = fileCfg.Opacity
		cm.yamlManagedKeys["opacity"] = true
	}
	if fileCfg.Background != "" {
		cm.Background = fileCfg.Background
		cm.yamlManagedKeys["background"] = true
	}
	return nil
}

// InitWithDB has been removed. Use SetDB(db) to inject a database connection.

// buildConfigMap flattens modules to module_field keys
func (cm *ConfigManager) buildConfigMap() map[string]string {
	out := make(map[string]string)
	for k, v := range cm.Base.ToMap() {
		out[k] = v
	}
	for k, v := range cm.Database.ToMap() {
		out[k] = v
	}
	for k, v := range cm.Transfer.ToMap() {
		out[k] = v
	}
	for k, v := range cm.Storage.ToMap() {
		out[k] = v
	}
	for k, v := range cm.User.ToMap() {
		out[k] = v
	}
	for k, v := range cm.MCP.ToMap() {
		out[k] = v
	}
	out["notify_title"] = cm.NotifyTitle
	out["notify_content"] = cm.NotifyContent
	out["admin_token"] = cm.AdminToken
	out["themes_select"] = cm.ThemesSelect
	out["robots_text"] = cm.RobotsText
	out["page_explain"] = cm.PageExplain
	out["show_admin_addr"] = fmt.Sprintf("%d", cm.ShowAdminAddr)
	out["opacity"] = fmt.Sprintf("%v", cm.Opacity)
	out["background"] = cm.Background
	return out
}

func (cm *ConfigManager) ReloadConfig() error {
	// ReloadConfig no longer reads configuration from the database.
	// Configuration should be provided via `config.yaml` and environment variables.
	// Preserve in-memory immutable fields across reload (port/admin token).
	curPort := cm.Base.Port
	curAdmin := cm.AdminToken
	cm.applyEnvironmentOverrides()
	cm.Base.Port = curPort
	cm.AdminToken = curAdmin
	return nil
}

func (cm *ConfigManager) applyEnvironmentOverrides() {
	if p := os.Getenv("PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			cm.Base.Port = n
		}
	}
	if t := os.Getenv("ADMIN_TOKEN"); t != "" {
		cm.AdminToken = t
	}
	if dp := os.Getenv("DATA_PATH"); dp != "" {
		cm.Base.DataPath = dp
	}
}

// Save saves the configuration to the database (if db is set).
// Save persists configuration. Persisting to DB is intentionally removed;
// this method returns an error to surface that saving is unsupported.
func (cm *ConfigManager) Save() error {
	return errors.New("saving configuration to database is not supported; use config.yaml and environment variables")
}

// Get helpers
func (cm *ConfigManager) GetAddress() string              { return cm.Base.GetAddress() }
func (cm *ConfigManager) GetDatabaseDSN() (string, error) { return cm.Database.GetDSN() }
func (cm *ConfigManager) IsUserSystemEnabled() bool       { return cm.User.IsUserSystemEnabled() }
func (cm *ConfigManager) IsMCPEnabled() bool              { return cm.MCP.IsMCPEnabled() }
func (cm *ConfigManager) Clone() *ConfigManager {
	nc := NewConfigManager()
	nc.Base = cm.Base.Clone()
	nc.Database = cm.Database.Clone()
	nc.Transfer = cm.Transfer.Clone()
	nc.Storage = cm.Storage.Clone()
	nc.User = cm.User.Clone()
	nc.MCP = cm.MCP.Clone()
	nc.NotifyTitle = cm.NotifyTitle
	nc.NotifyContent = cm.NotifyContent
	nc.AdminToken = cm.AdminToken
	return nc
}

// Validate validates all configuration modules.
func (cm *ConfigManager) Validate() error {
	if cm.Base == nil || cm.Database == nil || cm.Transfer == nil || cm.Storage == nil || cm.User == nil || cm.MCP == nil {
		return errors.New("配置模块未完全初始化")
	}
	if err := cm.Base.Validate(); err != nil {
		return err
	}
	if err := cm.Database.Validate(); err != nil {
		return err
	}
	return nil
}
