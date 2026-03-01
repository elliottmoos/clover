package tmux

import (
	"fmt"
	"os/exec"
	"strings"
)

// Entry represents a repo to be launched in a tmux window or pane.
type Entry struct {
	Name    string   // window/pane name (repo name)
	WorkDir string   // working directory
	Command []string // command to send (e.g. ["claude", "--model", "opus"])
}

// Session represents a tmux session to be created.
type Session struct {
	Name    string
	Entries []Entry
	Layout  string // "windows" or "panes"
}

// FindTmux returns the path to the tmux binary.
func FindTmux() (string, error) {
	path, err := exec.LookPath("tmux")
	if err != nil {
		return "", fmt.Errorf("tmux not found in PATH: %w", err)
	}
	return path, nil
}

// SessionExists checks if a tmux session with the given name exists.
func SessionExists(name string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", name)
	return cmd.Run() == nil
}

// BuildCommands generates the tmux commands needed to create a session.
func BuildCommands(s *Session) [][]string {
	var cmds [][]string

	if len(s.Entries) == 0 {
		return nil
	}

	switch s.Layout {
	case "panes":
		cmds = buildPanesLayout(s)
	default: // "windows"
		cmds = buildWindowsLayout(s)
	}

	return cmds
}

func buildWindowsLayout(s *Session) [][]string {
	var cmds [][]string

	first := s.Entries[0]
	// Create new session with first entry
	cmds = append(cmds, []string{
		"tmux", "new-session", "-d",
		"-s", s.Name,
		"-n", first.Name,
		"-c", first.WorkDir,
	})
	// Send command to first window
	cmds = append(cmds, sendKeysCmd(s.Name, first.Name, first.Command))

	// Create additional windows
	for _, e := range s.Entries[1:] {
		cmds = append(cmds, []string{
			"tmux", "new-window",
			"-t", s.Name,
			"-n", e.Name,
			"-c", e.WorkDir,
		})
		cmds = append(cmds, sendKeysCmd(s.Name, e.Name, e.Command))
	}

	// Select first window
	cmds = append(cmds, []string{"tmux", "select-window", "-t", s.Name + ":0"})

	return cmds
}

func buildPanesLayout(s *Session) [][]string {
	var cmds [][]string

	first := s.Entries[0]
	// Create new session with first pane
	cmds = append(cmds, []string{
		"tmux", "new-session", "-d",
		"-s", s.Name,
		"-n", "clover",
		"-c", first.WorkDir,
	})
	cmds = append(cmds, sendKeysCmd(s.Name, "", first.Command))

	// Split for additional entries
	for _, e := range s.Entries[1:] {
		cmds = append(cmds, []string{
			"tmux", "split-window",
			"-t", s.Name,
			"-c", e.WorkDir,
		})
		cmds = append(cmds, sendKeysCmd(s.Name, "", e.Command))
	}

	// Apply tiled layout
	cmds = append(cmds, []string{
		"tmux", "select-layout", "-t", s.Name, "tiled",
	})

	return cmds
}

func sendKeysCmd(session, window string, command []string) []string {
	target := session
	if window != "" {
		target = session + ":" + window
	}
	cmdStr := shellJoin(command)
	return []string{"tmux", "send-keys", "-t", target, cmdStr, "Enter"}
}

// shellJoin joins command parts with proper quoting for send-keys.
func shellJoin(parts []string) string {
	quoted := make([]string, len(parts))
	for i, p := range parts {
		if strings.ContainsAny(p, " \t'\"\\") {
			quoted[i] = "'" + strings.ReplaceAll(p, "'", "'\\''") + "'"
		} else {
			quoted[i] = p
		}
	}
	return strings.Join(quoted, " ")
}

// Execute runs a list of tmux commands sequentially.
func Execute(cmds [][]string) error {
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("running %v: %s: %w", args, string(out), err)
		}
	}
	return nil
}

// AttachCmd returns the command to attach to a tmux session.
func AttachCmd(session string) []string {
	return []string{"tmux", "attach-session", "-t", session}
}
