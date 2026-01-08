package backend

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockRepository implements the Repository interface for testing.
type mockRepository struct {
	path        string
	upstreamURL string
}

func (m *mockRepository) Path() string {
	return m.path
}

func (m *mockRepository) UpstreamURL() string {
	return m.upstreamURL
}

// mockLoader implements the Loader interface for testing.
type mockLoader struct {
	repos map[string]Repository
}

func (m *mockLoader) Load(repo string) (Repository, error) {
	if r, ok := m.repos[repo]; ok {
		return r, nil
	}
	return nil, errRepositoryNotFound
}

var errRepositoryNotFound = &repositoryNotFoundError{}

type repositoryNotFoundError struct{}

func (e *repositoryNotFoundError) Error() string {
	return "repository not found"
}

func TestBackend_ServeHTTP_NotFound(t *testing.T) {
	loader := &mockLoader{repos: make(map[string]Repository)}
	backend := NewBackend(loader)

	tests := []struct {
		name       string
		path       string
		method     string
		wantStatus int
	}{
		{
			name:       "unknown path",
			path:       "/unknown/path",
			method:     http.MethodGet,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "repo not found",
			path:       "/github.com/owner/repo/info/refs",
			method:     http.MethodGet,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			backend.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("ServeHTTP() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestBackend_ServeHTTP_MethodNotAllowed(t *testing.T) {
	loader := &mockLoader{
		repos: map[string]Repository{
			"github.com/owner/repo": &mockRepository{path: "/tmp/test", upstreamURL: "https://github.com/owner/repo.git"},
		},
	}
	backend := NewBackend(loader)

	tests := []struct {
		name   string
		path   string
		method string
	}{
		{
			name:   "POST to info/refs",
			path:   "/github.com/owner/repo/info/refs",
			method: http.MethodPost,
		},
		{
			name:   "GET to git-upload-pack",
			path:   "/github.com/owner/repo/git-upload-pack",
			method: http.MethodGet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			backend.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("ServeHTTP() status = %v, want %v", w.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

func TestBackend_Prefix(t *testing.T) {
	loader := &mockLoader{
		repos: map[string]Repository{
			"github.com/owner/repo": &mockRepository{path: "/tmp/test", upstreamURL: "https://github.com/owner/repo.git"},
		},
	}
	backend := NewBackend(loader)
	backend.Prefix = "/git"

	// Request without prefix should 404
	req := httptest.NewRequest(http.MethodGet, "/github.com/owner/repo/info/refs", nil)
	w := httptest.NewRecorder()
	backend.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("ServeHTTP() without prefix status = %v, want %v", w.Code, http.StatusNotFound)
	}
}
