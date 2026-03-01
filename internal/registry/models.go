package registry

import "time"

// Repo represents a registered git repository.
type Repo struct {
	Name       string    `yaml:"name"`
	Path       string    `yaml:"path"`
	Focused    bool      `yaml:"focused,omitempty"`
	AddedAt    time.Time `yaml:"added_at"`
	LastUsedAt time.Time `yaml:"last_used_at,omitempty"`
}
