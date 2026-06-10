package fileutil

import (
	"bufio"
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
func (rc *RegexCache) Get(pattern string) *regexp.Regexp {
	rc.mu.RLock()
	re, ok := rc.patterns[pattern]
	rc.mu.RUnlock()
	if ok {
		return re
	}

	rc.mu.Lock()
	defer rc.mu.Unlock()

	if re, ok := rc.patterns[pattern]; ok {
		return re
	}

	re = regexp.MustCompile(pattern)
	rc.patterns[pattern] = re
	return re
}

// Clear removes all cached patterns
func (rc *RegexCache) Clear() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.patterns = make(map[string]*regexp.Regexp)
}

// FileCache provides thread-safe caching of file contents (lines).
type FileCache struct {
	mu       sync.RWMutex
	buffers  map[string][]string
	maxFiles int
	curFiles int
}

// NewFileCache creates a new FileCache with the specified maximum number of cached files
func NewFileCache(maxFiles int) *FileCache {
	if maxFiles <= 0 {
		maxFiles = 20
	}
	return &FileCache{
		buffers:  make(map[string][]string),
		maxFiles: maxFiles,
	}
}

// GetLines returns the lines of a file, reading and caching them on first access.
func (fc *FileCache) GetLines(dirPath, fileName string) []string {
	cacheKey := filepath.Join(dirPath, fileName)

	fc.mu.RLock()
	lines, ok := fc.buffers[cacheKey]
	fc.mu.RUnlock()
	if ok {
		return lines
	}

	lines = fc.readFileLines(filepath.Join(dirPath, fileName))
	if lines == nil {
		return nil
	}

	fc.mu.Lock()
	defer fc.mu.Unlock()

	if lines, ok := fc.buffers[cacheKey]; ok {
		return lines
	}

	if fc.curFiles >= fc.maxFiles {
		for key := range fc.buffers {
			delete(fc.buffers, key)
			fc.curFiles--
			break
		}
	}

	fc.buffers[cacheKey] = lines
	fc.curFiles++
	return lines
}

// readFileLines reads a file and returns its lines.
// Files larger than 10MB are NOT cached.
func (fc *FileCache) readFileLines(filePath string) []string {
	fi, err := os.Stat(filePath)
	if err != nil {
		return nil
	}
	if fi.Size() > 10*1024*1024 {
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
	_, ok := fc.buffers[cacheKey]
	fc.mu.RUnlock()
	return ok
}

// Clear removes all cached file content
func (fc *FileCache) Clear() {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.buffers = make(map[string][]string)
	fc.curFiles = 0
}

// scannerPool provides reusable buffer allocations for scanners.
var ScannerPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 64*1024)
	},
}

// CreatePooledScanner creates a new bufio.Scanner with a buffer from the pool.
func CreatePooledScanner(f *os.File) *bufio.Scanner {
	buf := ScannerPool.Get().([]byte)
	scanner := bufio.NewScanner(f)
	scanner.Buffer(buf, 1024*1024)
	return scanner
}

// Global instances
var (
	GlobalRegexCache = NewRegexCache()
	GlobalFileCache  = NewFileCache(20)
)
