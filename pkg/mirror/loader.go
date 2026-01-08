package mirror

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/wzshiming/gitmirror/pkg/backend"
)

// Loader implements the backend.Loader interface for mirrored repositories.
type Loader struct {
	mirror   *Mirror
	resolver Resolver
	errorLog *log.Logger
	// FetchTimeout is the timeout for fetching from upstream.
	FetchTimeout time.Duration
}

// Resolver resolves repository paths to upstream URLs.
type Resolver interface {
	Resolve(repoPath string) (string, error)
}

// NewLoader creates a new Loader.
func NewLoader(mirror *Mirror, resolver Resolver, errorLog *log.Logger) *Loader {
	return &Loader{
		mirror:       mirror,
		resolver:     resolver,
		errorLog:     errorLog,
		FetchTimeout: 5 * time.Minute,
	}
}

// Load loads a repository from the mirror or fetches it from upstream.
func (l *Loader) Load(repoPath string) (backend.Repository, error) {
	// Resolve the repository path to an upstream URL
	upstreamURL, err := l.resolver.Resolve(repoPath)
	if err != nil {
		return nil, fmt.Errorf("resolving repository path: %w", err)
	}

	// Create a context with timeout for fetching
	ctx, cancel := context.WithTimeout(context.Background(), l.FetchTimeout)
	defer cancel()

	// Get or fetch the repository
	repo, err := l.mirror.GetOrFetch(ctx, upstreamURL)
	if err != nil {
		return nil, fmt.Errorf("getting or fetching repository: %w", err)
	}

	return repo, nil
}
