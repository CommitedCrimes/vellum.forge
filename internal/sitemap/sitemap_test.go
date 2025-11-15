package sitemap

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"vellum.forge/internal/content"
)

func TestGenerateSitemap(t *testing.T) {
	entries := []*SitemapEntry{
		{
			URL:        "https://example.com/",
			LastMod:    time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			ChangeFreq: Daily,
			Priority:   1.0,
		},
		{
			URL:        "https://example.com/blog/post-1",
			LastMod:    time.Date(2024, 1, 14, 12, 0, 0, 0, time.UTC),
			ChangeFreq: Monthly,
			Priority:   0.8,
		},
	}

	sitemapData, err := GenerateSitemap(entries)
	if err != nil {
		t.Fatalf("GenerateSitemap failed: %v", err)
	}

	// Parse the XML to verify structure
	var urlset URLSet
	err = xml.Unmarshal(sitemapData, &urlset)
	if err != nil {
		t.Fatalf("Failed to parse generated sitemap: %v", err)
	}

	// Verify URLSet
	if urlset.XMLNS != "http://www.sitemaps.org/schemas/sitemap/0.9" {
		t.Errorf("Expected XMLNS 'http://www.sitemaps.org/schemas/sitemap/0.9', got %s", urlset.XMLNS)
	}

	if len(urlset.URLs) != 2 {
		t.Errorf("Expected 2 URLs, got %d", len(urlset.URLs))
	}

	// Verify first URL
	if urlset.URLs[0].Loc != "https://example.com/" {
		t.Errorf("Expected loc 'https://example.com/', got %s", urlset.URLs[0].Loc)
	}

	if urlset.URLs[0].ChangeFreq != "daily" {
		t.Errorf("Expected changefreq 'daily', got %s", urlset.URLs[0].ChangeFreq)
	}

	if urlset.URLs[0].Priority != 1.0 {
		t.Errorf("Expected priority 1.0, got %f", urlset.URLs[0].Priority)
	}

	// Verify lastmod is formatted correctly
	if urlset.URLs[0].LastMod == "" {
		t.Error("LastMod should not be empty")
	}
}

func TestBuildSitemapFromContent(t *testing.T) {
	posts := []*content.Content{
		{
			Frontmatter: content.Frontmatter{
				Title: "Published Post",
				Slug:  "published-post",
				Date:  time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
				Draft: false,
			},
		},
		{
			Frontmatter: content.Frontmatter{
				Title: "Draft Post",
				Slug:  "draft-post",
				Date:  time.Date(2024, 1, 16, 10, 0, 0, 0, time.UTC),
				Draft: true,
			},
		},
	}

	pages := []*content.Content{
		{
			Frontmatter: content.Frontmatter{
				Title: "About Page",
				Slug:  "about",
				Date:  time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
				Draft: false,
			},
		},
	}

	baseURL := "https://example.com"
	postsModTime := time.Now()
	pagesModTime := time.Now()

	entries := BuildSitemapFromContent(baseURL, posts, postsModTime, pages, pagesModTime)

	// Should have:
	// - Home page
	// - Blog index
	// - 1 published post (draft excluded)
	// - 1 page
	// Total: 4 entries
	expectedCount := 4
	if len(entries) != expectedCount {
		t.Errorf("Expected %d entries, got %d", expectedCount, len(entries))
	}

	// Verify home page is first
	if entries[0].URL != "https://example.com/" {
		t.Errorf("Expected first entry to be home page, got %s", entries[0].URL)
	}

	if entries[0].Priority != 1.0 {
		t.Errorf("Expected home page priority 1.0, got %f", entries[0].Priority)
	}

	// Verify blog index is second
	if entries[1].URL != "https://example.com/blog" {
		t.Errorf("Expected second entry to be blog index, got %s", entries[1].URL)
	}

	// Verify published post is included
	foundPost := false
	for _, entry := range entries {
		if entry.URL == "https://example.com/blog/published-post" {
			foundPost = true
			if entry.Priority != 0.8 {
				t.Errorf("Expected post priority 0.8, got %f", entry.Priority)
			}
		}
	}
	if !foundPost {
		t.Error("Published post not found in sitemap")
	}

	// Verify draft post is NOT included
	for _, entry := range entries {
		if strings.Contains(entry.URL, "draft-post") {
			t.Error("Draft post should not be in sitemap")
		}
	}

	// Verify page is included
	foundPage := false
	for _, entry := range entries {
		if entry.URL == "https://example.com/about" {
			foundPage = true
			if entry.Priority != 0.7 {
				t.Errorf("Expected page priority 0.7, got %f", entry.Priority)
			}
		}
	}
	if !foundPage {
		t.Error("Page not found in sitemap")
	}
}

func TestAddTagPages(t *testing.T) {
	baseURL := "https://example.com"
	tags := []string{"golang", "web", "testing"}
	lastMod := time.Now()
	entries := []*SitemapEntry{}

	entries = AddTagPages(baseURL, tags, lastMod, entries)

	if len(entries) != 3 {
		t.Errorf("Expected 3 tag entries, got %d", len(entries))
	}

	// Verify tag URL format
	expectedURL := "https://example.com/tag/golang"
	if entries[0].URL != expectedURL {
		t.Errorf("Expected %s, got %s", expectedURL, entries[0].URL)
	}

	// Verify priority
	if entries[0].Priority != 0.6 {
		t.Errorf("Expected tag priority 0.6, got %f", entries[0].Priority)
	}
}

func TestAddAuthorPages(t *testing.T) {
	baseURL := "https://example.com"
	authors := []string{"john-doe", "jane-smith"}
	lastMod := time.Now()
	entries := []*SitemapEntry{}

	entries = AddAuthorPages(baseURL, authors, lastMod, entries)

	if len(entries) != 2 {
		t.Errorf("Expected 2 author entries, got %d", len(entries))
	}

	// Verify author URL format
	expectedURL := "https://example.com/author/john-doe"
	if entries[0].URL != expectedURL {
		t.Errorf("Expected %s, got %s", expectedURL, entries[0].URL)
	}
}

func TestGenerateSitemapValidXML(t *testing.T) {
	entries := []*SitemapEntry{
		{
			URL:      "https://example.com/",
			Priority: 1.0,
		},
	}

	sitemapData, err := GenerateSitemap(entries)
	if err != nil {
		t.Fatalf("GenerateSitemap failed: %v", err)
	}

	// Verify XML declaration is present
	if !strings.HasPrefix(string(sitemapData), "<?xml") {
		t.Error("Sitemap should start with XML declaration")
	}

	// Verify it's valid XML by parsing
	var urlset URLSet
	err = xml.Unmarshal(sitemapData, &urlset)
	if err != nil {
		t.Errorf("Generated sitemap is not valid XML: %v", err)
	}
}

func TestFormatW3CDatetime(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	formatted := formatW3CDatetime(testTime)

	// Should be in RFC3339 format
	expected := "2024-01-15T10:30:00Z"
	if formatted != expected {
		t.Errorf("Expected %s, got %s", expected, formatted)
	}

	// Test zero time
	zeroFormatted := formatW3CDatetime(time.Time{})
	if zeroFormatted != "" {
		t.Errorf("Expected empty string for zero time, got %s", zeroFormatted)
	}
}
