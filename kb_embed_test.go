// kb_embed_test.go
//
// Regression tests for embedded KB articles:
//   - the binary embeds all kb_articles/*.json at compile-time
//   - loadKBArticles succeeds without an external kb_articles/ directory
//   - an external directory overrides matching articles and appends new ones
package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKBArticles_EmbeddedCorpusNotEmpty(t *testing.T) {
	articles := loadEmbeddedKBArticles()
	if len(articles) == 0 {
		t.Fatalf("embedded KB corpus is empty — verify //go:embed kb_articles/*.json in kb.go")
	}
	foundKnown := false
	for _, a := range articles {
		if a.TIDID == "TID-000019823" || a.KbID == "000021149" {
			foundKnown = true
			break
		}
	}
	if !foundKnown {
		t.Errorf("expected well-known TID to be present in embedded articles, got %d articles", len(articles))
	}
}

func TestKBArticles_LoadKBArticlesWorksWithoutDir(t *testing.T) {
	// Must return embedded articles even when called with a nonexistent directory
	articles := loadKBArticles("/nonexistent/directory")
	if len(articles) == 0 {
		t.Fatalf("loadKBArticles returned no articles")
	}
}

func TestKBArticles_ExternalOverrideAndMerge(t *testing.T) {
	tmpDir := t.TempDir()
	extKB := filepath.Join(tmpDir, "kb_articles")
	if err := os.MkdirAll(extKB, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a custom article with an invented ID
	customJSON := `{
  "tid_id": "TID-9999999",
  "title": "Custom Test Article",
  "url": "https://example.test/kb/9999999",
  "log_patterns": ["custom test pattern"]
}`
	if err := os.WriteFile(filepath.Join(extKB, "TID-9999999.json"), []byte(customJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// Switch working dir temporarily to test external discovery
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origWD) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	merged := loadKBArticles()
	foundCustom := false
	for _, a := range merged {
		if a.TIDID == "TID-9999999" {
			foundCustom = true
			if a.Title != "Custom Test Article" {
				t.Errorf("expected overridden title, got %q", a.Title)
			}
			break
		}
	}
	if !foundCustom {
		t.Errorf("expected external article TID-9999999 to be merged into corpus")
	}
}
