package cache

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// FileWatcher watches for file changes and triggers cache invalidation
type FileWatcher struct {
	cache   *Cache
	logger  *slog.Logger
	paths   map[string]bool // paths being watched
	mu      sync.RWMutex
	stopCh  chan struct{}
	wg      sync.WaitGroup
	pollInt time.Duration // polling interval
}

// NewFileWatcher creates a new file watcher
func NewFileWatcher(cache *Cache, logger *slog.Logger) *FileWatcher {
	return &FileWatcher{
		cache:   cache,
		logger:  logger,
		paths:   make(map[string]bool),
		stopCh:  make(chan struct{}),
		pollInt: 10 * time.Second, // Poll every 10 seconds
	}
}

// Watch starts watching a directory for changes
func (fw *FileWatcher) Watch(path string) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	if !fw.paths[absPath] {
		fw.paths[absPath] = true
		fw.logger.Info("Started watching path for cache invalidation", "path", absPath)
	}

	return nil
}

// Start begins the file watching process
func (fw *FileWatcher) Start() {
	fw.wg.Add(1)
	go fw.watchLoop()
}

// Stop stops the file watcher
func (fw *FileWatcher) Stop() {
	close(fw.stopCh)
	fw.wg.Wait()
}

// watchLoop is the main watching loop using polling
func (fw *FileWatcher) watchLoop() {
	defer fw.wg.Done()

	// Track file modification times
	fileTimes := make(map[string]time.Time)

	ticker := time.NewTicker(fw.pollInt)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fw.checkForChanges(fileTimes)
		case <-fw.stopCh:
			return
		}
	}
}

// checkForChanges checks all watched paths for file changes
func (fw *FileWatcher) checkForChanges(fileTimes map[string]time.Time) {
	fw.mu.RLock()
	paths := make([]string, 0, len(fw.paths))
	for path := range fw.paths {
		paths = append(paths, path)
	}
	fw.mu.RUnlock()

	for _, watchPath := range paths {
		err := filepath.Walk(watchPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip files we can't access
			}

			// Only watch markdown files and template files
			if info.IsDir() || (!strings.HasSuffix(strings.ToLower(path), ".md") &&
				!strings.HasSuffix(strings.ToLower(path), ".jet") &&
				!strings.HasSuffix(strings.ToLower(path), ".tmpl")) {
				return nil
			}

			modTime := info.ModTime()
			lastTime, exists := fileTimes[path]

			if !exists {
				// First time seeing this file
				fileTimes[path] = modTime
				return nil
			}

			if modTime.After(lastTime) {
				// File has been modified
				fw.logger.Info("File changed, invalidating cache", "path", path)
				fileTimes[path] = modTime
				fw.handleFileChange(path)
			}

			return nil
		})

		if err != nil {
			fw.logger.Error("Error walking watched path", "path", watchPath, "error", err)
		}
	}
}

// handleFileChange processes a file change event
func (fw *FileWatcher) handleFileChange(filePath string) {
	absPath, _ := filepath.Abs(filePath)

	// Invalidate entries related to this file
	invalidated := fw.cache.InvalidateByFilePath(absPath)

	// For certain file types, also invalidate related caches
	if strings.Contains(absPath, "blog") {
		// If a blog post changed, also invalidate blog index
		blogIndexInvalidated := fw.cache.Invalidate("/blog")
		invalidated += blogIndexInvalidated

		// Also invalidate home page if it shows recent blog posts
		homeInvalidated := fw.cache.Invalidate("/")
		invalidated += homeInvalidated

		// Invalidate RSS feed and sitemap as they include blog posts
		rssInvalidated := fw.cache.Invalidate("rss:")
		invalidated += rssInvalidated
		sitemapInvalidated := fw.cache.Invalidate("sitemap:")
		invalidated += sitemapInvalidated
	}

	// If a page changed, invalidate sitemap
	if strings.Contains(absPath, "pages") {
		sitemapInvalidated := fw.cache.Invalidate("sitemap:")
		invalidated += sitemapInvalidated
	}

	// If theme files changed, invalidate everything
	if strings.Contains(absPath, "themes") || strings.Contains(absPath, "templates") {
		fw.logger.Info("Theme/template file changed, clearing entire cache", "path", filePath)
		fw.cache.Clear()
		return
	}

	if invalidated > 0 {
		fw.logger.Info("Cache entries invalidated", "file", filePath, "count", invalidated)
	}
}

// InvalidateForPath manually invalidates cache entries for a specific path
func (fw *FileWatcher) InvalidateForPath(path string) int {
	absPath, _ := filepath.Abs(path)
	return fw.cache.InvalidateByFilePath(absPath)
}

// CacheInvalidator provides methods to manually invalidate cache entries
type CacheInvalidator struct {
	cache  *Cache
	logger *slog.Logger
}

// NewCacheInvalidator creates a new cache invalidator
func NewCacheInvalidator(cache *Cache, logger *slog.Logger) *CacheInvalidator {
	return &CacheInvalidator{
		cache:  cache,
		logger: logger,
	}
}

// InvalidateByPath invalidates cache entries for a specific path pattern
func (ci *CacheInvalidator) InvalidateByPath(pathPattern string) int {
	count := ci.cache.Invalidate(pathPattern)
	if count > 0 {
		ci.logger.Info("Manual cache invalidation", "pattern", pathPattern, "count", count)
	}
	return count
}

// InvalidateByFile invalidates cache entries for a specific file
func (ci *CacheInvalidator) InvalidateByFile(filePath string) int {
	count := ci.cache.InvalidateByFilePath(filePath)
	if count > 0 {
		ci.logger.Info("Manual cache invalidation for file", "file", filePath, "count", count)
	}
	return count
}

// InvalidateBlogIndex invalidates the blog index cache
func (ci *CacheInvalidator) InvalidateBlogIndex() int {
	return ci.InvalidateByPath("/blog")
}

// InvalidateHome invalidates the home page cache
func (ci *CacheInvalidator) InvalidateHome() int {
	return ci.InvalidateByPath("/")
}

// InvalidateFeed invalidates the RSS feed cache
func (ci *CacheInvalidator) InvalidateFeed() int {
	return ci.InvalidateByPath("rss:")
}

// InvalidateSitemap invalidates the sitemap cache
func (ci *CacheInvalidator) InvalidateSitemap() int {
	return ci.InvalidateByPath("sitemap:")
}

// InvalidateAll clears the entire cache
func (ci *CacheInvalidator) InvalidateAll() {
	ci.cache.Clear()
	ci.logger.Info("Entire cache cleared")
}
