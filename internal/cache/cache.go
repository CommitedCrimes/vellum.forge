package cache

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Entry represents a cached response
type Entry struct {
	Body        []byte
	Headers     http.Header
	StatusCode  int
	ContentType string
	ETag        string
	LastMod     time.Time
	Size        int64
	CreatedAt   time.Time
	ExpiresAt   time.Time
	AccessedAt  time.Time
	AccessCount int64
}

// IsExpired checks if the cache entry has expired
func (e *Entry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// SizeBytes returns the total size of the entry in bytes
func (e *Entry) SizeBytes() int64 {
	size := int64(len(e.Body))

	// Add headers size estimation
	for k, values := range e.Headers {
		size += int64(len(k))
		for _, v := range values {
			size += int64(len(v))
		}
	}

	// Add other fields size estimation
	size += int64(len(e.ContentType))
	size += int64(len(e.ETag))
	size += 200 // rough estimate for other fields

	return size
}

// Touch updates the access time and count
func (e *Entry) Touch() {
	e.AccessedAt = time.Now()
	e.AccessCount++
}

// KeyParams contains parameters for generating cache keys
type KeyParams struct {
	Method         string
	NormalizedPath string
	AbsFilePath    string
	FileMTimeNs    int64
	Template       string
	ThemeID        string
	FeatureFlags   map[string]bool
	AcceptEncoding string
	ContentType    string
}

// Config contains cache configuration
type Config struct {
	MaxEntries   int           // Maximum number of cache entries
	MaxSizeBytes int64         // Maximum total cache size in bytes
	DefaultTTL   time.Duration // Default TTL for entries
	CleanupFreq  time.Duration // How often to run cleanup
}

// DefaultConfig returns a default cache configuration
func DefaultConfig() Config {
	return Config{
		MaxEntries:   1000,
		MaxSizeBytes: 100 * 1024 * 1024, // 100MB
		DefaultTTL:   time.Hour,
		CleanupFreq:  5 * time.Minute,
	}
}

// Cache represents an in-memory LRU+TTL cache
type Cache struct {
	config    Config
	entries   map[string]*Entry
	lruList   *lruList
	totalSize int64
	mu        sync.RWMutex
	stopCh    chan struct{}
	wg        sync.WaitGroup
}

// New creates a new cache with the given configuration
func New(config Config) *Cache {
	c := &Cache{
		config:  config,
		entries: make(map[string]*Entry),
		lruList: newLRUList(),
		stopCh:  make(chan struct{}),
	}

	// Start cleanup goroutine
	c.wg.Add(1)
	go c.janitor()

	return c
}

// GenerateKey generates a cache key from the given parameters
func GenerateKey(params KeyParams) string {
	h := sha256.New()

	// Write all parameters to hash
	h.Write([]byte(params.Method))
	h.Write([]byte("|"))
	h.Write([]byte(params.NormalizedPath))
	h.Write([]byte("|"))
	h.Write([]byte(params.AbsFilePath))
	h.Write([]byte("|"))
	h.Write([]byte(strconv.FormatInt(params.FileMTimeNs, 10)))
	h.Write([]byte("|"))
	h.Write([]byte(params.Template))
	h.Write([]byte("|"))
	h.Write([]byte(params.ThemeID))
	h.Write([]byte("|"))

	// Add feature flags in sorted order for consistency
	var flags []string
	for k, v := range params.FeatureFlags {
		flags = append(flags, fmt.Sprintf("%s:%t", k, v))
	}
	for _, flag := range flags {
		h.Write([]byte(flag))
		h.Write([]byte(","))
	}

	h.Write([]byte("|"))
	h.Write([]byte(params.AcceptEncoding))
	h.Write([]byte("|"))
	h.Write([]byte(params.ContentType))

	return fmt.Sprintf("%x", h.Sum(nil))
}

// Get retrieves an entry from the cache
func (c *Cache) Get(key string) (*Entry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if entry.IsExpired() {
		c.removeEntry(key)
		return nil, false
	}

	// Touch entry and move to front of LRU
	entry.Touch()
	c.lruList.moveToFront(key)

	return entry, true
}

// Set stores an entry in the cache
func (c *Cache) Set(key string, entry *Entry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If entry already exists, remove it first
	if existing, exists := c.entries[key]; exists {
		c.totalSize -= existing.SizeBytes()
		c.lruList.remove(key)
	}

	// Set expiration if not set
	if entry.ExpiresAt.IsZero() {
		entry.ExpiresAt = time.Now().Add(c.config.DefaultTTL)
	}

	// Set creation time if not set
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}

	// Touch the entry
	entry.Touch()

	// Add to cache
	c.entries[key] = entry
	c.totalSize += entry.SizeBytes()
	c.lruList.addToFront(key)

	// Evict if necessary
	c.evictIfNecessary()
}

// Delete removes an entry from the cache
func (c *Cache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, exists := c.entries[key]; exists {
		c.totalSize -= entry.SizeBytes()
		delete(c.entries, key)
		c.lruList.remove(key)
		return true
	}

	return false
}

// Invalidate removes entries based on a path pattern
func (c *Cache) Invalidate(pathPattern string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	var keysToRemove []string

	// Find keys that match the pattern
	for key := range c.entries {
		// For now, we'll match based on the path being contained in the key
		// This is a simple implementation - could be enhanced with more sophisticated matching
		if strings.Contains(key, pathPattern) {
			keysToRemove = append(keysToRemove, key)
		}
	}

	// Remove matching entries
	for _, key := range keysToRemove {
		c.removeEntry(key)
	}

	return len(keysToRemove)
}

// InvalidateByFilePath invalidates cache entries for a specific file path
func (c *Cache) InvalidateByFilePath(filePath string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	var keysToRemove []string
	absPath, _ := filepath.Abs(filePath)

	// Find keys that contain this file path
	for key := range c.entries {
		if strings.Contains(key, absPath) || strings.Contains(key, filePath) {
			keysToRemove = append(keysToRemove, key)
		}
	}

	// Remove matching entries
	for _, key := range keysToRemove {
		c.removeEntry(key)
	}

	return len(keysToRemove)
}

// Stats returns cache statistics
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return CacheStats{
		Entries:      len(c.entries),
		SizeBytes:    c.totalSize,
		MaxEntries:   c.config.MaxEntries,
		MaxSizeBytes: c.config.MaxSizeBytes,
	}
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*Entry)
	c.lruList = newLRUList()
	c.totalSize = 0
}

// Close stops the cache and cleanup routines
func (c *Cache) Close() {
	close(c.stopCh)
	c.wg.Wait()
}

// removeEntry removes an entry (must be called with lock held)
func (c *Cache) removeEntry(key string) {
	if entry, exists := c.entries[key]; exists {
		c.totalSize -= entry.SizeBytes()
		delete(c.entries, key)
		c.lruList.remove(key)
	}
}

// evictIfNecessary removes old entries to stay within limits (must be called with lock held)
func (c *Cache) evictIfNecessary() {
	// Evict by count
	for len(c.entries) > c.config.MaxEntries {
		oldest := c.lruList.removeOldest()
		if oldest != "" {
			c.removeEntry(oldest)
		}
	}

	// Evict by size
	for c.totalSize > c.config.MaxSizeBytes && len(c.entries) > 0 {
		oldest := c.lruList.removeOldest()
		if oldest != "" {
			c.removeEntry(oldest)
		}
	}
}

// janitor runs periodic cleanup
func (c *Cache) janitor() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.config.CleanupFreq)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopCh:
			return
		}
	}
}

// cleanup removes expired entries
func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiredKeys []string
	now := time.Now()

	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		c.removeEntry(key)
	}
}

// CacheStats contains cache statistics
type CacheStats struct {
	Entries      int
	SizeBytes    int64
	MaxEntries   int
	MaxSizeBytes int64
}

// ShouldBypass checks if cache should be bypassed based on request
func ShouldBypass(r *http.Request) bool {
	// Check query parameter
	if r.URL.Query().Get("nocache") == "1" {
		return true
	}

	// Check header
	if r.Header.Get("X-Bypass-Cache") == "1" {
		return true
	}

	return false
}

// GetFileMTime gets the modification time of a file in nanoseconds
func GetFileMTime(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return info.ModTime().UnixNano(), nil
}

// CreateCacheEntry creates a cache entry from response data
func CreateCacheEntry(body []byte, headers http.Header, statusCode int, ttl time.Duration) *Entry {
	now := time.Now()

	entry := &Entry{
		Body:        make([]byte, len(body)),
		Headers:     make(http.Header),
		StatusCode:  statusCode,
		Size:        int64(len(body)),
		CreatedAt:   now,
		ExpiresAt:   now.Add(ttl),
		AccessedAt:  now,
		AccessCount: 1,
	}

	// Copy body
	copy(entry.Body, body)

	// Copy headers
	for k, values := range headers {
		entry.Headers[k] = make([]string, len(values))
		copy(entry.Headers[k], values)
	}

	// Extract common headers
	entry.ContentType = headers.Get("Content-Type")
	entry.ETag = headers.Get("ETag")

	if lastModStr := headers.Get("Last-Modified"); lastModStr != "" {
		if parsed, err := http.ParseTime(lastModStr); err == nil {
			entry.LastMod = parsed
		}
	}

	return entry
}

// WriteEntryToResponse writes a cache entry to an HTTP response
func WriteEntryToResponse(w http.ResponseWriter, entry *Entry) error {
	// Copy headers
	for k, values := range entry.Headers {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}

	// Write status code
	w.WriteHeader(entry.StatusCode)

	// Write body
	_, err := w.Write(entry.Body)
	return err
}

// SupportsConditionalRequests checks if the entry supports conditional requests
func (e *Entry) SupportsConditionalRequests() bool {
	return e.ETag != "" || !e.LastMod.IsZero()
}

// MatchesConditions checks if the entry matches conditional request headers
func (e *Entry) MatchesConditions(r *http.Request) bool {
	// Check If-None-Match (ETag)
	if inm := r.Header.Get("If-None-Match"); inm != "" && e.ETag != "" {
		// Simple ETag comparison - could be enhanced for weak/strong ETags
		etags := strings.Split(inm, ",")
		for _, etag := range etags {
			etag = strings.TrimSpace(etag)
			if etag == "*" || etag == e.ETag {
				return true
			}
		}
	}

	// Check If-Modified-Since
	if ims := r.Header.Get("If-Modified-Since"); ims != "" && !e.LastMod.IsZero() {
		if since, err := http.ParseTime(ims); err == nil {
			// If entry hasn't been modified since the given time
			return !e.LastMod.After(since)
		}
	}

	return false
}
