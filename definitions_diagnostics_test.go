// definitions_diagnostics_test.go
//
// M0 regression tests: definition-loading problems (invalid YAML rules,
// malformed KB articles) must surface as ReportData.Diagnostics with file and
// line — never as invisible stdout noise — and the user/system definition
// roots must be discovered. Also locks the JSON schema stability guarantee:
// with no diagnostics, the "diagnostics" key must stay absent so existing
// report consumers keep working.
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sssd-inspector/config"
)

// withTempWorkingDir chdirs into a fresh temp dir and restores the original
// working directory on cleanup, so loader discovery cannot see the repo's own
// rules/ and kb_articles/ directories.
func withTempWorkingDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Errorf("restore wd %s: %v", oldWd, err)
		}
	})
	return dir
}

// setUserDefinitionsRoot points the user root at dir for one test; TestMain
// keeps it empty everywhere else (no test in this package runs in parallel).
func setUserDefinitionsRoot(t *testing.T, dir string) {
	t.Helper()
	prev := userDefinitionsRoot
	userDefinitionsRoot = func() string { return dir }
	t.Cleanup(func() { userDefinitionsRoot = prev })
}

// TestInvalidRuleProducesDiagnostic is the core M0 guarantee: a rule skipped
// for a bad severity must produce a Diagnostic carrying the file, the source
// line and a human-readable message naming the rule and the bad value.
func TestInvalidRuleProducesDiagnostic(t *testing.T) {
	dir := withTempWorkingDir(t)
	content := `rules:
  - name: "ok-rule"
    severity: "warning"
    message: "fine"
    patterns: ["keep-me"]
  - name: "bad-severity-rule"
    severity: "shout"
    message: "skipped"
    patterns: ["x"]
`
	rulesPath := filepath.Join(dir, "rules.yaml")
	if err := os.WriteFile(rulesPath, []byte(content), 0644); err != nil {
		t.Fatalf("write rules.yaml: %v", err)
	}

	rules, diags := loadAnalysisRules()
	if len(rules) != 1 || rules[0].Name != "ok-rule" {
		t.Fatalf("rules = %+v, want only ok-rule", rules)
	}
	if len(diags) != 1 {
		t.Fatalf("diagnostics = %+v, want exactly 1", diags)
	}
	d := diags[0]
	if d.File != rulesPath {
		t.Errorf("diagnostic file = %q, want %q", d.File, rulesPath)
	}
	if d.Line != 6 {
		t.Errorf("diagnostic line = %d, want 6 (start of the bad-severity rule)", d.Line)
	}
	if !strings.Contains(d.Message, "bad-severity-rule") {
		t.Errorf("diagnostic message %q does not name the rule", d.Message)
	}
	if !strings.Contains(d.Message, "shout") {
		t.Errorf("diagnostic message %q does not name the invalid severity", d.Message)
	}
	if d.Severity != SevWarning {
		t.Errorf("diagnostic severity = %d, want SevWarning", d.Severity)
	}
}

// TestInvalidRuleYAMLSyntaxProducesLineDiagnostic covers the parse-error
// path: a syntactically broken file must be reported with the YAML line.
// The file must be named rules.yaml — the discovery globs only pick
// rules.yaml / rules/*.{yaml,yml} candidates.
func TestInvalidRuleYAMLSyntaxProducesLineDiagnostic(t *testing.T) {
	dir := withTempWorkingDir(t)
	content := "rules:\n  - name: \"broken\"\n   severity: bad-indent\n"
	rulesPath := filepath.Join(dir, "rules.yaml")
	if err := os.WriteFile(rulesPath, []byte(content), 0644); err != nil {
		t.Fatalf("write rules.yaml: %v", err)
	}

	rules, diags := loadAnalysisRules()
	if len(rules) != 0 {
		t.Errorf("rules = %+v, want none from a broken file", rules)
	}
	if len(diags) != 1 {
		t.Fatalf("diagnostics = %+v, want exactly 1", diags)
	}
	d := diags[0]
	if d.File != rulesPath {
		t.Errorf("diagnostic file = %q, want %q", d.File, rulesPath)
	}
	if d.Line <= 0 {
		t.Errorf("diagnostic line = %d, want the YAML error line (> 0)", d.Line)
	}
	if !strings.Contains(d.Message, "invalid rules file") {
		t.Errorf("diagnostic message = %q, want it to flag an invalid rules file", d.Message)
	}
}

// TestMalformedKBJSONProducesDiagnostic: a corrupt external KB article must
// be skipped with a diagnostic (file + line from the JSON error offset), and
// the valid article next to it must still load.
func TestMalformedKBJSONProducesDiagnostic(t *testing.T) {
	dir := withTempWorkingDir(t)
	kbDir := filepath.Join(dir, "kb_articles")
	if err := os.MkdirAll(kbDir, 0755); err != nil {
		t.Fatalf("mkdir kb_articles: %v", err)
	}
	badPath := filepath.Join(kbDir, "TID-000000001.json")
	if err := os.WriteFile(badPath, []byte("{\n  \"tid_id\": \"TID-000000001\",\n  \"title\": oops\n}\n"), 0644); err != nil {
		t.Fatalf("write bad article: %v", err)
	}
	goodPath := filepath.Join(kbDir, "TID-000000002.json")
	if err := os.WriteFile(goodPath, []byte(`{"tid_id":"TID-000000002","title":"Good Article","log_patterns":["good pattern"]}`), 0644); err != nil {
		t.Fatalf("write good article: %v", err)
	}

	articles, diags := loadKBArticlesDiag()
	foundGood := false
	for _, a := range articles {
		if a.TIDID == "TID-000000002" {
			foundGood = true
		}
		if a.TIDID == "TID-000000001" {
			t.Errorf("malformed article was loaded anyway: %+v", a)
		}
	}
	if !foundGood {
		t.Errorf("valid article TID-000000002 not loaded; corpus=%d", len(articles))
	}
	if len(diags) != 1 {
		t.Fatalf("diagnostics = %+v, want exactly 1", diags)
	}
	d := diags[0]
	if d.File != badPath {
		t.Errorf("diagnostic file = %q, want %q", d.File, badPath)
	}
	if d.Line != 3 {
		t.Errorf("diagnostic line = %d, want 3 (line of the bad token)", d.Line)
	}
	if !strings.Contains(d.Message, "invalid KB article JSON") {
		t.Errorf("diagnostic message = %q", d.Message)
	}
}

// TestDefinitionsPathResolution locks the config.yaml-aligned search roots:
// user → $HOME/.sssd-inspector, system → /etc/sssd-inspector, and a rule
// dropped into the user root must be discovered by the loader.
func TestDefinitionsPathResolution(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if got, want := config.UserDefinitionsRoot(), filepath.Join(home, ".sssd-inspector"); got != want {
		t.Errorf("UserDefinitionsRoot() = %q, want %q", got, want)
	}
	if got := config.SystemDefinitionsRoot(); got != "/etc/sssd-inspector" {
		t.Errorf("SystemDefinitionsRoot() = %q, want /etc/sssd-inspector", got)
	}

	// A rule in the user root must load even though the working directory
	// (chdir'ed away below) and the system root hold nothing.
	userRoot := t.TempDir()
	setUserDefinitionsRoot(t, userRoot)
	withTempWorkingDir(t)

	rulesDir := filepath.Join(userRoot, "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("mkdir user rules dir: %v", err)
	}
	userRule := `rules:
  - name: "user-root-rule"
    severity: "warning"
    message: "found me"
    patterns: ["user-marker"]
`
	if err := os.WriteFile(filepath.Join(rulesDir, "site.yaml"), []byte(userRule), 0644); err != nil {
		t.Fatalf("write site.yaml: %v", err)
	}

	found := false
	for _, c := range ruleCandidatePaths() {
		if c == rulesDir {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ruleCandidatePaths() = %v, want it to include %q", ruleCandidatePaths(), rulesDir)
	}

	rules, diags := loadAnalysisRules()
	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics: %+v", diags)
	}
	if len(rules) != 1 || rules[0].Name != "user-root-rule" {
		t.Fatalf("rules = %+v, want user-root-rule from the user definitions root", rules)
	}
}

// TestDiagnosticsSurviveJSONRoundTrip: the Wails GUI round trip (Go → JS →
// Go) must not lose the diagnostics, and a report without diagnostics must
// not serialize the key at all (schema stability for existing consumers).
func TestDiagnosticsSurviveJSONRoundTrip(t *testing.T) {
	input := Diagnostic{
		File:     "/home/user/.sssd-inspector/rules/site.yaml",
		Line:     12,
		Message:  `rule "x": invalid severity "loud"`,
		Severity: SevWarning,
	}
	back := roundTripReport(t, ReportData{AppVersion: "test", Diagnostics: []Diagnostic{input}})
	if len(back.Diagnostics) != 1 {
		t.Fatalf("diagnostics after round trip = %+v, want 1 entry", back.Diagnostics)
	}
	if back.Diagnostics[0] != input {
		t.Errorf("diagnostic after round trip = %+v, want %+v", back.Diagnostics[0], input)
	}

	// Schema stability: no diagnostics → the key must be absent entirely.
	empty, err := json.Marshal(ReportData{AppVersion: "test"})
	if err != nil {
		t.Fatalf("marshal empty report: %v", err)
	}
	if strings.Contains(string(empty), "diagnostics") {
		t.Errorf("report without diagnostics contains a diagnostics key: %s", empty)
	}
}
