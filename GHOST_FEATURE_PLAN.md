# Ghost CMS Feature Parity Plan

This document outlines the roadmap to bring VellumForge to feature parity with Ghost CMS.

## Current State

VellumForge currently has:
- ✅ Markdown-based blog posts and pages
- ✅ Frontmatter support (title, date, tags, description, cover, draft, slug)
- ✅ Theme system with fallback support
- ✅ Advanced caching (LRU+TTL with auto-invalidation)
- ✅ Syntax highlighting and Mermaid diagrams
- ✅ Image attachments

## Phase 1: Author System

**Goal:** Support single and multiple authors with author profiles

### Frontmatter Changes
```yaml
---
title: "Post Title"
date: 2024-01-15
author: "john-doe"  # Single author (slug)
# OR
authors: ["john-doe", "jane-smith"]  # Multiple authors
tags: ["tech", "golang"]
---
```

### Author Profile Files
Create `data/authors/{author-slug}.md`:
```yaml
---
name: "John Doe"
slug: "john-doe"
email: "john@example.com"
bio: "Software engineer and blogger"
avatar: "/images/authors/john-doe.jpg"
website: "https://johndoe.com"
twitter: "@johndoe"
github: "johndoe"
location: "San Francisco, CA"
---

Extended bio in markdown format...
```

### Implementation Tasks
- [ ] Add `Author` and `Authors []string` fields to `Frontmatter` struct
- [ ] Create `internal/content/author.go` with `Author` struct and loader
- [ ] Add `LoadAuthor(authorSlug)` and `LoadAuthors()` methods
- [ ] Update handlers to pass author data to templates
- [ ] Create author page route: `/author/{slug}`
- [ ] Add author archive route: `/author/{slug}/page/{number}`
- [ ] Update templates to display author info on posts
- [ ] Add author byline to blog post template
- [ ] Create author profile page template
- [ ] Update cache key builder to include author modification times

### Template Variables
```go
data["Author"] = authorProfile      // Single author
data["Authors"] = []authorProfiles  // Multiple authors
data["Post"]["Author"] = author     // Author on post
```

---

## Phase 2: Content Discovery

**Goal:** Make content discoverable through RSS, sitemap, and search

### RSS/Atom Feeds

#### Routes
- `/rss` or `/feed` - Main blog RSS feed
- `/rss/{tag}` - Tag-specific RSS feed
- `/author/{slug}/rss` - Author-specific RSS feed

#### Implementation Tasks
- [ ] Create `internal/feed/rss.go` with RSS 2.0 generator
- [ ] Create `internal/feed/atom.go` with Atom feed generator
- [ ] Add RSS handler in `cmd/web/handlers.go`
- [ ] Support pagination in feeds (last N posts)
- [ ] Add feed auto-discovery links in HTML `<head>`
- [ ] Cache RSS feeds with appropriate TTL
- [ ] Add RSS link to templates/navigation

#### Feed Configuration
```bash
# Environment variables
FEED_TITLE="My Blog"
FEED_DESCRIPTION="Thoughts on technology and software"
FEED_LINK="https://example.com"
FEED_AUTHOR="John Doe"
FEED_AUTHOR_EMAIL="john@example.com"
FEED_ITEMS_COUNT=20  # Number of items in feed
```

### Sitemap.xml

#### Routes
- `/sitemap.xml` - Main sitemap
- `/sitemap-posts.xml` - Posts sitemap (if large site)
- `/sitemap-pages.xml` - Pages sitemap (if large site)

#### Implementation Tasks
- [ ] Create `internal/sitemap/sitemap.go`
- [ ] Add sitemap handler
- [ ] Include all posts, pages, author pages, tag pages
- [ ] Add `lastmod`, `changefreq`, `priority` fields
- [ ] Cache sitemap with file watching invalidation
- [ ] Add robots.txt with sitemap reference

### Search

#### Implementation Options
**Option A: Simple In-Memory Search**
- [ ] Create `internal/search/memory.go`
- [ ] Index posts on startup (title, content, tags)
- [ ] Add `/search` route with query parameter `?q=`
- [ ] Simple substring matching with ranking
- [ ] Cache search index, rebuild on content changes

**Option B: Full-Text Search (Future)**
- Consider integrating Bleve (Go full-text search)
- Or external service like Algolia/Meilisearch

#### Implementation Tasks (Option A)
- [ ] Create search index structure
- [ ] Implement tokenization and indexing
- [ ] Add search handler with pagination
- [ ] Create search results template
- [ ] Add search box to theme templates
- [ ] Return results as JSON for AJAX searches

---

## Phase 3: SEO & Social

**Goal:** Comprehensive SEO and social media optimization

### Meta Tags in Frontmatter
```yaml
---
title: "Post Title"
description: "SEO description"  # Already exists
cover: "/images/cover.jpg"      # Already exists
# New fields:
featured_image: "/images/social.jpg"  # Separate social image
og_title: "Custom OG Title"           # Override for Open Graph
og_description: "Custom OG desc"      # Override for Open Graph
twitter_title: "Custom Twitter Title"
canonical_url: "https://example.com/original"  # For syndicated content
seo_title: "Custom SEO Title | Site Name"
---
```

### Implementation Tasks

#### Open Graph Tags
- [ ] Add `internal/seo/opengraph.go`
- [ ] Generate OG tags: `og:title`, `og:description`, `og:image`, `og:url`, `og:type`, `og:site_name`
- [ ] Add OG tags to base template `<head>`
- [ ] Support article-specific tags: `article:published_time`, `article:author`, `article:tag`

#### Twitter Cards
- [ ] Add `internal/seo/twitter.go`
- [ ] Generate Twitter card tags: `twitter:card`, `twitter:title`, `twitter:description`, `twitter:image`
- [ ] Support `twitter:creator` and `twitter:site`
- [ ] Add Twitter card tags to base template

#### Structured Data (JSON-LD)
- [ ] Add `internal/seo/jsonld.go`
- [ ] Generate Article schema for blog posts
- [ ] Generate BlogPosting schema
- [ ] Generate Person schema for authors
- [ ] Generate WebSite schema for home page
- [ ] Generate BreadcrumbList for navigation
- [ ] Inject JSON-LD in template `<head>`

#### SEO Enhancements
- [ ] Auto-generate meta description from content if not provided
- [ ] Add canonical URL support
- [ ] Generate `robots` meta tag based on draft status
- [ ] Add `<link rel="canonical">` tags
- [ ] Support `noindex` for draft posts

#### Configuration
```bash
SITE_NAME="My Blog"
SITE_DESCRIPTION="A blog about technology"
SITE_URL="https://example.com"
TWITTER_HANDLE="@myblog"
FACEBOOK_APP_ID="123456789"
DEFAULT_SHARE_IMAGE="/images/default-share.jpg"
```

---

## Phase 4: Content Organization

**Goal:** Better content navigation through tags, pagination, and archives

### Tag Pages

#### Routes
- `/tag/{slug}` - Tag archive (first page)
- `/tag/{slug}/page/{number}` - Paginated tag archive

#### Implementation Tasks
- [ ] Add tag archive handler
- [ ] Load all posts with specific tag
- [ ] Sort by date (newest first)
- [ ] Implement pagination
- [ ] Create tag archive template
- [ ] Add tag cloud/list to sidebar or footer
- [ ] Update cache to invalidate tag pages when posts change

### Pagination

#### Configuration
```bash
POSTS_PER_PAGE=10
PAGINATION_CONTEXT=2  # Show 2 pages before/after current
```

#### Implementation Tasks
- [ ] Create `internal/pagination/paginator.go`
- [ ] Add pagination helper functions
- [ ] Support prev/next navigation
- [ ] Support page number links with context
- [ ] Add pagination to blog index
- [ ] Add pagination to tag archives
- [ ] Add pagination to author archives
- [ ] Add pagination to search results
- [ ] Update cache keys to include page number

#### Template Variables
```go
data["Pagination"] = Pagination{
    CurrentPage: 2,
    TotalPages: 10,
    PerPage: 10,
    TotalItems: 95,
    HasPrev: true,
    HasNext: true,
    PrevPage: 1,
    NextPage: 3,
    Pages: []PageLink{...},  // Page numbers to show
}
```

### Date-Based Archives

#### Routes
- `/archive/{year}` - Posts from year
- `/archive/{year}/{month}` - Posts from month
- `/archive/{year}/{month}/{day}` - Posts from day

#### Implementation Tasks
- [ ] Add date archive handlers
- [ ] Group posts by date
- [ ] Create archive templates
- [ ] Add archive navigation/calendar widget
- [ ] Support pagination in archives

### Navigation Menus

#### Configuration File
Create `data/navigation.yaml`:
```yaml
primary:
  - label: "Home"
    url: "/"
  - label: "Blog"
    url: "/blog"
  - label: "About"
    url: "/about"
  - label: "Contact"
    url: "/contact"

secondary:
  - label: "Privacy"
    url: "/privacy"
  - label: "Terms"
    url: "/terms"

social:
  - label: "Twitter"
    url: "https://twitter.com/example"
    icon: "twitter"
  - label: "GitHub"
    url: "https://github.com/example"
    icon: "github"
```

#### Implementation Tasks
- [ ] Create `internal/navigation/navigation.go`
- [ ] Parse navigation.yaml on startup
- [ ] Add navigation to template data
- [ ] Support nested menus
- [ ] Add active menu item detection
- [ ] Watch navigation.yaml for changes

---

## Phase 5: Enhanced Features

**Goal:** Content enhancements like reading time, excerpts, featured posts

### Reading Time Estimation

#### Implementation Tasks
- [ ] Create `internal/content/readingtime.go`
- [ ] Calculate reading time from word count (average 200-250 WPM)
- [ ] Add `ReadingTime` field to `Content` struct
- [ ] Display in blog post template ("5 min read")
- [ ] Add to RSS feed items

### Automatic Excerpts

#### Frontmatter
```yaml
---
title: "Post Title"
excerpt: "Custom excerpt text"  # Manual override
---
```

#### Implementation Tasks
- [ ] Auto-generate excerpt from first N words/characters if not provided
- [ ] Strip HTML tags from excerpt
- [ ] Smart truncation (don't break words)
- [ ] Add `<!--more-->` tag support for custom excerpt cutoff
- [ ] Add `Excerpt` field to `Content` struct
- [ ] Use in blog index, RSS feeds, meta descriptions

### Featured Posts

#### Frontmatter
```yaml
---
title: "Post Title"
featured: true  # Mark as featured
featured_order: 1  # Optional: control order of featured posts
---
```

#### Implementation Tasks
- [ ] Add `Featured` field to `Frontmatter` struct
- [ ] Filter featured posts in blog index
- [ ] Add featured posts section to home page
- [ ] Add featured badge/indicator in templates
- [ ] Support featured posts in specific positions

### Related Posts

#### Implementation Tasks
- [ ] Create `internal/content/related.go`
- [ ] Find related posts by shared tags
- [ ] Find related posts by same author
- [ ] Rank by relevance (more shared tags = higher rank)
- [ ] Limit to N most related posts (configurable)
- [ ] Add to blog post template
- [ ] Cache related posts

---

## Phase 6: API & Admin

**Goal:** Provide programmatic access and basic content management UI

### Content API

#### Routes (RESTful)
- `GET /api/posts` - List posts (with pagination, filtering)
- `GET /api/posts/{slug}` - Get single post
- `GET /api/pages/{slug}` - Get single page
- `GET /api/tags` - List all tags
- `GET /api/tags/{slug}` - Get tag with posts
- `GET /api/authors` - List authors
- `GET /api/authors/{slug}` - Get author with posts
- `GET /api/navigation` - Get navigation menus
- `GET /api/settings` - Get site settings

#### Query Parameters
```
?page=2&per_page=10          # Pagination
?tag=golang                   # Filter by tag
?author=john-doe              # Filter by author
?featured=true                # Only featured posts
?include=author,tags          # Include relations
?fields=title,slug,excerpt    # Select specific fields
?order=published_at desc      # Custom ordering
```

#### Implementation Tasks
- [ ] Create `internal/api/posts.go`
- [ ] Create `internal/api/pages.go`
- [ ] Create `internal/api/tags.go`
- [ ] Create `internal/api/authors.go`
- [ ] Add API routes with `/api` prefix
- [ ] Implement JSON serialization
- [ ] Add CORS support (configurable)
- [ ] Add rate limiting
- [ ] Support ETag/conditional requests
- [ ] Cache API responses
- [ ] Add API documentation endpoint (`/api/docs`)

#### API Response Format
```json
{
  "posts": [
    {
      "id": "hello-world",
      "slug": "hello-world",
      "title": "Hello World",
      "excerpt": "Welcome to my blog...",
      "html": "<h1>Welcome</h1>...",
      "published_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-16T12:00:00Z",
      "reading_time": 5,
      "featured": false,
      "author": {
        "slug": "john-doe",
        "name": "John Doe"
      },
      "tags": [
        {"slug": "golang", "name": "Golang"}
      ],
      "url": "https://example.com/blog/hello-world"
    }
  ],
  "meta": {
    "pagination": {
      "page": 1,
      "per_page": 10,
      "total": 95,
      "pages": 10,
      "prev": null,
      "next": 2
    }
  }
}
```

### Admin Dashboard (Basic)

**Note:** This is a significant undertaking. Consider these options:

**Option A: File-Based Admin**
- Simple web UI to browse files
- Edit markdown in browser
- Upload images
- No database needed

**Option B: Integration with External Tools**
- Use Forestry.io, Netlify CMS, or Decap CMS
- Git-based workflow
- No backend changes needed

**Option C: Custom Admin (Future Phase)**
- Full-featured admin panel
- User authentication
- WYSIWYG editor
- Media library
- Post scheduling UI

#### Implementation Tasks (Option A - Basic)
- [ ] Create admin routes under `/admin`
- [ ] Add basic authentication (configurable username/password)
- [ ] List posts and pages
- [ ] Edit markdown files in browser
- [ ] Upload images to `data/attachments`
- [ ] Simple markdown editor with preview
- [ ] Create new posts/pages
- [ ] Delete posts/pages
- [ ] Invalidate cache on edits

---

## Phase 7: Advanced Features

**Goal:** Comments, newsletters, analytics, and post scheduling

### Comments Integration

#### Implementation Options
**Option A: Third-Party Services**
- Disqus
- Commento
- utterances (GitHub issues)
- giscus (GitHub discussions)

**Option B: Self-Hosted (Future)**
- Custom comment system with database
- Moderation queue
- Email notifications

#### Implementation Tasks (Option A)
- [ ] Add comment configuration to frontmatter
```yaml
---
comments: true  # Enable/disable per post
comments_locked: false  # Lock comments
---
```
- [ ] Add comments section to blog post template
- [ ] Support multiple comment providers (configurable)
- [ ] Add environment variables for comment service config

### Newsletter/Subscriptions

#### Implementation Options
**Option A: Third-Party Integration**
- Mailchimp
- ConvertKit
- Buttondown
- Substack embed

**Option B: Self-Hosted (Future)**
- Email collection with database
- Email sending via SMTP or service (SendGrid, SES)
- Subscription management
- Newsletter scheduling

#### Implementation Tasks (Option A)
- [ ] Add newsletter signup form to templates
- [ ] Add subscription form partial
- [ ] Support embed codes for third-party services
- [ ] Add subscribe call-to-action on blog posts

### Analytics Integration

#### Implementation Tasks
- [ ] Add analytics configuration
```bash
GOOGLE_ANALYTICS_ID="G-XXXXXXXXXX"
PLAUSIBLE_DOMAIN="example.com"
FATHOM_SITE_ID="XXXXX"
SIMPLE_ANALYTICS_DOMAIN="example.com"
```
- [ ] Inject analytics scripts in base template
- [ ] Support multiple analytics providers
- [ ] Add privacy-friendly options (Plausible, Fathom)
- [ ] Support custom analytics code injection

### Post Scheduling

#### Frontmatter
```yaml
---
title: "Scheduled Post"
date: 2024-01-15
published_at: 2024-02-01T10:00:00Z  # Future date = scheduled
status: "scheduled"  # draft, scheduled, published
---
```

#### Implementation Tasks
- [ ] Add `PublishedAt` and `Status` fields to `Frontmatter`
- [ ] Add background job to check scheduled posts
- [ ] Only show published posts (where `published_at <= now`)
- [ ] Add status indicator in admin UI
- [ ] Invalidate cache when scheduled post becomes published
- [ ] Support timezone configuration

### Image Optimization

#### Implementation Tasks
- [ ] Automatic image resizing on upload
- [ ] Generate multiple sizes (thumbnail, small, medium, large)
- [ ] WebP conversion
- [ ] Lazy loading support in templates
- [ ] Responsive image srcset generation
- [ ] Image CDN integration (optional)

### Internationalization (i18n)

#### Implementation Tasks (Future Phase)
- [ ] Multi-language content support
- [ ] Language-specific URLs
- [ ] Translation files for UI
- [ ] Language switcher
- [ ] hreflang tags for SEO

---

## Implementation Priority

### High Priority (Core Ghost Features)
1. **Phase 2: Content Discovery** - RSS and sitemap are essential
2. **Phase 3: SEO & Social** - Critical for discoverability
3. **Phase 4: Pagination & Tags** - Essential for usability
4. **Phase 5: Reading Time & Excerpts** - Improves UX significantly

### Medium Priority
5. **Phase 1: Author System** - Important for multi-author blogs
6. **Phase 5: Featured Posts & Related Posts** - Nice content features
7. **Phase 6: Content API** - Enables headless usage

### Lower Priority (Can be added later)
8. **Phase 4: Date Archives** - Nice to have
9. **Phase 7: Comments** - Can use third-party initially
10. **Phase 7: Newsletter** - Can use third-party initially
11. **Phase 6: Admin Dashboard** - Complex, consider third-party tools

### Future Considerations
- Database support (SQLite, PostgreSQL) for better performance at scale
- Full-text search with Bleve or external service
- Custom admin panel
- Membership/subscription system (Ghost's premium feature)
- Email newsletter sending
- Webhooks
- Custom integrations
- Multi-language support

---

## Configuration Management

As features grow, consider moving from environment variables to a configuration file:

`config.yaml`:
```yaml
site:
  name: "My Blog"
  description: "Thoughts on technology"
  url: "https://example.com"
  timezone: "America/New_York"

content:
  posts_per_page: 10
  excerpt_length: 200
  reading_speed_wpm: 250

features:
  rss: true
  sitemap: true
  search: true
  comments:
    enabled: true
    provider: "disqus"  # disqus, commento, utterances
    disqus_shortname: "my-blog"
  analytics:
    provider: "plausible"  # google, plausible, fathom, simple
    plausible_domain: "example.com"

cache:
  enabled: true
  ttl: 3600
  max_entries: 1000
  max_size_mb: 100

theme:
  active: "default"
```

---

## Testing Strategy

For each phase, implement:
- [ ] Unit tests for new functions
- [ ] Integration tests for handlers
- [ ] Template rendering tests
- [ ] Cache invalidation tests
- [ ] Performance benchmarks
- [ ] E2E tests for critical paths

---

## Migration Path

When implementing these features, maintain backward compatibility:
1. All new frontmatter fields should be optional
2. Existing content should work without changes
3. New routes should not conflict with existing content slugs
4. Configuration should have sensible defaults
5. Cache should auto-invalidate with structure changes

---

## Documentation Updates

For each implemented phase:
- [ ] Update README.md with new features
- [ ] Update CLAUDE.md with architectural changes
- [ ] Add examples to documentation
- [ ] Update API documentation
- [ ] Create migration guides if needed
