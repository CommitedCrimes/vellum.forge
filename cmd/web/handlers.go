package main

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"vellum.forge/internal/cache"
	"vellum.forge/internal/version"
)

// renderWithCache is a helper function that handles caching for template rendering
func (app *application) renderWithCache(w http.ResponseWriter, r *http.Request, cacheKey string, renderFunc func(http.ResponseWriter) error) error {
	// Skip caching if cache is disabled or request should bypass cache
	if app.cache == nil || cache.ShouldBypass(r) {
		if cache.ShouldBypass(r) {
			app.logger.Info("Cache bypass", "path", r.URL.Path, "reason", "bypass requested")
		}
		return renderFunc(w)
	}

	// Try to get from cache first
	if entry, found := app.cache.Get(cacheKey); found {
		// Log cache hit
		app.logger.Info("Cache hit", "path", r.URL.Path, "size", len(entry.Body), "age", time.Since(entry.CreatedAt).Round(time.Millisecond))

		// Check for conditional requests
		if cache.HandleConditionalRequest(w, r, entry) {
			app.logger.Info("Conditional request - 304 Not Modified", "path", r.URL.Path, "etag", entry.ETag)
			return nil // 304 Not Modified was sent
		}

		// Set caching headers
		cache.SetCacheHeaders(w, entry, time.Duration(app.config.cacheTTL)*time.Second)

		// Write cached response
		return cache.WriteEntryToResponse(w, entry)
	}

	// Not in cache, capture response
	app.logger.Info("Cache miss", "path", r.URL.Path)
	responseCapture := cache.NewResponseCapture(w)
	err := renderFunc(responseCapture)
	if err != nil {
		return err
	}

	// Get captured data
	body, headers, statusCode := responseCapture.GetCapturedData()

	// Only cache successful responses
	if statusCode == http.StatusOK {
		// Create cache entry
		entry := cache.CreateCacheEntry(body, headers, statusCode, time.Duration(app.config.cacheTTL)*time.Second)

		// Set caching headers
		cache.SetCacheHeaders(responseCapture, entry, time.Duration(app.config.cacheTTL)*time.Second)

		// Store in cache
		app.cache.Set(cacheKey, entry)
		app.logger.Info("Cache store", "path", r.URL.Path, "size", len(body), "ttl", time.Duration(app.config.cacheTTL)*time.Second)
	}

	// Write response to client
	return responseCapture.Flush()
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	var cacheKey string
	var err error

	// Build cache key if caching is enabled
	if app.cache != nil && !cache.ShouldBypass(r) {
		cacheKey, err = app.cacheKeyBuilder.BuildKeyForHome(r)
		if err != nil {
			app.logger.Warn("Failed to build cache key for home page", "error", err)
		}
	}

	renderFunc := func(writer http.ResponseWriter) error {
		data := app.newTemplateData(r)
		return app.jetRenderer.RenderPage(writer, http.StatusOK, data, "pages/home.jet")
	}

	if cacheKey != "" {
		err = app.renderWithCache(w, r, cacheKey, renderFunc)
	} else {
		err = renderFunc(w)
	}

	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) blogIndex(w http.ResponseWriter, r *http.Request) {
	var cacheKey string
	var err error

	fmt.Println("blogIndex", app.cache != nil, cache.ShouldBypass(r))

	// Build cache key if caching is enabled
	if app.cache != nil && !cache.ShouldBypass(r) {
		cacheKey, err = app.cacheKeyBuilder.BuildKeyForBlogIndex(r)
		if err != nil {
			app.logger.Warn("Failed to build cache key for blog index", "error", err)
		}
	}

	renderFunc := func(writer http.ResponseWriter) error {
		data := app.newTemplateData(r)

		// Load blog posts from content directory (with app.config.dataDir as base directory)
		blogPosts, metas, err := app.contentLoader.LoadBlogPosts(app.config.dataDir)
		if err != nil {
			return err
		}

		data["BlogPosts"] = blogPosts
		return app.jetRenderer.RenderPage(writer, http.StatusOK, data, "pages/blog/index.jet")
	}

	if cacheKey != "" {
		err = app.renderWithCache(w, r, cacheKey, renderFunc)
	} else {
		err = renderFunc(w)
	}

	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) blogPost(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	// First check if the blog post exists - don't cache 404s
	blogPost, _, err := app.contentLoader.LoadBlogPost(app.config.dataDir, slug)
	if err != nil {
		app.notFound(w, r)
		return
	}

	// Now that we know it exists, build cache key if caching is enabled
	var cacheKey string
	if app.cache != nil && !cache.ShouldBypass(r) {
		cacheKey, err = app.cacheKeyBuilder.BuildKeyForBlogPost(r, slug)
		if err != nil {
			app.logger.Warn("Failed to build cache key for blog post", "slug", slug, "error", err)
		}
	}

	renderFunc := func(writer http.ResponseWriter) error {
		data := app.newTemplateData(r)
		data["Post"] = blogPost

		return app.jetRenderer.RenderPage(writer, http.StatusOK, data, "pages/blog/post.jet")
	}

	if cacheKey != "" {
		err = app.renderWithCache(w, r, cacheKey, renderFunc)
	} else {
		err = renderFunc(w)
	}

	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) page(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	// First check if the page exists - don't cache 404s
	page, _, err := app.contentLoader.LoadPage(app.config.dataDir, slug)
	if err != nil {
		app.notFound(w, r)
		return
	}

	// Now that we know it exists, build cache key if caching is enabled
	var cacheKey string
	if app.cache != nil && !cache.ShouldBypass(r) {
		cacheKey, err = app.cacheKeyBuilder.BuildKeyForPage(r, slug)
		if err != nil {
			app.logger.Warn("Failed to build cache key for page", "slug", slug, "error", err)
		}
	}

	renderFunc := func(writer http.ResponseWriter) error {
		data := app.newTemplateData(r)
		data["Page"] = page

		return app.jetRenderer.RenderPage(writer, http.StatusOK, data, "pages/page.jet")
	}

	if cacheKey != "" {
		err = app.renderWithCache(w, r, cacheKey, renderFunc)
	} else {
		err = renderFunc(w)
	}

	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	write, err := w.Write([]byte(fmt.Sprintf(`{"status":"ok","version":"%s"}`, version.Get())))
	if err != nil {
		app.logger.Error("Error writing health response", "error", err)
		app.serverError(w, r, err)
		return
	}
	app.logger.Info(fmt.Sprintf("Wrote %d bytes to health endpoint", write))
}

func (app *application) cacheStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if app.cache == nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"cache":"disabled"}`))
		return
	}

	stats := app.cache.Stats()
	response := fmt.Sprintf(`{
		"enabled": true,
		"entries": %d,
		"sizeBytes": %d,
		"sizeMB": %.2f,
		"maxEntries": %d,
		"maxSizeBytes": %d,
		"maxSizeMB": %.2f,
		"utilization": %.2f,
		"hits": %d,
		"misses": %d,
		"hitRate": %.2f
	}`,
		stats.Entries,
		stats.SizeBytes,
		float64(stats.SizeBytes)/(1024*1024),
		stats.MaxEntries,
		stats.MaxSizeBytes,
		float64(stats.MaxSizeBytes)/(1024*1024),
		float64(stats.Entries)/float64(stats.MaxEntries)*100,
		stats.Hits,
		stats.Misses,
		stats.HitRate,
	)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}

func (app *application) cacheClear(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if app.cache == nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"cache":"disabled"}`))
		return
	}

	// Only allow POST requests for cache clearing
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error":"method not allowed"}`))
		return
	}

	app.cache.Clear()
	app.logger.Info("Cache manually cleared via API")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true,"message":"cache cleared"}`))
}

func (app *application) themeAssets(w http.ResponseWriter, r *http.Request) {
	// Extract the requested path from the URL
	requestedPath := chi.URLParam(r, "*")

	// Validate and clean the path to prevent directory traversal
	cleanPath, err := app.validateAssetPath(requestedPath)
	if err != nil {
		app.notFound(w, r)
		return
	}

	// Build the full path to the theme asset
	themeAssetsDir := filepath.Join(app.config.themeDir, app.config.theme, "assets")
	fullPath := filepath.Join(themeAssetsDir, cleanPath)

	// Ensure the final path is still within the theme assets directory
	if !app.isPathSafe(themeAssetsDir, fullPath) {
		app.notFound(w, r)
		return
	}

	// Check if file exists and is not a directory
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			app.notFound(w, r)
			return
		}
		app.serverError(w, r, err)
		return
	}

	// Disable directory listing - only serve files
	if fileInfo.IsDir() {
		app.notFound(w, r)
		return
	}

	// Set appropriate content type based on file extension
	contentType := app.getContentType(cleanPath)
	w.Header().Set("Content-Type", contentType)

	// Set security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")

	// Serve the file
	http.ServeFile(w, r, fullPath)
}

// validateAssetPath validates and cleans a requested asset path to prevent directory traversal
func (app *application) validateAssetPath(requestedPath string) (string, error) {
	// Remove any leading slashes
	cleanPath := strings.TrimPrefix(requestedPath, "/")

	// Reject empty paths
	if cleanPath == "" {
		return "", fmt.Errorf("empty path")
	}

	// Clean the path to resolve any .. elements
	cleanPath = filepath.Clean(cleanPath)

	// Check for directory traversal attempts
	if strings.Contains(cleanPath, "..") || strings.HasPrefix(cleanPath, "/") {
		return "", fmt.Errorf("invalid path: %s", requestedPath)
	}

	return cleanPath, nil
}

// isPathSafe checks if the resolved path is within the expected base directory
func (app *application) isPathSafe(baseDir, requestedPath string) bool {
	// Get absolute paths for comparison
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return false
	}

	absRequested, err := filepath.Abs(requestedPath)
	if err != nil {
		return false
	}

	// Check if the requested path is within the base directory
	relPath, err := filepath.Rel(absBase, absRequested)
	if err != nil {
		return false
	}

	// If the relative path starts with "..", it's outside the base directory
	return !strings.HasPrefix(relPath, "..")
}

// getContentType returns the appropriate content type for a file based on its extension
func (app *application) getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	// Use Go's built-in mime type detection first
	contentType := mime.TypeByExtension(ext)
	if contentType != "" {
		return contentType
	}

	// Fallback for common web asset types
	switch ext {
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".ico":
		return "image/x-icon"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".eot":
		return "application/vnd.ms-fontobject"
	default:
		return "application/octet-stream"
	}
}
