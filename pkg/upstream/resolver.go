// Package upstream provides functionality to parse git-upload-pack requests
// and determine which repository to cache/mirror.
package upstream

import (
	"errors"
	"net/url"
	"path"
	"strings"
)

var (
	// ErrInvalidRequest is returned when the request is invalid.
	ErrInvalidRequest = errors.New("invalid request")
	// ErrUnsupportedHost is returned when the host is not supported.
	ErrUnsupportedHost = errors.New("unsupported host")
)

// SupportedHost represents a supported upstream git host.
type SupportedHost struct {
	Name    string
	BaseURL string
}

// DefaultHosts is the list of default supported hosts.
var DefaultHosts = []SupportedHost{
	{Name: "github.com", BaseURL: "https://github.com"},
	{Name: "gitlab.com", BaseURL: "https://gitlab.com"},
	{Name: "bitbucket.org", BaseURL: "https://bitbucket.org"},
	{Name: "gitee.com", BaseURL: "https://gitee.com"},
}

// Resolver resolves repository paths to upstream URLs.
type Resolver struct {
	hosts map[string]SupportedHost
}

// NewResolver creates a new Resolver with the given supported hosts.
func NewResolver(hosts []SupportedHost) *Resolver {
	if hosts == nil {
		hosts = DefaultHosts
	}
	r := &Resolver{
		hosts: make(map[string]SupportedHost),
	}
	for _, h := range hosts {
		r.hosts[h.Name] = h
	}
	return r
}

// Resolve resolves a repository path to an upstream URL.
// The path should be in the format: host/owner/repo
func (r *Resolver) Resolve(repoPath string) (string, error) {
	repoPath = strings.TrimPrefix(repoPath, "/")
	repoPath = strings.TrimSuffix(repoPath, ".git")

	parts := strings.SplitN(repoPath, "/", 2)
	if len(parts) < 2 {
		return "", ErrInvalidRequest
	}

	host := parts[0]
	repoName := parts[1]

	supportedHost, ok := r.hosts[host]
	if !ok {
		return "", ErrUnsupportedHost
	}

	u, err := url.Parse(supportedHost.BaseURL)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, repoName+".git")

	return u.String(), nil
}

// ParseRepoPath parses a URL path to extract the repository path.
// Returns the normalized repository path.
func ParseRepoPath(urlPath string) string {
	// Remove common git endpoints from the path
	urlPath = strings.TrimPrefix(urlPath, "/")
	urlPath = strings.TrimSuffix(urlPath, "/info/refs")
	urlPath = strings.TrimSuffix(urlPath, "/git-upload-pack")
	urlPath = strings.TrimSuffix(urlPath, "/git-receive-pack")
	urlPath = strings.TrimSuffix(urlPath, "/HEAD")
	return urlPath
}
