package main

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	err := app.jetRenderer.RenderPage(w, http.StatusOK, data, "pages/home.jet")
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) blogIndex(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	// Load blog posts from content directory (with app.config.dataDir as base directory)
	blogPosts, err := app.contentLoader.LoadBlogPosts(app.config.dataDir)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data["BlogPosts"] = blogPosts

	err = app.jetRenderer.RenderPage(w, http.StatusOK, data, "pages/blog/index.jet")
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) blogPost(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	blogPost, err := app.contentLoader.LoadBlogPost(app.config.dataDir, slug)
	if err != nil {
		app.notFound(w, r)
		return
	}

	data := app.newTemplateData(r)
	data["Post"] = blogPost

	err = app.jetRenderer.RenderPage(w, http.StatusOK, data, "pages/blog/post.jet")
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) page(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	page, err := app.contentLoader.LoadPage(app.config.dataDir, slug)
	if err != nil {
		app.notFound(w, r)
		return
	}

	data := app.newTemplateData(r)
	data["Page"] = page

	err = app.jetRenderer.RenderPage(w, http.StatusOK, data, "pages/page.jet")
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
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
