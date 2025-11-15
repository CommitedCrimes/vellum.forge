package feed

import (
	"encoding/xml"
	"fmt"
	"time"

	"vellum.forge/internal/content"
)

// RSS represents an RSS 2.0 feed
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Atom    string   `xml:"xmlns:atom,attr"`
	Content string   `xml:"xmlns:content,attr"`
	Channel *Channel `xml:"channel"`
}

// Channel represents the RSS channel
type Channel struct {
	Title         string `xml:"title"`
	Link          string `xml:"link"`
	Description   string `xml:"description"`
	Language      string `xml:"language,omitempty"`
	Copyright     string `xml:"copyright,omitempty"`
	ManagingEditor string `xml:"managingEditor,omitempty"`
	WebMaster     string `xml:"webMaster,omitempty"`
	PubDate       string `xml:"pubDate,omitempty"`
	LastBuildDate string `xml:"lastBuildDate,omitempty"`
	Generator     string `xml:"generator,omitempty"`
	AtomLink      *AtomLink `xml:"atom:link"`
	Items         []*Item   `xml:"item"`
}

// AtomLink represents the atom:link element for feed self-reference
type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

// Item represents an RSS item (blog post)
type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Author      string `xml:"author,omitempty"`
	Category    string `xml:"category,omitempty"`
	GUID        *GUID  `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Content     string `xml:"content:encoded,omitempty"`
}

// GUID represents the globally unique identifier for an item
type GUID struct {
	IsPermaLink bool   `xml:"isPermaLink,attr"`
	Value       string `xml:",chardata"`
}

// Config holds the configuration for RSS feed generation
type Config struct {
	Title          string
	Link           string
	Description    string
	Language       string
	Copyright      string
	ManagingEditor string
	WebMaster      string
	Generator      string
}

// GenerateRSS creates an RSS feed from blog posts
func GenerateRSS(posts []*content.Content, config Config, feedURL string, limit int) ([]byte, error) {
	// Limit the number of items
	if limit > 0 && len(posts) > limit {
		posts = posts[:limit]
	}

	// Find the most recent post date for lastBuildDate
	var lastBuildDate time.Time
	for _, post := range posts {
		if post.Frontmatter.Date.After(lastBuildDate) {
			lastBuildDate = post.Frontmatter.Date
		}
	}

	// Create the RSS feed
	rss := &RSS{
		Version: "2.0",
		Atom:    "http://www.w3.org/2005/Atom",
		Content: "http://purl.org/rss/1.0/modules/content/",
		Channel: &Channel{
			Title:          config.Title,
			Link:           config.Link,
			Description:    config.Description,
			Language:       config.Language,
			Copyright:      config.Copyright,
			ManagingEditor: config.ManagingEditor,
			WebMaster:      config.WebMaster,
			Generator:      config.Generator,
			LastBuildDate:  formatRFC822(lastBuildDate),
			AtomLink: &AtomLink{
				Href: feedURL,
				Rel:  "self",
				Type: "application/rss+xml",
			},
			Items: make([]*Item, 0, len(posts)),
		},
	}

	// Add items
	for _, post := range posts {
		// Skip draft posts
		if post.Frontmatter.Draft {
			continue
		}

		item := &Item{
			Title:       post.Frontmatter.Title,
			Link:        fmt.Sprintf("%s/blog/%s", config.Link, post.Frontmatter.Slug),
			Description: post.Frontmatter.Description,
			PubDate:     formatRFC822(post.Frontmatter.Date),
			GUID: &GUID{
				IsPermaLink: true,
				Value:       fmt.Sprintf("%s/blog/%s", config.Link, post.Frontmatter.Slug),
			},
		}

		// Add content:encoded for full HTML content
		if post.HTML != "" {
			item.Content = post.HTML
		}

		// Add first tag as category if available
		if len(post.Frontmatter.Tags) > 0 {
			item.Category = post.Frontmatter.Tags[0]
		}

		rss.Channel.Items = append(rss.Channel.Items, item)
	}

	// Marshal to XML
	output, err := xml.MarshalIndent(rss, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal RSS feed: %w", err)
	}

	// Add XML declaration
	xmlDeclaration := []byte(xml.Header)
	return append(xmlDeclaration, output...), nil
}

// formatRFC822 formats a time.Time to RFC822 format (required by RSS 2.0)
func formatRFC822(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC1123Z)
}
