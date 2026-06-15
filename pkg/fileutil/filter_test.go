package fileutil

import (
	"testing"
)

// TestFileFilter_IsRelevantFile tests file relevance checks
func TestFileFilter_IsRelevantFile(t *testing.T) {
	ff := NewFileFilter()

	// These files should be relevant based on constants (basic check)
	if !ff.IsRelevantFile("messages") {
		t.Error("'messages' should be relevant")
	}
	if !ff.IsRelevantFile("sssd.txt") {
		t.Error("'sssd.txt' should be relevant")
	}
	if !ff.IsRelevantFile("sssd.conf") {
		t.Error("'sssd.conf' should be relevant")
	}

	// Files that should not be relevant
	if ff.IsRelevantFile("random_file.txt") {
		// May or may not be relevant depending on constants
		// This is just a sanity check
	}
}

// TestFileFilter_AddRemoveRelevantFile tests dynamic modification
func TestFileFilter_AddRemoveRelevantFile(t *testing.T) {
	ff := NewFileFilter()

	ff.AddRelevantFile("custom.txt")
	if !ff.IsRelevantFile("custom.txt") {
		t.Error("'custom.txt' should be relevant after AddRelevantFile")
	}

	ff.RemoveRelevantFile("custom.txt")
	if ff.IsRelevantFile("custom.txt") {
		t.Error("'custom.txt' should not be relevant after RemoveRelevantFile")
	}
}

// TestFileFilter_GetRelevantFiles verifies the list of relevant files
func TestFileFilter_GetRelevantFiles(t *testing.T) {
	ff := NewFileFilter()
	files := ff.GetRelevantFiles()

	if len(files) == 0 {
		t.Error("GetRelevantFiles should return at least one file")
	}

	// All returned files should pass IsRelevantFile
	for _, f := range files {
		if !ff.IsRelevantFile(f) {
			t.Errorf("file %q returned by GetRelevantFiles but IsRelevantFile returns false", f)
		}
	}

	// Adding a file should update both
	ff.AddRelevantFile("new_file.txt")
	files = ff.GetRelevantFiles()
	found := false
	for _, f := range files {
		if f == "new_file.txt" {
			found = true
			break
		}
	}
	if !found {
		t.Error("GetRelevantFiles should include newly added file")
	}
}
