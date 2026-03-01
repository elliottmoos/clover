package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, "sonnet", cfg.Claude.Model)
	assert.Equal(t, "windows", cfg.Session.Layout)
	assert.Equal(t, "clover", cfg.Session.SessionName)
}

func TestMergeScalars(t *testing.T) {
	dst := DefaultConfig()
	src := &Config{
		Claude: ClaudeFlags{Model: "opus"},
	}
	merge(&dst, src)
	assert.Equal(t, "opus", dst.Claude.Model)
	// Unchanged defaults should persist
	assert.Equal(t, "windows", dst.Session.Layout)
}

func TestMergeEmptyDoesNotOverride(t *testing.T) {
	dst := DefaultConfig()
	src := &Config{}
	merge(&dst, src)
	assert.Equal(t, "sonnet", dst.Claude.Model)
}

func TestMergeAdditionalFlagsAppend(t *testing.T) {
	dst := Config{
		Claude: ClaudeFlags{AdditionalFlags: []string{"--verbose"}},
	}
	src := &Config{
		Claude: ClaudeFlags{AdditionalFlags: []string{"--debug"}},
	}
	merge(&dst, src)
	assert.Equal(t, []string{"--verbose", "--debug"}, dst.Claude.AdditionalFlags)
}

func TestLoadWithRepoOverride(t *testing.T) {
	// Set up temp XDG dir
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Create global config
	configDir := filepath.Join(tmpDir, "clover")
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	globalCfg := []byte("claude:\n  model: opus\nsession:\n  layout: panes\n")
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.yaml"), globalCfg, 0o644))

	// Create repo config
	repoDir := t.TempDir()
	repoCfg := []byte("claude:\n  model: haiku\n  additional_flags:\n    - \"--verbose\"\n")
	require.NoError(t, os.WriteFile(filepath.Join(repoDir, ".clover.yaml"), repoCfg, 0o644))

	cfg, err := Load(repoDir)
	require.NoError(t, err)

	// Repo overrides global
	assert.Equal(t, "haiku", cfg.Claude.Model)
	// Global value persists where repo doesn't override
	assert.Equal(t, "panes", cfg.Session.Layout)
	// Additional flags from repo
	assert.Equal(t, []string{"--verbose"}, cfg.Claude.AdditionalFlags)
}

func TestLoadMissingFiles(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg, err := Load("")
	require.NoError(t, err)
	// Should get defaults
	assert.Equal(t, "sonnet", cfg.Claude.Model)
}

func TestGetSet(t *testing.T) {
	cfg := DefaultConfig()

	require.NoError(t, Set(&cfg, "claude.model", "opus"))
	val, err := Get(&cfg, "claude.model")
	require.NoError(t, err)
	assert.Equal(t, "opus", val)

	require.NoError(t, Set(&cfg, "session.layout", "panes"))
	val, err = Get(&cfg, "session.layout")
	require.NoError(t, err)
	assert.Equal(t, "panes", val)
}

func TestSetInvalidLayout(t *testing.T) {
	cfg := DefaultConfig()
	err := Set(&cfg, "session.layout", "invalid")
	assert.ErrorContains(t, err, "must be")
}

func TestGetUnknownKey(t *testing.T) {
	cfg := DefaultConfig()
	_, err := Get(&cfg, "unknown.key")
	assert.ErrorContains(t, err, "unknown config key")
}

func TestSetUnknownKey(t *testing.T) {
	cfg := DefaultConfig()
	err := Set(&cfg, "unknown.key", "value")
	assert.ErrorContains(t, err, "unknown config key")
}
