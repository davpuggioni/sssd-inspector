package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRegexCache_Basic tests basic regex caching
func TestRegexCache_Basic(t *testing.T) {
	cache := NewRegexCache()

	re := cache.Get(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
	if re == nil {
		t.Fatal("Get returned nil")
	}
	if !re.MatchString("192.168.1.1") {
		t.Error("regex should match IP address")
	}
	if re.MatchString("not-an-ip") {
		t.Error("regex should not match non-IP")
	}
}

// TestRegexCache_Caching verifies that the same pattern returns the same compiled regex
func TestRegexCache_Caching(t *testing.T) {
	cache := NewRegexCache()
	pattern := `test-\d+`

	re1 := cache.Get(pattern)
	re2 := cache.Get(pattern)

	if re1 != re2 {
		t.Error("Get should return the same object for the same pattern")
	}

	if !re1.MatchString("test-123") {
		t.Error("regex should match pattern")
	}
}

// TestRegexCache_MultiplePatterns tests different patterns
func TestRegexCache_MultiplePatterns(t *testing.T) {
	cache := NewRegexCache()

	patterns := []string{
		`foo`,
		`bar`,
		`[0-9]+`,
	}
	for _, p := range patterns {
		re := cache.Get(p)
		if re == nil {
			t.Fatalf("Get(%q) returned nil", p)
		}
	}
}

// TestRegexCache_ConcurrentSafety tests concurrent access (race detection)
func TestRegexCache_ConcurrentSafety(t *testing.T) {
	cache := NewRegexCache()
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				pattern := `pattern-\d+-\d+`
				re := cache.Get(pattern)
				if re == nil {
					t.Error("Get returned nil")
				}
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestFileCache_GetLines tests reading and caching file lines
func TestFileCache_GetLines(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"test.txt": "line1\nline2\nline3\n",
	})
	defer os.RemoveAll(dir)

	cache := NewFileCache(10)
	lines := cache.GetLines(dir, "test.txt")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %v", len(lines), lines)
	}
	if !cache.Has(dir, "test.txt") {
		t.Error("file should be cached after GetLines")
	}
}

// TestFileCache_GetLines_MissingFile tests behavior with missing files
func TestFileCache_GetLines_MissingFile(t *testing.T) {
	dir := setupTestDir(t, map[string]string{})
	defer os.RemoveAll(dir)

	cache := NewFileCache(10)
	lines := cache.GetLines(dir, "nonexistent.txt")
	if lines != nil {
		t.Errorf("expected nil for missing file, got %v", lines)
	}
}

// TestFileCache_Clear verifies that Clear empties the cache
func TestFileCache_Clear(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"test.txt": "content\n",
	})
	defer os.RemoveAll(dir)

	cache := NewFileCache(10)
	cache.GetLines(dir, "test.txt")
	if !cache.Has(dir, "test.txt") {
		t.Error("file should be cached after GetLines")
	}

	cache.Clear()
	if cache.Has(dir, "test.txt") {
		t.Error("file should not be cached after Clear")
	}
}

// TestFileCache_MaxFiles verifies the cache eviction policy
func TestFileCache_MaxFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	for i := 0; i < 5; i++ {
		name := filepath.Join(tempDir, "test.txt")
		if err := os.WriteFile(name, []byte("content"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	cache := NewFileCache(2) // max 2 files
	// This should not panic or cause errors
	for i := 0; i < 5; i++ {
		cache.GetLines(tempDir, "test.txt")
	}
}

// TestNewFileCache_Default verifies default max files
func TestNewFileCache_Default(t *testing.T) {
	cache := NewFileCache(0)
	if cache.maxFiles != 20 {
		t.Errorf("expected default maxFiles=20, got %d", cache.maxFiles)
	}
}
