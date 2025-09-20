package content

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Loader handles loading and parsing content files
type Loader struct {
	parser *MarkdownParser
}

// NewLoader creates a new content loader
func NewLoader() *Loader {
	return &Loader{
		parser: NewMarkdownParser(),
	}
}

// LoadContent loads and parses a single content file
func (l *Loader) LoadContent(filePath string) (*Content, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	parsed, err := l.parser.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse content from %s: %w", filePath, err)
	}

	// Generate slug from filename if not provided in frontmatter
	if parsed.Frontmatter.Slug == "" {
		base := filepath.Base(filePath)
		ext := filepath.Ext(base)
		parsed.Frontmatter.Slug = strings.TrimSuffix(base, ext)
	}

	return parsed, nil
}

// LoadContentFromDir loads all content files from a directory
func (l *Loader) LoadContentFromDir(dirPath string) ([]*Content, error) {
	var contents []*Content

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-markdown files
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}

		content, err := l.LoadContent(path)
		if err != nil {
			return fmt.Errorf("failed to load content from %s: %w", path, err)
		}

		// Skip draft content
		if content.Frontmatter.Draft {
			return nil
		}

		contents = append(contents, content)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", dirPath, err)
	}

	// Sort by date (newest first)
	sort.Slice(contents, func(i, j int) bool {
		return contents[i].GetDate().After(contents[j].GetDate())
	})

	return contents, nil
}

// LoadBlogPosts loads blog posts from the content directory
func (l *Loader) LoadBlogPosts(contentDir string) ([]*Content, error) {
	blogDir := filepath.Join(contentDir, "blog")
	return l.LoadContentFromDir(blogDir)
}

// LoadPage loads a single page by slug
func (l *Loader) LoadPage(contentDir, slug string) (*Content, error) {
	pagesDir := filepath.Join(contentDir, "pages")

	// Try to find the file by slug
	var foundPath string
	err := filepath.WalkDir(pagesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}

		// Check if the file matches the slug
		base := filepath.Base(path)
		ext := filepath.Ext(base)
		fileSlug := strings.TrimSuffix(base, ext)

		if fileSlug == slug {
			foundPath = path
			return filepath.SkipDir // Stop walking
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search for page %s: %w", slug, err)
	}

	if foundPath == "" {
		return nil, fmt.Errorf("page not found: %s", slug)
	}

	return l.LoadContent(foundPath)
}

// LoadBlogPost loads a single blog post by slug
func (l *Loader) LoadBlogPost(contentDir, slug string) (*Content, error) {
	blogDir := filepath.Join(contentDir, "blog")

	// Try to find the file by slug
	var foundPath string
	err := filepath.WalkDir(blogDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}

		// Check if the file matches the slug
		base := filepath.Base(path)
		ext := filepath.Ext(base)
		fileSlug := strings.TrimSuffix(base, ext)

		if fileSlug == slug {
			foundPath = path
			return filepath.SkipDir // Stop walking
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search for blog post %s: %w", slug, err)
	}

	if foundPath == "" {
		return nil, fmt.Errorf("blog post not found: %s", slug)
	}

	return l.LoadContent(foundPath)
}
