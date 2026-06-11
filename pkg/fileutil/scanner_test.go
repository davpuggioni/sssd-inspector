// Package fileutil provides file system utilities for SSSD Inspector
package fileutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestDir creates a temporary directory with the given files for testing
func setupTestDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "fileutil-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			os.RemoveAll(dir)
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}
	return dir
}

// TestFileProcessor_AnyFileContains tests the AnyFileContains method
func TestFileProcessor_AnyFileContains(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"file1.txt": "hello world\nfoo bar\n",
		"file2.txt": "baz qux\n",
	})
	defer os.RemoveAll(dir)

	fp := NewFileProcessor()

	tests := []struct {
		name     string
		files    []string
		search   string
		expected bool
	}{
		{"find in first file", []string{"file1.txt", "file2.txt"}, "hello", true},
		{"find in second file", []string{"file1.txt", "file2.txt"}, "baz", true},
		{"not found", []string{"file1.txt", "file2.txt"}, "nonexistent", false},
		{"empty files list", []string{}, "hello", false},
		{"non-existent file", []string{"missing.txt"}, "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fp.AnyFileContains(dir, tt.files, tt.search)
			if got != tt.expected {
				t.Errorf("AnyFileContains(%v, %q) = %v, want %v", tt.files, tt.search, got, tt.expected)
			}
		})
	}
}

// TestFileProcessor_ScanFiles tests the ScanFiles method
func TestFileProcessor_ScanFiles(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"test.txt": "line1\nline2\nline3\n",
	})
	defer os.RemoveAll(dir)

	fp := NewFileProcessor()
	var lines []string
	err := fp.ScanFiles(dir, []string{"test.txt"}, func(line string) {
		lines = append(lines, line)
	})
	if err != nil {
		t.Fatalf("ScanFiles failed: %v", err)
	}
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d: %v", len(lines), lines)
	}
}

// TestFileProcessor_ScanFiles_MissingFile tests scanning a non-existent file
func TestFileProcessor_ScanFiles_MissingFile(t *testing.T) {
	dir := setupTestDir(t, map[string]string{})
	defer os.RemoveAll(dir)

	fp := NewFileProcessor()
	var lines []string
	err := fp.ScanFiles(dir, []string{"nonexistent.txt"}, func(line string) {
		lines = append(lines, line)
	})
	if err != nil {
		t.Fatalf("ScanFiles with missing file should not error: %v", err)
	}
	if len(lines) != 0 {
		t.Errorf("expected 0 lines for missing file, got %d", len(lines))
	}
}

// TestSectionExtractor_ExtractSection tests extracting a section from a file
func TestSectionExtractor_ExtractSection(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"config.txt": `header line
# /etc/sssd/sssd.conf
[domain/ad]
id_provider = ad
enumerate = false
#==[ Command ]======================================
more data
`,
	})
	defer os.RemoveAll(dir)

	se := NewSectionExtractor()
	section := se.ExtractSection(dir, "config.txt", "# /etc/sssd/sssd.conf")

	if !strings.Contains(section, "id_provider = ad") {
		t.Errorf("section should contain 'id_provider = ad', got:\n%s", section)
	}
	if strings.Contains(section, "more data") {
		t.Errorf("section should not contain data past the boundary marker")
	}
}

// TestSectionExtractor_ExtractSection_NotFound tests extracting a non-existent section
func TestSectionExtractor_ExtractSection_NotFound(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"file.txt": "some data\n",
	})
	defer os.RemoveAll(dir)

	se := NewSectionExtractor()
	section := se.ExtractSection(dir, "file.txt", "# nonexistent header")
	if section != "" {
		t.Errorf("expected empty section for non-existent header, got %q", section)
	}
}

// TestSafeFileReader_ReadFileSafe tests reading file lines
func TestSafeFileReader_ReadFileSafe(t *testing.T) {
	sfr := NewSafeFileReader()

	tests := []struct {
		name     string
		lines    []string
		expected string
	}{
		{"empty lines", nil, ""},
		{"single line", []string{"hello"}, "hello"},
		{"multiple lines", []string{"a", "b", "c"}, "a\nb\nc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sfr.ReadFileSafe(tt.lines)
			if got != tt.expected {
				t.Errorf("ReadFileSafe(%v) = %q, want %q", tt.lines, got, tt.expected)
			}
		})
	}
}
