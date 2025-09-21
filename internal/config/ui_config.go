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
