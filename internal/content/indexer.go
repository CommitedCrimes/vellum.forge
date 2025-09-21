package content

import (
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ContentMeta holds lightweight metadata for a content file
type ContentMeta struct {
	Path    string
	Slug    string
	Title   string
	Date    time.Time
	Draft   bool
	ModTime time.Time
	Size    int64
	IsBlog  bool
}

// Changes describes index changes since last update
type Changes struct {
	AddedOrUpdated []*ContentMeta
	Removed        []*ContentMeta
}

// IndexerOptions configures the Indexer behaviour
type IndexerOptions struct {
	Debounce       time.Duration
	RescanInterval time.Duration
}

// Indexer maintains slug -> meta maps for pages and blog, plus a sorted blog list
type Indexer struct {
	dataDir string
	logger  *slog.Logger

	parser *MarkdownParser

	mu           sync.RWMutex
	pagesBySlug  map[string]*ContentMeta
	postsBySlug  map[string]*ContentMeta
	blogSorted   []*ContentMeta // sorted by Date desc
	fileSnapshot map[string]os.FileInfo

	options  IndexerOptions
	stopCh   chan struct{}
	wg       sync.WaitGroup
	onChange func(Changes)
}

// NewIndexer creates a new Indexer instance
func NewIndexer(dataDir string, logger *slog.Logger, opts IndexerOptions) *Indexer {
	if opts.Debounce <= 0 {
		opts.Debounce = 250 * time.Millisecond
	}
	if opts.RescanInterval <= 0 {
		opts.RescanInterval = 15 * time.Second
	}

	return &Indexer{
		dataDir:      dataDir,
		logger:       logger,
		parser:       NewMarkdownParser(),
		pagesBySlug:  make(map[string]*ContentMeta),
		postsBySlug:  make(map[string]*ContentMeta),
		blogSorted:   make([]*ContentMeta, 0),
		fileSnapshot: make(map[string]os.FileInfo),
		options:      opts,
		stopCh:       make(chan struct{}),
	}
}

// SetOnChange registers a callback invoked after coalesced updates are applied
func (ix *Indexer) SetOnChange(cb func(Changes)) {
	ix.onChange = cb
}

// Start begins watching and periodic rescans
func (ix *Indexer) Start() {
	// Initial build
	ix.Rescan()

	ix.wg.Add(1)
	go ix.watchLoop()
}

// Stop terminates the indexer
func (ix *Indexer) Stop() {
	close(ix.stopCh)
	ix.wg.Wait()
}

// GetBlogList returns a copy of the current sorted blog list (frontmatter-only)
func (ix *Indexer) GetBlogList() []*Content {
	ix.mu.RLock()
	defer ix.mu.RUnlock()

	list := make([]*Content, 0, len(ix.blogSorted))
	for _, meta := range ix.blogSorted {
		fm := Frontmatter{Title: meta.Title, Date: meta.Date, Draft: meta.Draft, Slug: meta.Slug}
		list = append(list, &Content{Frontmatter: fm})
	}
	return list
}

// GetBlogPathBySlug returns absolute path for a blog post
func (ix *Indexer) GetBlogPathBySlug(slug string) (string, bool) {
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	if meta, ok := ix.postsBySlug[slug]; ok {
		return meta.Path, true
	}
	return "", false
}

// GetPagePathBySlug returns absolute path for a page
func (ix *Indexer) GetPagePathBySlug(slug string) (string, bool) {
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	if meta, ok := ix.pagesBySlug[slug]; ok {
		return meta.Path, true
	}
	return "", false
}

// Rescan rebuilds the index from disk
func (ix *Indexer) Rescan() {
	changes := ix.rebuildFromDisk()
	if ix.onChange != nil && (len(changes.AddedOrUpdated) > 0 || len(changes.Removed) > 0) {
		ix.onChange(changes)
	}
}

func (ix *Indexer) watchLoop() {
	defer ix.wg.Done()

	debounceTimer := time.NewTimer(ix.options.Debounce)
	if !debounceTimer.Stop() {
		<-debounceTimer.C
	}

	var pending bool
	rescanTicker := time.NewTicker(ix.options.RescanInterval)
	defer rescanTicker.Stop()

	for {
		select {
		case <-time.After(1 * time.Second):
			if ix.detectFileSystemChange() {
				pending = true
				resetTimer(debounceTimer, ix.options.Debounce)
			}
		case <-debounceTimer.C:
			if pending {
				pending = false
				ix.Rescan()
			}
		case <-rescanTicker.C:
			ix.Rescan()
		case <-ix.stopCh:
			return
		}
	}
}

func resetTimer(t *time.Timer, d time.Duration) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
	t.Reset(d)
}

// detectFileSystemChange checks for any path add/remove/mtime change compared to snapshot
func (ix *Indexer) detectFileSystemChange() bool {
	current := ix.scanFiles()

	ix.mu.RLock()
	prev := ix.fileSnapshot
	ix.mu.RUnlock()

	// Compare keys and mod times/sizes
	if len(current) != len(prev) {
		ix.mu.Lock()
		ix.fileSnapshot = current
		ix.mu.Unlock()
		return true
	}

	for p, info := range current {
		if old, ok := prev[p]; !ok || !info.ModTime().Equal(old.ModTime()) || info.Size() != old.Size() {
			ix.mu.Lock()
			ix.fileSnapshot = current
			ix.mu.Unlock()
			return true
		}
	}

	return false
}

func (ix *Indexer) rebuildFromDisk() Changes {
	files := ix.scanFiles()

	// Build new maps
	newPages := make(map[string]*ContentMeta)
	newPosts := make(map[string]*ContentMeta)
	var newBlogList []*ContentMeta

	for path, info := range files {
		isBlog := strings.Contains(path, string(filepath.Separator)+"blog"+string(filepath.Separator))
		isPage := strings.Contains(path, string(filepath.Separator)+"pages"+string(filepath.Separator))
		if !isBlog && !isPage {
			continue
		}

		// Read file for frontmatter-only parsing
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		fm, err := ix.parser.ParseFrontmatterOnly(data)
		if err != nil {
			continue
		}

		// Derive slug
		slug := fm.Slug
		if slug == "" {
			base := filepath.Base(path)
			ext := filepath.Ext(base)
			slug = strings.TrimSuffix(base, ext)
		}

		meta := &ContentMeta{
			Path:    path,
			Slug:    slug,
			Title:   fm.Title,
			Date:    fm.Date,
			Draft:   fm.Draft,
			ModTime: info.ModTime(),
			Size:    info.Size(),
			IsBlog:  isBlog,
		}

		if meta.Draft {
			continue
		}

		if isBlog {
			newPosts[slug] = meta
			newBlogList = append(newBlogList, meta)
		} else if isPage {
			newPages[slug] = meta
		}
	}

	// Sort blog list by date desc; fallback to mod time if date is zero
	sort.Slice(newBlogList, func(i, j int) bool {
		di := newBlogList[i].Date
		dj := newBlogList[j].Date
		if di.IsZero() {
			di = newBlogList[i].ModTime
		}
		if dj.IsZero() {
			dj = newBlogList[j].ModTime
		}
		return di.After(dj)
	})

	// Compute changes vs previous
	ix.mu.Lock()
	oldPages := ix.pagesBySlug
	oldPosts := ix.postsBySlug
	oldFiles := ix.fileSnapshot
	ix.pagesBySlug = newPages
	ix.postsBySlug = newPosts
	ix.blogSorted = newBlogList
	ix.fileSnapshot = files
	ix.mu.Unlock()

	var ch Changes

	// Added or updated posts
	for slug, meta := range newPosts {
		if old, ok := oldPosts[slug]; !ok || !meta.ModTime.Equal(old.ModTime) || meta.Size != old.Size {
			ch.AddedOrUpdated = append(ch.AddedOrUpdated, meta)
		}
	}
	// Added or updated pages
	for slug, meta := range newPages {
		if old, ok := oldPages[slug]; !ok || !meta.ModTime.Equal(old.ModTime) || meta.Size != old.Size {
			ch.AddedOrUpdated = append(ch.AddedOrUpdated, meta)
		}
	}

	// Removed posts
	for slug, old := range oldPosts {
		if _, ok := newPosts[slug]; !ok {
			ch.Removed = append(ch.Removed, old)
		}
	}
	// Removed pages
	for slug, old := range oldPages {
		if _, ok := newPages[slug]; !ok {
			ch.Removed = append(ch.Removed, old)
		}
	}

	// Detect file-level changes not mapped (rename changes path)
	_ = oldFiles

	return ch
}

// scanFiles walks data/pages and data/blog and returns a snapshot of .md files
func (ix *Indexer) scanFiles() map[string]os.FileInfo {
	result := make(map[string]os.FileInfo)
	root := ix.dataDir

	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		lower := strings.ToLower(path)
		if !strings.HasSuffix(lower, ".md") {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		// Only include pages/ and blog/
		if strings.Contains(path, string(filepath.Separator)+"pages"+string(filepath.Separator)) ||
			strings.Contains(path, string(filepath.Separator)+"blog"+string(filepath.Separator)) {
			abs, err := filepath.Abs(path)
			if err != nil {
				abs = path
			}
			result[abs] = info
		}
		return nil
	})

	return result
}
