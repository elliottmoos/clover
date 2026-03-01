package launcher

import (
	"fmt"
	"os/exec"

	"github.com/elliottmoos/clover/internal/config"
)

// BuildCommand constructs the claude CLI argument list from a merged config.
// The returned slice is suitable for syscall.Exec (first element is "claude").
func BuildCommand(cfg *config.Config) []string {
	args := []string{"claude"}

	if cfg.Claude.Model != "" {
		args = append(args, "--model", cfg.Claude.Model)
	}
	if cfg.Claude.Print {
		args = append(args, "--print")
	}
	if cfg.Claude.Continue {
		args = append(args, "--continue")
	}
	args = append(args, cfg.Claude.AdditionalFlags...)

	return args
}

// FindClaude returns the absolute path to the claude binary.
func FindClaude() (string, error) {
	path, err := exec.LookPath("claude")
	if err != nil {
		return "", fmt.Errorf("claude not found in PATH: %w", err)
	}
	return path, nil
}
