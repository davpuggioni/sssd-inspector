// rules_engine_unified_test.go
//
// Regression tests for the single-pass rules engine (applyAnalysisRules):
// rules that share a file set must be evaluated in ONE scan with a posting
// list fanning canonical AC hits out to (rule, position) pairs, without
// changing any observable semantics of the old one-scan-per-rule loop.
package main

import (
	"testing"
)

// TestApplyRulesUnified_MultiRuleSingleGroup pins the core win: several
// literal rules over the same default file set all fire, each attributed to
// its own RuleID/evidence, in input order.
func TestApplyRulesUnified_MultiRuleSingleGroup(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"sssd.conf": "[domain/example.test]\nid_provider = ad\n",
		"sssd.txt":  "(2026-01-01 10:00:00) sssd: first-marker here\n(2026-01-01 10:00:01) sssd: second-marker here\n",
	})
	rules := []AnalysisRule{
		{Name: "r-first", Severity: "warning", Category: "cat-a", Message: "first", Patterns: []string{"first-marker"}},
		{Name: "r-second", Severity: "error", Category: "cat-b", Message: "second", Patterns: []string{"second-marker"}},
		{Name: "r-absent", Severity: "warning", Category: "cat-c", Message: "absent", Patterns: []string{"never-present-marker"}},
	}
	var report ReportData
	applyAnalysisRules(dir, &report, rules)

	if len(report.ConfigFindings) != 2 {
		t.Fatalf("expected exactly 2 findings, got %d: %+v", len(report.ConfigFindings), report.ConfigFindings)
	}
	if report.ConfigFindings[0].RuleID != "rule:r-first" || report.ConfigFindings[1].RuleID != "rule:r-second" {
		t.Errorf("findings must be emitted in input-rule order, got %s, %s",
			report.ConfigFindings[0].RuleID, report.ConfigFindings[1].RuleID)
	}
	if report.ConfigFindings[0].Evidence != "(2026-01-01 10:00:00) sssd: first-marker here" {
		t.Errorf("wrong evidence for r-first: %q", report.ConfigFindings[0].Evidence)
	}
	if report.ConfigFindings[1].Severity != SevError {
		t.Errorf("r-second must keep its own severity, got %v", report.ConfigFindings[1].Severity)
	}
}

// TestApplyRulesUnified_SatisfiedRuleDoesNotStarveOthers: in the unified scan
// a satisfied rule must stop consuming marks, but the remaining rules must
// keep being evaluated on LATER lines.
func TestApplyRulesUnified_SatisfiedRuleDoesNotStarveOthers(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"sssd.txt": "line1 early-bird found\nline2 nothing\nline3 late-bird found\n",
	})
	rules := []AnalysisRule{
		{Name: "early", Severity: "warning", Category: "c", Message: "e", Patterns: []string{"early-bird"}},
		{Name: "late", Severity: "warning", Category: "c", Message: "l", Patterns: []string{"late-bird"}},
	}
	var report ReportData
	applyAnalysisRules(dir, &report, rules)
	if len(report.ConfigFindings) != 2 {
		t.Fatalf("both rules must fire, got %d: %+v", len(report.ConfigFindings), report.ConfigFindings)
	}
	if report.ConfigFindings[1].Evidence != "line3 late-bird found" {
		t.Errorf("late rule must capture its own (later) evidence, got %q", report.ConfigFindings[1].Evidence)
	}
}

// TestApplyRulesUnified_MatchAllIsPerLineInSharedGroup: match:all must still
// require every pattern on the SAME line even when other rules in the group
// are marked on different lines.
func TestApplyRulesUnified_MatchAllIsPerLineInSharedGroup(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"sssd.txt": "alpha-only and noise\nmore noise beta-only\n",
	})
	rules := []AnalysisRule{
		{Name: "split", Severity: "warning", Category: "c", Message: "m", Match: "all", Patterns: []string{"alpha-only", "beta-only"}},
		{Name: "other", Severity: "warning", Category: "c", Message: "o", Patterns: []string{"noise"}},
	}
	var report ReportData
	applyAnalysisRules(dir, &report, rules)
	found := false
	for _, f := range report.ConfigFindings {
		if f.RuleID == "rule:split" {
			t.Errorf("match:all must not fire when patterns are on different lines: %+v", f)
		}
		if f.RuleID == "rule:other" {
			found = true
		}
	}
	if !found {
		t.Errorf("the any-mode rule in the same group must still fire: %+v", report.ConfigFindings)
	}
}

// TestApplyRulesUnified_DistinctFileSetsStaySeparate: a rule scoped to
// krb5.conf must evaluate only that file even though other rules travel
// through the same engine (one scan per distinct file set).
func TestApplyRulesUnified_DistinctFileSetsStaySeparate(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"sssd.conf": "# note\n[domain/example.test]\nid_provider = ad\n",
		"krb5.conf": "[libdefaults]\n default_realm = EXAMPLE.TEST\n",
	})
	rules := []AnalysisRule{
		{Name: "on-sssd", Severity: "warning", Category: "c", Message: "m", Files: []string{"sssd.conf"}, Patterns: []string{"id_provider"}},
		{Name: "on-krb5", Severity: "warning", Category: "c", Message: "m", Files: []string{"krb5.conf"}, Patterns: []string{"default_realm"}},
		{Name: "on-krb5-bis", Severity: "warning", Category: "c", Message: "m", Files: []string{"krb5.conf"}, Patterns: []string{"EXAMPLE"}},
	}
	var report ReportData
	applyAnalysisRules(dir, &report, rules)
	if len(report.ConfigFindings) != 3 {
		t.Fatalf("expected 3 findings (two sharing the krb5 group), got %d: %+v",
			len(report.ConfigFindings), report.ConfigFindings)
	}
	if report.ConfigFindings[1].RuleID != "rule:on-krb5" || report.ConfigFindings[2].RuleID != "rule:on-krb5-bis" {
		t.Errorf("krb5-group rules must both fire with input order preserved: %s, %s",
			report.ConfigFindings[1].RuleID, report.ConfigFindings[2].RuleID)
	}
}

// TestApplyRulesUnified_MixedLiteralAndRegexInOneGroup: a literal rule and a
// regex rule sharing the default file set must both fire from the same pass.
func TestApplyRulesUnified_MixedLiteralAndRegexInOneGroup(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"sssd.txt": "(2026-01-01 10:00:00) sssd: literal-token present\n(2026-01-01 10:00:01) sssd: xyzzy 12345\n",
	})
	rules := []AnalysisRule{
		{Name: "lit", Severity: "warning", Category: "c", Message: "m", Patterns: []string{"literal-token"}},
		{Name: "re", Severity: "error", Category: "c", Message: "m", PatternType: "regex", Match: "all", Patterns: []string{"xyzzy \\d+", "12345"}},
	}
	var report ReportData
	applyAnalysisRules(dir, &report, rules)
	if len(report.ConfigFindings) != 2 {
		t.Fatalf("literal and regex rules in one group must both fire, got %d: %+v",
			len(report.ConfigFindings), report.ConfigFindings)
	}
	if report.ConfigFindings[1].RuleID != "rule:re" || report.ConfigFindings[1].Severity != SevError {
		t.Errorf("regex rule lost its identity/severity: %+v", report.ConfigFindings[1])
	}
}

// TestApplyRulesUnified_EmptyRuleListAndNoMatch: the engine must be silent on
// empty input and on rules that match nothing (no stray findings).
func TestApplyRulesUnified_EmptyRuleListAndNoMatch(t *testing.T) {
	dir := setupMockDir(t, map[string]string{"sssd.conf": "[sssd]\nservices = nss, pam\n"})
	var report ReportData
	applyAnalysisRules(dir, &report, nil)
	applyAnalysisRules(dir, &report, []AnalysisRule{
		{Name: "nope", Severity: "warning", Category: "c", Message: "m", Patterns: []string{"not-in-file"}},
	})
	if len(report.ConfigFindings) != 0 {
		t.Errorf("expected no findings, got %+v", report.ConfigFindings)
	}
}
