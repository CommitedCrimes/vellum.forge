package sitemap

import (
	"encoding/xml"
	"fmt"
	"time"

	"vellum.forge/internal/content"
)

// URLSet represents the root element of a sitemap
type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	XMLNS   string   `xml:"xmlns,attr"`
	URLs    []*URL   `xml:"url"`
}

// URL represents a single URL entry in the sitemap
type URL struct {
	Loc        string  `xml:"loc"`
	LastMod    string  `xml:"lastmod,omitempty"`
	ChangeFreq string  `xml:"changefreq,omitempty"`
	Priority   float64 `xml:"priority,omitempty"`
}

// ChangeFreq represents how frequently a page is likely to change
type ChangeFreq string

const (
	Always  ChangeFreq = "always"
	Hourly  ChangeFreq = "hourly"
	Daily   ChangeFreq = "daily"
	Weekly  ChangeFreq = "weekly"
	Monthly ChangeFreq = "monthly"
	Yearly  ChangeFreq = "yearly"
	Never   ChangeFreq = "never"
)

// Config holds configuration for sitemap generation
type Config struct {
	BaseURL string
}

// SitemapEntry represents a page to be included in the sitemap
type SitemapEntry struct {
	URL        string
	LastMod    time.Time
	ChangeFreq ChangeFreq
	Priority   float64
}

// GenerateSitemap creates an XML sitemap from the provided entries
func GenerateSitemap(entries []*SitemapEntry) ([]byte, error) {
	urlset := &URLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  make([]*URL, 0, len(entries)),
	}

	for _, entry := range entries {
		url := &URL{
			Loc:      entry.URL,
			Priority: entry.Priority,
		}

		if !entry.LastMod.IsZero() {
			url.LastMod = formatW3CDatetime(entry.LastMod)
		}

		if entry.ChangeFreq != "" {
			url.ChangeFreq = string(entry.ChangeFreq)
		}

		urlset.URLs = append(urlset.URLs, url)
	}

	// Marshal to XML
	output, err := xml.MarshalIndent(urlset, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sitemap: %w", err)
	}

	// Add XML declaration
	xmlDeclaration := []byte(xml.Header)
	return append(xmlDeclaration, output...), nil
}

// BuildSitemapFromContent builds sitemap entries from blog posts and pages
func BuildSitemapFromContent(
	baseURL string,
	posts []*content.Content,
	postsModTime time.Time,
	pages []*content.Content,
	pagesModTime time.Time,
) []*SitemapEntry {
	entries := make([]*SitemapEntry, 0)

	// Add home page
	entries = append(entries, &SitemapEntry{
		URL:        baseURL + "/",
		LastMod:    postsModTime, // Home page changes when posts change
		ChangeFreq: Daily,
		Priority:   1.0,
	})

	// Add blog index
	entries = append(entries, &SitemapEntry{
		URL:        baseURL + "/blog",
		LastMod:    postsModTime,
		ChangeFreq: Daily,
		Priority:   0.9,
	})

	// Add blog posts
	for _, post := range posts {
		// Skip draft posts
		if post.Frontmatter.Draft {
			continue
		}

		entry := &SitemapEntry{
			URL:        fmt.Sprintf("%s/blog/%s", baseURL, post.Frontmatter.Slug),
			LastMod:    post.Frontmatter.Date,
			ChangeFreq: Monthly,
			Priority:   0.8,
		}

		entries = append(entries, entry)
	}

	// Add static pages
	for _, page := range pages {
		// Skip draft pages
		if page.Frontmatter.Draft {
			continue
		}

		entry := &SitemapEntry{
			URL:        fmt.Sprintf("%s/%s", baseURL, page.Frontmatter.Slug),
			LastMod:    page.Frontmatter.Date,
			ChangeFreq: Weekly,
			Priority:   0.7,
		}

		entries = append(entries, entry)
	}

	return entries
}

// AddTagPages adds tag archive pages to the sitemap entries
func AddTagPages(baseURL string, tags []string, lastMod time.Time, entries []*SitemapEntry) []*SitemapEntry {
	for _, tag := range tags {
		entry := &SitemapEntry{
			URL:        fmt.Sprintf("%s/tag/%s", baseURL, tag),
			LastMod:    lastMod,
			ChangeFreq: Weekly,
			Priority:   0.6,
		}
		entries = append(entries, entry)
	}
	return entries
}

// AddAuthorPages adds author pages to the sitemap entries
func AddAuthorPages(baseURL string, authors []string, lastMod time.Time, entries []*SitemapEntry) []*SitemapEntry {
	for _, author := range authors {
		entry := &SitemapEntry{
			URL:        fmt.Sprintf("%s/author/%s", baseURL, author),
			LastMod:    lastMod,
			ChangeFreq: Weekly,
			Priority:   0.6,
		}
		entries = append(entries, entry)
	}
	return entries
}

// formatW3CDatetime formats a time.Time to W3C datetime format (required by sitemaps)
// Format: YYYY-MM-DD or YYYY-MM-DDThh:mm:ss+00:00
func formatW3CDatetime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	// Use ISO 8601 / RFC3339 format
	return t.Format(time.RFC3339)
}
