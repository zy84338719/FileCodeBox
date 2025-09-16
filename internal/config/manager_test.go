package config

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestLoadFromYAML ensures YAML fields are loaded and marked as yamlManaged
func TestLoadFromYAML(t *testing.T) {
	data := map[string]interface{}{
		"base":          map[string]interface{}{"name": "TCB", "port": 12345},
		"themes_select": "themes/test",
		"page_explain":  "test explain",
		"notify_title":  "nt",
	}
	b, err := yaml.Marshal(data)
	if err != nil {
		t.Fatalf("yaml marshal: %v", err)
	}
	f := "./test_config.yaml"
	if err := os.WriteFile(f, b, 0644); err != nil {
		t.Fatalf("write tmp yaml: %v", err)
	}
	defer os.Remove(f)

	cm := NewConfigManager()
	if err := cm.LoadFromYAML(f); err != nil {
		t.Fatalf("LoadFromYAML failed: %v", err)
	}
	if cm.Base == nil || cm.Base.Name != "TCB" {
		t.Fatalf("expected base.name TCB, got %#v", cm.Base)
	}
	if cm.ThemesSelect != "themes/test" {
		t.Fatalf("expected themes_select themes/test, got %s", cm.ThemesSelect)
	}
	// Ensure at least one of Base.ToMap() keys is present in yamlManagedKeys
	found := false
	for k := range cm.Base.ToMap() {
		if cm.yamlManagedKeys[k] {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected some base.* key to be recorded in yamlManagedKeys, got none")
	}
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
	defer os.Remove(f)

	os.Setenv("PORT", "9090")
	defer os.Unsetenv("PORT")

	cm := NewConfigManager()
	if err := cm.LoadFromYAML(f); err != nil {
		t.Fatalf("LoadFromYAML failed: %v", err)
	}
	cm.applyEnvironmentOverrides()
	if cm.Base.Port != 9090 {
		t.Fatalf("expected PORT env to override to 9090, got %d", cm.Base.Port)
	}
}
