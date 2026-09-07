// kb_ingest_test.go - Tests for dual-schema TID JSON ingestion
// (curated kb_articles/ format + kbscraper TIDs/ format).

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// scraperArticleJSON mirrors the kbscraper TID schema.
const scraperArticleJSON = `{
  "kb_id": "000021001",
  "title": "AppArmor automatically enabled during upgrade | SUSE | Support Center",
  "url": "https://support.scc.suse.com/s/kb/apparmor-automatically-enabled-during-upgrade",
  "plain_text": "AppArmor automatically enabled during upgrade\n\nEnvironment\nSUSE Linux Enterprise Server 12 SP5\n\nSituation\nIf the AppArmor systemd service is disabled in SLES12-SP3, it is changed to \"enabled\" during the upgrade.\n\nResolution\nSet a systemd mask on the AppArmor service file.",
  "environment": "SUSE Linux Enterprise Server 12 SP5",
  "situation": "If the AppArmor systemd service is disabled in SLES12-SP3, it is changed to \"enabled\" during the upgrade.\n\nSLES12 SP5 Example:\nsystemctl status apparmor.service",
  "resolution": "Set a systemd \"mask\" on the AppArmor service file, this will prevent AppArmor from running.",
  "cause": "Due to a change in the systemd vendor preset file, AppArmor is enabled during OS upgrade.",
  "created": "2023-03-03T22:47:00Z",
  "changed": "2023-03-03T22:47:00Z"
}`

// curatedArticleJSON mirrors the curated kb_articles/ schema.
const curatedArticleJSON = `{
  "tid_id": "TID-000019823",
  "title": "SSSD fails with 'service key not available' due to AD RC4 deprecation",
  "url": "https://support.scc.suse.com/s/article/000019823",
  "conditions": {
    "log_patterns": ["service key not available", "TGT failed verification"],
    "config_patterns": [],
    "min_sssd_version": "2.9.0"
  },
  "description": "Microsoft AD enforces RC4 encryption when the operatingSystemVersion attribute starts with a number lower than 6."
}`

func parseArticle(t *testing.T, raw string) TIDArticle {
	t.Helper()
	var article TIDArticle
	if err := json.Unmarshal([]byte(raw), &article); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	normalizeTIDArticle(&article)
	return article
}

func TestIngestScraperFormat(t *testing.T) {
	article := parseArticle(t, scraperArticleJSON)

	if article.TIDID != "TID-000021001" {
		t.Errorf("TIDID = %q, want TID-000021001", article.TIDID)
	}
	if article.KbID != "000021001" {
		t.Errorf("KbID = %q, want 000021001", article.KbID)
	}
	if article.Title == "" || article.URL == "" {
		t.Error("title/url must be populated from the scraper schema")
	}
	// Description derived from situation.
	if !strings.Contains(article.Description, "changed to") {
		t.Errorf("Description should derive from situation, got %q", article.Description)
	}
	if len(article.Description) > 0 && strings.Contains(article.Description, "\n") {
		t.Errorf("Description should be a single collapsed paragraph, got %q", article.Description)
	}
	// No matching patterns for scraped articles.
	if len(article.LogPatterns) != 0 || len(article.ConfigPatterns) != 0 {
		t.Errorf("scraped articles must not gain patterns, got log=%v config=%v", article.LogPatterns, article.ConfigPatterns)
	}
	if article.PlainText == "" || article.Situation == "" || article.Resolution == "" {
		t.Error("scraper content fields must be preserved")
	}
}

func TestIngestScraperNoSituation(t *testing.T) {
	var article TIDArticle
	raw := `{"kb_id":"000021002","title":"T","url":"u","plain_text":"First paragraph here.\n\nSecond paragraph.","cause":"the cause"}`
	if err := json.Unmarshal([]byte(raw), &article); err != nil {
		t.Fatal(err)
	}
	normalizeTIDArticle(&article)

	if article.TIDID != "TID-000021002" {
		t.Errorf("TIDID = %q", article.TIDID)
	}
	// No situation → falls back to cause.
	if article.Description != "the cause" {
		t.Errorf("Description = %q, want fallback to cause", article.Description)
	}
}

func TestIngestCuratedFormatUnchanged(t *testing.T) {
	article := parseArticle(t, curatedArticleJSON)

	if article.TIDID != "TID-000019823" {
		t.Errorf("TIDID = %q", article.TIDID)
	}
	if article.Description != "Microsoft AD enforces RC4 encryption when the operatingSystemVersion attribute starts with a number lower than 6." {
		t.Errorf("curated description must not be overwritten, got %q", article.Description)
	}
	if len(article.LogPatterns) != 2 || article.LogPatterns[0] != "service key not available" {
		t.Errorf("curated log patterns must be preserved, got %v", article.LogPatterns)
	}
	if article.KbID != "" {
		t.Errorf("curated articles should not populate scraper fields, got KbID=%q", article.KbID)
	}
}

// TestIngestCuratedNestedConditions verifies the alternate curated schema
// (patterns nested under "conditions", as in kb_articles/TID-000019823.json)
// now has its patterns promoted — previously they were silently ignored.
func TestIngestCuratedNestedConditions(t *testing.T) {
	nested := `{
  "tid_id": "TID-000019823",
  "title": "SSSD fails with 'service key not available' due to AD RC4 deprecation",
  "url": "https://support.scc.suse.com/s/article/000019823",
  "conditions": {
    "log_patterns": ["service key not available", "TGT failed verification"],
    "config_patterns": ["ldap_id_mapping"],
    "min_sssd_version": "2.9.0"
  },
  "description": "desc"
}`
	article := parseArticle(t, nested)

	if len(article.LogPatterns) != 2 {
		t.Errorf("nested log_patterns not promoted, got %v", article.LogPatterns)
	}
	if len(article.ConfigPatterns) != 1 || article.ConfigPatterns[0] != "ldap_id_mapping" {
		t.Errorf("nested config_patterns not promoted, got %v", article.ConfigPatterns)
	}
	if article.Description != "desc" {
		t.Errorf("existing description must not be overwritten, got %q", article.Description)
	}
}

// TestIngestAllScraperFiles validates ingestion against the real scraped
// corpus (if present on this machine). Skipped in CI / other machines.
func TestIngestAllScraperFiles(t *testing.T) {
	dir := os.Getenv("SCRAPER_TID_DIR")
	if dir == "" {
		t.Skip("SCRAPER_TID_DIR not set; skipping real-corpus validation")
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no JSON files found in %s", dir)
	}
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Errorf("%s: read: %v", file, err)
			continue
		}
		var article TIDArticle
		if err := json.Unmarshal(content, &article); err != nil {
			t.Errorf("%s: invalid JSON: %v", file, err)
			continue
		}
		normalizeTIDArticle(&article)
		if !strings.HasPrefix(article.TIDID, "TID-") {
			t.Errorf("%s: TIDID = %q, want TID-<kb_id>", file, article.TIDID)
		}
		if article.Title == "" || article.URL == "" {
			t.Errorf("%s: missing title/url", file)
		}
	}
}

// project to guarantee no regression in the curated path.
func TestIngestRealCuratedFiles(t *testing.T) {
	files, err := filepath.Glob("kb_articles/*.json")
	if err != nil || len(files) == 0 {
		t.Skip("no kb_articles/ found")
	}
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("%s: %v", file, err)
		}
		var article TIDArticle
		if err := json.Unmarshal(content, &article); err != nil {
			t.Errorf("%s: invalid JSON: %v", file, err)
			continue
		}
		normalizeTIDArticle(&article)
		if article.TIDID == "" {
			t.Errorf("%s: TIDID is empty", file)
		}
	}
}

func TestIngestMixedKBDirectory(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"curated.json": curatedArticleJSON,
		"scraped.json": scraperArticleJSON,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil || len(entries) != 2 {
		t.Fatalf("glob failed: %v (%d entries)", err, len(entries))
	}

	loaded := 0
	for _, file := range entries {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var article TIDArticle
		if err := json.Unmarshal(content, &article); err != nil {
			t.Fatalf("%s: %v", file, err)
		}
		normalizeTIDArticle(&article)
		if article.TIDID == "" {
			t.Errorf("%s: TIDID not derived", file)
		}
		loaded++
	}
	if loaded != 2 {
		t.Errorf("expected 2 articles loaded, got %d", loaded)
	}
}
