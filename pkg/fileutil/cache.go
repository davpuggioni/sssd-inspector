package fileutil

import (
	"bufio"
	"container/list"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

// RegexCache provides thread-safe caching of compiled regular expressions.
type RegexCache struct {
	mu       sync.RWMutex
	patterns map[string]*regexp.Regexp
}

// NewRegexCache creates a new RegexCache
func NewRegexCache() *RegexCache {
	return &RegexCache{
		patterns: make(map[string]*regexp.Regexp),
	}
}

// Get returns a compiled regex for the given pattern, compiling it on first access.
// Returns nil if the pattern is invalid (no panic).
func (rc *RegexCache) Get(pattern string) *regexp.Regexp {
	rc.mu.RLock()
	re, ok := rc.patterns[pattern]
	rc.mu.RUnlock()
	if ok {
		return re
	}

	rc.mu.Lock()
	defer rc.mu.Unlock()

	// Double-check after acquiring write lock
	if re, ok := rc.patterns[pattern]; ok {
		return re
	}

	// Use Compile instead of MustCompile to avoid panic on invalid patterns
	re, err := regexp.Compile(pattern)
	if err != nil {
		// Log the error and return a non-nil regex that matches nothing or re-throw
		// We return nil; callers should check for nil
		rc.patterns[pattern] = nil
		return nil
	}
	rc.patterns[pattern] = re
	return re
}

// Clear removes all cached patterns
func (rc *RegexCache) Clear() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.patterns = make(map[string]*regexp.Regexp)
}

// lruEntry holds a cache entry with its key for LRU eviction
type lruEntry struct {
	key   string
	lines []string
}

// FileCache provides thread-safe caching of file contents (lines) with LRU eviction.
type FileCache struct {
	mu       sync.RWMutex
	entries  map[string]*list.Element
	lruList  *list.List
	maxFiles int
}

// NewFileCache creates a new FileCache with the specified maximum number of cached files
func NewFileCache(maxFiles int) *FileCache {
	if maxFiles <= 0 {
		maxFiles = 20
	}
	return &FileCache{
		entries:  make(map[string]*list.Element),
		lruList:  list.New(),
		maxFiles: maxFiles,
	}
}

// GetLines returns the lines of a file, reading and caching them on first access.
func (fc *FileCache) GetLines(dirPath, fileName string) []string {
	cacheKey := filepath.Join(dirPath, fileName)

	fc.mu.RLock()
	elem, ok := fc.entries[cacheKey]
	if ok {
		// Move to front (most recently used)
		fc.lruList.MoveToFront(elem)
		lines := elem.Value.(*lruEntry).lines
		fc.mu.RUnlock()
		return lines
	}
	fc.mu.RUnlock()

	// Not in cache, read the file
	lines := fc.readFileLines(filepath.Join(dirPath, fileName))
	if lines == nil {
		return nil
	}

	fc.mu.Lock()
	defer fc.mu.Unlock()

	// Double-check after acquiring write lock
	if elem, ok := fc.entries[cacheKey]; ok {
		fc.lruList.MoveToFront(elem)
		return elem.Value.(*lruEntry).lines
	}

	// Evict least recently used if at capacity
	if fc.lruList.Len() >= fc.maxFiles {
		oldest := fc.lruList.Back()
		if oldest != nil {
			fc.lruList.Remove(oldest)
			delete(fc.entries, oldest.Value.(*lruEntry).key)
		}
	}

	// Add new entry
	entry := &lruEntry{key: cacheKey, lines: lines}
	elem = fc.lruList.PushFront(entry)
	fc.entries[cacheKey] = elem
	return lines
}

// readFileLines reads a file and returns its lines.
// To prevent OOM, files larger than MaxCacheFileSize (10MB) are NOT cached;
// they should only be accessed via streaming. This function returns nil
// for such files, forcing callers to use the streaming scanner instead.
func (fc *FileCache) readFileLines(filePath string) []string {
	// Check file size before reading to prevent caching large files in RAM
	fi, err := os.Stat(filePath)
	if err != nil {
		return nil
	}
	if fi.Size() > 10*1024*1024 { // 10MB - large files must be streamed
		return nil
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil
	}

	return lines
}

// Has checks if a file is already cached
func (fc *FileCache) Has(dirPath, fileName string) bool {
	cacheKey := filepath.Join(dirPath, fileName)
	fc.mu.RLock()
	_, ok := fc.entries[cacheKey]
	fc.mu.RUnlock()
	return ok
}

// Clear removes all cached file content
func (fc *FileCache) Clear() {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.entries = make(map[string]*list.Element)
	fc.lruList.Init()
}

// Pool-related global variables

// scannerPool provides reusable buffer allocations for scanners.
var ScannerPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 64*1024)
	},
}

// CreatePooledScanner creates a new bufio.Scanner with a buffer from the pool.
func CreatePooledScanner(f *os.File) *bufio.Scanner {
	bufPtr := ScannerPool.Get()
	buf, ok := bufPtr.([]byte)
	if !ok {
		return bufio.NewScanner(f)
	}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(buf, 1024*1024)
	return scanner
}

// Global instances
var (
	GlobalRegexCache = NewRegexCache()
	GlobalFileCache  = NewFileCache(20)
)
