package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"

	"vellum.forge/internal/content"
	"vellum.forge/internal/env"
	"vellum.forge/internal/response"
	"vellum.forge/internal/version"

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
	cacheTTL int
	dataDir  string
	themeDir string
}

type application struct {
	config        config
	logger        *slog.Logger
	wg            sync.WaitGroup
	contentLoader *content.Loader
	jetRenderer   *response.JetRenderer
}

func run(logger *slog.Logger) error {
	var cfg config

	cfg.baseURL = env.GetString("BASE_URL", "http://localhost:6886")
	cfg.httpPort = env.GetInt("PORT", 6886)
	cfg.cookie.secretKey = env.GetString("COOKIE_SECRET_KEY", "fredbzsw2qsqb3mto3xfxnclebdt4hht")
	cfg.cacheTTL = env.GetInt("CACHE_TTL", 3600)
	cfg.dataDir = env.GetString("DATA_DIR", "data")
	cfg.themeDir = env.GetString("THEME_DIR", "themes")
	cfg.theme = env.GetString("THEME", "default")

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

	return app.serveHTTP()
}
