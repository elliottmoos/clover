package cmd

import (
	"fmt"

	"github.com/elliottmoos/clover/internal/config"
	"github.com/elliottmoos/clover/internal/registry"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage clover configuration",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create default global configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.InitGlobal()
		if err != nil {
			return err
		}
		fmt.Printf("Created config at %s\n", path)
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show merged configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoName, _ := cmd.Flags().GetString("repo")

		var repoPath string
		if repoName != "" {
			reg, err := registry.Load()
			if err != nil {
				return err
			}
			repo := reg.Find(repoName)
			if repo == nil {
				return fmt.Errorf("repo %q not found", repoName)
			}
			repoPath = repo.Path
		}

		cfg, err := config.Load(repoPath)
		if err != nil {
			return err
		}

		data, err := yaml.Marshal(cfg)
		if err != nil {
			return err
		}
		fmt.Print(string(data))
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a global configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]

		cfg, err := config.LoadGlobal()
		if err != nil {
			return err
		}

		if err := config.Set(cfg, key, value); err != nil {
			return err
		}

		if err := config.SaveGlobal(cfg); err != nil {
			return err
		}

		fmt.Printf("Set %s = %s\n", key, value)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load("")
		if err != nil {
			return err
		}

		val, err := config.Get(cfg, args[0])
		if err != nil {
			return err
		}

		fmt.Println(val)
		return nil
	},
}

func init() {
	configShowCmd.Flags().String("repo", "", "Show config merged with a specific repo's .clover.yaml")
	configCmd.AddCommand(configInitCmd, configShowCmd, configSetCmd, configGetCmd)
	rootCmd.AddCommand(configCmd)
}
