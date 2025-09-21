package cache

import (
	"net/http"
	"testing"
	"time"
)

func TestCache_BasicOperations(t *testing.T) {
	config := Config{
		MaxEntries:   10,
		MaxSizeBytes: 1024,
		DefaultTTL:   time.Hour,
		CleanupFreq:  time.Minute,
	}
	cache := New(config)
	defer cache.Close()

	// Test Set and Get
	entry := &Entry{
		Body:        []byte("test content"),
		Headers:     make(http.Header),
		StatusCode:  200,
		ContentType: "text/html",
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(time.Hour),
	}

	key := "test-key"
	cache.Set(key, entry)

	retrieved, found := cache.Get(key)
	if !found {
		t.Error("Expected to find cached entry")
	}

	if string(retrieved.Body) != "test content" {
		t.Errorf("Expected 'test content', got '%s'", string(retrieved.Body))
	}

	// Test Delete
	deleted := cache.Delete(key)
	if !deleted {
		t.Error("Expected entry to be deleted")
	}

	_, found = cache.Get(key)
	if found {
		t.Error("Expected entry to be gone after deletion")
	}
}

func TestCache_LRUEviction(t *testing.T) {
	config := Config{
		MaxEntries:   2,
		MaxSizeBytes: 1024,
		DefaultTTL:   time.Hour,
		CleanupFreq:  time.Minute,
	}
	cache := New(config)
	defer cache.Close()

	// Add entries beyond max capacity
	for i := 0; i < 3; i++ {
		entry := &Entry{
			Body:        []byte("content"),
			Headers:     make(http.Header),
			StatusCode:  200,
			ContentType: "text/html",
			CreatedAt:   time.Now(),
			ExpiresAt:   time.Now().Add(time.Hour),
		}
		cache.Set(string(rune('a'+i)), entry)
		time.Sleep(time.Millisecond) // Ensure different timestamps
	}

	// First entry should be evicted
	_, found := cache.Get("a")
	if found {
		t.Error("Expected oldest entry 'a' to be evicted")
	}

	// Last two should still be there
	_, found = cache.Get("b")
	if !found {
		t.Error("Expected entry 'b' to still exist")
	}

	_, found = cache.Get("c")
	if !found {
		t.Error("Expected entry 'c' to still exist")
	}
}

func TestCache_TTLExpiration(t *testing.T) {
	config := Config{
		MaxEntries:   10,
		MaxSizeBytes: 1024,
		DefaultTTL:   10 * time.Millisecond,
		CleanupFreq:  5 * time.Millisecond,
	}
	cache := New(config)
	defer cache.Close()

	entry := &Entry{
		Body:        []byte("content"),
		Headers:     make(http.Header),
		StatusCode:  200,
		ContentType: "text/html",
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(10 * time.Millisecond),
	}

	cache.Set("test", entry)

	// Should be found immediately
	_, found := cache.Get("test")
	if !found {
		t.Error("Expected entry to be found")
	}

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Should be expired now
	_, found = cache.Get("test")
	if found {
		t.Error("Expected entry to be expired")
	}
}

func TestGenerateKey(t *testing.T) {
	params1 := KeyParams{
		Method:         "GET",
		NormalizedPath: "/blog/test",
		AbsFilePath:    "/path/to/file.md",
		FileMTimeNs:    123456789,
		Template:       "blog.jet",
		ThemeID:        "default",
		FeatureFlags:   map[string]bool{"mermaid": true},
		AcceptEncoding: "gzip",
		ContentType:    "text/html",
	}

	params2 := params1 // Same params
	params3 := params1
	params3.Method = "POST" // Different method

	key1 := GenerateKey(params1)
	key2 := GenerateKey(params2)
	key3 := GenerateKey(params3)

	if key1 != key2 {
		t.Error("Expected same parameters to generate same key")
	}

	if key1 == key3 {
		t.Error("Expected different parameters to generate different keys")
	}
}

func TestShouldBypass(t *testing.T) {
	// Test query parameter bypass
	req1, _ := http.NewRequest("GET", "/?nocache=1", nil)
	if !ShouldBypass(req1) {
		t.Error("Expected to bypass cache with nocache=1")
	}

	// Test header bypass
	req2, _ := http.NewRequest("GET", "/", nil)
	req2.Header.Set("X-Bypass-Cache", "1")
	if !ShouldBypass(req2) {
		t.Error("Expected to bypass cache with X-Bypass-Cache header")
	}

	// Test normal request
	req3, _ := http.NewRequest("GET", "/", nil)
	if ShouldBypass(req3) {
		t.Error("Expected normal request to not bypass cache")
	}
}

func TestEntry_MatchesConditions(t *testing.T) {
	entry := &Entry{
		ETag:    `"abc123"`,
		LastMod: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	// Test If-None-Match
	req1, _ := http.NewRequest("GET", "/", nil)
	req1.Header.Set("If-None-Match", `"abc123"`)
	if !entry.MatchesConditions(req1) {
		t.Error("Expected ETag to match")
	}

	// Test If-Modified-Since
	req2, _ := http.NewRequest("GET", "/", nil)
	req2.Header.Set("If-Modified-Since", "Sun, 01 Jan 2023 13:00:00 GMT")
	if !entry.MatchesConditions(req2) {
		t.Error("Expected not modified since condition to match")
	}

	// Test no match
	req3, _ := http.NewRequest("GET", "/", nil)
	req3.Header.Set("If-None-Match", `"different"`)
	if entry.MatchesConditions(req3) {
		t.Error("Expected different ETag to not match")
	}
}
