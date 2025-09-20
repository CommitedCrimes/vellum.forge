package content

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/microcosm-cc/bluemonday"
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
	md        goldmark.Markdown
	sanitizer *bluemonday.Policy
}

// NewMarkdownParser creates a new markdown parser with all configured extensions
func NewMarkdownParser() *MarkdownParser {
	md := goldmark.New(
		goldmark.WithExtensions(
			&frontmatter.Extender{},
			extension.GFM, // Tables, strikethrough, task lists, autolink, emoji
			highlighting.NewHighlighting(
				highlighting.WithStyle("autumn"),
				highlighting.WithFormatOptions(
					html.WithClasses(false),
					html.WithLineNumbers(true),
				),
			),
			&mermaid.Extender{},
		),
		goldmark.WithRendererOptions(
			goldmarkHTML.WithUnsafe(), // Allow raw HTML
		),
	)

	// Create a permissive but safe HTML sanitizer
	sanitizer := bluemonday.NewPolicy()

	// Allow common HTML elements for rich content
	sanitizer.AllowElements("p", "br", "strong", "em", "u", "s", "del", "ins", "mark")
	sanitizer.AllowElements("h1", "h2", "h3", "h4", "h5", "h6")
	sanitizer.AllowElements("ul", "ol", "li", "dl", "dt", "dd")
	sanitizer.AllowElements("blockquote", "pre", "code")
	sanitizer.AllowElements("a", "img", "figure", "figcaption")
	sanitizer.AllowElements("table", "thead", "tbody", "tfoot", "tr", "th", "td")
	sanitizer.AllowElements("div", "span", "section", "article", "header", "footer", "main")

	// Allow attributes
	sanitizer.AllowAttrs("href", "title").OnElements("a")
	sanitizer.AllowAttrs("src", "alt", "title", "width", "height").OnElements("img")
	sanitizer.AllowAttrs("class", "id").Globally()
	sanitizer.AllowAttrs("style").OnElements("pre", "code", "span") // For syntax highlighting

	// Allow data attributes for Mermaid diagrams
	sanitizer.AllowAttrs("class").Matching(regexp.MustCompile(`^mermaid$`)).OnElements("pre")

	// Allow specific CSS properties for syntax highlighting
	sanitizer.AllowStyles("color", "background-color", "font-weight", "font-style", "text-decoration").Globally()

	return &MarkdownParser{
		md:        md,
		sanitizer: sanitizer,
	}
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

	// Sanitize the HTML output
	sanitizedHTML := p.sanitizer.Sanitize(htmlBuf.String())

	return &Content{
		Frontmatter: frontmatterData,
		Body:        body,
		HTML:        sanitizedHTML,
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
