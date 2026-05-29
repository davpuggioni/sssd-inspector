// Package main tests for parallel file processing utilities
package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// BenchmarkParallelFileProcessor benchmarks parallel file scanning
func BenchmarkParallelFileProcessor(b *testing.B) {
	// Setup: Create temporary directory with test files
	dir, err := os.MkdirTemp("", "bench-parallel-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Create multiple test files with content
	testFiles := make([]string, 10)
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("test-%d.txt", i)
		path := filepath.Join(dir, name)
		var content strings.Builder
		for j := 0; j < 1000; j++ {
			content.WriteString(fmt.Sprintf("Line %d: This is a test line for benchmarking purposes\n", j))
		}
		if err := os.WriteFile(path, []byte(content.String()), 0644); err != nil {
			b.Fatal(err)
		}
		testFiles[i] = name
	}

	b.ResetTimer()

	// Benchmark parallel processing
	for i := 0; i < b.N; i++ {
		pfp := NewParallelFileProcessor(4)
		var lineCount int
		var mu sync.Mutex
		pfp.ScanFilesParallel(dir, testFiles, func(line string) {
			mu.Lock()
			lineCount++
			mu.Unlock()
		})
		_ = lineCount
	}
}

// BenchmarkSequentialFileProcessor benchmarks sequential file scanning for comparison
func BenchmarkSequentialFileProcessor(b *testing.B) {
	// Setup: Create temporary directory with test files
	dir, err := os.MkdirTemp("", "bench-sequential-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Create multiple test files with content
	testFiles := make([]string, 10)
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("test-%d.txt", i)
		path := filepath.Join(dir, name)
		var content strings.Builder
		for j := 0; j < 1000; j++ {
			content.WriteString(fmt.Sprintf("Line %d: This is a test line for benchmarking purposes\n", j))
		}
		if err := os.WriteFile(path, []byte(content.String()), 0644); err != nil {
			b.Fatal(err)
		}
		testFiles[i] = name
	}

	b.ResetTimer()

	// Benchmark sequential processing
	for i := 0; i < b.N; i++ {
		fp := NewFileProcessor()
		var lineCount int
		fp.ScanFiles(dir, testFiles, func(line string) {
			lineCount++
		})
		_ = lineCount
	}
}

// TestParallelFileProcessor tests the parallel file processor
func TestParallelFileProcessor(t *testing.T) {
	dir, err := os.MkdirTemp("", "test-parallel-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Create test files
	files := []string{"file1.txt", "file2.txt", "file3.txt"}
	for _, name := range files {
		content := fmt.Sprintf("Content of %s\nLine 2\nLine 3\n", name)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Test parallel scanning
	pfp := NewParallelFileProcessor(2)
	var mu sync.Mutex
	linesFound := make(map[string]bool)
	pfp.ScanFilesParallel(dir, files, func(line string) {
		mu.Lock()
		linesFound[line] = true
		mu.Unlock()
	})

	expectedLines := []string{
		"Content of file1.txt",
		"Content of file2.txt",
		"Content of file3.txt",
		"Line 2",
		"Line 3",
	}

	for _, expected := range expectedLines {
		if !linesFound[expected] {
			t.Errorf("Expected line not found: %s", expected)
		}
	}
}

// TestParallelFileProcessorEmptyFiles tests empty file list
func TestParallelFileProcessorEmptyFiles(t *testing.T) {
	pfp := NewParallelFileProcessor(2)
	// Should not panic with empty files
	pfp.ScanFilesParallel("/tmp", []string{}, func(line string) {})
}

// TestBatchProcessor tests the batch processor
func TestBatchProcessor(t *testing.T) {
	bp := NewBatchProcessor(2)
	items := []string{"a", "b", "c", "d", "e"}

	results := bp.ProcessBatch(items, func(item string) string {
		return strings.ToUpper(item)
	})

	expected := map[string]bool{"A": true, "B": true, "C": true, "D": true, "E": true}
	for _, r := range results {
		if !expected[r] {
			t.Errorf("Unexpected result: %s", r)
		}
		delete(expected, r)
	}

	if len(expected) > 0 {
		t.Errorf("Missing results: %v", expected)
	}
}

// TestBatchProcessorEmpty tests empty batch
func TestBatchProcessorEmpty(t *testing.T) {
	bp := NewBatchProcessor(2)
	results := bp.ProcessBatch([]string{}, func(item string) string {
		return item
	})
	if results != nil {
		t.Errorf("Expected nil for empty batch, got: %v", results)
	}
}
