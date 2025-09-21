# VellumForge Cache System

VellumForge includes a comprehensive in-memory cache system with LRU+TTL capabilities designed to optimize the performance of static site generation and content delivery.

## Features

### Core Cache Functionality
- **LRU (Least Recently Used) Eviction**: Automatically removes least recently used entries when cache reaches capacity
- **TTL (Time To Live)**: Configurable expiration time for cache entries
- **Size Management**: Configurable limits on both number of entries and total memory usage
- **Automatic Cleanup**: Background janitor process removes expired entries

### Cache Key Generation
Cache keys are generated based on multiple factors to ensure proper cache invalidation:

- HTTP method (GET, POST, etc.)
- Normalized request path
- Absolute file path of content files
- File modification time (nanosecond precision)
- Template name being used
- Theme ID
- Feature flags (e.g., Mermaid diagram support)
- Accept-Encoding header (for compression variants)

### Cache Bypass Support
- **Query Parameter**: Add `?nocache=1` to any URL to bypass cache
- **HTTP Header**: Send `X-Bypass-Cache: 1` header to bypass cache
- Useful for testing and debugging

### Conditional Requests (304 Not Modified)
- Supports `If-None-Match` (ETag-based)
- Supports `If-Modified-Since` (timestamp-based)
- Automatically returns 304 responses when content hasn't changed
- Reduces bandwidth usage and improves client performance

### Auto-invalidation
- **File System Monitoring**: Watches content and theme directories for changes
- **Smart Invalidation**: 
  - Blog post changes invalidate specific post + blog index + home page
  - Theme/template changes clear entire cache
  - Page changes invalidate specific page only
- **Configurable Polling**: File system polling with configurable interval (default: 2 seconds)

## Configuration

### Environment Variables

```bash
# Cache enabled/disabled (default: true)
CACHE_ENABLED=true

# Cache TTL in seconds (default: 3600 = 1 hour)
CACHE_TTL=3600

# Maximum number of cache entries (default: 1000)
CACHE_MAX_ENTRIES=1000

# Maximum cache size in MB (default: 100)
CACHE_MAX_SIZE_MB=100
```

### Cache Configuration in Code

```go
cacheConfig := cache.Config{
    MaxEntries:   1000,
    MaxSizeBytes: 100 * 1024 * 1024, // 100MB
    DefaultTTL:   time.Hour,
    CleanupFreq:  5 * time.Minute,
}
```

## API Endpoints

### Cache Statistics
```
GET /cache/stats
```

Returns JSON with cache statistics:
```json
{
    "enabled": true,
    "entries": 42,
    "sizeBytes": 1048576,
    "sizeMB": 1.00,
    "maxEntries": 1000,
    "maxSizeBytes": 104857600,
    "maxSizeMB": 100.00,
    "utilization": 4.20
}
```

### Manual Cache Clearing
```
POST /cache/clear
```

Clears the entire cache manually:
```json
{
    "success": true,
    "message": "cache cleared"
}
```

## Cache Entry Structure

Each cache entry stores:
- **Body**: Rendered response body (bytes)
- **Headers**: HTTP response headers
- **Status Code**: HTTP status code
- **Content Type**: MIME type
- **ETag**: Entity tag for conditional requests
- **Last Modified**: Last modification timestamp
- **Size**: Entry size in bytes
- **Creation Time**: When entry was created
- **Expiration Time**: When entry expires
- **Access Statistics**: Last access time and access count

## Implementation Details

### LRU Algorithm
The cache uses a doubly-linked list combined with a hash map for O(1) access and O(1) LRU operations:

```go
type lruList struct {
    head  *lruNode
    tail  *lruNode
    nodes map[string]*lruNode
}
```

### Memory Management
- **Entry Size Calculation**: Includes body, headers, and metadata overhead
- **Total Size Tracking**: Real-time tracking of total cache memory usage
- **Eviction Policy**: Size-based and count-based eviction when limits exceeded

### File Watching
- **Polling-based**: Uses file system polling instead of OS events for reliability
- **Path Monitoring**: Watches both content and theme directories
- **Change Detection**: Compares file modification times to detect changes

### Response Capture
The cache middleware captures responses before they're sent to clients:

```go
type ResponseCapture struct {
    http.ResponseWriter
    statusCode int
    body       *bytes.Buffer
    headers    http.Header
}
```

## Usage Examples

### Cache Bypass for Testing
```bash
# Bypass cache with query parameter
curl "http://localhost:6886/blog/my-post?nocache=1"

# Bypass cache with header
curl -H "X-Bypass-Cache: 1" "http://localhost:6886/blog/my-post"
```

### Conditional Requests
```bash
# First request gets full response with ETag
curl -v "http://localhost:6886/blog/my-post"

# Subsequent request with If-None-Match returns 304 if unchanged
curl -H 'If-None-Match: "abc123"' "http://localhost:6886/blog/my-post"
```

### Manual Cache Management
```bash
# Get cache statistics
curl "http://localhost:6886/cache/stats"

# Clear cache
curl -X POST "http://localhost:6886/cache/clear"
```

## Performance Benefits

1. **Reduced Template Rendering**: Cached responses eliminate repeated template rendering
2. **Faster File I/O**: Avoid repeated file system access for content files
3. **Improved Response Times**: Serve cached content in microseconds vs milliseconds
4. **Bandwidth Optimization**: 304 responses reduce network traffic
5. **CPU Usage Reduction**: Less processing for cached requests

## Cache Invalidation Strategies

### Automatic Invalidation
- **Content Changes**: File modification triggers specific cache invalidation
- **Theme Updates**: Template/theme changes clear entire cache
- **Time-based**: TTL ensures content freshness

### Manual Invalidation
- **API Endpoint**: `/cache/clear` for full cache clearing
- **Programmatic**: `cacheInvalidator.InvalidateByPath()` for specific patterns

### Smart Invalidation
- **Blog Posts**: Invalidates post + index + home page
- **Pages**: Invalidates specific page only
- **Global Changes**: Theme modifications clear all cached content

## Monitoring and Debugging

### Logging
The cache system provides detailed logging:
- Cache hits/misses
- Invalidation events
- File system changes
- Performance metrics

### Statistics
Real-time statistics available via API:
- Entry count and size
- Hit/miss ratios
- Memory utilization
- Eviction events

### Debug Headers
When caching is active, responses include debug headers:
- `ETag`: Entity tag for conditional requests
- `Last-Modified`: Content modification time
- `Cache-Control`: Caching directives
- `Vary`: Response variation indicators
