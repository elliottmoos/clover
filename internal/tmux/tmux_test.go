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
