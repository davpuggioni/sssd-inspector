// analysis_scoring_test.go
package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// Regression test for the Aho-Corasick literal/regex split: patterns written
// with regular-expression syntax (e.g. "Attribute .* not allowed for user")
// previously failed SILENTLY because the trie matched them byte-by-byte.
func TestSinglePass_RegexPatternRouting(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"sssd.txt": "(2026-01-01 10:00:00) [sssd[ifp]] Attribute manager not allowed for user\n",
	})
	defer os.RemoveAll(dir)

	result := performSinglePassScan(dir, "None", nil)

	found := false
	for _, e := range result.SSSDLogErrors {
		if strings.Contains(e.Description, "IFP Attribute Filter") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("regex-routed pattern 'Attribute .* not allowed for user' was not matched (silent AC failure)")
	}
}

// The pre-filter must keep passing lines that contain SSSD-related keywords
// (behaviour parity with the previous strings.Contains chain).
func TestSinglePass_PrefilterKeepsRelevantLines(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"sssd.txt": "(2026-01-01 10:00:00) [sssd[be[EXAMPLE.COM]]] Preauthentication failed\n",
	})
	defer os.RemoveAll(dir)

	result := performSinglePassScan(dir, "None", nil)

	found := false
	for _, e := range result.SSSDLogErrors {
		if strings.Contains(e.Description, "Preauthentication failed") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("pre-filter dropped a relevant SSSD line: Preauthentication pattern not detected")
	}
}

func TestComputeExecutiveSummary_ScoringAndHeadline(t *testing.T) {
	r := &ReportData{
		ConfigFindings: []ConfigFinding{
			{Severity: SevCritical, Category: "krb5_realm", Message: "realm mismatch"},
			{Severity: SevError, Category: "dns", Message: "search domain mismatch"},
			{Severity: SevWarning, Category: "case_sensitive", Message: "case sensitive"},
		},
		Problems:      []string{"problem one"},
		Warnings:      []string{"hint one"},
		SSSDLogErrors: []SSSDLogError{{Description: "err"}},
	}
	computeExecutiveSummary(r)

	s := r.Summary
	if s.CriticalCount != 1 || s.ErrorCount != 1 {
		t.Errorf("unexpected counts: %+v", s)
	}
	// 100 - 25 (critical) - 12 (error) - 2 (warning finding) - 2 (warning list)
	//    - 3 (log error pattern) - 4 (problem string)
	expectedScore := 100 - 25 - 12 - 2 - 2 - 3 - 4
	if s.HealthScore != expectedScore {
		t.Errorf("health score = %d, want %d", s.HealthScore, expectedScore)
	}
	if s.TopCategory != "krb5_realm" {
		t.Errorf("dominant category = %q, want krb5_realm (critical outranks others)", s.TopCategory)
	}
	if !strings.Contains(s.Headline, "CRITICAL") {
		t.Errorf("headline = %q, want a [CRITICAL] triage statement", s.Headline)
	}
}

func TestComputeExecutiveSummary_HealthyReport(t *testing.T) {
	r := &ReportData{}
	computeExecutiveSummary(r)
	if r.Summary.HealthScore != 100 {
		t.Errorf("clean report health score = %d, want 100", r.Summary.HealthScore)
	}
	if !strings.Contains(r.Summary.Headline, "HEALTHY") {
		t.Errorf("clean report headline = %q, want HEALTHY", r.Summary.Headline)
	}
}

// The unified KB config matching must find config patterns with a single
// scan of sssd.conf (functional parity with the per-pattern legacy scans).
func TestMatchKBArticlesWithEvidence_ConfigPattern(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"sssd.conf": "[domain/example.com]\nid_provider = ad\ncase_sensitive = true\n",
	})
	defer os.RemoveAll(dir)

	kb := []TIDArticle{
		{
			TIDID:          "TID-TEST-1",
			Title:          "Case sensitivity breaks lookups",
			LogPatterns:    []string{},
			ConfigPatterns: []string{"case_sensitive"},
		},
		{
			TIDID:          "TID-TEST-2",
			Title:          "Unrelated article",
			LogPatterns:    []string{},
			ConfigPatterns: []string{"nonexistent_option_xyz"},
		},
	}

	report := &ReportData{MACType: "None"}
	matchKBArticlesWithEvidence(dir, report, kb, map[string][]string{})

	if len(report.MatchedTIDs) != 1 || report.MatchedTIDs[0].TIDID != "TID-TEST-1" {
		t.Errorf("expected only TID-TEST-1 to match, got %+v", report.MatchedTIDs)
	}
}

// SELinux articles must still be suppressed on AppArmor systems after the
// rewrite of the matcher.
func TestMatchKBArticlesWithEvidence_AppArmorSuppression(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"sssd.conf": "[domain/example.com]\ncase_sensitive = true\n",
	})
	defer os.RemoveAll(dir)

	kb := []TIDArticle{
		{
			TIDID:          "TID-SELINUX-1",
			Title:          "SELinux blocks SSSD",
			ConfigPatterns: []string{"case_sensitive"},
		},
	}
	report := &ReportData{MACType: "AppArmor"}
	matchKBArticlesWithEvidence(dir, report, kb, map[string][]string{})

	if len(report.MatchedTIDs) != 0 {
		t.Errorf("SELinux article must be suppressed on AppArmor systems")
	}
}

func TestBuildJSONReport_ContainsSummaryAndFindings(t *testing.T) {
	r := ReportData{
		AppVersion: "0.2.0",
		ConfigFindings: []ConfigFinding{
			{Severity: SevCritical, Category: "krb5_realm", Message: "mismatch", SourcePath: "sssd.conf", SourceLine: 7},
		},
	}
	computeExecutiveSummary(&r)

	data, err := buildJSONReport(r)
	if err != nil {
		t.Fatalf("buildJSONReport failed: %v", err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON report is not valid JSON: %v", err)
	}
	if _, ok := decoded["summary"]; !ok {
		t.Errorf("JSON report missing 'summary' block")
	}
	if _, ok := decoded["config_findings"]; !ok {
		t.Errorf("JSON report missing 'config_findings' block")
	}
}

// The fingerprint-based cache must return the same automaton instance for an
// identical pattern set (cache-hit path).
func TestACCache_FingerprintReturnsSameInstance(t *testing.T) {
	patterns := []string{"alpha", "beta"}
	m1 := globalACCache.Get(patterns)
	m2 := globalACCache.Get([]string{"alpha", "beta"})
	if m1 != m2 {
		t.Errorf("ACCache returned different automaton instances for the same pattern set")
	}
	if m1.Len() != 2 {
		t.Errorf("ACCache pattern count = %d, want 2", m1.Len())
	}
}

func TestIsRegexPatternClassification(t *testing.T) {
	regexPatterns := []string{"Attribute .* not allowed", "Illegal ID \\d+ for search"}
	for _, p := range regexPatterns {
		if !isRegexPattern(p) {
			t.Errorf("pattern %q should be classified as regex", p)
		}
	}
	literalPatterns := []string{"Preauthentication failed", "[13][Permission denied]", "KVNO Principal"}
	for _, p := range literalPatterns {
		if isRegexPattern(p) {
			t.Errorf("pattern %q should be classified as literal", p)
		}
	}
}

// Ensure findings and the summary survive the PII scrubbing pipeline.
func TestAnonymizeReport_ScrubsSummaryHeadline(t *testing.T) {
	r := &ReportData{
		SearchDomain: "mycompany.com",
		Problems:     []string{"[CRITICAL] realm mismatch against mycompany.com"},
		ConfigFindings: []ConfigFinding{
			{Severity: SevCritical, Category: "join", Message: "no realm for MYCOMPANY.COM", Evidence: "ad_domain = mycompany.com"},
		},
	}
	computeExecutiveSummary(r)
	r.Summary.Headline = "[CRITICAL] Host mycompany.com has a broken realm"
	anonymizeReport(r)

	if strings.Contains(r.Summary.Headline, "mycompany.com") {
		t.Errorf("summary headline was not anonymized: %q", r.Summary.Headline)
	}
	for _, f := range r.ConfigFindings {
		if strings.Contains(f.Message, "mycompany.com") || strings.Contains(f.Message, "MYCOMPANY.COM") {
			t.Errorf("finding message was not anonymized: %q", f.Message)
		}
		if strings.Contains(f.Evidence, "mycompany.com") {
			t.Errorf("finding evidence was not anonymized: %q", f.Evidence)
		}
	}
}
