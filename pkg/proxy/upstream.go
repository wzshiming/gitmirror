// Package proxy provides HTTP proxy functionality for git repositories.
// It intercepts git-upload-pack requests, parses them, and forwards them
// to upstream repositories while caching the responses.
package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Upstream represents an upstream git repository.
type Upstream struct {
	// URL is the base URL of the upstream repository (e.g., "https://github.com/owner/repo.git").
	URL string
	// Client is the HTTP client used to make requests. If nil, http.DefaultClient is used.
	Client *http.Client
}

// NewUpstream creates a new Upstream with the given URL.
func NewUpstream(url string) *Upstream {
	return &Upstream{
		URL:    strings.TrimSuffix(url, "/"),
		Client: http.DefaultClient,
	}
}

// InfoRefs fetches the info/refs from the upstream repository.
func (u *Upstream) InfoRefs(ctx context.Context, service string) ([]byte, error) {
	url := fmt.Sprintf("%s/info/refs?service=%s", u.URL, service)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "git/gitmirror")
	req.Header.Set("Git-Protocol", "version=2")

	resp, err := u.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching info/refs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upstream returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return body, nil
}

// UploadPack sends an upload-pack request to the upstream and returns the response.
// The requestBody should contain the pkt-line encoded request.
func (u *Upstream) UploadPack(ctx context.Context, requestBody []byte) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/git-upload-pack", u.URL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-git-upload-pack-request")
	req.Header.Set("Accept", "application/x-git-upload-pack-result")
	req.Header.Set("User-Agent", "git/gitmirror")
	req.Header.Set("Git-Protocol", "version=2")

	resp, err := u.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending upload-pack request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("upstream returned status %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// StreamUploadPack sends an upload-pack request and streams the response to the writer.
// This is useful for proxying responses directly to the client.
func (u *Upstream) StreamUploadPack(ctx context.Context, requestBody []byte, w io.Writer) error {
	body, err := u.UploadPack(ctx, requestBody)
	if err != nil {
		return err
	}
	defer body.Close()

	_, err = io.Copy(w, body)
	return err
}
