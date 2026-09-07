// analysis_advanced_test.go
// Tests for Phase 2 (advanced AD/Kerberos/DNS checks), Phase 3 (temporal
// clustering + fuzzy KB suggestions) and Phase 5 (YAML rules engine).
package main

import (
	"os"
	"strings"
	"testing"
)

// --- Phase 2: Kerberos [libdefaults] encryption checks ---

func TestAnalyzeKerberosConfig_WeakCrypto(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"etc.txt": "#==[ Command ]==\n# /etc/krb5.conf\n[libdefaults]\n    default_realm = EXAMPLE.COM\n    allow_weak_crypto = true\n#==[ Command ]==\n",
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzeKerberosConfig(dir, &report)

	found := false
	for _, w := range report.Warnings {
		if strings.Contains(w, "allow_weak_crypto") {
			found = true
		}
	}
	if !found {
		t.Errorf("allow_weak_crypto was not detected: %+v", report.Warnings)
	}
	if report.KerberosRealm != "EXAMPLE.COM" {
		t.Errorf("default_realm parsing broken: %q", report.KerberosRealm)
	}
}

func TestAnalyzeKerberosConfig_RC4OnlyEnctypes(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"etc.txt": "#==[ Command ]==\n# /etc/krb5.conf\n[libdefaults]\n    default_tgs_enctypes = arcfour-hmac-md5 rc4-hmac\n#==[ Command ]==\n",
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzeKerberosConfig(dir, &report)

	found := false
	for _, w := range report.Warnings {
		if strings.Contains(w, "[AD CRYPTO]") && strings.Contains(w, "default_tgs_enctypes") {
			found = true
		}
	}
	if !found {
		t.Errorf("RC4-only enctypes were not detected: %+v", report.Warnings)
	}
}

// --- Phase 2: advanced AD section options ---

func TestValidateADAdvancedOptions_GPOInvalidValue(t *testing.T) {
	cfg := parseSssdConfig("[domain/example.com]\nid_provider = ad\nad_gpo_access_control = bogus\n")
	var report ReportData
	report.AdDomain = "example.com"
	validateADConfig(cfg, &report)

	found := false
	for _, f := range report.ConfigFindings {
		if f.Category == "gpo" {
			found = true
		}
	}
	if !found {
		t.Errorf("invalid ad_gpo_access_control was not detected")
	}
}

func TestValidateADAdvancedOptions_AdHostnameMismatch(t *testing.T) {
	cfg := parseSssdConfig("[domain/example.com]\nid_provider = ad\nad_hostname = wrong.example.com\n")
	var report ReportData
	report.AdDomain = "example.com"
	report.Hostname = "host01.example.com"
	validateADConfig(cfg, &report)

	found := false
	for _, f := range report.ConfigFindings {
		if f.Category == "ad_hostname" && strings.Contains(f.Message, "does not match the system hostname") {
			found = true
		}
	}
	if !found {
		t.Errorf("ad_hostname mismatch was not detected")
	}
}

func TestValidateADAdvancedOptions_MachineAccountMalformed(t *testing.T) {
	cfg := parseSssdConfig("[domain/example.com]\nid_provider = ad\nad_machine_account_password_renewal_opts = soon\n")
	var report ReportData
	report.AdDomain = "example.com"
	validateADConfig(cfg, &report)

	found := false
	for _, f := range report.ConfigFindings {
		if f.Category == "machine_account" {
			found = true
		}
	}
	if !found {
		t.Errorf("malformed machine account renewal opts were not detected")
	}
}

// --- Phase 2: loopback DNS resolver ---

func TestAnalyzeDNS_LoopbackNameserver(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"network.txt": "#==[ Command ]==\n# /etc/resolv.conf\nnameserver 127.0.0.1\nsearch example.com\n#==[ Command ]==\n",
	})
	defer os.RemoveAll(dir)

	var report ReportData
	analyzeDNS(dir, &report)

	found := false
	for _, w := range report.Warnings {
		if strings.Contains(w, "[DNS] Nameserver 127.0.0.1 is the loopback") {
			found = true
		}
	}
	if !found {
		t.Errorf("loopback nameserver was not detected: %+v", report.Warnings)
	}
}

// --- Phase 3: temporal sliding-window clusters ---

func TestAnalyzeTemporalClusters_BurstDetected(t *testing.T) {
	timeline := []TimelineEvent{
		{Timestamp: "2026-01-01 10:00:00", Message: "Kerberos: Clock skew too great", RawLog: "raw1"},
		{Timestamp: "2026-01-01 10:00:10", Message: "Kerberos: Clock skew too great", RawLog: "raw2"},
		{Timestamp: "2026-01-01 10:00:20", Message: "Kerberos: Clock skew too great", RawLog: "raw3"},
		// Far away in time: must NOT join the same cluster.
		{Timestamp: "2026-01-01 10:30:00", Message: "Kerberos: Clock skew too great", RawLog: "raw4"},
		// Different description: must not be clustered with the others.
		{Timestamp: "2026-01-01 10:00:30", Message: "Other event", RawLog: "raw5"},
	}
	clusters := analyzeTemporalClusters(timeline)

	if len(clusters) != 1 {
		t.Fatalf("expected 1 cluster, got %d: %+v", len(clusters), clusters)
	}
	c := clusters[0]
	if c.EventCount != 3 {
		t.Errorf("cluster count = %d, want 3", c.EventCount)
	}
	if c.WindowStart != "2026-01-01 10:00:00" || c.WindowEnd != "2026-01-01 10:00:20" {
		t.Errorf("window = %s .. %s, unexpected", c.WindowStart, c.WindowEnd)
	}
	if c.SampleRawLog != "raw1" {
		t.Errorf("sample = %q, want raw1", c.SampleRawLog)
	}
}

func TestAnalyzeTemporalClusters_NoBurst(t *testing.T) {
	timeline := []TimelineEvent{
		{Timestamp: "2026-01-01 10:00:00", Message: "Event", RawLog: "a"},
		{Timestamp: "2026-01-01 10:00:40", Message: "Event", RawLog: "b"},
	}
	if clusters := analyzeTemporalClusters(timeline); len(clusters) != 0 {
		t.Errorf("two isolated events must not form a cluster: %+v", clusters)
	}
}

// --- Phase 3: TF-IDF fuzzy KB suggestions ---

func TestAnalyzeKBSuggestions_SimilarArticle(t *testing.T) {
	kb := []TIDArticle{
		{
			TIDID:       "TID-100",
			Title:       "Dynamic DNS updates fail with TSIG verify failure",
			Description: "nsupdate dynamic DNS update rejected by the AD DNS server because of a TSIG signature verification failure",
		},
		{
			TIDID:       "TID-200",
			Title:       "Totally unrelated topic",
			Description: "How to configure autofs maps for NFS home directories",
		},
	}
	timeline := []TimelineEvent{
		{Timestamp: "2026-01-01 10:00:00", Message: "Unmatched event", RawLog: "sssd[be[x]]: tsig verify failure on dns update request for host01.example.com"},
	}
	suggestions := analyzeKBSuggestions(timeline, kb, nil)

	if len(suggestions) == 0 {
		t.Fatalf("expected at least one suggestion")
	}
	if suggestions[0].TIDID != "TID-100" {
		t.Errorf("top suggestion = %s, want TID-100", suggestions[0].TIDID)
	}
	if suggestions[0].Score <= 0 || suggestions[0].Score > 1 {
		t.Errorf("similarity score out of range: %v", suggestions[0].Score)
	}
}

func TestAnalyzeKBSuggestions_SkipsAlreadyMatched(t *testing.T) {
	kb := []TIDArticle{
		{TIDID: "TID-100", Title: "Dynamic DNS update TSIG verify failure and nsupdate"},
	}
	timeline := []TimelineEvent{
		{RawLog: "tsig verify failure on dynamic dns update request"},
	}
	matched := []TIDArticle{{TIDID: "TID-100"}}
	if s := analyzeKBSuggestions(timeline, kb, matched); len(s) != 0 {
		t.Errorf("already-matched article must not be suggested again: %+v", s)
	}
}

func TestAnalyzeKBSuggestions_EmptyInputs(t *testing.T) {
	if s := analyzeKBSuggestions(nil, nil, nil); s != nil {
		t.Errorf("empty inputs must yield no suggestions")
	}
}

// --- Phase 5: YAML rules engine ---

func TestApplyAnalysisRules_LiteralAny(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"sssd.conf": "[domain/example.com]\nid_provider = ad\nignore_group_members = false\n",
	})
	defer os.RemoveAll(dir)

	rules := []AnalysisRule{
		{
			Name:     "group-members-not-ignored",
			Severity: "warning",
			Category: "tuning",
			Files:    []string{"sssd.conf"},
			Patterns: []string{"ignore_group_members = false"},
			Message:  "Large AD groups are fully expanded; consider 'ignore_group_members = True'.",
		},
	}
	var report ReportData
	applyAnalysisRules(dir, &report, rules)

	if len(report.ConfigFindings) != 1 {
		t.Fatalf("expected 1 rule finding, got %d", len(report.ConfigFindings))
	}
	f := report.ConfigFindings[0]
	if !strings.Contains(f.Message, "[RULE: group-members-not-ignored]") {
		t.Errorf("finding message = %q", f.Message)
	}
	if f.Severity != SevWarning {
		t.Errorf("severity = %v, want SevWarning", f.Severity)
	}
	if f.Evidence == "" {
		t.Errorf("evidence line was not captured")
	}
}

func TestApplyAnalysisRules_RegexAllModeNotTriggered(t *testing.T) {
	// match: all requires BOTH patterns on the SAME line; only one is present.
	dir := setupMockDir(t, map[string]string{
		"sssd.txt": "(2026-01-01 10:00:00) sssd: ldap id processing context init failed\n",
	})
	defer os.RemoveAll(dir)

	rules := []AnalysisRule{
		{
			Name:        "combo",
			Severity:    "error",
			Category:    "combo",
			PatternType: "regex",
			Match:       "all",
			Patterns:    []string{`ldap id processing .* failed`, `totally absent token`},
			Message:     "must not fire",
		},
	}
	var report ReportData
	applyAnalysisRules(dir, &report, rules)

	if len(report.ConfigFindings) != 0 {
		t.Errorf("all-match rule fired on partial match: %+v", report.ConfigFindings)
	}
}

func TestApplyAnalysisRules_RegexAllModeTriggered(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"sssd.txt": "(2026-01-01 10:00:00) sssd: ldap id processing context init failed\n",
	})
	defer os.RemoveAll(dir)

	rules := []AnalysisRule{
		{
			Name:        "combo",
			Severity:    "error",
			Category:    "combo",
			PatternType: "regex",
			Match:       "all",
			Patterns:    []string{`ldap id processing .* failed`, `context`},
			Message:     "both tokens present on the same line",
		},
	}
	var report ReportData
	applyAnalysisRules(dir, &report, rules)

	if len(report.ConfigFindings) != 1 {
		t.Fatalf("all-match rule did not fire: %+v", report.ConfigFindings)
	}
}

func TestRuleSeverity_Mapping(t *testing.T) {
	cases := map[string]struct {
		want Severity
		ok   bool
	}{
		"critical": {SevCritical, true},
		"error":    {SevError, true},
		"warning":  {SevWarning, true},
		"warn":     {SevWarning, true},
		"bogus":    {SevWarning, false},
	}
	for in, want := range cases {
		got, ok := ruleSeverity(in)
		if got != want.want || ok != want.ok {
			t.Errorf("ruleSeverity(%q) = (%v, %v), want (%v, %v)", in, got, ok, want.want, want.ok)
		}
	}
}

func TestLoadAnalysisRules_FromWorkingDirectory(t *testing.T) {
	dir := t.TempDir()
	content := `rules:
  - name: "test-rule"
    severity: "error"
    category: "test"
    message: "test message"
    patterns:
      - "marker-token"
  - name: "bad-severity"
    severity: "loud"
    message: "should be skipped"
    patterns:
      - "x"
  - name: "no-patterns"
    severity: "error"
    message: "should be skipped"
`
	if err := os.WriteFile(dir+"/rules.yaml", []byte(content), 0644); err != nil {
		t.Fatalf("failed to write rules.yaml: %v", err)
	}

	oldWd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	defer os.Chdir(oldWd)

	rules := loadAnalysisRules()
	if len(rules) != 1 {
		t.Fatalf("expected 1 valid rule, got %d: %+v", len(rules), rules)
	}
	if rules[0].Name != "test-rule" {
		t.Errorf("loaded rule = %q", rules[0].Name)
	}
}
