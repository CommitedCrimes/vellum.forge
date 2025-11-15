# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

VellumForge is a Go-based blog engine that renders markdown content with a focus on performance and flexibility. It uses:
- **Goldmark** for markdown parsing with frontmatter, syntax highlighting, GFM, and Mermaid diagram support
- **Jet templates** for HTML rendering with theme fallback support
- **chi** router for HTTP routing
- **In-memory LRU+TTL cache** with file watching for automatic invalidation

Module path: `vellum.forge`

## Common Commands

### Development
```bash
# Install dependencies
go mod tidy

# Build the application
make build

# Run the application
make run

# Run with live reload (watches .go, .tmpl, .html, .css, .js, .sql, image files)
make run/live

# Format and tidy
make tidy
```

### Testing
```bash
# Run all tests
make test

# Run tests with coverage report
make test/cover

# Run quality control checks (tests, vet, staticcheck, govulncheck)
make audit
```

### Running
```bash
# Run directly
go run ./cmd/web

# Run on custom port
export HTTP_PORT=9999
go run ./cmd/web

# View version
go run ./cmd/web -version
```

## Architecture

### Application Structure

The application follows the `application` struct pattern where all dependencies are injected:

```go
type application struct {
    config           config
    logger           *slog.Logger
    contentLoader    *content.Loader      // Loads markdown files
    jetRenderer      *response.JetRenderer // Renders Jet templates
    cache            *cache.Cache          // LRU+TTL cache
    cacheKeyBuilder  *cache.CacheKeyBuilder
    cacheInvalidator *cache.CacheInvalidator
    fileWatcher      *cache.FileWatcher    // Watches data/ and themes/
}
```

All handlers, middleware, and helpers are methods on this struct, giving them access to dependencies.

### Content System

**Content Loading Flow:**
1. Markdown files are stored in `data/blog/` (blog posts) and `data/pages/` (static pages)
2. `content.Loader` reads files and parses them with `MarkdownParser`
3. `MarkdownParser` extracts YAML frontmatter and converts markdown to HTML using Goldmark
4. Content is sanitized with bluemonday policy (configured in `internal/content/markdown.go`)

**Frontmatter Structure:**
```yaml
---
title: "Post Title"
slug: "custom-slug"  # Optional, defaults to filename
date: 2024-01-15
draft: false
tags: ["tag1", "tag2"]
---
```

**Markdown Extensions Enabled:**
- GFM (tables, strikethrough, task lists, autolinks, emoji)
- Syntax highlighting with Chroma (autumn theme, line numbers, CSS classes)
- Mermaid diagrams
- Raw HTML allowed (but sanitized)

### Template System (Jet)

**Template Resolution:**
VellumForge uses a fallback template loader:
1. First tries `themes/{THEME}/` (from `THEME` env var, default: "default")
2. Falls back to `themes/default/` if template not found in active theme

**Template Structure:**
- `pages/` - Full page templates (home.jet, blog.jet, blog-post.jet, page.jet)
- `partials/` - Reusable template fragments
- `layouts/` - Base layouts

**Custom Template Functions:**
- `version()` - Application version
- `now()` - Current time
- `formatDate(time, layout)` - Format time with Go layout
- `humanizeTime(time)` - "2 days ago" style formatting
- `safeHTML(string)` - Mark string as safe HTML
- `truncate(string, length)` - Truncate with ellipsis

**Rendering in Handlers:**
```go
data := app.newTemplateData(r)
data["Title"] = "My Page"
err := app.jetRenderer.RenderPage(w, http.StatusOK, data, "pages/home.jet")
```

### Cache System

**Architecture:**
- **LRU eviction** when entry count or size limits exceeded
- **TTL expiration** with background cleanup every 5 minutes
- **File watching** auto-invalidates cache when content/theme files change
- **Conditional requests** support (ETag, If-None-Match, If-Modified-Since)

**Cache Key Composition:**
Cache keys include:
- HTTP method
- Request path (normalized)
- File path + modification time (nanosecond precision)
- Template name
- Theme ID
- Feature flags (e.g., Mermaid support)
- Accept-Encoding header

**Cache Bypass:**
- Query param: `?nocache=1`
- Header: `X-Bypass-Cache: 1`

**Cache Invalidation:**
Auto-invalidation watches `data/` and `themes/{theme}/` directories:
- **Blog post changes** → invalidate specific post + blog index + home page
- **Page changes** → invalidate specific page only
- **Theme/template changes** → clear entire cache

**Cache API Endpoints:**
- `GET /cache/stats` - View cache statistics (entries, size, utilization)
- `POST /cache/clear` - Manually clear entire cache

**Helper for Caching in Handlers:**
```go
cacheKey, _ := app.cacheKeyBuilder.BuildKeyForHome(r)
err := app.renderWithCache(w, r, cacheKey, func(w http.ResponseWriter) error {
    // Render logic here
    return app.jetRenderer.RenderPage(w, http.StatusOK, data, "pages/home.jet")
})
```

### Routing

Routes are defined in `cmd/web/routes.go` using chi router:
- `/` - Home page
- `/blog` - Blog index
- `/blog/{slug}` - Individual blog post
- `/{slug}` - Static page
- `/static/*` - Static assets (embedded from `assets/static/`)
- `/themes/*` - Theme assets (from `themes/{theme}/assets/`)
- `/images/*` - User attachments (from `data/attachments/`)
- `/cache/stats` - Cache statistics
- `/cache/clear` - Clear cache
- `/health` - Health check

**Middleware Chain (in order):**
1. `logAccess` - Request logging
2. `recoverPanic` - Panic recovery
3. `securityHeaders` - Security headers
4. `cacheControlMiddleware` - Cache-Control headers
5. `contentTypeMiddleware` - Content-Type handling
6. `RequestID` (chi) - Request ID generation
7. `RealIP` (chi) - Real IP extraction
8. `Logger` (chi) - Chi's logger
9. `Recoverer` (chi) - Chi's recoverer
10. `Compress(5)` (chi) - gzip compression
11. `StripSlashes` (chi) - Strip trailing slashes
12. `cacheMiddleware` - Custom cache middleware

### Configuration

All configuration is via environment variables (loaded from `.env` if present):

```bash
BASE_URL="http://localhost:6886"      # Base URL
PORT=6886                              # HTTP port
THEME="default"                        # Active theme
DATA_DIR="data"                        # Content directory
THEME_DIR="themes"                     # Themes directory
CACHE_ENABLED=true                     # Enable/disable cache
CACHE_TTL=3600                         # Cache TTL in seconds
CACHE_MAX_ENTRIES=1000                 # Max cache entries
CACHE_MAX_SIZE_MB=100                  # Max cache size in MB
COOKIE_SECRET_KEY="..."                # Secret for signed/encrypted cookies
```

Configuration is parsed in `cmd/web/main.go` using `internal/env` helpers.

### Content Directory Structure

```
data/
├── blog/          # Blog posts (markdown files)
├── pages/         # Static pages (markdown files)
└── attachments/   # User-uploaded images/attachments
```

### Embedded Assets

Static assets in `assets/static/` are embedded into the binary via `assets/efs.go` using Go's `embed` package.

## Important Patterns

### Error Handling in Handlers
```go
func (app *application) yourHandler(w http.ResponseWriter, r *http.Request) {
    // Use helper methods for consistent error responses
    if err != nil {
        app.serverError(w, r, err)  // 500 Internal Server Error
        return
    }

    // Other error helpers:
    // app.notFound(w, r)           // 404 Not Found
    // app.badRequest(w, r, err)    // 400 Bad Request
}
```

### Background Tasks
Use `app.backgroundTask()` for async operations:
```go
app.backgroundTask(r, func() error {
    // Background work here
    return nil
})
```

Graceful shutdown waits for all background tasks to complete.

### Testing Patterns
- Handler tests use `testutils_test.go` helpers
- Use `httptest.NewRecorder()` for response capture
- Tests are in `*_test.go` files alongside source files
- Run tests with race detector: `go test -race ./...`

## Key Implementation Details

### Cache Key Generation
Cache keys must be invalidated when:
- File content changes (tracked via modification time)
- Theme changes
- Template changes
- Query parameters change

The `cache.CacheKeyBuilder` handles this complexity - use it rather than building keys manually.

### File Watching Implementation
File watching uses **polling** (not OS events) for reliability:
- Default poll interval: 2 seconds
- Compares modification times to detect changes
- Tracks directories: `data/` and `themes/{theme}/`

### Template Fallback
When a template is not found in the active theme, the system automatically falls back to `themes/default/`. This allows themes to override only specific templates.

### Markdown Sanitization
HTML in markdown is allowed but sanitized via bluemonday. The policy in `internal/content/markdown.go` defines which HTML elements/attributes are permitted. Modify this policy if you need to allow additional HTML elements.

## Development Notes

- Entry point is `cmd/web/main.go` - start reading there
- HTTP server startup/shutdown logic is in `cmd/web/server.go`
- Handlers are in `cmd/web/handlers.go` (can split into multiple files for larger apps)
- All errors are logged with structured logging (slog with tint)
- HTTP server errors are automatically logged at `Warn` level
- Live reload watches: `.go`, `.tpl`, `.tmpl`, `.html`, `.css`, `.js`, `.sql`, and image files

## Cache Implementation Details

See `CACHE.md` for comprehensive cache system documentation including:
- LRU+TTL algorithm details
- Cache key generation logic
- Auto-invalidation mechanisms
- Conditional request handling
- Performance optimization strategies
