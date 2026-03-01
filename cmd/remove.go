package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/elliottmoos/clover/internal/registry"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Unregister a repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		force, _ := cmd.Flags().GetBool("force")

		reg, err := registry.Load()
		if err != nil {
			return err
		}

		repo := reg.Find(name)
		if repo == nil {
			return fmt.Errorf("repo %q not found", name)
		}

		if !force {
			fmt.Printf("Remove %q (%s)? [y/N] ", name, repo.Path)
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(answer)) != "y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := reg.Remove(name); err != nil {
			return err
		}

		if err := reg.Save(); err != nil {
			return err
		}

		fmt.Printf("Removed %q\n", name)
		return nil
	},
}

func init() {
	removeCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	rootCmd.AddCommand(removeCmd)
}
