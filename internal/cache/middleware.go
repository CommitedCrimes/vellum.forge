package cache

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ResponseCapture captures response data for caching
type ResponseCapture struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
	headers    http.Header
	size       int
}

// NewResponseCapture creates a new response capture wrapper
func NewResponseCapture(w http.ResponseWriter) *ResponseCapture {
	return &ResponseCapture{
		ResponseWriter: w,
		body:           new(bytes.Buffer),
		headers:        make(http.Header),
		statusCode:     http.StatusOK,
	}
}

// WriteHeader captures the status code
func (rc *ResponseCapture) WriteHeader(statusCode int) {
	rc.statusCode = statusCode
	// Don't write to underlying response yet - we may need to modify it
}

// Write captures the response body
func (rc *ResponseCapture) Write(data []byte) (int, error) {
	rc.size += len(data)
	return rc.body.Write(data)
}

// Header returns the header map that will be sent
func (rc *ResponseCapture) Header() http.Header {
	// Return our captured headers, but fall back to underlying writer's headers
	if len(rc.headers) > 0 {
		return rc.headers
	}
	return rc.ResponseWriter.Header()
}

// Flush writes the captured response to the underlying ResponseWriter
func (rc *ResponseCapture) Flush() error {
	// Copy our captured headers to the underlying response writer
	for k, values := range rc.headers {
		for _, v := range values {
			rc.ResponseWriter.Header().Add(k, v)
		}
	}

	// Also copy any headers that were set directly on the underlying writer
	underlyingHeaders := rc.ResponseWriter.Header()
	for k, values := range underlyingHeaders {
		// Only copy if we don't already have this header
		if len(rc.headers[k]) == 0 {
			for _, v := range values {
				rc.headers[k] = append(rc.headers[k], v)
			}
		}
	}

	// Write status code
	rc.ResponseWriter.WriteHeader(rc.statusCode)

	// Write body
	_, err := rc.ResponseWriter.Write(rc.body.Bytes())
	return err
}

// GetCapturedData returns the captured response data
func (rc *ResponseCapture) GetCapturedData() ([]byte, http.Header, int) {
	// Ensure we have the latest headers from the underlying writer
	underlyingHeaders := rc.ResponseWriter.Header()
	finalHeaders := make(http.Header)

	// Copy underlying headers first
	for k, values := range underlyingHeaders {
		finalHeaders[k] = make([]string, len(values))
		copy(finalHeaders[k], values)
	}

	// Override with our captured headers
	for k, values := range rc.headers {
		finalHeaders[k] = make([]string, len(values))
		copy(finalHeaders[k], values)
	}

	return rc.body.Bytes(), finalHeaders, rc.statusCode
}

// CacheKeyBuilder helps build cache keys for different types of requests
type CacheKeyBuilder struct {
	themeID  string
	dataDir  string
	themeDir string
}

// NewCacheKeyBuilder creates a new cache key builder
func NewCacheKeyBuilder(themeID, dataDir, themeDir string) *CacheKeyBuilder {
	return &CacheKeyBuilder{
		themeID:  themeID,
		dataDir:  dataDir,
		themeDir: themeDir,
	}
}

// BuildKey builds a cache key for the given request and context
func (ckb *CacheKeyBuilder) BuildKey(r *http.Request, template string, filePaths []string) (string, error) {
	// Normalize path
	normalizedPath := strings.ToLower(strings.TrimRight(r.URL.Path, "/"))
	if normalizedPath == "" {
		normalizedPath = "/"
	}

	// Get file modification times
	var maxMTime int64
	var absFilePaths []string

	for _, filePath := range filePaths {
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			continue // Skip files we can't resolve
		}
		absFilePaths = append(absFilePaths, absPath)

		if mtime, err := GetFileMTime(absPath); err == nil && mtime > maxMTime {
			maxMTime = mtime
		}
	}

	// Create combined file path string
	combinedFilePath := strings.Join(absFilePaths, ";")

	// Detect feature flags (for now, just check for mermaid in query params or headers)
	featureFlags := make(map[string]bool)
	if r.URL.Query().Get("mermaid") == "1" || r.Header.Get("X-Enable-Mermaid") == "1" {
		featureFlags["mermaid"] = true
	}

	// Get accept-encoding
	acceptEncoding := r.Header.Get("Accept-Encoding")

	// Build cache key parameters
	params := KeyParams{
		Method:         r.Method,
		NormalizedPath: normalizedPath,
		AbsFilePath:    combinedFilePath,
		FileMTimeNs:    maxMTime,
		Template:       template,
		ThemeID:        ckb.themeID,
		FeatureFlags:   featureFlags,
		AcceptEncoding: acceptEncoding,
		ContentType:    "text/html",
	}

	return GenerateKey(params), nil
}

// BuildKeyForBlogIndex builds a cache key for the blog index page
func (ckb *CacheKeyBuilder) BuildKeyForBlogIndex(r *http.Request) (string, error) {
	blogDir := filepath.Join(ckb.dataDir, "blog")

	var filePaths []string
	err := filepath.Walk(blogDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
			filePaths = append(filePaths, path)
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to walk blog directory: %w", err)
	}

	return ckb.BuildKey(r, "pages/blog/index.jet", filePaths)
}

// BuildKeyForBlogPost builds a cache key for a specific blog post
func (ckb *CacheKeyBuilder) BuildKeyForBlogPost(r *http.Request, slug string) (string, error) {
	blogDir := filepath.Join(ckb.dataDir, "blog")

	// Find the specific blog post file
	var filePath string
	err := filepath.Walk(blogDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
			base := filepath.Base(path)
			ext := filepath.Ext(base)
			fileSlug := strings.TrimSuffix(base, ext)
			if fileSlug == slug {
				filePath = path
				return filepath.SkipDir
			}
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to find blog post: %w", err)
	}

	if filePath == "" {
		return "", fmt.Errorf("blog post not found: %s", slug)
	}

	return ckb.BuildKey(r, "pages/blog/post.jet", []string{filePath})
}

// BuildKeyForPage builds a cache key for a regular page
func (ckb *CacheKeyBuilder) BuildKeyForPage(r *http.Request, slug string) (string, error) {
	pagesDir := filepath.Join(ckb.dataDir, "pages")

	// Find the specific page file
	var filePath string
	err := filepath.Walk(pagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
			base := filepath.Base(path)
			ext := filepath.Ext(base)
			fileSlug := strings.TrimSuffix(base, ext)
			if fileSlug == slug {
				filePath = path
				return filepath.SkipDir
			}
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to find page: %w", err)
	}

	if filePath == "" {
		return "", fmt.Errorf("page not found: %s", slug)
	}

	return ckb.BuildKey(r, "pages/page.jet", []string{filePath})
}

// BuildKeyForHome builds a cache key for the home page
func (ckb *CacheKeyBuilder) BuildKeyForHome(r *http.Request) (string, error) {
	// Home page might depend on recent blog posts, so include blog directory
	blogDir := filepath.Join(ckb.dataDir, "blog")

	var filePaths []string
	err := filepath.Walk(blogDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
			filePaths = append(filePaths, path)
		}
		return nil
	})

	if err != nil {
		// If we can't read blog posts, just cache based on template
		return ckb.BuildKey(r, "pages/home.jet", []string{})
	}

	return ckb.BuildKey(r, "pages/home.jet", filePaths)
}

// GenerateETag generates an ETag for response content
func GenerateETag(content []byte) string {
	hash := md5.Sum(content)
	return fmt.Sprintf(`"%x"`, hash)
}

// SetCacheHeaders sets appropriate caching headers on the response
func SetCacheHeaders(w http.ResponseWriter, entry *Entry, maxAge time.Duration) {
	// Set ETag if not already set
	if entry.ETag == "" {
		entry.ETag = GenerateETag(entry.Body)
	}
	w.Header().Set("ETag", entry.ETag)

	// Set Last-Modified if not already set
	if entry.LastMod.IsZero() {
		entry.LastMod = entry.CreatedAt
	}
	w.Header().Set("Last-Modified", entry.LastMod.UTC().Format(http.TimeFormat))

	// Set Cache-Control
	if maxAge > 0 {
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(maxAge.Seconds())))
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}

	// Set Vary header to indicate response varies by encoding
	w.Header().Set("Vary", "Accept-Encoding")
}

// HandleConditionalRequest handles conditional requests (If-None-Match, If-Modified-Since)
// Returns true if a 304 Not Modified response was sent
func HandleConditionalRequest(w http.ResponseWriter, r *http.Request, entry *Entry) bool {
	if !entry.SupportsConditionalRequests() {
		return false
	}

	if entry.MatchesConditions(r) {
		// Copy essential headers for 304 response
		if entry.ETag != "" {
			w.Header().Set("ETag", entry.ETag)
		}
		if !entry.LastMod.IsZero() {
			w.Header().Set("Last-Modified", entry.LastMod.UTC().Format(http.TimeFormat))
		}

		w.WriteHeader(http.StatusNotModified)
		return true
	}

	return false
}
