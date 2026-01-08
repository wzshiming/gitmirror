// gitmirror is a read-only git mirror server that proxies git repositories
// from upstream sources and caches them locally.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/wzshiming/gitmirror/pkg/backend"
	"github.com/wzshiming/gitmirror/pkg/mirror"
	"github.com/wzshiming/gitmirror/pkg/upstream"
)

func main() {
	var (
		addr     = flag.String("addr", ":8080", "HTTP server address")
		cacheDir = flag.String("cache-dir", "./cache", "Directory to store mirrored repositories")
		prefix   = flag.String("prefix", "", "URL path prefix to strip")
	)
	flag.Parse()

	errorLog := log.New(os.Stderr, "[gitmirror] ", log.LstdFlags)

	// Create the mirror
	m, err := mirror.NewMirror(*cacheDir)
	if err != nil {
		errorLog.Fatalf("Failed to create mirror: %v", err)
	}

	// Create the resolver for upstream URLs
	resolver := upstream.NewResolver(nil)

	// Create the loader
	loader := mirror.NewLoader(m, resolver, errorLog)

	// Create the backend handler
	handler := backend.NewBackend(loader)
	handler.ErrorLog = errorLog
	handler.Prefix = *prefix

	log.Printf("Starting gitmirror server on %s", *addr)
	log.Printf("Cache directory: %s", *cacheDir)
	if *prefix != "" {
		log.Printf("URL prefix: %s", *prefix)
	}

	if err := http.ListenAndServe(*addr, handler); err != nil {
		errorLog.Fatalf("Failed to start server: %v", err)
	}
}
