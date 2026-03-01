package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"syscall"
	"time"

	"github.com/elliottmoos/clover/internal/paths"
	"gopkg.in/yaml.v3"
)

// Registry manages the set of registered repositories.
type Registry struct {
	Repos []Repo `yaml:"repos"`
	path  string
}

// Load reads the registry from the default file path.
// Returns an empty registry if the file does not exist.
func Load() (*Registry, error) {
	p, err := paths.RegistryFile()
	if err != nil {
		return nil, err
	}
	return LoadFrom(p)
}

// LoadFrom reads the registry from a specific file path.
func LoadFrom(path string) (*Registry, error) {
	r := &Registry{path: path}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return r, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading registry: %w", err)
	}
	if err := yaml.Unmarshal(data, r); err != nil {
		return nil, fmt.Errorf("parsing registry: %w", err)
	}
	return r, nil
}

// Save writes the registry to disk with advisory file locking.
func (r *Registry) Save() error {
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	f, err := os.OpenFile(r.path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return fmt.Errorf("opening registry: %w", err)
	}
	defer f.Close()

	// Advisory flock to prevent concurrent writes.
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("locking registry: %w", err)
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	data, err := yaml.Marshal(r)
	if err != nil {
		return fmt.Errorf("marshaling registry: %w", err)
	}

	if err := f.Truncate(0); err != nil {
		return fmt.Errorf("truncating registry: %w", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return fmt.Errorf("seeking registry: %w", err)
	}
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("writing registry: %w", err)
	}
	return nil
}

// Add registers a new repo. Returns an error on duplicate name.
func (r *Registry) Add(name, repoPath string) error {
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	for _, repo := range r.Repos {
		if repo.Name == name {
			return fmt.Errorf("repo %q already registered", name)
		}
		if repo.Path == absPath {
			return fmt.Errorf("path %q already registered as %q", absPath, repo.Name)
		}
	}

	r.Repos = append(r.Repos, Repo{
		Name:    name,
		Path:    absPath,
		AddedAt: time.Now(),
	})
	return nil
}

// Remove unregisters a repo by name.
func (r *Registry) Remove(name string) error {
	for i, repo := range r.Repos {
		if repo.Name == name {
			r.Repos = append(r.Repos[:i], r.Repos[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("repo %q not found", name)
}

// Find returns a pointer to the repo with the given name, or nil.
func (r *Registry) Find(name string) *Repo {
	for i := range r.Repos {
		if r.Repos[i].Name == name {
			return &r.Repos[i]
		}
	}
	return nil
}

// List returns all repos, with focused repos sorted first.
func (r *Registry) List() []Repo {
	sorted := make([]Repo, len(r.Repos))
	copy(sorted, r.Repos)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Focused != sorted[j].Focused {
			return sorted[i].Focused
		}
		return sorted[i].Name < sorted[j].Name
	})
	return sorted
}

// Focused returns only repos with Focused=true.
func (r *Registry) Focused() []Repo {
	var result []Repo
	for _, repo := range r.Repos {
		if repo.Focused {
			result = append(result, repo)
		}
	}
	return result
}

// ToggleFocus flips the focus state of the named repo.
func (r *Registry) ToggleFocus(name string) error {
	repo := r.Find(name)
	if repo == nil {
		return fmt.Errorf("repo %q not found", name)
	}
	repo.Focused = !repo.Focused
	return nil
}

// TouchLastUsed updates last_used_at for the named repo.
func (r *Registry) TouchLastUsed(name string) error {
	repo := r.Find(name)
	if repo == nil {
		return fmt.Errorf("repo %q not found", name)
	}
	repo.LastUsedAt = time.Now()
	return nil
}
