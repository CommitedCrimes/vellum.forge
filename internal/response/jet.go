package response

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/CloudyKit/jet/v6"
	"vellum.forge/assets"
	"vellum.forge/internal/version"
)

// JetRenderer handles Jet template rendering with theme directory support
type JetRenderer struct {
	views *jet.Set
}

// NewJetRenderer creates a new Jet template renderer
func NewJetRenderer(themeDir string) (*JetRenderer, error) {
	// Create a new Jet set with the theme directory
	views := jet.NewSet(
		jet.NewOSFileSystemLoader(themeDir),
		jet.InDevelopmentMode(), // Enable for development
	)

	// Add custom functions with proper types
	addJetFunctions(views)

	return &JetRenderer{
		views: views,
	}, nil
}

// NewJetRendererWithEmbedded creates a Jet renderer using embedded templates
func NewJetRendererWithEmbedded() (*JetRenderer, error) {
	// Create custom loader for embedded templates
	loader := &embeddedJetLoader{}

	views := jet.NewSet(
		loader,
		jet.InDevelopmentMode(),
	)

	// Add custom functions with proper types
	addJetFunctions(views)

	return &JetRenderer{
		views: views,
	}, nil
}

// addJetFunctions adds all the custom functions to the Jet view set
func addJetFunctions(views *jet.Set) {
	// Basic utility functions
	views.AddGlobal("version", func() string {
		return version.Get()
	})

	views.AddGlobal("now", func() time.Time {
		return time.Now()
	})

	// Date/time formatting functions
	views.AddGlobal("formatDate", func(t time.Time, layout string) string {
		return t.Format(layout)
	})

	views.AddGlobal("humanizeTime", func(t time.Time) string {
		duration := time.Since(t)
		return approxDuration(duration) + " ago"
	})

	// String functions
	views.AddGlobal("safeHTML", func(s string) template.HTML {
		return template.HTML(s)
	})

	views.AddGlobal("truncate", func(s string, length int) string {
		if len(s) <= length {
			return s
		}
		return s[:length] + "..."
	})
}

// approxDuration returns a human-readable approximation of a duration
func approxDuration(d time.Duration) string {
	const (
		day  = 24 * time.Hour
		year = 365 * day
	)

	switch {
	case d >= year:
		years := int(d / year)
		if years == 1 {
			return "1 year"
		}
		return fmt.Sprintf("%d years", years)
	case d >= day:
		days := int(d / day)
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	case d >= time.Hour:
		hours := int(d / time.Hour)
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	case d >= time.Minute:
		minutes := int(d / time.Minute)
		if minutes == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", minutes)
	case d >= time.Second:
		seconds := int(d / time.Second)
		if seconds == 1 {
			return "1 second"
		}
		return fmt.Sprintf("%d seconds", seconds)
	default:
		return "less than 1 second"
	}
}

// RenderPage renders a page template with the given data
func (jr *JetRenderer) RenderPage(w http.ResponseWriter, status int, data any, templatePath string) error {
	return jr.RenderPageWithHeaders(w, status, data, nil, templatePath)
}

// RenderPageWithHeaders renders a page template with custom headers
func (jr *JetRenderer) RenderPageWithHeaders(w http.ResponseWriter, status int, data any, headers http.Header, templatePath string) error {
	// Load the template
	tmpl, err := jr.views.GetTemplate(templatePath)
	if err != nil {
		return fmt.Errorf("failed to load template %s: %w", templatePath, err)
	}

	// Create template variables
	vars := make(jet.VarMap)

	// Convert data to jet variables
	if data != nil {
		if dataMap, ok := data.(map[string]any); ok {
			for key, value := range dataMap {
				vars.Set(key, value)
			}
		} else {
			vars.Set("Data", data)
		}
	}

	// Set headers
	for key, values := range headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	// Execute template
	return tmpl.Execute(w, vars, nil)
}

// RenderPartial renders a partial template
func (jr *JetRenderer) RenderPartial(w io.Writer, templatePath string, data any) error {
	tmpl, err := jr.views.GetTemplate(templatePath)
	if err != nil {
		return fmt.Errorf("failed to load partial template %s: %w", templatePath, err)
	}

	vars := make(jet.VarMap)
	if data != nil {
		if dataMap, ok := data.(map[string]any); ok {
			for key, value := range dataMap {
				vars.Set(key, value)
			}
		} else {
			vars.Set("Data", data)
		}
	}

	return tmpl.Execute(w, vars, nil)
}

// embeddedJetLoader implements jet.Loader for embedded file systems
type embeddedJetLoader struct{}

func (l *embeddedJetLoader) Open(name string) (io.ReadCloser, error) {
	// Ensure the path starts with templates/
	if !strings.HasPrefix(name, "templates/") {
		name = filepath.Join("templates", name)
	}

	file, err := assets.EmbeddedFiles.Open(name)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (l *embeddedJetLoader) Exists(name string) bool {
	// Ensure the path starts with templates/
	if !strings.HasPrefix(name, "templates/") {
		name = filepath.Join("templates", name)
	}

	file, err := assets.EmbeddedFiles.Open(name)
	if err != nil {
		return false
	}
	file.Close()
	return true
}
