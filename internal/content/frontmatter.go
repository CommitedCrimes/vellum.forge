package content

import (
	"time"
)

// Frontmatter represents the YAML frontmatter structure for content files
type Frontmatter struct {
	Title       string    `yaml:"title"`
	Date        time.Time `yaml:"date"`
	Tags        []string  `yaml:"tags"`
	Description string    `yaml:"description"`
	Cover       string    `yaml:"cover"`
	Draft       bool      `yaml:"draft"`
	Slug        string    `yaml:"slug"`
}

// Content represents a parsed content file with frontmatter and body
type Content struct {
	Frontmatter Frontmatter
	Body        string
	HTML        string
}

// GetSlug returns the slug from frontmatter or generates one from title
func (c *Content) GetSlug() string {
	if c.Frontmatter.Slug != "" {
		return c.Frontmatter.Slug
	}
	// TODO: Implement slug generation from title
	return ""
}

// GetDate returns the date from frontmatter or current time as fallback
func (c *Content) GetDate() time.Time {
	if !c.Frontmatter.Date.IsZero() {
		return c.Frontmatter.Date
	}
	return time.Now()
}
