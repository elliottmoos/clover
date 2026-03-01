package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/elliottmoos/clover/internal/registry"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <path>",
	Short: "Register a git repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoPath := args[0]

		absPath, err := filepath.Abs(repoPath)
		if err != nil {
			return fmt.Errorf("resolving path: %w", err)
		}

		// Validate .git directory exists
		gitDir := filepath.Join(absPath, ".git")
		info, err := os.Stat(gitDir)
		if err != nil || !info.IsDir() {
			return fmt.Errorf("%s is not a git repository (no .git directory)", absPath)
		}

		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			name = filepath.Base(absPath)
		}

		reg, err := registry.Load()
		if err != nil {
			return err
		}

		if err := reg.Add(name, absPath); err != nil {
			return err
		}

		if err := reg.Save(); err != nil {
			return err
		}

		fmt.Printf("Added %q (%s)\n", name, absPath)
		return nil
	},
}

func init() {
	addCmd.Flags().StringP("name", "n", "", "Override the repository name (defaults to directory basename)")
	rootCmd.AddCommand(addCmd)
}
