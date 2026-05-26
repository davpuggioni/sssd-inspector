// Package main provides caching utilities for the scanning engine
package main

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

// RegexCache provides thread-safe caching of compiled regular expressions.
// Compiles each pattern only once, reusing it for the entire application lifetime.
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
// Thread-safe: multiple goroutines can call Get concurrently.
func (rc *RegexCache) Get(pattern string) *regexp.Regexp {
	// Fast path: read lock
	rc.mu.RLock()
	re, ok := rc.patterns[pattern]
	rc.mu.RUnlock()
	if ok {
		return re
	}

	// Slow path: compile and store (with write lock)
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// Double-check after acquiring write lock
	if re, ok := rc.patterns[pattern]; ok {
		return re
	}

	re = regexp.MustCompile(pattern)
	rc.patterns[pattern] = re
	return re
}

// MustCompile is a convenience wrapper that compiles with panic on error,
// matching the standard regexp.MustCompile signature but using the cache.
func (rc *RegexCache) MustCompile(pattern string) *regexp.Regexp {
	return rc.Get(pattern)
}

// Clear removes all cached patterns, freeing memory
func (rc *RegexCache) Clear() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.patterns = make(map[string]*regexp.Regexp)
}

// FileCache provides thread-safe caching of file contents (lines).
// It caches file content to avoid reading the same file multiple times
// during a single analysis pass. The cache should be cleared between analyses.
type FileCache struct {
	mu       sync.RWMutex
	buffers  map[string][]string // filePath -> lines
	maxFiles int                 // maximum number of files to cache
	curFiles int                 // current count of cached files
}

// NewFileCache creates a new FileCache with the specified maximum number of cached files
func NewFileCache(maxFiles int) *FileCache {
	if maxFiles <= 0 {
		maxFiles = 20 // Default: cache up to 20 files
	}
	return &FileCache{
		buffers:  make(map[string][]string),
		maxFiles: maxFiles,
	}
}

// GetLines returns the lines of a file, reading and caching them on first access.
// Thread-safe: multiple goroutines can call GetLines concurrently.
func (fc *FileCache) GetLines(dirPath, fileName string) []string {
	cacheKey := filepath.Join(dirPath, fileName)

	// Fast path: read lock
	fc.mu.RLock()
	lines, ok := fc.buffers[cacheKey]
	fc.mu.RUnlock()
	if ok {
		return lines
	}

	// Slow path: read file and cache
	lines = fc.readFileLines(filepath.Join(dirPath, fileName))
	if lines == nil {
		return nil
	}

	fc.mu.Lock()
	defer fc.mu.Unlock()

	// Double-check after acquiring write lock
	if lines, ok := fc.buffers[cacheKey]; ok {
		return lines
	}

	// Evict oldest if at capacity
	if fc.curFiles >= fc.maxFiles {
		for key := range fc.buffers {
			delete(fc.buffers, key)
			fc.curFiles--
			break // Evict just one entry
		}
	}

	fc.buffers[cacheKey] = lines
	fc.curFiles++
	return lines
}

// readFileLines reads a file and returns its lines
func (fc *FileCache) readFileLines(filePath string) []string {
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
	_, ok := fc.buffers[cacheKey]
	fc.mu.RUnlock()
	return ok
}

// Clear removes all cached file content, freeing memory.
// Should be called between analysis runs.
func (fc *FileCache) Clear() {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.buffers = make(map[string][]string)
	fc.curFiles = 0
}

// globalRegexCache is the application-wide regex cache
var globalRegexCache = NewRegexCache()

// globalFileCache is the application-wide file cache (used during analysis)
var globalFileCache = NewFileCache(20)
