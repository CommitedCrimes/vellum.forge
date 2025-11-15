package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"

	"vellum.forge/internal/cache"
	"vellum.forge/internal/content"
	"vellum.forge/internal/env"
	"vellum.forge/internal/response"
	"vellum.forge/internal/version"

	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
)

func main() {
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug}))

	err := run(logger)
	if err != nil {
		trace := string(debug.Stack())
		logger.Error(err.Error(), "trace", trace)
		os.Exit(1)
	}
}

type config struct {
	baseURL  string
	httpPort int
	theme    string
	cookie   struct {
		secretKey string
	}
	cacheTTL        int
	dataDir         string
	themeDir        string
	cacheEnabled    bool
	cacheMaxSize    int64
	cacheMaxEntries int
}

type application struct {
	config           config
	logger           *slog.Logger
	wg               sync.WaitGroup
	contentLoader    *content.Loader
	jetRenderer      *response.JetRenderer
	cache            *cache.Cache
	cacheKeyBuilder  *cache.CacheKeyBuilder
	cacheInvalidator *cache.CacheInvalidator
	fileWatcher      *cache.FileWatcher
}

func run(logger *slog.Logger) error {
	// Load .env file if it exists (silently ignore if it doesn't)
	_ = godotenv.Load()

	var cfg config

	cfg.baseURL = env.GetString("BASE_URL", "http://localhost:6886")
	cfg.httpPort = env.GetInt("PORT", 6886)
	cfg.cookie.secretKey = env.GetString("COOKIE_SECRET_KEY", "fredbzsw2qsqb3mto3xfxnclebdt4hht")
	cfg.cacheTTL = env.GetInt("CACHE_TTL", 3600)
	cfg.dataDir = env.GetString("DATA_DIR", "data")
	cfg.themeDir = env.GetString("THEME_DIR", "themes")
	cfg.theme = env.GetString("THEME", "default")
	cfg.cacheEnabled = env.GetBool("CACHE_ENABLED", true)
	cfg.cacheMaxSize = int64(env.GetInt("CACHE_MAX_SIZE_MB", 100)) * 1024 * 1024 // Convert MB to bytes
	cfg.cacheMaxEntries = env.GetInt("CACHE_MAX_ENTRIES", 1000)

	showVersion := flag.Bool("version", false, "display version and exit")

	flag.Parse()

	if *showVersion {
		fmt.Printf("version: %s\n", version.Get())
		return nil
	}

	// Initialize Jet renderer
	themeDir := filepath.Join(cfg.themeDir, cfg.theme)
	jetRenderer, err := response.NewJetRenderer(themeDir)
	if err != nil {
		return fmt.Errorf("failed to initialize Jet renderer: %w", err)
	}

	app := &application{
		config:        cfg,
		logger:        logger,
		contentLoader: content.NewLoader(),
		jetRenderer:   jetRenderer,
	}

	// Initialize cache if enabled
	if cfg.cacheEnabled {
		cacheConfig := cache.Config{
			MaxEntries:   cfg.cacheMaxEntries,
			MaxSizeBytes: cfg.cacheMaxSize,
			DefaultTTL:   time.Duration(cfg.cacheTTL) * time.Second,
			CleanupFreq:  5 * time.Minute,
		}

		app.cache = cache.NewWithLogger(cacheConfig, logger)
		app.cacheKeyBuilder = cache.NewCacheKeyBuilder(cfg.theme, cfg.dataDir, cfg.themeDir)
		app.cacheInvalidator = cache.NewCacheInvalidator(app.cache, logger)

		// Initialize file watcher for auto-invalidation
		app.fileWatcher = cache.NewFileWatcher(app.cache, logger)
		err = app.fileWatcher.Watch(cfg.dataDir)
		if err != nil {
			logger.Warn("Failed to watch data directory for cache invalidation", "error", err)
		}
		err = app.fileWatcher.Watch(themeDir)
		if err != nil {
			logger.Warn("Failed to watch theme directory for cache invalidation", "error", err)
		}
		app.fileWatcher.Start()

		logger.Info("Cache initialized",
			"maxEntries", cfg.cacheMaxEntries,
			"maxSizeMB", cfg.cacheMaxSize/(1024*1024),
			"ttlSeconds", cfg.cacheTTL)
	}

	return app.serveHTTP()
}
