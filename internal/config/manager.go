package config

import (
	"errors"
	"os"
	"strconv"
	"strings"

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

	// UI config grouped under `ui` in config.yaml. We keep top-level
	// compatibility fields but prefer `UI` when loading/saving YAML.
	UI *UIConfig `yaml:"ui" json:"ui"`

	// UI / Theme / Page top-level compatibility fields
	ThemesSelect  string  `yaml:"themes_select" json:"themes_select"`
	RobotsText    string  `yaml:"robots_text" json:"robots_text"`
	PageExplain   string  `yaml:"page_explain" json:"page_explain"`
	ShowAdminAddr int     `yaml:"show_admin_addr" json:"show_admin_addr"`
	Opacity       float64 `yaml:"opacity" json:"opacity"`
	Background    string  `yaml:"background" json:"background"`

	// Persistent runtime metadata
	SysStart string `yaml:"sys_start" json:"sys_start"`

	// rate limit / business fields (kept at top-level for backwards compatibility)
	UploadMinute int      `yaml:"upload_minute" json:"upload_minute"`
	UploadCount  int      `yaml:"upload_count" json:"upload_count"`
	ErrorMinute  int      `yaml:"error_minute" json:"error_minute"`
	ErrorCount   int      `yaml:"error_count" json:"error_count"`
	ExpireStyle  []string `yaml:"expire_style" json:"expire_style"`

	db *gorm.DB

	// ConfigManager now reads/writes the whole struct directly to YAML.
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
	if fileCfg.User != nil {
		// Merge user config fields rather than clobbering defaults with an empty struct
		if fileCfg.User.AllowUserRegistration != 0 {
			cm.User.AllowUserRegistration = fileCfg.User.AllowUserRegistration
		}
		if fileCfg.User.RequireEmailVerify != 0 {
			cm.User.RequireEmailVerify = fileCfg.User.RequireEmailVerify
		}
		if fileCfg.User.UserUploadSize != 0 {
			cm.User.UserUploadSize = fileCfg.User.UserUploadSize
		}
		if fileCfg.User.UserStorageQuota != 0 {
			cm.User.UserStorageQuota = fileCfg.User.UserStorageQuota
		}
		if fileCfg.User.SessionExpiryHours != 0 {
			cm.User.SessionExpiryHours = fileCfg.User.SessionExpiryHours
		}
		if fileCfg.User.MaxSessionsPerUser != 0 {
			cm.User.MaxSessionsPerUser = fileCfg.User.MaxSessionsPerUser
		}
		if strings.TrimSpace(fileCfg.User.JWTSecret) != "" {
			cm.User.JWTSecret = fileCfg.User.JWTSecret
		}
	}
	if fileCfg.MCP != nil {
		cm.MCP = fileCfg.MCP
	}
	if fileCfg.NotifyTitle != "" {
		cm.NotifyTitle = fileCfg.NotifyTitle
	}
	if fileCfg.NotifyContent != "" {
		cm.NotifyContent = fileCfg.NotifyContent
	}

	// Prefer structured UI block if present; otherwise fallback to legacy top-level fields.
	if fileCfg.UI != nil {
		cm.UI = fileCfg.UI
		// sync top-level compatibility fields
		cm.ThemesSelect = cm.UI.ThemesSelect
		cm.Background = cm.UI.Background
		cm.PageExplain = cm.UI.PageExplain
		cm.Opacity = cm.UI.Opacity
		cm.RobotsText = cm.UI.RobotsText
		cm.ShowAdminAddr = cm.UI.ShowAdminAddr
	} else {
		if fileCfg.ThemesSelect != "" {
			cm.ThemesSelect = fileCfg.ThemesSelect
			cm.UI.ThemesSelect = fileCfg.ThemesSelect
		}
		if fileCfg.RobotsText != "" {
			cm.RobotsText = fileCfg.RobotsText
			cm.UI.RobotsText = fileCfg.RobotsText
		}
		if fileCfg.PageExplain != "" {
			cm.PageExplain = fileCfg.PageExplain
			cm.UI.PageExplain = fileCfg.PageExplain
		}
		if fileCfg.ShowAdminAddr != 0 {
			cm.ShowAdminAddr = fileCfg.ShowAdminAddr
			cm.UI.ShowAdminAddr = fileCfg.ShowAdminAddr
		}
		if fileCfg.Opacity != 0 {
			cm.Opacity = fileCfg.Opacity
			cm.UI.Opacity = fileCfg.Opacity
		}
		if fileCfg.Background != "" {
			cm.Background = fileCfg.Background
			cm.UI.Background = fileCfg.Background
		}
	}

	// Persistent runtime metadata
	if fileCfg.SysStart != "" {
		cm.SysStart = fileCfg.SysStart
	}
	// no runtime KeyValues persisted here anymore
	// Backwards-compat: some configs place UI-related fields under a `ui` map
	// (e.g. config.yaml uses `ui: { themes_select: themes/2025 }`). Parse the
	// raw YAML and copy `ui.themes_select` into the top-level ThemesSelect when
	// present so ServeAdminPage and static file routes resolve correctly.
	var raw map[string]any
	if err := yaml.Unmarshal(b, &raw); err == nil && raw != nil {
		if uiRaw, ok := raw["ui"]; ok {
			if uiMap, ok2 := uiRaw.(map[string]any); ok2 {
				if ts, ok3 := uiMap["themes_select"].(string); ok3 && ts != "" {
					cm.ThemesSelect = ts
				}
				if bg, ok3 := uiMap["background"].(string); ok3 && bg != "" {
					cm.Background = bg
				}
				if pe, ok3 := uiMap["page_explain"].(string); ok3 && pe != "" {
					cm.PageExplain = pe
				}
				if opacityVal, ok3 := uiMap["opacity"]; ok3 {
					// Keep existing numeric parsing in callers; only set when simple types present
					switch v := opacityVal.(type) {
					case float64:
						if v != 0 {
							cm.Opacity = v
						}
					case int:
						if float64(v) != 0 {
							cm.Opacity = float64(v)
						}
					}
				}
			}
		}
	}
	return nil
}

// InitWithDB has been removed. Use SetDB(db) to inject a database connection.

func (cm *ConfigManager) ReloadConfig() error {
	// ReloadConfig no longer reads configuration from the database.
	// Configuration should be provided via `config.yaml` and environment variables.
	// Preserve in-memory immutable fields across reload (port).
	curPort := cm.Base.Port
	cm.applyEnvironmentOverrides()
	cm.Base.Port = curPort
	return nil
}

// PersistYAML writes the current ConfigManager to the YAML config file (CONFIG_PATH or ./config.yaml).
// It serializes the entire struct (yaml tags) and overwrites the file.
func (cm *ConfigManager) PersistYAML() error {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "./config.yaml"
	}

	// Ensure UI block reflects top-level compatibility fields before marshalling
	if cm.UI == nil {
		cm.UI = &UIConfig{}
	}
	cm.UI.ThemesSelect = cm.ThemesSelect
	cm.UI.Background = cm.Background
	cm.UI.PageExplain = cm.PageExplain
	cm.UI.Opacity = cm.Opacity
	cm.UI.RobotsText = cm.RobotsText
	cm.UI.ShowAdminAddr = cm.ShowAdminAddr

	out, err := yaml.Marshal(cm)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

// Runtime key/value helpers removed: runtime arbitrary key storage is no longer
// persisted inside `ConfigManager`. Configuration should be updated via the
// structured module Update(...) methods and persisted with PersistYAML().

func (cm *ConfigManager) applyEnvironmentOverrides() {
	if p := os.Getenv("PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			cm.Base.Port = n
		}
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
	if cm.UI != nil {
		ui := *cm.UI
		nc.UI = &ui
	}
	nc.SysStart = cm.SysStart
	// AdminToken removed
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
