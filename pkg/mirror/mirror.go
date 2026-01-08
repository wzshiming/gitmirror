// Package mirror provides functionality to mirror remote git repositories.
package mirror

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// Mirror manages local mirror repositories for remote git repositories.
type Mirror struct {
	// BaseDir is the directory where mirrored repositories are stored.
	BaseDir string
	// mu protects concurrent access to repositories.
	mu sync.RWMutex
}

// NewMirror creates a new Mirror with the given base directory.
func NewMirror(baseDir string) (*Mirror, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("creating base directory: %w", err)
	}
	return &Mirror{
		BaseDir: baseDir,
	}, nil
}

// Repository represents a mirrored git repository.
type Repository struct {
	path string
}

// Path returns the path to the repository's git directory.
func (r *Repository) Path() string {
	return r.path
}

// GetOrFetch returns a local mirror of the remote repository.
// If the repository already exists locally, it returns it directly.
// If not, it clones the repository as a bare mirror.
func (m *Mirror) GetOrFetch(ctx context.Context, remoteURL string) (*Repository, error) {
	localPath := m.repoPath(remoteURL)

	m.mu.RLock()
	exists := m.exists(localPath)
	m.mu.RUnlock()

	if exists {
		return &Repository{path: localPath}, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if m.exists(localPath) {
		return &Repository{path: localPath}, nil
	}

	// Clone as bare mirror
	if err := m.clone(ctx, remoteURL, localPath); err != nil {
		return nil, fmt.Errorf("cloning repository: %w", err)
	}

	return &Repository{path: localPath}, nil
}

// Fetch updates an existing mirrored repository from its remote.
func (m *Mirror) Fetch(ctx context.Context, remoteURL string) error {
	localPath := m.repoPath(remoteURL)

	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.exists(localPath) {
		return fmt.Errorf("repository not found: %s", remoteURL)
	}

	cmd := exec.CommandContext(ctx, "git", "-C", localPath, "fetch", "--all", "--prune")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("fetching repository: %w: %s", err, string(output))
	}

	return nil
}

// repoPath converts a remote URL to a local path.
func (m *Mirror) repoPath(remoteURL string) string {
	// Normalize the URL to create a consistent path
	path := remoteURL
	path = strings.TrimPrefix(path, "https://")
	path = strings.TrimPrefix(path, "http://")
	path = strings.TrimPrefix(path, "git://")
	path = strings.TrimSuffix(path, ".git")
	path = path + ".git"
	return filepath.Join(m.BaseDir, path)
}

// exists checks if a repository exists at the given path.
func (m *Mirror) exists(path string) bool {
	info, err := os.Stat(filepath.Join(path, "HEAD"))
	return err == nil && !info.IsDir()
}

// clone clones a remote repository as a bare mirror.
func (m *Mirror) clone(ctx context.Context, remoteURL, localPath string) error {
	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return fmt.Errorf("creating parent directory: %w", err)
	}

	cmd := exec.CommandContext(ctx, "git", "clone", "--bare", "--mirror", remoteURL, localPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("cloning repository: %w: %s", err, string(output))
	}

	return nil
}
