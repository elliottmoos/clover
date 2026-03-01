package cmd

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/elliottmoos/clover/internal/config"
	"github.com/elliottmoos/clover/internal/launcher"
	"github.com/elliottmoos/clover/internal/registry"
	"github.com/elliottmoos/clover/internal/tmux"
	"github.com/spf13/cobra"
)

var sessionCmd = &cobra.Command{
	Use:   "session [names...]",
	Short: "Launch a tmux session with Claude Code in multiple repos",
	RunE: func(cmd *cobra.Command, args []string) error {
		focusFlag, _ := cmd.Flags().GetBool("focus")
		allFlag, _ := cmd.Flags().GetBool("all")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		layout, _ := cmd.Flags().GetString("layout")
		sessionName, _ := cmd.Flags().GetString("name")
		instancesFlag, _ := cmd.Flags().GetInt("instances")

		reg, err := registry.Load()
		if err != nil {
			return err
		}

		// Determine which repos to include
		var repos []registry.Repo
		switch {
		case allFlag:
			repos = reg.List()
		case focusFlag:
			repos = reg.Focused()
		case len(args) > 0:
			for _, name := range args {
				repo := reg.Find(name)
				if repo == nil {
					return fmt.Errorf("repo %q not found", name)
				}
				repos = append(repos, *repo)
			}
		default:
			return fmt.Errorf("specify repo names, --focus, or --all")
		}

		if len(repos) == 0 {
			return fmt.Errorf("no repos matched")
		}

		// Find tmux
		tmuxPath, err := tmux.FindTmux()
		if err != nil {
			return err
		}

		// Load session defaults from config
		globalCfg, err := config.Load("")
		if err != nil {
			return err
		}

		if layout == "" {
			layout = globalCfg.Session.Layout
		}
		if sessionName == "" {
			sessionName = globalCfg.Session.SessionName
		}

		// Validate CLI --instances flag against global max
		if instancesFlag > 0 && instancesFlag > globalCfg.Session.MaxInstances {
			return fmt.Errorf("--instances %d exceeds max_instances (%d)", instancesFlag, globalCfg.Session.MaxInstances)
		}

		// Build entries
		var entries []tmux.Entry
		for _, repo := range repos {
			repoCfg, err := config.Load(repo.Path)
			if err != nil {
				return fmt.Errorf("loading config for %q: %w", repo.Name, err)
			}

			// Determine instances: CLI flag > per-repo config
			instances := repoCfg.Session.Instances
			if instancesFlag > 0 {
				instances = instancesFlag
			} else if instances > repoCfg.Session.MaxInstances {
				return fmt.Errorf("repo %q: instances (%d) exceeds max_instances (%d)", repo.Name, instances, repoCfg.Session.MaxInstances)
			}

			cmdArgs := launcher.BuildCommand(repoCfg)
			entries = append(entries, tmux.Entry{
				Name:      repo.Name,
				WorkDir:   repo.Path,
				Command:   cmdArgs,
				Instances: instances,
			})
		}

		session := &tmux.Session{
			Name:    sessionName,
			Layout:  layout,
			Entries: entries,
		}

		// Build tmux commands: reconcile if session exists (windows layout only),
		// otherwise create from scratch. For panes layout with an existing session
		// we skip modification and just attach.
		var cmds [][]string
		sessionExists := !dryRun && tmux.SessionExists(sessionName)
		if sessionExists && layout == "windows" {
			existing, err := tmux.ListWindows(sessionName)
			if err != nil {
				return err
			}
			cmds = tmux.BuildReconcileCommands(session, existing)
		} else if !sessionExists {
			cmds = tmux.BuildCommands(session)
		}

		if dryRun {
			for _, c := range cmds {
				fmt.Println(strings.Join(c, " "))
			}
			return nil
		}

		// Execute tmux commands
		if err := tmux.Execute(cmds); err != nil {
			return err
		}

		// Attach to the session (replaces this process)
		attachArgs := tmux.AttachCmd(sessionName)
		return syscall.Exec(tmuxPath, attachArgs, os.Environ())
	},
}

func init() {
	sessionCmd.Flags().Bool("focus", false, "Include only focused repos")
	sessionCmd.Flags().Bool("all", false, "Include all registered repos")
	sessionCmd.Flags().Bool("dry-run", false, "Print tmux commands without executing")
	sessionCmd.Flags().String("layout", "", "Layout: windows or panes (defaults to config)")
	sessionCmd.Flags().String("name", "", "Tmux session name (defaults to config)")
	sessionCmd.Flags().Int("instances", 0, "Number of claude panes per window (0 = use config)")
	rootCmd.AddCommand(sessionCmd)
}
