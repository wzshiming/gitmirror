// Package backend provides a Git HTTP backend handler for serving git repositories
// over HTTP using the Smart-HTTP protocol as a read-only mirror.
package backend

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/wzshiming/gitmirror/pkg/proxy"
)

type contextKey string

// Service represents a Git transport service.
type Service string

// String returns the string representation of the service.
func (s Service) String() string {
	return string(s)
}

// Name returns the name of the service without the "git-" prefix.
func (s Service) Name() string {
	return strings.TrimPrefix(string(s), "git-")
}

// Git service names.
const (
	UploadPackService Service = "git-upload-pack"
)

type route struct {
	pattern *regexp.Regexp
	method  string
	handler http.HandlerFunc
	svc     Service
}

var routes = []route{
	{regexp.MustCompile("(.*?)/HEAD$"), http.MethodGet, getTextFile, ""},
	{regexp.MustCompile("(.*?)/info/refs$"), http.MethodGet, getInfoRefs, ""},
	{regexp.MustCompile("(.*?)/objects/info/alternates$"), http.MethodGet, getTextFile, ""},
	{regexp.MustCompile("(.*?)/objects/info/http-alternates$"), http.MethodGet, getTextFile, ""},
	{regexp.MustCompile("(.*?)/objects/info/packs$"), http.MethodGet, getInfoPacks, ""},
	{regexp.MustCompile("(.*?)/objects/[0-9a-f]{2}/[0-9a-f]{38}$"), http.MethodGet, getLooseObject, ""},
	{regexp.MustCompile("(.*?)/objects/[0-9a-f]{2}/[0-9a-f]{62}$"), http.MethodGet, getLooseObject, ""},
	{regexp.MustCompile(`(.*?)/objects/pack/pack-[0-9a-f]{40}\.pack$`), http.MethodGet, getPackFile, ""},
	{regexp.MustCompile(`(.*?)/objects/pack/pack-[0-9a-f]{64}\.pack$`), http.MethodGet, getPackFile, ""},
	{regexp.MustCompile(`(.*?)/objects/pack/pack-[0-9a-f]{40}\.idx$`), http.MethodGet, getIdxFile, ""},
	{regexp.MustCompile(`(.*?)/objects/pack/pack-[0-9a-f]{64}\.idx$`), http.MethodGet, getIdxFile, ""},

	// git-upload-pack for read-only clone/fetch (smart HTTP)
	{regexp.MustCompile("(.*?)/git-upload-pack$"), http.MethodPost, serviceRPC, UploadPackService},
}

// Loader loads repository storage based on a repository path.
type Loader interface {
	// Load loads a repository given a path.
	// Returns an error if the repository does not exist.
	Load(repo string) (Repository, error)
}

// Repository represents a git repository for serving.
type Repository interface {
	// Path returns the path to the repository's local git directory.
	// May be empty if the repository is fully proxied.
	Path() string
	// UpstreamURL returns the URL of the upstream repository.
	UpstreamURL() string
}

// Backend represents a Git HTTP handler.
type Backend struct {
	// Loader is used to load repositories from the given path.
	Loader Loader
	// ErrorLog is the logger used to log errors. If nil, no errors are logged.
	ErrorLog *log.Logger
	// Prefix is a path prefix that will be stripped from the URL path before
	// matching the route patterns.
	Prefix string
}

// NewBackend returns a Git HTTP handler that serves git repositories over
// HTTP as a read-only mirror.
func NewBackend(loader Loader) *Backend {
	return &Backend{
		Loader: loader,
	}
}

// ServeHTTP implements the [http.Handler] interface.
func (b *Backend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path
	urlPath = strings.TrimPrefix(urlPath, b.Prefix)
	for _, rt := range routes {
		if m := rt.pattern.FindStringSubmatch(urlPath); m != nil {
			if r.Method != rt.method {
				renderStatusError(w, http.StatusMethodNotAllowed)
				return
			}

			repo := strings.TrimPrefix(m[1], "/")
			file := strings.Replace(urlPath, repo+"/", "", 1)

			st, err := b.Loader.Load(repo)
			if err != nil {
				logf(b.ErrorLog, "error loading repository: %v", err)
				renderStatusError(w, http.StatusNotFound)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, contextKey("errorLog"), b.ErrorLog)
			ctx = context.WithValue(ctx, contextKey("repo"), repo)
			ctx = context.WithValue(ctx, contextKey("file"), file)
			ctx = context.WithValue(ctx, contextKey("service"), rt.svc)
			ctx = context.WithValue(ctx, contextKey("storer"), st)

			rt.handler(w, r.WithContext(ctx))
			return
		}
	}

	// If no route matched, return 404.
	renderStatusError(w, http.StatusNotFound)
}

// logf logs the given message to the error log if it is set.
func logf(logger *log.Logger, format string, v ...any) {
	if logger != nil {
		logger.Printf(format, v...)
	}
}

func serviceRPC(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	st, ok := ctx.Value(contextKey("storer")).(Repository)
	if !ok {
		renderStatusError(w, http.StatusInternalServerError)
		return
	}
	svc, ok := ctx.Value(contextKey("service")).(Service)
	if !ok {
		renderStatusError(w, http.StatusInternalServerError)
		return
	}
	errorLog, _ := ctx.Value(contextKey("errorLog")).(*log.Logger)
	contentType := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))

	expectedContentType := strings.ToLower(fmt.Sprintf("application/x-git-%s-request", svc.Name()))
	if contentType != expectedContentType {
		renderStatusError(w, http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-result", svc.Name()))
	w.Header().Set("Connection", "Keep-Alive")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	var reader io.ReadCloser
	var err error
	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(r.Body)
		if err != nil {
			logf(errorLog, "error creating gzip reader: %v", err)
			renderStatusError(w, http.StatusInternalServerError)
			return
		}
		defer func() { _ = reader.Close() }()
	default:
		reader = r.Body
	}

	// Read the request body to forward to upstream
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		logf(errorLog, "error reading request body: %v", err)
		renderStatusError(w, http.StatusInternalServerError)
		return
	}

	// Proxy the request to upstream
	upstream := proxy.NewUpstream(st.UpstreamURL())
	err = upstream.StreamUploadPack(ctx, requestBody, w)
	if err != nil {
		logf(errorLog, "error proxying upload-pack request: %v", err)
		// Don't write error if we've already started writing
		return
	}
}

func getTextFile(w http.ResponseWriter, r *http.Request) {
	hdrNocache(w)
	sendFile(w, r, "text/plain; charset=utf-8")
}

func getInfoRefs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	st, ok := ctx.Value(contextKey("storer")).(Repository)
	if !ok {
		renderStatusError(w, http.StatusInternalServerError)
		return
	}
	errorLog, _ := ctx.Value(contextKey("errorLog")).(*log.Logger)

	service := Service(r.URL.Query().Get("service"))

	if service == UploadPackService {
		hdrNocache(w)
		w.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-advertisement", service.Name()))

		// Proxy info/refs to upstream
		upstream := proxy.NewUpstream(st.UpstreamURL())
		body, err := upstream.InfoRefs(ctx, string(service))
		if err != nil {
			logf(errorLog, "error fetching info/refs from upstream: %v", err)
			renderStatusError(w, http.StatusBadGateway)
			return
		}

		if _, err := w.Write(body); err != nil {
			logf(errorLog, "error writing info/refs response: %v", err)
			return
		}
	} else {
		hdrNocache(w)
		sendFile(w, r, "text/plain; charset=utf-8")
	}
}

func getInfoPacks(w http.ResponseWriter, r *http.Request) {
	hdrCacheForever(w)
	sendFile(w, r, "text/plain; charset=utf-8")
}

func getLooseObject(w http.ResponseWriter, r *http.Request) {
	hdrCacheForever(w)
	sendFile(w, r, "application/x-git-loose-object")
}

func getPackFile(w http.ResponseWriter, r *http.Request) {
	hdrCacheForever(w)
	sendFile(w, r, "application/x-git-packed-objects")
}

func getIdxFile(w http.ResponseWriter, r *http.Request) {
	hdrCacheForever(w)
	sendFile(w, r, "application/x-git-packed-objects-toc")
}

func sendFile(w http.ResponseWriter, r *http.Request, contentType string) {
	ctx := r.Context()
	st, ok := ctx.Value(contextKey("storer")).(Repository)
	if !ok {
		renderStatusError(w, http.StatusInternalServerError)
		return
	}
	file, ok := ctx.Value(contextKey("file")).(string)
	if !ok {
		renderStatusError(w, http.StatusInternalServerError)
		return
	}

	// Use filepath.Join to safely construct the path and prevent path traversal
	fullPath := filepath.Join(st.Path(), file)
	// Verify the path is within the repository directory
	if !strings.HasPrefix(fullPath, filepath.Clean(st.Path())+string(filepath.Separator)) {
		renderStatusError(w, http.StatusBadRequest)
		return
	}
	http.ServeFile(w, r, fullPath)
}

func renderStatusError(w http.ResponseWriter, code int) {
	http.Error(w, fmt.Sprintf("%d %s", code, http.StatusText(code)), code)
}

func hdrNocache(w http.ResponseWriter) {
	w.Header().Set("Expires", "Fri, 01 Jan 1980 00:00:00 GMT")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Cache-Control", "no-cache, max-age=0, must-revalidate")
}

func hdrCacheForever(w http.ResponseWriter) {
	now := time.Now()
	expires := now.Add(365 * 24 * time.Hour)
	w.Header().Set("Date", now.Format(http.TimeFormat))
	w.Header().Set("Expires", expires.Format(http.TimeFormat))
	w.Header().Set("Cache-Control", "public, max-age=31536000")
}
