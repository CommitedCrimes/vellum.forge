package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"vellum.forge/internal/response"

	"github.com/tomasen/realip"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			pv := recover()
			if pv != nil {
				app.serverError(w, r, fmt.Errorf("%v", pv))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")

		next.ServeHTTP(w, r)
	})
}

func (app *application) logAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mw := response.NewMetricsResponseWriter(w)
		next.ServeHTTP(mw, r)

		var (
			ip     = realip.FromRequest(r)
			method = r.Method
			url    = r.URL.String()
			proto  = r.Proto
		)

		userAttrs := slog.Group("user", "ip", ip)
		requestAttrs := slog.Group("request", "method", method, "url", url, "proto", proto)
		responseAttrs := slog.Group("response", "status", mw.StatusCode, "size", mw.BytesCount)

		app.logger.Info("access", userAttrs, requestAttrs, responseAttrs)
	})
}

func (app *application) cacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This middleware just passes through - the actual caching logic is in the handlers
		// We could add cache-related headers or logging here if needed
		next.ServeHTTP(w, r)
	})
}

// cacheControlMiddleware sets appropriate cache control headers
func (app *application) cacheControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Set Vary header for User-Agent (Accept-Encoding is set by compression middleware)
		w.Header().Set("Vary", "User-Agent")

		// Set cache control based on content type and path
		if strings.HasPrefix(path, "/static/") {
			// Static assets - long cache with immutable
			w.Header().Set("Cache-Control", "public, max-age=15768000, immutable") // 0.5 years
		} else if strings.HasPrefix(path, "/themes/") {
			// Theme assets - medium cache
			w.Header().Set("Cache-Control", "public, max-age=86400, must-revalidate") // 1 day
		} else if isApiEndpoint(path) {
			// API endpoints - no cache
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		} else {
			// HTML pages - short cache with validation
			w.Header().Set("Cache-Control", "public, max-age=300, must-revalidate") // 5 minutes
		}

		next.ServeHTTP(w, r)
	})
}

// isApiEndpoint checks if the path is an API endpoint
func isApiEndpoint(path string) bool {
	apiPaths := []string{"/cache/stats", "/cache/clear", "/health"}
	for _, apiPath := range apiPaths {
		if path == apiPath {
			return true
		}
	}
	return false
}

// contentTypeMiddleware ensures correct Content-Type headers
func (app *application) contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Set Content-Type based on file extension if not already set
		if w.Header().Get("Content-Type") == "" {
			contentType := getContentType(path)
			if contentType != "" {
				w.Header().Set("Content-Type", contentType)
			}
		}

		next.ServeHTTP(w, r)
	})
}

// getContentType returns the appropriate Content-Type for a given path
func getContentType(path string) string {
	// Find the last dot in the path
	dotIndex := strings.LastIndex(path, ".")
	if dotIndex == -1 {
		// No extension found
		return ""
	}

	ext := strings.ToLower(path[dotIndex:])

	contentTypes := map[string]string{
		".css":   "text/css; charset=utf-8",
		".js":    "application/javascript; charset=utf-8",
		".json":  "application/json; charset=utf-8",
		".xml":   "application/xml; charset=utf-8",
		".html":  "text/html; charset=utf-8",
		".htm":   "text/html; charset=utf-8",
		".txt":   "text/plain; charset=utf-8",
		".md":    "text/markdown; charset=utf-8",
		".svg":   "image/svg+xml",
		".ico":   "image/x-icon",
		".png":   "image/png",
		".jpg":   "image/jpeg",
		".jpeg":  "image/jpeg",
		".gif":   "image/gif",
		".webp":  "image/webp",
		".avif":  "image/avif",
		".woff":  "font/woff",
		".woff2": "font/woff2",
		".ttf":   "font/ttf",
		".eot":   "application/vnd.ms-fontobject",
	}

	return contentTypes[ext]
}
