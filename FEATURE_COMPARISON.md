# VellumForge vs Ghost CMS - Feature Comparison

## Legend
- ✅ Implemented
- 🔶 Partially implemented
- ⏳ Planned (see GHOST_FEATURE_PLAN.md)
- ❌ Not planned
- 🔧 Needs third-party integration

---

## Content Management

| Feature | VellumForge | Ghost | Notes |
|---------|-------------|-------|-------|
| Markdown posts | ✅ | ✅ | VellumForge uses file-based markdown |
| WYSIWYG editor | ❌ | ✅ | Ghost has Koenig editor; VellumForge is file-based |
| Draft posts | ✅ | ✅ | Via frontmatter `draft: true` |
| Scheduled posts | ⏳ | ✅ | Planned with `published_at` frontmatter |
| Static pages | ✅ | ✅ | Both support pages |
| Tags | ✅ | ✅ | VellumForge has basic tags |
| Categories | ❌ | ❌ | Neither has categories (tags serve this purpose) |
| Featured posts | ⏳ | ✅ | Planned with `featured: true` frontmatter |
| Custom excerpts | ⏳ | ✅ | Planned with auto-generation |
| Post settings | 🔶 | ✅ | VellumForge via frontmatter; Ghost has UI |

---

## Authors & Team

| Feature | VellumForge | Ghost | Notes |
|---------|-------------|-------|-------|
| Single author | 🔶 | ✅ | Currently no author system |
| Multiple authors | ⏳ | ✅ | Planned with author profiles |
| Author profiles | ⏳ | ✅ | Planned with markdown files |
| Author pages | ⏳ | ✅ | Planned `/author/{slug}` |
| Team management | ❌ | ✅ | Ghost has user roles; VellumForge is file-based |
| User roles | ❌ | ✅ | Not applicable for file-based system |

---

## Content Organization

| Feature | VellumForge | Ghost | Notes |
|---------|-------------|-------|-------|
| Blog index | ✅ | ✅ | Both have blog listing |
| Pagination | ⏳ | ✅ | Planned for blog, tags, archives |
| Tag pages | ⏳ | ✅ | Planned `/tag/{slug}` |
| Tag archives | ⏳ | ✅ | Planned with pagination |
| Date archives | ⏳ | 🔶 | Ghost doesn't have by default |
| Collections | ❌ | ✅ | Ghost has custom collections |
| Related posts | ⏳ | ✅ | Planned based on tags |
| Reading time | ⏳ | ✅ | Planned calculation |

---

## SEO & Discovery

| Feature | VellumForge | Ghost | Notes |
|---------|-------------|-------|-------|
| Meta descriptions | ✅ | ✅ | Via frontmatter `description` |
| Meta titles | 🔶 | ✅ | Uses page title currently |
| Open Graph tags | ⏳ | ✅ | Planned |
| Twitter Cards | ⏳ | ✅ | Planned |
| Structured data (JSON-LD) | ⏳ | ✅ | Planned |
| Sitemap.xml | ⏳ | ✅ | Planned |
| robots.txt | ⏳ | ✅ | Planned |
| RSS feed | ⏳ | ✅ | Planned |
| Canonical URLs | ⏳ | ✅ | Planned |
| AMP support | ❌ | ✅ | Not planned |

---

## Theming

| Feature | VellumForge | Ghost | Notes |
|---------|-------------|-------|-------|
| Themes | ✅ | ✅ | Both have theme support |
| Theme fallback | ✅ | ❌ | VellumForge unique feature |
| Template engine | ✅ (Jet) | ✅ (Handlebars) | Different engines |
| Custom templates | ✅ | ✅ | Both support custom templates |
| Theme marketplace | ❌ | ✅ | Ghost has marketplace |
| Dynamic routing | ❌ | ✅ | Ghost has routes.yaml |
| Custom template functions | ✅ | ✅ | VellumForge has several built-in |

---

## Media & Assets

| Feature | VellumForge | Ghost | Notes |
|---------|-------------|-------|-------|
| Image uploads | 🔶 | ✅ | VellumForge manual upload to data/attachments |
| Media library | ❌ | ✅ | Ghost has full media management |
| Image optimization | ⏳ | ✅ | Planned for VellumForge |
| Responsive images | ⏳ | ✅ | Planned with srcset |
| CDN integration | ⏳ | ✅ | Planned for VellumForge |
| Unsplash integration | ❌ | ✅ | Ghost feature |

---

## Performance

| Feature | VellumForge | Ghost | Notes |
|---------|-------------|-------|-------|
| Caching | ✅ | ✅ | VellumForge has advanced LRU+TTL cache |
| File watching | ✅ | ❌ | VellumForge auto-invalidates cache |
| Cache invalidation | ✅ | 🔶 | VellumForge has smart invalidation |
| ETag support | ✅ | ✅ | Both support conditional requests |
| 304 responses | ✅ | ✅ | Both support Not Modified |
| Compression | ✅ | ✅ | VellumForge via chi middleware |
| Database | ❌ | ✅ | Ghost uses database; VellumForge is file-based |

---

## API & Integrations

| Feature | VellumForge | Ghost | Notes |
|---------|-------------|-------|-------|
| Content API | ⏳ | ✅ | Planned JSON API |
| Admin API | ❌ | ✅ | Ghost has full admin API |
| Webhooks | ❌ | ✅ | Ghost feature |
| Custom integrations | ❌ | ✅ | Ghost has integration directory |
| REST API | ⏳ | ✅ | Planned for VellumForge |
| GraphQL | ❌ | ❌ | Neither has GraphQL |

---

## Admin & Management

| Feature | VellumForge | Ghost | Notes |
|---------|-------------|-------|-------|
| Admin dashboard | ⏳ | ✅ | Planned basic version for VellumForge |
| Content editor | ❌ | ✅ | VellumForge is file-based (use external editor) |
| User management | ❌ | ✅ | Ghost has full user system |
| Markdown editor | ⏳ | ✅ | Planned browser-based editor |
| Media upload UI | ⏳ | ✅ | Planned for VellumForge |
| Settings UI | ❌ | ✅ | VellumForge uses env vars and files |
| Code injection | ⏳ | ✅ | Planned for analytics/scripts |

---

## Membership & Monetization

| Feature | VellumForge | Ghost | Notes |
|---------|-------------|-------|-------|
| Memberships | ❌ | ✅ | Ghost premium feature |
| Paid subscriptions | ❌ | ✅ | Ghost premium feature |
| Member tiers | ❌ | ✅ | Ghost feature |
| Email newsletters | 🔧 | ✅ | VellumForge can integrate third-party |
| Built-in email sending | ❌ | ✅ | Ghost has built-in |
| Stripe integration | ❌ | ✅ | Ghost feature |
| Member portal | ❌ | ✅ | Ghost feature |

---

## Email & Newsletter

| Feature | VellumForge | Ghost | Notes |
|---------|-------------|-------|-------|
| Email sending | ❌ | ✅ | Ghost has built-in |
| Newsletter subscriptions | 🔧 | ✅ | VellumForge via third-party (Mailchimp, etc.) |
| Email templates | ❌ | ✅ | Ghost feature |
| Bulk email | ❌ | ✅ | Ghost feature |
| Email analytics | ❌ | ✅ | Ghost feature |
| Segmentation | ❌ | ✅ | Ghost feature |

---

## Comments & Community

| Feature | VellumForge | Ghost | Notes |
|---------|-------------|-------|-------|
| Native comments | ❌ | ✅ | Ghost has built-in comments |
| Third-party comments | 🔧 | 🔧 | Both can integrate Disqus, etc. |
| Comment moderation | ❌ | ✅ | Ghost feature |
| Member comments | ❌ | ✅ | Ghost feature |

---

## Search

| Feature | VellumForge | Ghost | Notes |
|---------|-------------|-------|-------|
| Built-in search | ⏳ | ❌ | Planned in-memory search for VellumForge |
| Search API | ⏳ | ✅ | Ghost has search API |
| Third-party search | 🔧 | 🔧 | Both can integrate Algolia, etc. |

---

## Analytics & Insights

| Feature | VellumForge | Ghost | Notes |
|---------|-------------|-------|-------|
| Built-in analytics | ❌ | ✅ | Ghost has analytics dashboard |
| Post analytics | ❌ | ✅ | Ghost feature |
| Member analytics | ❌ | ✅ | Ghost feature |
| Third-party analytics | 🔧 | 🔧 | Both can integrate GA, Plausible, etc. |
| Cache stats | ✅ | ❌ | VellumForge has `/cache/stats` |

---

## Development

| Feature | VellumForge | Ghost | Notes |
|---------|-------------|-------|-------|
| Local development | ✅ | ✅ | Both support local dev |
| Live reload | ✅ | ✅ | VellumForge has `make run/live` |
| Theme development | ✅ | ✅ | Both support theme dev |
| Version control friendly | ✅ | 🔶 | VellumForge is entirely file-based (Git-friendly) |
| Database required | ❌ | ✅ | VellumForge advantage: no DB needed |
| Built with | Go | Node.js | Different tech stacks |
| Extensibility | 🔶 | ✅ | Ghost has more extension points |

---

## Deployment & Hosting

| Feature | VellumForge | Ghost | Notes |
|---------|-------------|-------|-------|
| Self-hosted | ✅ | ✅ | Both can be self-hosted |
| Official hosting | ❌ | ✅ | Ghost has Ghost(Pro) |
| Docker support | 🔶 | ✅ | VellumForge can be dockerized |
| Static export | ❌ | ❌ | Neither exports to static HTML |
| Single binary | ✅ | ❌ | VellumForge compiles to single binary |
| Easy updates | ✅ | 🔶 | VellumForge: replace binary; Ghost: npm update |

---

## Special Features

### VellumForge Unique Features
- ✅ **Advanced caching system** with LRU+TTL and smart invalidation
- ✅ **File watching** with automatic cache invalidation
- ✅ **No database required** - entirely file-based
- ✅ **Single binary deployment** - easy to deploy
- ✅ **Theme fallback** - override only what you need
- ✅ **Mermaid diagrams** built-in
- ✅ **Syntax highlighting** with line numbers
- ✅ **Git-friendly** - all content in version control

### Ghost Unique Features
- ✅ **Built-in membership system** with paid subscriptions
- ✅ **Email newsletter sending** with segmentation
- ✅ **Admin dashboard** with WYSIWYG editor
- ✅ **User management** with roles and permissions
- ✅ **Built-in analytics** and insights
- ✅ **Official hosting** (Ghost Pro)
- ✅ **Extensive API** for integrations
- ✅ **Member portal** and authentication

---

## Summary

### VellumForge Strengths
1. **Performance** - Advanced caching, no database overhead
2. **Simplicity** - File-based, single binary, no dependencies
3. **Developer-friendly** - Git-friendly, easy to version control
4. **Deployment** - Single binary, easy updates
5. **Cost** - No database server needed

### Ghost Strengths
1. **Feature-rich** - Memberships, newsletters, monetization
2. **User-friendly** - Full admin UI, WYSIWYG editor
3. **Ecosystem** - Large theme marketplace, integrations
4. **Managed hosting** - Ghost Pro available
5. **Team collaboration** - User roles and permissions

### Recommendation
- **Choose VellumForge if you want:**
  - Simple, fast blog with file-based content
  - Git-friendly workflow
  - No database complexity
  - Single developer or small team
  - Maximum performance with caching
  - Easy deployment and updates

- **Choose Ghost if you want:**
  - Full-featured CMS with admin UI
  - Membership and subscription features
  - Built-in email newsletters
  - Team collaboration with roles
  - Extensive third-party integrations
  - Managed hosting option

---

## Making VellumForge More Ghost-like

The **GHOST_FEATURE_PLAN.md** outlines how to bring VellumForge closer to Ghost's feature set while maintaining its file-based, simple architecture. The key is to focus on:

1. **Content features** (RSS, sitemap, SEO, pagination)
2. **Organization features** (tags, authors, archives)
3. **Discovery features** (search, related posts)
4. **API access** (Content API for headless usage)

For features that require significant infrastructure (memberships, email sending, user management), VellumForge can integrate with third-party services rather than building them in-house.
