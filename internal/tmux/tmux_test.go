package tmux

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCommandsWindowsLayout(t *testing.T) {
	s := &Session{
		Name:   "clover",
		Layout: "windows",
		Entries: []Entry{
			{Name: "repo-a", WorkDir: "/code/a", Command: []string{"claude", "--model", "opus"}},
			{Name: "repo-b", WorkDir: "/code/b", Command: []string{"claude", "--model", "sonnet"}},
		},
	}

	cmds := BuildCommands(s)
	require.NotEmpty(t, cmds)

	// First command: new-session
	assert.Equal(t, []string{
		"tmux", "new-session", "-d",
		"-s", "clover",
		"-n", "repo-a",
		"-c", "/code/a",
	}, cmds[0])

	// Second: send-keys for first window
	assert.Equal(t, "tmux", cmds[1][0])
	assert.Equal(t, "send-keys", cmds[1][1])
	assert.Contains(t, cmds[1][4], "claude --model opus")

	// Third: new-window for second entry
	assert.Equal(t, []string{
		"tmux", "new-window",
		"-t", "clover",
		"-n", "repo-b",
		"-c", "/code/b",
	}, cmds[2])

	// Fourth: send-keys for second window
	assert.Contains(t, cmds[3][4], "claude --model sonnet")

	// Last: select first window
	assert.Equal(t, []string{"tmux", "select-window", "-t", "clover:0"}, cmds[len(cmds)-1])
}

func TestBuildCommandsPanesLayout(t *testing.T) {
	s := &Session{
		Name:   "test",
		Layout: "panes",
		Entries: []Entry{
			{Name: "a", WorkDir: "/code/a", Command: []string{"claude"}},
			{Name: "b", WorkDir: "/code/b", Command: []string{"claude"}},
		},
	}

	cmds := BuildCommands(s)
	require.NotEmpty(t, cmds)

	// First: new-session
	assert.Equal(t, "new-session", cmds[0][1])
	// Second: send-keys
	assert.Equal(t, "send-keys", cmds[1][1])
	// Third: split-window
	assert.Equal(t, "split-window", cmds[2][1])
	// Fourth: send-keys
	assert.Equal(t, "send-keys", cmds[3][1])
	// Last: select-layout tiled
	last := cmds[len(cmds)-1]
	assert.Equal(t, []string{"tmux", "select-layout", "-t", "test", "tiled"}, last)
}

func TestBuildCommandsEmpty(t *testing.T) {
	s := &Session{Name: "empty", Layout: "windows"}
	cmds := BuildCommands(s)
	assert.Nil(t, cmds)
}

func TestBuildCommandsSingleEntry(t *testing.T) {
	s := &Session{
		Name:   "single",
		Layout: "windows",
		Entries: []Entry{
			{Name: "only", WorkDir: "/code/only", Command: []string{"claude"}},
		},
	}

	cmds := BuildCommands(s)
	// new-session, send-keys, select-window
	assert.Len(t, cmds, 3)
}

func TestShellJoin(t *testing.T) {
	assert.Equal(t, "claude --model opus", shellJoin([]string{"claude", "--model", "opus"}))
	assert.Equal(t, "claude '--flag=with space'", shellJoin([]string{"claude", "--flag=with space"}))
}

func TestAttachCmd(t *testing.T) {
	assert.Equal(t, []string{"tmux", "attach-session", "-t", "my-session"}, AttachCmd("my-session"))
}

func TestBuildReconcileCommandsNewWindow(t *testing.T) {
	s := &Session{
		Name:   "clover",
		Layout: "windows",
		Entries: []Entry{
			{Name: "repo-a", WorkDir: "/code/a", Command: []string{"claude"}, Instances: 1},
		},
	}
	existing := map[string]int{} // session exists but window is absent

	cmds := BuildReconcileCommands(s, existing)
	require.NotEmpty(t, cmds)

	assert.Equal(t, "new-window", cmds[0][1])
	assert.Contains(t, cmds[0], "repo-a")
	assert.Equal(t, "send-keys", cmds[1][1])
}

func TestBuildReconcileCommandsMissingPanes(t *testing.T) {
	s := &Session{
		Name:   "clover",
		Layout: "windows",
		Entries: []Entry{
			{Name: "repo-a", WorkDir: "/code/a", Command: []string{"claude"}, Instances: 3},
		},
	}
	existing := map[string]int{"repo-a": 1} // window exists with 1 pane, want 3

	cmds := BuildReconcileCommands(s, existing)
	require.NotEmpty(t, cmds)

	var splits, sends int
	var hasTiled bool
	for _, c := range cmds {
		switch c[1] {
		case "split-window":
			splits++
		case "send-keys":
			sends++
		case "select-layout":
			if c[len(c)-1] == "tiled" {
				hasTiled = true
			}
		}
	}
	assert.Equal(t, 2, splits, "2 splits to go from 1 to 3 panes")
	assert.Equal(t, 2, sends, "1 send-keys per new pane")
	assert.True(t, hasTiled)
}

func TestBuildReconcileCommandsNoop(t *testing.T) {
	s := &Session{
		Name:   "clover",
		Layout: "windows",
		Entries: []Entry{
			{Name: "repo-a", WorkDir: "/code/a", Command: []string{"claude"}, Instances: 2},
		},
	}
	existing := map[string]int{"repo-a": 2} // already at desired count

	cmds := BuildReconcileCommands(s, existing)
	assert.Empty(t, cmds)
}

func TestBuildReconcileCommandsMixed(t *testing.T) {
	s := &Session{
		Name:   "clover",
		Layout: "windows",
		Entries: []Entry{
			{Name: "repo-a", WorkDir: "/code/a", Command: []string{"claude"}, Instances: 2}, // missing window
			{Name: "repo-b", WorkDir: "/code/b", Command: []string{"claude"}, Instances: 3}, // has 1, needs 3
			{Name: "repo-c", WorkDir: "/code/c", Command: []string{"claude"}, Instances: 1}, // satisfied
		},
	}
	existing := map[string]int{
		"repo-b": 1,
		"repo-c": 2, // more than desired — leave alone
	}

	cmds := BuildReconcileCommands(s, existing)
	require.NotEmpty(t, cmds)

	// repo-a: new-window + send-keys (instances=2 → splitPaneCmds adds 1 split + tiled)
	// repo-b: 2 splits + 2 sends + tiled
	// repo-c: nothing
	var newWindows, splits int
	for _, c := range cmds {
		switch c[1] {
		case "new-window":
			newWindows++
		case "split-window":
			splits++
		}
	}
	assert.Equal(t, 1, newWindows, "only repo-a gets a new window")
	assert.Equal(t, 3, splits, "1 for repo-a extra pane + 2 for repo-b")
}

func TestBuildCommandsWindowsMultiInstance(t *testing.T) {
	s := &Session{
		Name:   "clover",
		Layout: "windows",
		Entries: []Entry{
			{Name: "repo-a", WorkDir: "/code/a", Command: []string{"claude"}, Instances: 3},
		},
	}

	cmds := BuildCommands(s)
	require.NotEmpty(t, cmds)

	// Count command types
	var splitWindows, sendKeys int
	var hasLayout bool
	for _, c := range cmds {
		switch c[1] {
		case "split-window":
			splitWindows++
		case "send-keys":
			sendKeys++
		case "select-layout":
			if len(c) >= 5 && c[4] == "tiled" {
				hasLayout = true
			}
		}
	}

	assert.Equal(t, 2, splitWindows, "expect 2 split-windows for 3 instances")
	assert.Equal(t, 3, sendKeys, "expect 3 send-keys for 3 instances")
	assert.True(t, hasLayout, "expect tiled layout when instances > 1")
}

func TestBuildCommandsWindowsMultiInstanceMultiRepo(t *testing.T) {
	s := &Session{
		Name:   "clover",
		Layout: "windows",
		Entries: []Entry{
			{Name: "repo-a", WorkDir: "/code/a", Command: []string{"claude"}, Instances: 2},
			{Name: "repo-b", WorkDir: "/code/b", Command: []string{"claude"}, Instances: 3},
		},
	}

	cmds := BuildCommands(s)
	require.NotEmpty(t, cmds)

	var splitWindows, sendKeys int
	for _, c := range cmds {
		switch c[1] {
		case "split-window":
			splitWindows++
		case "send-keys":
			sendKeys++
		}
	}

	// repo-a: 1 split; repo-b: 2 splits
	assert.Equal(t, 3, splitWindows)
	// repo-a: 2 send-keys; repo-b: 3 send-keys
	assert.Equal(t, 5, sendKeys)
}

func TestBuildCommandsWindowsDefaultInstances(t *testing.T) {
	s := &Session{
		Name:   "clover",
		Layout: "windows",
		Entries: []Entry{
			{Name: "repo-a", WorkDir: "/code/a", Command: []string{"claude"}}, // Instances zero-value
		},
	}

	cmds := BuildCommands(s)
	require.NotEmpty(t, cmds)

	for _, c := range cmds {
		assert.NotEqual(t, "split-window", c[1], "zero Instances should produce no splits")
		if c[1] == "select-layout" {
			assert.NotEqual(t, "tiled", c[len(c)-1], "no tiled layout when single instance")
		}
	}
}
