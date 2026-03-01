package launcher

import (
	"testing"

	"github.com/elliottmoos/clover/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestBuildCommandDefaults(t *testing.T) {
	cfg := config.DefaultConfig()
	args := BuildCommand(&cfg)
	assert.Equal(t, []string{"claude", "--model", "sonnet"}, args)
}

func TestBuildCommandAllFlags(t *testing.T) {
	cfg := &config.Config{
		Claude: config.ClaudeFlags{
			Model:           "opus",
			Print:           true,
			Continue:        true,
			AdditionalFlags: []string{"--verbose", "--no-cache"},
		},
	}
	args := BuildCommand(cfg)
	assert.Equal(t, []string{
		"claude",
		"--model", "opus",
		"--print",
		"--continue",
		"--verbose",
		"--no-cache",
	}, args)
}

func TestBuildCommandNoModel(t *testing.T) {
	cfg := &config.Config{}
	args := BuildCommand(cfg)
	assert.Equal(t, []string{"claude"}, args)
}

func TestBuildCommandPrintOnly(t *testing.T) {
	cfg := &config.Config{
		Claude: config.ClaudeFlags{
			Print: true,
		},
	}
	args := BuildCommand(cfg)
	assert.Equal(t, []string{"claude", "--print"}, args)
}
