package upstream

import (
	"testing"
)

func TestResolver_Resolve(t *testing.T) {
	tests := []struct {
		name     string
		repoPath string
		want     string
		wantErr  error
	}{
		{
			name:     "github repository",
			repoPath: "github.com/owner/repo",
			want:     "https://github.com/owner/repo.git",
		},
		{
			name:     "github repository with .git suffix",
			repoPath: "github.com/owner/repo.git",
			want:     "https://github.com/owner/repo.git",
		},
		{
			name:     "github repository with leading slash",
			repoPath: "/github.com/owner/repo",
			want:     "https://github.com/owner/repo.git",
		},
		{
			name:     "gitlab repository",
			repoPath: "gitlab.com/owner/repo",
			want:     "https://gitlab.com/owner/repo.git",
		},
		{
			name:     "bitbucket repository",
			repoPath: "bitbucket.org/owner/repo",
			want:     "https://bitbucket.org/owner/repo.git",
		},
		{
			name:     "unsupported host",
			repoPath: "unsupported.com/owner/repo",
			wantErr:  ErrUnsupportedHost,
		},
		{
			name:     "invalid path - no repo",
			repoPath: "github.com",
			wantErr:  ErrInvalidRequest,
		},
	}

	resolver := NewResolver(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolver.Resolve(tt.repoPath)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Resolve() unexpected error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Resolve() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRepoPath(t *testing.T) {
	tests := []struct {
		name    string
		urlPath string
		want    string
	}{
		{
			name:    "info/refs endpoint",
			urlPath: "/github.com/owner/repo/info/refs",
			want:    "github.com/owner/repo",
		},
		{
			name:    "git-upload-pack endpoint",
			urlPath: "/github.com/owner/repo/git-upload-pack",
			want:    "github.com/owner/repo",
		},
		{
			name:    "HEAD endpoint",
			urlPath: "/github.com/owner/repo/HEAD",
			want:    "github.com/owner/repo",
		},
		{
			name:    "simple path",
			urlPath: "/github.com/owner/repo",
			want:    "github.com/owner/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseRepoPath(tt.urlPath); got != tt.want {
				t.Errorf("ParseRepoPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
