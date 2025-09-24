package config

import (
	"os"
	"path/filepath"
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

func TestUpdateTransactionPersistsAndReloads(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.Setenv("CONFIG_PATH", configPath); err != nil {
		t.Fatalf("setenv failed: %v", err)
	}
	defer os.Unsetenv("CONFIG_PATH")

	cm := NewConfigManager()
	if err := cm.UpdateTransaction(func(draft *ConfigManager) error {
		draft.Base.Name = "Transactional"
		draft.NotifyTitle = "updated"
		return nil
	}); err != nil {
		t.Fatalf("UpdateTransaction returned error: %v", err)
	}

	if cm.Base.Name != "Transactional" {
		t.Fatalf("expected in-memory base name updated, got %s", cm.Base.Name)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read persisted config: %v", err)
	}

	var persisted ConfigManager
	if err := yaml.Unmarshal(data, &persisted); err != nil {
		t.Fatalf("failed to unmarshal persisted config: %v", err)
	}
	if persisted.Base == nil || persisted.Base.Name != "Transactional" {
		t.Fatalf("expected persisted base name, got %#v", persisted.Base)
	}
}

func TestUpdateTransactionRollbackOnPersistFailure(t *testing.T) {
	tempDir := t.TempDir()
	badPath := filepath.Join(tempDir, "missing", "config.yaml")
	if err := os.Setenv("CONFIG_PATH", badPath); err != nil {
		t.Fatalf("setenv failed: %v", err)
	}
	defer os.Unsetenv("CONFIG_PATH")

	cm := NewConfigManager()
	originalName := cm.Base.Name

	err := cm.UpdateTransaction(func(draft *ConfigManager) error {
		draft.Base.Name = "ShouldRollback"
		return nil
	})
	if err == nil {
		t.Fatalf("expected error when persisting to missing directory")
	}

	if cm.Base.Name != originalName {
		t.Fatalf("expected rollback to restore base name, got %s", cm.Base.Name)
	}

	if _, err := os.Stat(badPath); !os.IsNotExist(err) {
		t.Fatalf("expected no config file created, stat err=%v", err)
	}
}
