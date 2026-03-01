package cmd

import (
	"fmt"

	"github.com/elliottmoos/clover/internal/registry"
	"github.com/spf13/cobra"
)

var focusCmd = &cobra.Command{
	Use:   "focus <name>",
	Short: "Toggle focus status on a repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		reg, err := registry.Load()
		if err != nil {
			return err
		}

		if err := reg.ToggleFocus(name); err != nil {
			return err
		}

		if err := reg.Save(); err != nil {
			return err
		}

		repo := reg.Find(name)
		state := "unfocused"
		if repo.Focused {
			state = "focused"
		}
		fmt.Printf("%q is now %s\n", name, state)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(focusCmd)
}
