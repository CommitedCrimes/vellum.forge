package content

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldmarkHTML "github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/frontmatter"
	"go.abhg.dev/goldmark/mermaid"
	"gopkg.in/yaml.v3"
)

// MarkdownParser handles parsing markdown with frontmatter and extensions
type MarkdownParser struct {
	md goldmark.Markdown
}

// NewMarkdownParser creates a new markdown parser with all configured extensions
func NewMarkdownParser() *MarkdownParser {
	md := goldmark.New(
		goldmark.WithExtensions(
			&frontmatter.Extender{},
			extension.GFM, // Tables, strikethrough, task lists, autolink, emoji
			highlighting.NewHighlighting(
				highlighting.WithStyle("vim"),
				highlighting.WithFormatOptions(
					html.WithLineNumbers(true),
				),
			),
			&mermaid.Extender{},
		),
		goldmark.WithRendererOptions(
			goldmarkHTML.WithUnsafe(), // Allow raw HTML
		),
	)

	return &MarkdownParser{md: md}
}

// Parse parses markdown content with frontmatter
func (p *MarkdownParser) Parse(content []byte) (*Content, error) {
	// Create parser context
	ctx := parser.NewContext()

	// Convert markdown to HTML
	var htmlBuf bytes.Buffer
	if err := p.md.Convert(content, &htmlBuf, parser.WithContext(ctx)); err != nil {
		return nil, fmt.Errorf("failed to convert markdown: %w", err)
	}

	// Extract frontmatter
	fm := frontmatter.Get(ctx)
	var frontmatterData Frontmatter
	if fm != nil {
		if err := fm.Decode(&frontmatterData); err != nil {
			return nil, fmt.Errorf("failed to decode frontmatter: %w", err)
		}
	}

	// Extract body content (everything after frontmatter)
	body := p.extractBody(content)

	return &Content{
		Frontmatter: frontmatterData,
		Body:        body,
		HTML:        htmlBuf.String(),
	}, nil
}

// extractBody extracts the markdown body content after frontmatter
func (p *MarkdownParser) extractBody(content []byte) string {
	contentStr := string(content)

	// Find the end of frontmatter
	if strings.HasPrefix(contentStr, "---\n") {
		// Find the closing ---
		parts := strings.SplitN(contentStr, "\n---\n", 2)
		if len(parts) > 1 {
			return strings.TrimSpace(parts[1])
		}
	}

	return contentStr
}

// ParseFrontmatterOnly parses only the frontmatter from content
func (p *MarkdownParser) ParseFrontmatterOnly(content []byte) (*Frontmatter, error) {
	contentStr := string(content)

	// Check if content has frontmatter
	if !strings.HasPrefix(contentStr, "---\n") {
		return &Frontmatter{}, nil
	}

	// Extract frontmatter section
	parts := strings.SplitN(contentStr, "\n---\n", 2)
	if len(parts) < 2 {
		return &Frontmatter{}, nil
	}

	frontmatterStr := parts[0][4:] // Remove the opening ---

	var frontmatterData Frontmatter
	if err := yaml.Unmarshal([]byte(frontmatterStr), &frontmatterData); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return &frontmatterData, nil
}
