package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewUpstream(t *testing.T) {
	upstream := NewUpstream("https://github.com/owner/repo.git")
	if upstream.URL != "https://github.com/owner/repo.git" {
		t.Errorf("URL = %q, want %q", upstream.URL, "https://github.com/owner/repo.git")
	}
	if upstream.Client == nil {
		t.Error("Client should not be nil")
	}
}

func TestNewUpstream_TrailingSlash(t *testing.T) {
	upstream := NewUpstream("https://github.com/owner/repo.git/")
	if upstream.URL != "https://github.com/owner/repo.git" {
		t.Errorf("URL = %q, want %q", upstream.URL, "https://github.com/owner/repo.git")
	}
}

func TestUpstream_InfoRefs(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/info/refs" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("service") != "git-upload-pack" {
			t.Errorf("unexpected service: %s", r.URL.Query().Get("service"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mock refs"))
	}))
	defer server.Close()

	upstream := NewUpstream(server.URL)
	body, err := upstream.InfoRefs(context.Background(), "git-upload-pack")
	if err != nil {
		t.Fatalf("InfoRefs() error = %v", err)
	}
	if string(body) != "mock refs" {
		t.Errorf("InfoRefs() = %q, want %q", body, "mock refs")
	}
}

func TestUpstream_InfoRefs_Error(t *testing.T) {
	// Create a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	upstream := NewUpstream(server.URL)
	_, err := upstream.InfoRefs(context.Background(), "git-upload-pack")
	if err == nil {
		t.Error("InfoRefs() expected error, got nil")
	}
}

func TestUpstream_UploadPack(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/git-upload-pack" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/x-git-upload-pack-request" {
			t.Errorf("unexpected content-type: %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pack data"))
	}))
	defer server.Close()

	upstream := NewUpstream(server.URL)
	body, err := upstream.UploadPack(context.Background(), []byte("request"))
	if err != nil {
		t.Fatalf("UploadPack() error = %v", err)
	}
	defer body.Close()

	// Read the response
	buf := make([]byte, 100)
	n, _ := body.Read(buf)
	if string(buf[:n]) != "pack data" {
		t.Errorf("UploadPack() response = %q, want %q", buf[:n], "pack data")
	}
}
