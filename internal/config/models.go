package config

// ClaudeFlags holds settings that map to claude CLI flags.
type ClaudeFlags struct {
	Model           string   `yaml:"model,omitempty"`
	Print           bool     `yaml:"print,omitempty"`
	Continue        bool     `yaml:"continue,omitempty"`
	AdditionalFlags []string `yaml:"additional_flags,omitempty"`
}

// SessionDefaults holds default settings for tmux sessions.
type SessionDefaults struct {
	Layout      string `yaml:"layout,omitempty"`        // "windows" or "panes"
	SessionName string `yaml:"session_name,omitempty"`  // default tmux session name
	Instances   int    `yaml:"instances,omitempty"`     // number of claude panes per window
	MaxInstances int   `yaml:"max_instances,omitempty"` // upper bound for instances
}

// Config is the full configuration for clover.
type Config struct {
	Claude  ClaudeFlags     `yaml:"claude"`
	Session SessionDefaults `yaml:"session"`
}

// DefaultConfig returns the built-in default configuration.
func DefaultConfig() Config {
	return Config{
		Claude: ClaudeFlags{
			Model: "sonnet",
		},
		Session: SessionDefaults{
			Layout:       "windows",
			SessionName:  "clover",
			Instances:    1,
			MaxInstances: 10,
		},
	}
}
