package main

import (
	"net/http"

	"vellum.forge/assets"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()
	mux.NotFound(app.notFound)

	mux.Use(app.logAccess)
	mux.Use(app.recoverPanic)
	mux.Use(app.securityHeaders)
	mux.Use(app.cacheControlMiddleware)
	mux.Use(app.contentTypeMiddleware)
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.Compress(5)) // gzip compression
	mux.Use(middleware.StripSlashes)
	mux.Use(app.cacheMiddleware)

	// Static assets
	fileServer := http.FileServer(http.FS(assets.EmbeddedFiles))
	mux.Handle("/static/*", fileServer)

	// Theme assets (css, js, images, etc from themes/{theme}/assets/)
	mux.Handle("/themes/*", http.HandlerFunc(app.themeAssets))

	// User attachment images (from data/attachments)
	mux.Handle("/images/*", http.HandlerFunc(app.attachmentImages))

	// Routes
	mux.Get("/", app.home)
	mux.Get("/blog", app.blogIndex)
	mux.Get("/blog/{slug}", app.blogPost)
	mux.Get("/{slug}", app.page)
	mux.Get("/health", app.health)

	// RSS and sitemap
	mux.Get("/rss", app.rssFeed)
	mux.Get("/feed", app.rssFeed)        // Alternative RSS URL
	mux.Get("/sitemap.xml", app.sitemap)
	mux.Get("/robots.txt", app.robotsTxt)

	// Cache stats and clear
	mux.Get("/cache/stats", app.cacheStats)
	mux.Post("/cache/clear", app.cacheClear)

	return mux
}
