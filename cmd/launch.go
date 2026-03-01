package cmd

import (
	"fmt"
	"os"
	"syscall"

	"github.com/elliottmoos/clover/internal/config"
	"github.com/elliottmoos/clover/internal/launcher"
	"github.com/elliottmoos/clover/internal/registry"
	"github.com/spf13/cobra"
)

var launchCmd = &cobra.Command{
	Use:   "launch <name>",
	Short: "Launch Claude Code in a registered repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		reg, err := registry.Load()
		if err != nil {
			return err
		}

		repo := reg.Find(name)
		if repo == nil {
			return fmt.Errorf("repo %q not found", name)
		}

		// Verify repo path still exists
		if _, err := os.Stat(repo.Path); err != nil {
			return fmt.Errorf("repo path %s no longer exists: %w", repo.Path, err)
		}

		// Load merged config
		cfg, err := config.Load(repo.Path)
		if err != nil {
			return err
		}

		// Apply CLI flag overrides
		if model, _ := cmd.Flags().GetString("model"); model != "" {
			cfg.Claude.Model = model
		}
		if printFlag, _ := cmd.Flags().GetBool("print"); printFlag {
			cfg.Claude.Print = true
		}
		if cont, _ := cmd.Flags().GetBool("continue"); cont {
			cfg.Claude.Continue = true
		}
		if flags, _ := cmd.Flags().GetStringArray("flag"); len(flags) > 0 {
			cfg.Claude.AdditionalFlags = append(cfg.Claude.AdditionalFlags, flags...)
		}

		// Find claude binary
		claudePath, err := launcher.FindClaude()
		if err != nil {
			return err
		}

		// Update last_used_at
		_ = reg.TouchLastUsed(name)
		_ = reg.Save()

		// Build command
		cmdArgs := launcher.BuildCommand(cfg)

		// Change to repo directory
		if err := os.Chdir(repo.Path); err != nil {
			return fmt.Errorf("changing to repo dir: %w", err)
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "Launching: %v in %s\n", cmdArgs, repo.Path)
		}

		// Replace process with claude
		return syscall.Exec(claudePath, cmdArgs, os.Environ())
	},
}

func init() {
	launchCmd.Flags().StringP("model", "m", "", "Override the claude model")
	launchCmd.Flags().BoolP("print", "p", false, "Use claude --print mode")
	launchCmd.Flags().BoolP("continue", "c", false, "Continue previous conversation")
	launchCmd.Flags().StringArray("flag", nil, "Additional flags to pass to claude")
	rootCmd.AddCommand(launchCmd)
}
