package config

// UIConfig holds theme and page related configuration and is stored under `ui` in config.yaml.
type UIConfig struct {
	ThemesSelect  string  `yaml:"themes_select" json:"themes_select"`
	Background    string  `yaml:"background" json:"background"`
	PageExplain   string  `yaml:"page_explain" json:"page_explain"`
	RobotsText    string  `yaml:"robots_text" json:"robots_text"`
	ShowAdminAddr int     `yaml:"show_admin_addr" json:"show_admin_addr"`
	Opacity       float64 `yaml:"opacity" json:"opacity"`
}

func NewUIConfig() *UIConfig {
	return &UIConfig{}
}

func (u *UIConfig) Clone() *UIConfig {
	if u == nil {
		return NewUIConfig()
	}
	nu := *u
	return &nu
}

// Update applies values from a map to the UIConfig. It supports values typed as
// simple primitives (string/number).
func (u *UIConfig) Update(m map[string]interface{}) error {
	if u == nil {
		return nil
	}
	if v, ok := m["themes_select"].(string); ok {
		u.ThemesSelect = v
	}
	if v, ok := m["background"].(string); ok {
		u.Background = v
	}
	if v, ok := m["page_explain"].(string); ok {
		u.PageExplain = v
	}
	if v, ok := m["robots_text"].(string); ok {
		u.RobotsText = v
	}
	if v, ok := m["show_admin_addr"].(int); ok {
		u.ShowAdminAddr = v
	} else if v2, ok2 := m["show_admin_addr"].(float64); ok2 {
		u.ShowAdminAddr = int(v2)
	}
	if v, ok := m["opacity"].(float64); ok {
		u.Opacity = v
	} else if v2, ok2 := m["opacity"].(int); ok2 {
		u.Opacity = float64(v2)
	}
	return nil
}
