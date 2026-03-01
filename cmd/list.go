package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/elliottmoos/clover/internal/registry"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered repositories",
	RunE: func(cmd *cobra.Command, args []string) error {
		focusOnly, _ := cmd.Flags().GetBool("focus")
		format, _ := cmd.Flags().GetString("format")

		reg, err := registry.Load()
		if err != nil {
			return err
		}

		var repos []registry.Repo
		if focusOnly {
			repos = reg.Focused()
		} else {
			repos = reg.List()
		}

		if len(repos) == 0 {
			fmt.Println("No repositories registered.")
			return nil
		}

		switch format {
		case "json":
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(repos)
		default:
			return printTable(repos)
		}
	},
}

func printTable(repos []registry.Repo) error {
	tw := tablewriter.NewTable(os.Stdout)

	tw.Header("Name", "Path", "Focused")

	for _, r := range repos {
		focus := ""
		if r.Focused {
			focus = "*"
		}
		tw.Append(r.Name, r.Path, focus)
	}

	tw.Render()
	return nil
}

func init() {
	listCmd.Flags().Bool("focus", false, "Show only focused repositories")
	listCmd.Flags().String("format", "table", "Output format: table or json")
	rootCmd.AddCommand(listCmd)
}
