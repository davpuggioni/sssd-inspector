// analysis_rootcause_test.go
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRootCauseBreakdown_GroupsSignals verifies config and log signals that
// point at the same coarse root cause are grouped into a single row (the P4a
// "one root cause, many symptoms" view) rather than a flat list.
func TestRootCauseBreakdown_GroupsSignals(t *testing.T) {
	var r ReportData
	// Two DNS-related config findings and two DNS log errors -> one "dns" row.
	r.ConfigFindings = append(r.ConfigFindings,
		ConfigFinding{Severity: SevError, Category: "dns", Message: "search domain mismatch"},
		ConfigFinding{Severity: SevError, Category: "dns", Message: "resolv.conf has no nameservers"},
	)
	r.SSSDLogErrors = append(r.SSSDLogErrors,
		SSSDLogError{Description: "DNS Resolution Error: SRV fallback failed."},
		SSSDLogError{Description: "SRV Lookup Failure: unable to resolve _ldap._tcp."},
	)

	rows := rootCauseBreakdown(r)
	// Expect exactly two groups: dns (config+log) and srv (log only).
	if len(rows) != 2 {
		t.Fatalf("expected 2 grouped rows, got %d: %+v", len(rows), rows)
	}
	var dnsRow, srvRow *rootCauseGroup
	for i := range rows {
		switch rows[i].Category {
		case "dns":
			dnsRow = &rows[i]
		case "srv":
			srvRow = &rows[i]
		}
	}
	if dnsRow == nil || srvRow == nil {
		t.Fatalf("expected both 'dns' and 'srv' groups, got: %+v", rows)
	}
	if len(dnsRow.ConfigSignals) != 2 {
		t.Errorf("dns config signals = %d, want 2", len(dnsRow.ConfigSignals))
	}
	if len(dnsRow.LogSignals) != 1 {
		t.Errorf("dns log signals = %d, want 1", len(dnsRow.LogSignals))
	}
	if len(srvRow.LogSignals) != 1 {
		t.Errorf("srv log signals = %d, want 1", len(srvRow.LogSignals))
	}
	// The heavier dns group (config findings) must rank above srv.
	if rows[0].Category != "dns" {
		t.Errorf("top-ranked root cause = %q, want dns", rows[0].Category)
	}
}

// TestBuildTextReport_IncludesRootCause verifies the plain-text report now
// contains the grouped root-cause breakdown section.
func TestBuildTextReport_IncludesRootCause(t *testing.T) {
	r := ReportData{
		ConfigFindings: []ConfigFinding{
			{Severity: SevCritical, Category: "join", Message: "machine account missing"},
		},
	}
	txt := buildTextReport(r)
	if !strings.Contains(txt, "ROOT-CAUSE BREAKDOWN") {
		t.Error("text report missing ROOT-CAUSE BREAKDOWN header")
	}
	if !strings.Contains(txt, "join") {
		t.Error("text report root-cause breakdown missing 'join' category")
	}
}

// TestWriteHTMLReportFile_RootBreakdown verifies the HTML template renders
// successfully with the new rootBreakdown section and produces a valid file.
func TestWriteHTMLReportFile_RootBreakdown(t *testing.T) {
	dir := t.TempDir()
	fname := filepath.Join(dir, "report.html")
	r := ReportData{
		AppVersion: "test",
		ConfigFindings: []ConfigFinding{
			{Severity: SevError, Category: "dns", Message: "resolv.conf broken"},
		},
	}
	writeHTMLReportFile(r, fname)
	data, err := os.ReadFile(fname)
	if err != nil {
		t.Fatalf("failed to read generated HTML: %v", err)
	}
	if !strings.Contains(string(data), "Root-Cause Breakdown") {
		t.Error("generated HTML missing Root-Cause Breakdown section")
	}
	if !strings.Contains(string(data), "Search findings & root causes") {
		t.Error("generated HTML missing interactive search field")
	}
	if !strings.Contains(string(data), "Severity:") {
		t.Error("generated HTML missing severity filter buttons")
	}
}

// TestWriteHTMLReportFile_P6Sections verifies the P6 report upgrades:
// table of contents, incident cards, and the bounded event-timeline table
// render alongside the legacy sections.
func TestWriteHTMLReportFile_P6Sections(t *testing.T) {
	dir := t.TempDir()
	fname := filepath.Join(dir, "report.html")
	r := ReportData{
		AppVersion: "test",
		ConfigFindings: []ConfigFinding{
			{Severity: SevError, Category: "dns", Message: "resolv.conf broken"},
		},
		Timeline: []TimelineEvent{
			{Timestamp: "2024-01-01 00:00:01", Message: "Backend Offline: marked OFFLINE.", RawLog: "sssd: Going offline!", Occurrences: 7},
		},
		TemporalClusters: []TemporalCluster{
			{Description: "Backend Offline: marked OFFLINE.", EventCount: 7, WindowStart: "2024-01-01 00:00:01", WindowEnd: "2024-01-01 00:05:00"},
		},
	}
	writeHTMLReportFile(r, fname)
	data, err := os.ReadFile(fname)
	if err != nil {
		t.Fatalf("failed to read generated HTML: %v", err)
	}
	html := string(data)
	for _, frag := range []string{
		"Contents",
		"Root-Cause Incidents",
		"Root-Cause Breakdown",
		"Event Timeline",
		"7 total occurrence(s)",
		"Temporal Clusters",
		"id=\"triage\"",
		"id=\"incidents\"",
		"id=\"timeline\"",
	} {
		if !strings.Contains(html, frag) {
			t.Errorf("generated HTML missing %q", frag)
		}
	}
}

// TestAnalyzeData_EndToEnd_RootCause is a full-pipeline regression test: it
// feeds a mock supportconfig (sssd.conf + a messages log carrying a DNS/SRV
// failover cascade) through analyzeData and asserts the correlated root-cause
// finding, the grouped breakdown, and the interactive HTML all come out wired.
func TestAnalyzeData_EndToEnd_RootCause(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"sssd.conf": "[sssd]\nservices = nss, pam, ssh, sudo\nconfig_file_version = 2\ndomains = example.com\n\n[domain/example.com]\nid_provider = ad\nad_domain = example.com\n\n",
		"messages": "" +
			"Feb 14 12:00:01 host sssd[be[example.com]]: Service resolving timeout reached for _kerberos._tcp.example.com\n" +
			"Feb 14 12:00:05 host sssd[be[example.com]]: Unable to resolve SRV [10]: Name or service not known\n" +
			"Feb 14 12:00:09 host sssd[be[example.com]]: Unable to establish connection [10.0.0.5]: Connection refused\n" +
			"Feb 14 12:00:20 host sssd[be[example.com]]: Going offline!\n",
	})
	defer os.RemoveAll(dir)

	report := analyzeData(dir, false, nil)

	// 1. The sequence engine must have collapsed the chain into a "dns" root cause.
	foundRC := false
	for _, f := range report.ConfigFindings {
		if f.Category == "dns" && strings.Contains(f.Message, "[CORRELATED ROOT CAUSE]") {
			foundRC = true
		}
	}
	if !foundRC {
		t.Errorf("DNS failover chain was not collapsed into a correlated root-cause finding")
	}

	// 2. The grouped breakdown must surface the correlated root cause ("dns") with
	// its config signal, and the failover log signals must appear under the
	// DNS-family categories (dns/srv/net/offline) — never vanish.
	rows := rootCauseBreakdown(report)
	dnsConfigSeen, cascadeLogSeen := false, false
	for _, row := range rows {
		if row.Category == "dns" && len(row.ConfigSignals) > 0 {
			dnsConfigSeen = true
		}
		switch row.Category {
		case "dns", "srv", "net", "offline":
			if len(row.LogSignals) > 0 {
				cascadeLogSeen = true
			}
		}
	}
	if !dnsConfigSeen {
		t.Errorf("root-cause breakdown missing the correlated 'dns' config signal")
	}
	if !cascadeLogSeen {
		t.Errorf("root-cause breakdown missing failover log signals under dns/srv/net/offline")
	}

	// 3. The text report must expose the grouped breakdown.
	if !strings.Contains(buildTextReport(report), "ROOT-CAUSE BREAKDOWN") {
		t.Errorf("text report missing ROOT-CAUSE BREAKDOWN section")
	}

	// 4. HTML renders with the breakdown + interactive toolbar.
	dir2 := t.TempDir()
	fname := filepath.Join(dir2, "report.html")
	writeHTMLReportFile(report, fname)
	html, err := os.ReadFile(fname)
	if err != nil {
		t.Fatalf("read html: %v", err)
	}
	for _, frag := range []string{"Root-Cause Breakdown", "Search findings & root causes", "Severity:"} {
		if !strings.Contains(string(html), frag) {
			t.Errorf("html missing %q", frag)
		}
	}
}
