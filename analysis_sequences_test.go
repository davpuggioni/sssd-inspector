// analysis_sequences_test.go
package main

import (
	"strings"
	"testing"
)

// TestCategoryFor confirms the coarse root-cause classifier maps pattern
// descriptions (and free-text fallback) to the SSSD failure domains used by
// the sequence engine and the scoring pipeline.
func TestCategoryFor(t *testing.T) {
	cases := map[string]string{
		"DNS Resolution Error: SSSD could not resolve a server hostname.":     "dns",
		"SRV Lookup Failure: _ldap._tcp could not be resolved.":               "srv",
		"TLS Handshake Failure: ldap_install_tls failed (cert chain).":        "tls",
		"Kerberos: Preauthentication failed for user@AD.EXAMPLE.COM.":         "krb5",
		"Backend Offline: SSSD marked the data provider backend OFFLINE.":     "offline",
		"Keytab Principal Missing: no keytab entry for host.":                 "keytab",
		"ID Mapping Error: SID to UID translation failed.":                    "idmap",
		"GPO Permission Denied: machine account cannot read GPOs.":            "gpo",
		"AD Account Disabled: userAccountControl marks the account disabled.": "access",
		"AD CRYPTO BUG: service key not available.":                           "crypto",
	}
	for desc, want := range cases {
		if got := categoryFor(desc); got != want {
			t.Errorf("categoryFor(%q) = %q, want %q", desc, got, want)
		}
	}
	// Defensive keyword fallback for unknown-but-descriptive patterns.
	if got := categoryFor("Clock skew too great on request"); got != "krb5" {
		t.Errorf("categoryFor fallback = %q, want krb5", got)
	}
}

// TestCategoryFor_newPatterns verifies the patterns added from the SSSD C
// source (connect/SRV/offline) map to the domain they should fuel.
func TestCategoryFor_newPatterns(t *testing.T) {
	desc := "DNS SRV Timeout: Service (SRV) record resolution timed out."
	if got := categoryFor(desc); got != "srv" {
		t.Errorf("categoryFor(SRV timeout) = %q, want srv", got)
	}
	desc = "Network Error: SSSD could not establish a TCP connection."
	if got := categoryFor(desc); got != "net" {
		t.Errorf("categoryFor(establish connection) = %q, want net", got)
	}
	desc = "Backend Offline: SSSD marked the data provider backend as OFFLINE."
	if got := categoryFor(desc); got != "offline" {
		t.Errorf("categoryFor(offline) = %q, want offline", got)
	}
}

// TestWeightedLogBiasDominant verifies the scoring engine now seeds the
// dominant root-cause with log-error category weights (dns burst) rather than
// being overpowered by a long tail of benign config hints.
func TestWeightedLogBiasDominant(t *testing.T) {
	var report ReportData
	// A few benign config warnings in a non-dominant category.
	for i := 0; i < 6; i++ {
		addConfigFinding(&report, SevWarning, "enumerate", "[HINT] enumerate = true, tune it", "sssd.conf", "x", 0, "")
	}
	// A cluster of resolution errors should dominate.
	for i := 0; i < 8; i++ {
		report.SSSDLogErrors = append(report.SSSDLogErrors, SSSDLogError{Description: "DNS Resolution Error: SRV fallback failed.", Examples: []string{"Service resolving timeout reached"}})
	}
	computeExecutiveSummary(&report)
	if report.Summary.TopCategory != "dns" {
		t.Errorf("dominant category = %q, want dns (log-weighted root cause)", report.Summary.TopCategory)
	}
}

// TestHeadline_sequenceOverride verifies that when a correlated root-cause
// finding exists, buildHeadline surfaces it ahead of the generic headlines.
func TestHeadline_sequenceOverride(t *testing.T) {
	var report ReportData
	report.ConfigFindings = append(report.ConfigFindings, ConfigFinding{
		Severity: SevError,
		Category: "dns",
		Message:  "[CORRELATED ROOT CAUSE] DNS/SRV service discovery failure",
	})
	computeExecutiveSummary(&report)
	if !strings.Contains(report.Summary.Headline, "[ROOT CAUSE]") {
		t.Errorf("headline did not override to root cause: %q", report.Summary.Headline)
	}
}

// TestCorrelateSequences_DNSChain verifies that a DNS/SRV→connect→offline
// cascade detected in the sorted timeline is collapsed into a single synthetic
// root-cause finding instead of remaining N unrelated errors.
func TestCorrelateSequences_DNSChain(t *testing.T) {
	timeline := []TimelineEvent{
		{Timestamp: "2024-01-01 00:00:01", Message: "DNS SRV Timeout: Service (SRV) record resolution timed out.", RawLog: "sssd: Service resolving timeout reached: _kerberos._tcp.example.com"},
		{Timestamp: "2024-01-01 00:00:05", Message: "SRV Lookup Failure: _ldap._tcp could not be resolved.", RawLog: "sssd: Unable to resolve SRV [10]: Name or service not known"},
		{Timestamp: "2024-01-01 00:00:09", Message: "Network Error: SSSD could not establish a TCP connection to the backend server.", RawLog: "sssd: Unable to establish connection [10.0.0.1]: Connection refused"},
		{Timestamp: "2024-01-01 00:00:20", Message: "Backend Offline: SSSD marked the data provider backend as OFFLINE.", RawLog: "sssd: Going offline!"},
	}
	var report ReportData
	correlateSequences(timeline, &report)

	found := false
	for _, f := range report.ConfigFindings {
		if f.Category == "dns" {
			found = true
			if !containsString(report.Problems, f.Message) {
				t.Errorf("root-cause finding not mirrored into report.Problems")
			}
		}
	}
	if !found {
		t.Fatalf("expected a consolidated DNS/SRV root-cause finding, got %d findings", len(report.ConfigFindings))
	}
}

// TestCorrelateSequences_sparse verifies that a too-sparse timeline (below the
// minTimelineEvents floor) never produces a root-cause finding (no false pos).
func TestCorrelateSequences_sparse(t *testing.T) {
	timeline := []TimelineEvent{
		{Timestamp: "2024-01-01 00:00:01", Message: "DNS SRV Resolution: Service (SRV) record resolution timed out.", RawLog: "Service resolving timeout reached"},
	}
	var report ReportData
	correlateSequences(timeline, &report)
	if len(report.ConfigFindings) != 0 {
		t.Errorf("sparse timeline produced %d findings, want 0", len(report.ConfigFindings))
	}
}

// TestCorrelateSequences_offline collates the offline burst chain.
func TestCorrelateSequences_offline(t *testing.T) {
	timeline := []TimelineEvent{
		{Timestamp: "2024-01-01 00:00:01", Message: "Network Error: SSSD could not establish a TCP connection to the backend server.", RawLog: "unable to establish connection [10.0.0.1]:550"},
		{Timestamp: "2024-01-01 00:00:02", Message: "Backend Offline: data provider backend marked OFFLINE.", RawLog: "sssd: Going offline!"},
		{Timestamp: "2024-01-01 00:00:03", Message: "Offline Auth Failure: cached-credential login failed.", RawLog: "Offline authentication failed"},
	}
	var report ReportData
	correlateSequences(timeline, &report)
	found := false
	for _, f := range report.ConfigFindings {
		if f.Category == "offline" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected an offline root-cause finding in a burst")
	}
}

// isExcluded guards unrelated daemons that share SSSD vocabulary.
func TestIsExcluded(t *testing.T) {
	if !isExcluded("sssd: winbindd: failed to do something") {
		t.Error("expected winbindd line to be excluded")
	}
	if isExcluded("sssd: Going offline after retry") {
		t.Error("did not expect a real SSSD line to be excluded")
	}
}
