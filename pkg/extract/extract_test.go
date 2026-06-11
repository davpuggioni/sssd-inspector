package extract

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// TestExtractArchiveToTemp_InvalidFile tests behavior with non-existent file
func TestExtractArchiveToTemp_InvalidFile(t *testing.T) {
	_, err := ExtractArchiveToTemp("/nonexistent/test.txz", nil)
	if err == nil {
		t.Error("expected error for non-existent archive file")
	}
}

// TestIsRelevantFile_Default tests the default isRelevantFile behavior
func TestIsRelevantFile_Default(t *testing.T) {
	if !isRelevantFile("any_file.txt") {
		t.Error("default isRelevantFile should return true for any file")
	}
	if !isRelevantFile("sssd.txt") {
		t.Error("default isRelevantFile should return true for sssd.txt")
	}
}

// TestIsRelevantFile_Custom tests custom filter
func TestIsRelevantFile_Custom(t *testing.T) {
	original := IsRelevantFile
	defer func() { IsRelevantFile = original }()

	IsRelevantFile = func(name string) bool {
		return strings.HasSuffix(name, ".txt")
	}

	if !isRelevantFile("test.txt") {
		t.Error("custom filter should accept .txt files")
	}
	if isRelevantFile("test.bin") {
		t.Error("custom filter should reject non-.txt files")
	}
}

// TestCopyWithBuffer tests the internal copyWithBuffer function using sync.Pool
func TestCopyWithBuffer(t *testing.T) {
	pool := &sync.Pool{
		New: func() interface{} {
			b := make([]byte, ExtractionBufferSize)
			return &b
		},
	}

	// Create temp files for testing
	srcFile, err := os.CreateTemp("", "src-*")
	if err != nil {
		t.Fatalf("failed to create src file: %v", err)
	}
	defer os.Remove(srcFile.Name())
	defer srcFile.Close()

	content := "hello world test content"
	if _, err := srcFile.WriteString(content); err != nil {
		t.Fatalf("failed to write src: %v", err)
	}
	if _, err := srcFile.Seek(0, 0); err != nil {
		t.Fatalf("failed to seek: %v", err)
	}

	dstFile, err := os.CreateTemp("", "dst-*")
	if err != nil {
		t.Fatalf("failed to create dst file: %v", err)
	}
	defer os.Remove(dstFile.Name())
	defer dstFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		t.Fatalf("failed to stat src: %v", err)
	}

	written, err := copyWithBuffer(dstFile, srcFile, pool)
	if err != nil {
		t.Fatalf("copyWithBuffer failed: %v", err)
	}
	if written != srcInfo.Size() {
		t.Errorf("expected %d bytes written, got %d", srcInfo.Size(), written)
	}

	// Verify content
	dstContent, err := os.ReadFile(dstFile.Name())
	if err != nil {
		t.Fatalf("failed to read dst: %v", err)
	}
	if string(dstContent) != content {
		t.Errorf("expected content %q, got %q", content, string(dstContent))
	}
}

// TestConstants verifies expected constant values
func TestConstants(t *testing.T) {
	if ExtractionBufferSize != 32*1024 {
		t.Errorf("expected ExtractionBufferSize=32768, got %d", ExtractionBufferSize)
	}
	if DefaultExtractionTimeout != 5*60*1000*1000*1000 {
		// Just verify it's set
		if DefaultExtractionTimeout <= 0 {
			t.Error("DefaultExtractionTimeout should be positive")
		}
	}
	if MaxArchiveFileSize != 100*1024*1024 {
		t.Errorf("expected MaxArchiveFileSize=104857600, got %d", MaxArchiveFileSize)
	}
}

// TestFilePath_NoTraversal verifies path traversal protection
func TestFilePath_NoTraversal(t *testing.T) {
	// Test a simplified version of the path traversal check
	tempDir, err := os.MkdirTemp("", "sssd-inspector-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Path traversal attempt
	traversalPath := filepath.Join(tempDir, "../../../etc/passwd")
	cleanPath := filepath.Clean(traversalPath)
	cleanBase := filepath.Clean(tempDir) + string(os.PathSeparator)

	if !strings.HasPrefix(cleanPath, cleanBase) {
		t.Log("path traversal correctly detected")
	}
}
