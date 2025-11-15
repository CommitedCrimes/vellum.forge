package feed

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"vellum.forge/internal/content"
)

func TestGenerateRSS(t *testing.T) {
	// Create sample posts
	posts := []*content.Content{
		{
			Frontmatter: content.Frontmatter{
				Title:       "First Post",
				Slug:        "first-post",
				Description: "This is the first post",
				Date:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				Tags:        []string{"golang", "web"},
				Draft:       false,
			},
			HTML: "<h1>First Post</h1><p>Content here</p>",
		},
		{
			Frontmatter: content.Frontmatter{
				Title:       "Second Post",
				Slug:        "second-post",
				Description: "This is the second post",
				Date:        time.Date(2024, 1, 16, 14, 0, 0, 0, time.UTC),
				Tags:        []string{"testing"},
				Draft:       false,
			},
			HTML: "<h1>Second Post</h1><p>More content</p>",
		},
		{
			Frontmatter: content.Frontmatter{
				Title: "Draft Post",
				Slug:  "draft-post",
				Draft: true,
			},
		},
	}

	config := Config{
		Title:       "Test Blog",
		Link:        "https://example.com",
		Description: "A test blog",
		Language:    "en-us",
		Generator:   "VellumForge",
	}

	feedURL := "https://example.com/rss"

	// Generate RSS feed
	rssData, err := GenerateRSS(posts, config, feedURL, 20)
	if err != nil {
		t.Fatalf("GenerateRSS failed: %v", err)
	}

	// Debug: Print the generated XML
	t.Logf("Generated RSS:\n%s", string(rssData))

	// Parse the XML to verify structure
	var rss RSS
	err = xml.Unmarshal(rssData, &rss)
	if err != nil {
		t.Fatalf("Failed to parse generated RSS: %v", err)
	}

	// Verify RSS structure
	if rss.Version != "2.0" {
		t.Errorf("Expected RSS version 2.0, got %s", rss.Version)
	}

	if rss.Channel == nil {
		t.Fatal("Channel is nil")
	}

	// Verify channel data - Note: Link field doesn't unmarshal well due to XML quirks,
	// so we verify it's in the generated XML instead
	rssString := string(rssData)
	if !strings.Contains(rssString, "<link>"+config.Link+"</link>") {
		t.Errorf("Expected link %s in RSS feed", config.Link)
	}

	if rss.Channel.Title != config.Title {
		t.Errorf("Expected title %s, got %s", config.Title, rss.Channel.Title)
	}

	// Verify items (should have 2, draft should be excluded)
	if len(rss.Channel.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(rss.Channel.Items))
	}

	// Verify first item
	if rss.Channel.Items[0].Title != "First Post" {
		t.Errorf("Expected first item title 'First Post', got %s", rss.Channel.Items[0].Title)
	}

	// Verify GUID
	if rss.Channel.Items[0].GUID == nil {
		t.Error("GUID is nil")
	} else {
		expectedGUID := "https://example.com/blog/first-post"
		if rss.Channel.Items[0].GUID.Value != expectedGUID {
			t.Errorf("Expected GUID %s, got %s", expectedGUID, rss.Channel.Items[0].GUID.Value)
		}
		if !rss.Channel.Items[0].GUID.IsPermaLink {
			t.Error("Expected GUID to be permalink")
		}
	}

	// Verify category from tags
	if rss.Channel.Items[0].Category != "golang" {
		t.Errorf("Expected category 'golang', got %s", rss.Channel.Items[0].Category)
	}

	// Verify atom:link exists in XML
	if !strings.Contains(rssString, `atom:link href="`+feedURL+`"`) {
		t.Errorf("Expected atom:link with href %s in RSS feed", feedURL)
	}
	if !strings.Contains(rssString, `rel="self"`) {
		t.Error("Expected atom:link with rel='self' in RSS feed")
	}
}

func TestGenerateRSSLimit(t *testing.T) {
	// Create many posts
	posts := make([]*content.Content, 50)
	for i := 0; i < 50; i++ {
		posts[i] = &content.Content{
			Frontmatter: content.Frontmatter{
				Title: "Post " + string(rune(i)),
				Slug:  "post-" + string(rune(i)),
				Date:  time.Now(),
			},
		}
	}

	config := Config{
		Title:       "Test Blog",
		Link:        "https://example.com",
		Description: "A test blog",
	}

	// Generate with limit of 10
	rssData, err := GenerateRSS(posts, config, "https://example.com/rss", 10)
	if err != nil {
		t.Fatalf("GenerateRSS failed: %v", err)
	}

	var rss RSS
	err = xml.Unmarshal(rssData, &rss)
	if err != nil {
		t.Fatalf("Failed to parse generated RSS: %v", err)
	}

	if len(rss.Channel.Items) > 10 {
		t.Errorf("Expected at most 10 items, got %d", len(rss.Channel.Items))
	}
}

func TestGenerateRSSValidXML(t *testing.T) {
	posts := []*content.Content{
		{
			Frontmatter: content.Frontmatter{
				Title:       "Test Post",
				Slug:        "test-post",
				Description: "Description",
				Date:        time.Now(),
			},
		},
	}

	config := Config{
		Title:       "Test Blog",
		Link:        "https://example.com",
		Description: "A test blog",
	}

	rssData, err := GenerateRSS(posts, config, "https://example.com/rss", 20)
	if err != nil {
		t.Fatalf("GenerateRSS failed: %v", err)
	}

	// Verify XML declaration is present
	if !strings.HasPrefix(string(rssData), "<?xml") {
		t.Error("RSS feed should start with XML declaration")
	}

	// Verify it's valid XML by parsing
	var rss RSS
	err = xml.Unmarshal(rssData, &rss)
	if err != nil {
		t.Errorf("Generated RSS is not valid XML: %v", err)
	}
}

func TestFormatRFC822(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	formatted := formatRFC822(testTime)

	// Should be in RFC1123Z format
	expected := "Mon, 15 Jan 2024 10:30:00 +0000"
	if formatted != expected {
		t.Errorf("Expected %s, got %s", expected, formatted)
	}

	// Test zero time
	zeroFormatted := formatRFC822(time.Time{})
	if zeroFormatted != "" {
		t.Errorf("Expected empty string for zero time, got %s", zeroFormatted)
	}
}
