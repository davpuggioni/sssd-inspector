// utils_filter_test.go
//
// Tests for the FileFilter (relevant-file set, 0% coverage) and the
// context-aware scanning layer (utils.go).
package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFileFilter_DefaultsContainAnalyzerFiles(t *testing.T) {
	ff := NewFileFilter()
	for _, name := range []string{"sssd.conf", "sssd.txt", "messages", "nsswitch.conf", "rpm.txt"} {
		if !ff.IsRelevantFile(name) {
			t.Errorf("expected %s to be a default relevant file", name)
		}
	}
	if ff.IsRelevantFile("random-debug.bin") {
		t.Errorf("random-debug.bin must not be a relevant file")
	}
}

func TestFileFilter_AddRemoveGet(t *testing.T) {
	ff := NewFileFilter()
	ff.AddRelevantFile("custom-debug.log")
	if !ff.IsRelevantFile("custom-debug.log") {
		t.Errorf("AddRelevantFile did not take effect")
	}
	before := len(ff.GetRelevantFiles())
	ff.RemoveRelevantFile("custom-debug.log")
	if ff.IsRelevantFile("custom-debug.log") {
		t.Errorf("RemoveRelevantFile did not take effect")
	}
	if got := len(ff.GetRelevantFiles()); got != before-1 {
		t.Errorf("GetRelevantFiles: expected %d entries, got %d", before-1, got)
	}
	// Removing an unknown name must be a no-op, not a panic.
	ff.RemoveRelevantFile("never-existed.log")
}

// TestIsRelevantFile_LegacyGlobal pins the package-level helper used by the
// archive extractor (loaders.go) and the scanners.
func TestIsRelevantFile_LegacyGlobal(t *testing.T) {
	if !isRelevantFile("sssd.conf") {
		t.Errorf("isRelevantFile(sssd.conf) = false, want true")
	}
	if isRelevantFile("") {
		t.Errorf("isRelevantFile(\"\") = true, want false")
	}
}

func TestNewFileProcessorWithConfig_AnyFileContains(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("first line\nneedle here\n"), 0644); err != nil {
		t.Fatal(err)
	}
	fp := NewFileProcessorWithConfig(64*1024, 1024*1024)
	if !fp.AnyFileContains(dir, []string{"a.txt", "missing.txt"}, "needle") {
		t.Errorf("AnyFileContains missed an existing match (missing files must be skipped)")
	}
	if fp.AnyFileContains(dir, []string{"a.txt"}, "no-such-pattern") {
		t.Errorf("AnyFileContains reported a phantom match")
	}
}

func TestScanFileWithContext_MissingFileIsNil(t *testing.T) {
	if err := scanFileWithContext(context.Background(), filepath.Join(t.TempDir(), "nope.txt"), func(string) {}); err != nil {
		t.Errorf("missing file must be nil, got %v", err)
	}
}

func TestScanFileWithContext_CancelStopsScan(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "big.txt")
	var sb []byte
	for i := 0; i < 5000; i++ {
		sb = append(sb, []byte("log line number with some content\n")...)
	}
	if err := os.WriteFile(path, sb, 0644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	seen := 0
	err := scanFileWithContext(ctx, path, func(string) {
		seen++
		if seen == 3 {
			cancel()
		}
	})
	if err == nil {
		t.Errorf("expected a cancellation error, got nil")
	}
	if seen > 10 {
		t.Errorf("scan did not stop promptly after cancel: %d lines seen", seen)
	}
}

func TestAnyFileContainsWithContext(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "m.txt"), []byte("alpha\nbeta needle\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if !anyFileContainsWithContext(context.Background(), dir, []string{"m.txt"}, "needle") {
		t.Errorf("expected match")
	}
	if anyFileContainsWithContext(context.Background(), dir, []string{"m.txt"}, "absent") {
		t.Errorf("expected no match")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if anyFileContainsWithContext(ctx, dir, []string{"m.txt"}, "needle") {
		t.Errorf("cancelled context must yield false")
	}
}

func TestScanFilesWithContext_CancelSkipsWork(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "m.txt"), []byte("line\n"), 0644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	scanFilesWithContext(ctx, dir, []string{"m.txt"}, func(string) { called = true })
	if called {
		t.Errorf("cancelled scan must not invoke the callback")
	}
}
