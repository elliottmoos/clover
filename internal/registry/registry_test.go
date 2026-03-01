package registry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tempRegistryPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "registry.yaml")
}

func TestAddAndFind(t *testing.T) {
	r := &Registry{path: tempRegistryPath(t)}

	// Create a fake git repo dir
	repoDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repoDir, ".git"), 0o755))

	require.NoError(t, r.Add("myrepo", repoDir))
	assert.Len(t, r.Repos, 1)

	found := r.Find("myrepo")
	require.NotNil(t, found)
	assert.Equal(t, "myrepo", found.Name)

	absDir, _ := filepath.Abs(repoDir)
	assert.Equal(t, absDir, found.Path)
}

func TestAddDuplicateName(t *testing.T) {
	r := &Registry{path: tempRegistryPath(t)}
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	require.NoError(t, r.Add("foo", dir1))
	err := r.Add("foo", dir2)
	assert.ErrorContains(t, err, "already registered")
}

func TestAddDuplicatePath(t *testing.T) {
	r := &Registry{path: tempRegistryPath(t)}
	dir := t.TempDir()

	require.NoError(t, r.Add("foo", dir))
	err := r.Add("bar", dir)
	assert.ErrorContains(t, err, "already registered")
}

func TestRemove(t *testing.T) {
	r := &Registry{path: tempRegistryPath(t)}
	dir := t.TempDir()

	require.NoError(t, r.Add("foo", dir))
	require.NoError(t, r.Remove("foo"))
	assert.Empty(t, r.Repos)
}

func TestRemoveNotFound(t *testing.T) {
	r := &Registry{path: tempRegistryPath(t)}
	err := r.Remove("nope")
	assert.ErrorContains(t, err, "not found")
}

func TestListSortsFocusedFirst(t *testing.T) {
	r := &Registry{path: tempRegistryPath(t)}
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	dir3 := t.TempDir()

	require.NoError(t, r.Add("bravo", dir1))
	require.NoError(t, r.Add("alpha", dir2))
	require.NoError(t, r.Add("charlie", dir3))
	require.NoError(t, r.ToggleFocus("charlie"))

	list := r.List()
	assert.Equal(t, "charlie", list[0].Name)
	assert.Equal(t, "alpha", list[1].Name)
	assert.Equal(t, "bravo", list[2].Name)
}

func TestToggleFocus(t *testing.T) {
	r := &Registry{path: tempRegistryPath(t)}
	dir := t.TempDir()

	require.NoError(t, r.Add("foo", dir))
	assert.False(t, r.Find("foo").Focused)

	require.NoError(t, r.ToggleFocus("foo"))
	assert.True(t, r.Find("foo").Focused)

	require.NoError(t, r.ToggleFocus("foo"))
	assert.False(t, r.Find("foo").Focused)
}

func TestToggleFocusNotFound(t *testing.T) {
	r := &Registry{path: tempRegistryPath(t)}
	err := r.ToggleFocus("nope")
	assert.ErrorContains(t, err, "not found")
}

func TestFocused(t *testing.T) {
	r := &Registry{path: tempRegistryPath(t)}
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	require.NoError(t, r.Add("foo", dir1))
	require.NoError(t, r.Add("bar", dir2))
	require.NoError(t, r.ToggleFocus("bar"))

	focused := r.Focused()
	assert.Len(t, focused, 1)
	assert.Equal(t, "bar", focused[0].Name)
}

func TestSaveAndLoad(t *testing.T) {
	path := tempRegistryPath(t)
	r := &Registry{path: path}
	dir := t.TempDir()

	require.NoError(t, r.Add("foo", dir))
	require.NoError(t, r.Save())

	loaded, err := LoadFrom(path)
	require.NoError(t, err)
	assert.Len(t, loaded.Repos, 1)
	assert.Equal(t, "foo", loaded.Repos[0].Name)
}

func TestLoadNonExistent(t *testing.T) {
	r, err := LoadFrom(filepath.Join(t.TempDir(), "nope.yaml"))
	require.NoError(t, err)
	assert.Empty(t, r.Repos)
}
