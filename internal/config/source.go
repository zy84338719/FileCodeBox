package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConfigSource represents a configuration input that can mutate the manager state.
type ConfigSource interface {
	Apply(*ConfigManager) error
}

// ConfigSourceFunc allows plain functions to be used as ConfigSource.
type ConfigSourceFunc func(*ConfigManager) error

// Apply executes the underlying function.
func (f ConfigSourceFunc) Apply(cm *ConfigManager) error { return f(cm) }

// YAMLFileSource loads configuration values from a YAML file.
type YAMLFileSource struct {
	Path string
}

// Apply reads and merges YAML content into the manager.
func (s YAMLFileSource) Apply(cm *ConfigManager) error {
	if strings.TrimSpace(s.Path) == "" {
		return errors.New("config: YAML path is empty")
	}

	data, err := os.ReadFile(s.Path)
	if err != nil {
		return err
	}

	var fileCfg ConfigManager
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return fmt.Errorf("config: unmarshal %s: %w", s.Path, err)
	}

	cm.mergeConfigModules(&fileCfg)
	cm.mergeUserConfig(fileCfg.User)
	cm.mergeSimpleFields(&fileCfg)
	return nil
}

type envOverride struct {
	key   string
	apply func(string, *ConfigManager) error
}

// EnvSource mutates configuration using environment variables.
type EnvSource struct {
	overrides []envOverride
	lookup    func(string) string
}

// NewDefaultEnvSource returns the built-in environment overrides.
func NewDefaultEnvSource() EnvSource {
	return EnvSource{
		overrides: []envOverride{
			{key: "PORT", apply: applyPortOverride},
			{key: "DATA_PATH", apply: applyDataPathOverride},
			{key: "ENABLE_MCP_SERVER", apply: applyMCPEnabledOverride},
			{key: "MCP_PORT", apply: applyMCPPortOverride},
			{key: "MCP_HOST", apply: applyMCPHostOverride},
		},
	}
}

// Apply applies every configured override.
func (s EnvSource) Apply(cm *ConfigManager) error {
	lookup := s.lookup
	if lookup == nil {
		lookup = os.Getenv
	}

	var errs []error
	for _, override := range s.overrides {
		if value := lookup(override.key); value != "" {
			if err := override.apply(value, cm); err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", override.key, err))
			}
		}
	}

	return errors.Join(errs...)
}

func applyPortOverride(val string, cm *ConfigManager) error {
	port, err := strconv.Atoi(val)
	if err != nil {
		return fmt.Errorf("invalid port %q", val)
	}
	cm.Base.Port = port
	return nil
}

func applyDataPathOverride(val string, cm *ConfigManager) error {
	if strings.TrimSpace(val) == "" {
		return errors.New("data path cannot be blank")
	}
	cm.Base.DataPath = val
	return nil
}

func applyMCPEnabledOverride(val string, cm *ConfigManager) error {
	enabled, err := parseBool(val)
	if err != nil {
		return err
	}
	if enabled {
		cm.MCP.EnableMCPServer = 1
	} else {
		cm.MCP.EnableMCPServer = 0
	}
	return nil
}

func applyMCPPortOverride(val string, cm *ConfigManager) error {
	cm.MCP.MCPPort = val
	return nil
}

func applyMCPHostOverride(val string, cm *ConfigManager) error {
	cm.MCP.MCPHost = val
	return nil
}

func parseBool(val string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "1", "true", "t", "yes", "y":
		return true, nil
	case "0", "false", "f", "no", "n":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value %q", val)
	}
}
