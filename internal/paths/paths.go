package paths

import (
	"os"
	"path/filepath"
)

// ConfigDir returns the clover configuration directory.
// Respects XDG_CONFIG_HOME if set, otherwise defaults to ~/.config/clover.
func ConfigDir() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "clover"), nil
}

// RegistryFile returns the path to the registry YAML file.
func RegistryFile() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "registry.yaml"), nil
}

// GlobalConfigFile returns the path to the global config YAML file.
func GlobalConfigFile() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}
