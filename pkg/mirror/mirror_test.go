package mirror

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewMirror(t *testing.T) {
	tmpDir := t.TempDir()

	m, err := NewMirror(tmpDir)
	if err != nil {
		t.Fatalf("NewMirror() error = %v", err)
	}

	if m.BaseDir != tmpDir {
		t.Errorf("NewMirror() BaseDir = %v, want %v", m.BaseDir, tmpDir)
	}
}

func TestMirror_repoPath(t *testing.T) {
	m := &Mirror{BaseDir: "/base"}

	tests := []struct {
		name      string
		remoteURL string
		want      string
	}{
		{
			name:      "https URL",
			remoteURL: "https://github.com/owner/repo.git",
			want:      "/base/github.com/owner/repo.git",
		},
		{
			name:      "https URL without .git",
			remoteURL: "https://github.com/owner/repo",
			want:      "/base/github.com/owner/repo.git",
		},
		{
			name:      "http URL",
			remoteURL: "http://github.com/owner/repo.git",
			want:      "/base/github.com/owner/repo.git",
		},
		{
			name:      "git URL",
			remoteURL: "git://github.com/owner/repo.git",
			want:      "/base/github.com/owner/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := m.repoPath(tt.remoteURL); got != tt.want {
				t.Errorf("repoPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMirror_exists(t *testing.T) {
	tmpDir := t.TempDir()
	m := &Mirror{BaseDir: tmpDir}

	// Create a fake git repo
	repoPath := filepath.Join(tmpDir, "test.git")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	// Should not exist (no HEAD file)
	if m.exists(repoPath) {
		t.Errorf("exists() = true, want false (no HEAD file)")
	}

	// Create HEAD file
	headPath := filepath.Join(repoPath, "HEAD")
	if err := os.WriteFile(headPath, []byte("ref: refs/heads/main\n"), 0644); err != nil {
		t.Fatalf("failed to create HEAD file: %v", err)
	}

	// Should exist now
	if !m.exists(repoPath) {
		t.Errorf("exists() = false, want true")
	}
}

func TestRepository_Path(t *testing.T) {
	r := &Repository{path: "/test/path", upstreamURL: "https://github.com/owner/repo.git"}
	if got := r.Path(); got != "/test/path" {
		t.Errorf("Path() = %v, want /test/path", got)
	}
}

func TestRepository_UpstreamURL(t *testing.T) {
	r := &Repository{path: "/test/path", upstreamURL: "https://github.com/owner/repo.git"}
	if got := r.UpstreamURL(); got != "https://github.com/owner/repo.git" {
		t.Errorf("UpstreamURL() = %v, want https://github.com/owner/repo.git", got)
	}
}
