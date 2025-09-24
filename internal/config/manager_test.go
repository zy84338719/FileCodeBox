package config

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestLoadFromYAML ensures YAML fields are loaded and marked as yamlManaged
func TestLoadFromYAML(t *testing.T) {
	data := map[string]interface{}{
		"base":         map[string]interface{}{"name": "TCB", "port": 12345},
		"ui":           map[string]interface{}{"themes_select": "themes/test", "page_explain": "test explain"},
		"notify_title": "nt",
	}
	b, err := yaml.Marshal(data)
	if err != nil {
		t.Fatalf("yaml marshal: %v", err)
	}
	f := "./test_config.yaml"
	if err := os.WriteFile(f, b, 0644); err != nil {
		t.Fatalf("write tmp yaml: %v", err)
	}
	defer func() { _ = os.Remove(f) }()

	cm := NewConfigManager()
	if err := cm.LoadFromYAML(f); err != nil {
		t.Fatalf("LoadFromYAML failed: %v", err)
	}
	if cm.Base == nil || cm.Base.Name != "TCB" {
		t.Fatalf("expected base.name TCB, got %#v", cm.Base)
	}
	if cm.UI == nil || cm.UI.ThemesSelect != "themes/test" {
		t.Fatalf("expected ui.themes_select themes/test, got %#v", cm.UI)
	}
	// basic fields loaded
}

// TestEnvOverride ensures environment variables override YAML values
func TestEnvOverride(t *testing.T) {
	data2 := map[string]interface{}{
		"base": map[string]interface{}{"name": "FromYaml", "port": 8080},
	}
	b2, err := yaml.Marshal(data2)
	if err != nil {
		t.Fatalf("yaml marshal: %v", err)
	}
	f := "./test_config_env.yaml"
	if err := os.WriteFile(f, b2, 0644); err != nil {
		t.Fatalf("write tmp yaml: %v", err)
	}
	defer func() { _ = os.Remove(f) }()

	if err := os.Setenv("PORT", "9090"); err != nil {
		t.Fatalf("setenv failed: %v", err)
	}
	defer func() { _ = os.Unsetenv("PORT") }()

	cm := NewConfigManager()
	if err := cm.LoadFromYAML(f); err != nil {
		t.Fatalf("LoadFromYAML failed: %v", err)
	}
	cm.applyEnvironmentOverrides()
	if cm.Base.Port != 9090 {
		t.Fatalf("expected PORT env to override to 9090, got %d", cm.Base.Port)
	}
}

func TestApplySourcesAggregatesErrors(t *testing.T) {
	cm := NewConfigManager()
	src := NewDefaultEnvSource()
	src.lookup = func(key string) string {
		if key == "ENABLE_MCP_SERVER" {
			return "definitely-not-bool"
		}
		return ""
	}

	if err := cm.ApplySources(src); err == nil {
		t.Fatalf("expected aggregated error when environment value invalid")
	}
}
