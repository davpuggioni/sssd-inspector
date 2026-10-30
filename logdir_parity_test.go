package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// rawLogDir writes a single-file raw SSSD log directory and returns it.
func rawLogDir(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// withDefinitionsRoot points the user definitions root at a temp dir for the
// duration of the test, so KB articles and YAML rules can be injected without
// touching the developer's real ~/.sssd-inspector.
func withDefinitionsRoot(t *testing.T, root string) {
	t.Helper()
	prev := userDefinitionsRoot
	userDefinitionsRoot = func() string { return root }
	t.Cleanup(func() { userDefinitionsRoot = prev })
}

// TestRawMode_MatchesKBArticles is the parity guard for the raw-log path.
//
// analyzeLogsOnly used to skip matchKBArticlesWithEvidence even though the
// single-pass scanner had already collected the log evidence, so
// report.MatchedTIDs was ALWAYS empty under -logdir and
// analyzeKBSuggestions re-suggested articles that had in fact matched.
// KB articles are a documented feature of the mode, so this pins it.
func TestRawMode_MatchesKBArticles(t *testing.T) {
	root := t.TempDir()
	kbDir := filepath.Join(root, "kb_articles")
	if err := os.MkdirAll(kbDir, 0755); err != nil {
		t.Fatal(err)
	}
	article := `{
	  "tid_id": "TID-RAWMODE-TEST",
	  "title": "Synthetic article for the raw-log parity test",
	  "url": "https://example.invalid/tid",
	  "description": "matches a distinctive raw log line",
	  "log_patterns": ["Zqxjkv marker for parity"],
	  "config_patterns": ["id_provider = ad"]
	}`
	if err := os.WriteFile(filepath.Join(kbDir, "rawmode.json"), []byte(article), 0644); err != nil {
		t.Fatal(err)
	}
	withDefinitionsRoot(t, root)

	// Config pattern must NOT be satisfiable here (no sssd.conf in raw mode),
	// so only the log half may drive the match.
	dir := rawLogDir(t, "sssd_example.com.log",
		"(2026-04-01 10:00:00) [sssd[be[EXAMPLE.COM]]] Zqxjkv marker for parity\n")

	report := analyzeLogsOnly(context.Background(), dir, []string{"sssd_example.com.log"}, false, nil)

	if len(report.MatchedTIDs) != 1 || report.MatchedTIDs[0].TIDID != "TID-RAWMODE-TEST" {
		t.Fatalf("raw-log mode must match KB articles on log patterns, got %+v", report.MatchedTIDs)
	}
	if len(report.MatchedTIDs[0].Evidence) == 0 {
		t.Errorf("the matched article must carry the log line that matched: %+v", report.MatchedTIDs[0])
	}
	// An article already matched must not also be offered as a fuzzy suggestion.
	for _, s := range report.KBSuggestions {
		if s.TIDID == "TID-RAWMODE-TEST" {
			t.Errorf("a matched article must not reappear in KBSuggestions: %+v", s)
		}
	}
}

// TestRawMode_AppliesYAMLRules is the same parity guard for the data-driven
// rules engine: a site rule targeting the raw log files is meaningful in this
// mode, which has those files, and was silently skipped.
func TestRawMode_AppliesYAMLRules(t *testing.T) {
	root := t.TempDir()
	rule := "rules:\n" +
		"  - name: rawmode-rule\n" +
		"    severity: error\n" +
		"    category: site\n" +
		"    files: [\"sssd_example.com.log\"]\n" +
		"    match: any\n" +
		"    patterns: [\"Zqxjkv site marker\"]\n" +
		"    message: \"raw-log mode must run YAML rules\"\n"
	if err := os.WriteFile(filepath.Join(root, "rules.yaml"), []byte(rule), 0644); err != nil {
		t.Fatal(err)
	}
	withDefinitionsRoot(t, root)

	dir := rawLogDir(t, "sssd_example.com.log",
		"(2026-04-01 10:00:00) [sssd[be[EXAMPLE.COM]]] Zqxjkv site marker\n")

	report := analyzeLogsOnly(context.Background(), dir, []string{"sssd_example.com.log"}, false, nil)

	found := false
	for _, f := range report.ConfigFindings {
		if f.RuleID == "rule:rawmode-rule" {
			found = true
		}
	}
	if !found {
		t.Errorf("raw-log mode did not apply the YAML rule: %+v", report.ConfigFindings)
	}
}

// TestRawMode_KeytabAbsenceIsNotAFinding pins a deliberate negative: raw-log
// mode must NOT report a missing machine account when the logs carry no keytab
// evidence.
//
// collectLogFiles accepts a single file or a rotated subset, so a user can pass
// just pam_child.log. Absence of keytab evidence in a partial log set is not
// evidence of absence, and reporting it would be a false positive about a check
// that never ran. The positive case (evidence present) is still honoured.
func TestRawMode_KeytabAbsenceIsNotAFinding(t *testing.T) {
	t.Run("no keytab evidence stays silent", func(t *testing.T) {
		dir := rawLogDir(t, "sssd_example.com.log",
			"(2026-04-01 10:00:00) [sssd[be[EXAMPLE.COM]]] Service resolving timeout reached\n")
		report := analyzeLogsOnly(context.Background(), dir, []string{"sssd_example.com.log"}, false, nil)
		if strings.Contains(strings.Join(report.Problems, "|"), "No Kerberos Keytab") {
			t.Errorf("raw-log mode must not infer a missing keytab from absent evidence: %+v", report.Problems)
		}
	})

	t.Run("keytab evidence is still reported", func(t *testing.T) {
		dir := rawLogDir(t, "sssd_example.com.log",
			"(2026-04-01 10:00:00) [sssd[be[EXAMPLE.COM]]] KVNO Principal: host/srv@EXAMPLE.COM\n")
		report := analyzeLogsOnly(context.Background(), dir, []string{"sssd_example.com.log"}, false, nil)
		if !report.KeytabFound {
			t.Errorf("KeytabFound must be true on KVNO evidence")
		}
	})
}
