package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/elliottmoos/clover/internal/paths"
	"gopkg.in/yaml.v3"
)

// Load returns the merged config: built-in defaults < global config < per-repo config.
// repoPath may be empty to skip per-repo loading.
func Load(repoPath string) (*Config, error) {
	cfg := DefaultConfig()

	// Layer global config
	globalPath, err := paths.GlobalConfigFile()
	if err != nil {
		return nil, err
	}
	global, err := loadFile(globalPath)
	if err != nil {
		return nil, fmt.Errorf("loading global config: %w", err)
	}
	if global != nil {
		merge(&cfg, global)
	}

	// Layer per-repo config
	if repoPath != "" {
		repoConfigPath := filepath.Join(repoPath, ".clover.yaml")
		repo, err := loadFile(repoConfigPath)
		if err != nil {
			return nil, fmt.Errorf("loading repo config: %w", err)
		}
		if repo != nil {
			merge(&cfg, repo)
		}
	}

	return &cfg, nil
}

// LoadGlobal loads only the global config file (without merging defaults or repo config).
func LoadGlobal() (*Config, error) {
	globalPath, err := paths.GlobalConfigFile()
	if err != nil {
		return nil, err
	}
	cfg, err := loadFile(globalPath)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		empty := Config{}
		return &empty, nil
	}
	return cfg, nil
}

// SaveGlobal writes the config to the global config file.
func SaveGlobal(cfg *Config) error {
	globalPath, err := paths.GlobalConfigFile()
	if err != nil {
		return err
	}

	dir := filepath.Dir(globalPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return os.WriteFile(globalPath, data, 0o644)
}

// InitGlobal writes the default config to the global config file.
// Returns an error if the file already exists.
func InitGlobal() (string, error) {
	globalPath, err := paths.GlobalConfigFile()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(globalPath); err == nil {
		return globalPath, fmt.Errorf("config already exists at %s", globalPath)
	}

	cfg := DefaultConfig()
	if err := SaveGlobal(&cfg); err != nil {
		return "", err
	}
	return globalPath, nil
}

// Get retrieves a config value by dot-separated key (e.g. "claude.model").
func Get(cfg *Config, key string) (string, error) {
	switch key {
	case "claude.model":
		return cfg.Claude.Model, nil
	case "claude.print":
		return fmt.Sprintf("%v", cfg.Claude.Print), nil
	case "claude.continue":
		return fmt.Sprintf("%v", cfg.Claude.Continue), nil
	case "claude.additional_flags":
		return strings.Join(cfg.Claude.AdditionalFlags, ","), nil
	case "session.layout":
		return cfg.Session.Layout, nil
	case "session.session_name":
		return cfg.Session.SessionName, nil
	case "session.instances":
		return fmt.Sprintf("%d", cfg.Session.Instances), nil
	case "session.max_instances":
		return fmt.Sprintf("%d", cfg.Session.MaxInstances), nil
	default:
		return "", fmt.Errorf("unknown config key %q", key)
	}
}

// Set updates a config value by dot-separated key.
func Set(cfg *Config, key, value string) error {
	switch key {
	case "claude.model":
		cfg.Claude.Model = value
	case "claude.print":
		cfg.Claude.Print = value == "true"
	case "claude.continue":
		cfg.Claude.Continue = value == "true"
	case "claude.additional_flags":
		if value == "" {
			cfg.Claude.AdditionalFlags = nil
		} else {
			cfg.Claude.AdditionalFlags = strings.Split(value, ",")
		}
	case "session.layout":
		if value != "windows" && value != "panes" {
			return fmt.Errorf("layout must be 'windows' or 'panes'")
		}
		cfg.Session.Layout = value
	case "session.session_name":
		cfg.Session.SessionName = value
	case "session.instances":
		n, err := strconv.Atoi(value)
		if err != nil || n < 1 || n > cfg.Session.MaxInstances {
			return fmt.Errorf("session.instances must be between 1 and %d", cfg.Session.MaxInstances)
		}
		cfg.Session.Instances = n
	case "session.max_instances":
		n, err := strconv.Atoi(value)
		if err != nil || n < 1 {
			return fmt.Errorf("session.max_instances must be >= 1")
		}
		cfg.Session.MaxInstances = n
	default:
		return fmt.Errorf("unknown config key %q", key)
	}
	return nil
}

func loadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return &cfg, nil
}

// merge applies non-zero values from src onto dst.
// Scalars: src replaces dst if non-zero.
// AdditionalFlags: appended (not replaced).
func merge(dst, src *Config) {
	if src.Claude.Model != "" {
		dst.Claude.Model = src.Claude.Model
	}
	if src.Claude.Print {
		dst.Claude.Print = true
	}
	if src.Claude.Continue {
		dst.Claude.Continue = true
	}
	if len(src.Claude.AdditionalFlags) > 0 {
		dst.Claude.AdditionalFlags = append(dst.Claude.AdditionalFlags, src.Claude.AdditionalFlags...)
	}

	if src.Session.Layout != "" {
		dst.Session.Layout = src.Session.Layout
	}
	if src.Session.SessionName != "" {
		dst.Session.SessionName = src.Session.SessionName
	}
	if src.Session.Instances != 0 {
		dst.Session.Instances = src.Session.Instances
	}
	if src.Session.MaxInstances != 0 {
		dst.Session.MaxInstances = src.Session.MaxInstances
	}
}
