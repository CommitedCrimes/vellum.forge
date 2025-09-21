package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"vellum.forge/internal/cache"
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
		// Skip caching if cache is disabled or request should bypass cache
		if app.cache == nil || cache.ShouldBypass(r) {
			next.ServeHTTP(w, r)
			return
		}

		// Only cache GET requests
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		// Skip caching for certain paths
		path := r.URL.Path
		if path == "/health" || strings.HasPrefix(path, "/cache/") || strings.HasPrefix(path, "/static/") || strings.HasPrefix(path, "/themes/") {
			next.ServeHTTP(w, r)
			return
		}

		// Try to build cache key (this is route-specific, so we'll handle it in handlers)
		// For now, just capture the response and serve normally
		responseCapture := cache.NewResponseCapture(w)
		next.ServeHTTP(responseCapture, r)

		// Flush the captured response to the client
		responseCapture.Flush()
	})
}
