# Quick Start: Essential Ghost Features

This is a condensed guide for implementing the most critical Ghost-like features first.

## Week 1: RSS Feeds & Sitemap (Phase 2)

### RSS Feed Implementation

**1. Create feed package**
```bash
touch internal/feed/rss.go
```

**2. Add to config**
```go
// cmd/web/main.go - Add to config struct
type config struct {
    // ... existing fields
    site struct {
        title       string
        description string
        link        string
        author      string
    }
}

// In run() function
cfg.site.title = env.GetString("SITE_TITLE", "VellumForge Blog")
cfg.site.description = env.GetString("SITE_DESCRIPTION", "A blog built with VellumForge")
cfg.site.link = cfg.baseURL
cfg.site.author = env.GetString("SITE_AUTHOR", "VellumForge")
```

**3. Add route**
```go
// cmd/web/routes.go
mux.Get("/rss", app.rssFeed)
mux.Get("/feed", app.rssFeed)  // Alternative URL
```

**4. Create handler skeleton**
```go
// cmd/web/handlers.go
func (app *application) rssFeed(w http.ResponseWriter, r *http.Request) {
    // Load recent posts (limit to 20)
    // Generate RSS XML
    // Cache the feed
    // Set Content-Type: application/rss+xml
}
```

### Sitemap Implementation

**1. Create sitemap package**
```bash
touch internal/sitemap/sitemap.go
```

**2. Add routes**
```go
// cmd/web/routes.go
mux.Get("/sitemap.xml", app.sitemap)
mux.Get("/robots.txt", app.robotsTxt)
```

**3. Create handlers**
```go
// cmd/web/handlers.go
func (app *application) sitemap(w http.ResponseWriter, r *http.Request) {
    // Load all posts and pages
    // Generate sitemap XML
    // Cache the sitemap
}

func (app *application) robotsTxt(w http.ResponseWriter, r *http.Request) {
    // Return robots.txt with sitemap reference
}
```

---

## Week 2: SEO Meta Tags (Phase 3)

### Open Graph & Twitter Cards

**1. Create SEO package**
```bash
touch internal/seo/metatags.go
```

**2. Extend frontmatter**
```go
// internal/content/frontmatter.go
type Frontmatter struct {
    // ... existing fields
    OGTitle       string `yaml:"og_title"`
    OGDescription string `yaml:"og_description"`
    OGImage       string `yaml:"og_image"`
    TwitterCard   string `yaml:"twitter_card"`  // summary, summary_large_image
}
```

**3. Add to template data**
```go
// cmd/web/helpers.go
func (app *application) newTemplateData(r *http.Request) map[string]any {
    return map[string]any{
        "CurrentPath": r.URL.Path,
        "BaseURL":     app.config.baseURL,
        "SiteName":    app.config.site.title,
        // Add meta tag helper
        "Meta": app.buildMetaTags(r),
    }
}
```

**4. Update base template**
```html
<!-- themes/default/base.jet -->
<head>
    <!-- Basic meta -->
    <title>{{.Meta.Title}}</title>
    <meta name="description" content="{{.Meta.Description}}">

    <!-- Open Graph -->
    <meta property="og:type" content="{{.Meta.OGType}}">
    <meta property="og:title" content="{{.Meta.OGTitle}}">
    <meta property="og:description" content="{{.Meta.OGDescription}}">
    <meta property="og:image" content="{{.Meta.OGImage}}">
    <meta property="og:url" content="{{.Meta.CanonicalURL}}">

    <!-- Twitter Card -->
    <meta name="twitter:card" content="{{.Meta.TwitterCard}}">
    <meta name="twitter:title" content="{{.Meta.TwitterTitle}}">
    <meta name="twitter:description" content="{{.Meta.TwitterDescription}}">
    <meta name="twitter:image" content="{{.Meta.TwitterImage}}">

    <!-- Canonical -->
    <link rel="canonical" href="{{.Meta.CanonicalURL}}">
</head>
```

---

## Week 3: Pagination & Tag Pages (Phase 4)

### Pagination

**1. Create pagination package**
```bash
touch internal/pagination/paginator.go
```

**2. Add pagination to blog index**
```go
// cmd/web/handlers.go
func (app *application) blogIndex(w http.ResponseWriter, r *http.Request) {
    // Get page number from query param
    page := getPageNumber(r)
    perPage := 10

    // Load all posts
    allPosts, _, err := app.contentLoader.LoadBlogPosts(app.config.dataDir)

    // Paginate
    paginator := pagination.New(allPosts, page, perPage)

    data["Posts"] = paginator.Items()
    data["Pagination"] = paginator.Metadata()
}
```

**3. Update blog index route**
```go
// cmd/web/routes.go
mux.Get("/blog", app.blogIndex)
mux.Get("/blog/page/{number}", app.blogIndex)  // Paginated
```

**4. Add pagination template partial**
```html
<!-- themes/default/partials/pagination.jet -->
{{if .Pagination.TotalPages > 1}}
<nav class="pagination">
    {{if .Pagination.HasPrev}}
        <a href="{{.Pagination.PrevURL}}">← Previous</a>
    {{end}}

    {{range .Pagination.Pages}}
        <a href="{{.URL}}" class="{{if .Current}}active{{end}}">{{.Number}}</a>
    {{end}}

    {{if .Pagination.HasNext}}
        <a href="{{.Pagination.NextURL}}">Next →</a>
    {{end}}
</nav>
{{end}}
```

### Tag Pages

**1. Add tag routes**
```go
// cmd/web/routes.go
mux.Get("/tag/{slug}", app.tagArchive)
mux.Get("/tag/{slug}/page/{number}", app.tagArchive)
```

**2. Create tag handler**
```go
// cmd/web/handlers.go
func (app *application) tagArchive(w http.ResponseWriter, r *http.Request) {
    tagSlug := chi.URLParam(r, "slug")
    page := getPageNumber(r)

    // Load all posts with this tag
    posts := app.contentLoader.LoadPostsByTag(app.config.dataDir, tagSlug)

    // Paginate
    paginator := pagination.New(posts, page, 10)

    data["Tag"] = tagSlug
    data["Posts"] = paginator.Items()
    data["Pagination"] = paginator.Metadata()

    app.jetRenderer.RenderPage(w, http.StatusOK, data, "pages/tag.jet")
}
```

**3. Add filter method to content loader**
```go
// internal/content/loader.go
func (l *Loader) LoadPostsByTag(dataDir, tag string) ([]*Content, error) {
    allPosts, _, err := l.LoadBlogPosts(dataDir)
    // Filter posts that have the tag
    // Return filtered list
}
```

---

## Week 4: Reading Time & Excerpts (Phase 5)

### Reading Time

**1. Create reading time calculator**
```bash
touch internal/content/readingtime.go
```

```go
package content

import "strings"

const averageWPM = 250

func CalculateReadingTime(content string) int {
    words := len(strings.Fields(content))
    minutes := words / averageWPM
    if minutes < 1 {
        return 1
    }
    return minutes
}
```

**2. Add to Content struct**
```go
// internal/content/frontmatter.go
type Content struct {
    Frontmatter Frontmatter
    Body        string
    HTML        string
    ReadingTime int  // New field
}
```

**3. Calculate during parsing**
```go
// internal/content/loader.go
func (l *Loader) LoadContent(filePath string) (*Content, os.FileInfo, error) {
    // ... existing parsing code

    // Calculate reading time
    parsed.ReadingTime = CalculateReadingTime(parsed.Body)

    return parsed, fi, nil
}
```

**4. Display in template**
```html
<!-- themes/default/pages/blog/post.jet -->
<span class="reading-time">{{.Post.ReadingTime}} min read</span>
```

### Auto-Generated Excerpts

**1. Add excerpt generation**
```go
// internal/content/excerpt.go
package content

import "strings"

const defaultExcerptLength = 200

func GenerateExcerpt(content string, customLength ...int) string {
    length := defaultExcerptLength
    if len(customLength) > 0 {
        length = customLength[0]
    }

    // Strip HTML
    text := stripHTML(content)

    // Truncate at word boundary
    if len(text) <= length {
        return text
    }

    truncated := text[:length]
    lastSpace := strings.LastIndex(truncated, " ")
    if lastSpace > 0 {
        truncated = truncated[:lastSpace]
    }

    return truncated + "..."
}
```

**2. Extend frontmatter**
```go
type Frontmatter struct {
    // ... existing fields
    Excerpt string `yaml:"excerpt"`  // Optional manual excerpt
}
```

**3. Add to Content struct**
```go
type Content struct {
    // ... existing fields
    Excerpt string  // Auto-generated or from frontmatter
}
```

**4. Generate during parsing**
```go
// internal/content/loader.go
func (l *Loader) LoadContent(filePath string) (*Content, os.FileInfo, error) {
    // ... existing code

    // Set excerpt
    if parsed.Frontmatter.Excerpt != "" {
        parsed.Excerpt = parsed.Frontmatter.Excerpt
    } else {
        parsed.Excerpt = GenerateExcerpt(parsed.Body)
    }

    return parsed, fi, nil
}
```

**5. Use in blog index**
```html
<!-- themes/default/pages/blog/index.jet -->
{{range .BlogPosts}}
<article>
    <h2><a href="/blog/{{.Frontmatter.Slug}}">{{.Frontmatter.Title}}</a></h2>
    <p>{{.Excerpt}}</p>
    <span>{{.ReadingTime}} min read</span>
</article>
{{end}}
```

---

## Testing Each Feature

### RSS Feed Test
```bash
curl http://localhost:6886/rss
# Should return valid RSS XML
```

### Sitemap Test
```bash
curl http://localhost:6886/sitemap.xml
# Should return valid sitemap XML
```

### Meta Tags Test
```bash
curl http://localhost:6886/blog/your-post | grep "og:"
# Should see Open Graph tags
```

### Pagination Test
```bash
curl http://localhost:6886/blog/page/2
# Should show page 2 of posts
```

### Tag Archive Test
```bash
curl http://localhost:6886/tag/golang
# Should show posts tagged with 'golang'
```

---

## Environment Variables to Add

```bash
# .env
SITE_TITLE="My Blog"
SITE_DESCRIPTION="Thoughts on technology and software"
SITE_AUTHOR="Your Name"
TWITTER_HANDLE="@yourblog"
DEFAULT_SHARE_IMAGE="/images/default-share.jpg"
POSTS_PER_PAGE=10
EXCERPT_LENGTH=200
```

---

## Cache Invalidation Updates

For each new feature, update cache invalidation:

```go
// Tag pages need invalidation when posts change
app.cacheInvalidator.InvalidatePattern("/tag/")

// Pagination needs invalidation
app.cacheInvalidator.InvalidatePattern("/blog/page/")

// RSS/Sitemap need invalidation on any content change
app.cacheInvalidator.Invalidate("/rss")
app.cacheInvalidator.Invalidate("/sitemap.xml")
```

---

## What You'll Have After 4 Weeks

- ✅ RSS feed for blog posts
- ✅ Sitemap.xml for SEO
- ✅ robots.txt
- ✅ Open Graph tags for social sharing
- ✅ Twitter Cards
- ✅ Pagination for blog index
- ✅ Tag archive pages with pagination
- ✅ Reading time estimation
- ✅ Auto-generated excerpts
- ✅ Better meta tags and SEO

This gives you 80% of Ghost's content features with minimal complexity!

---

## Next Steps (Weeks 5-8)

1. **Author system** - Multi-author support
2. **Featured posts** - Highlight important content
3. **Related posts** - Content discovery
4. **Basic search** - In-memory search index
5. **Content API** - JSON API for headless usage

See `GHOST_FEATURE_PLAN.md` for complete roadmap.
