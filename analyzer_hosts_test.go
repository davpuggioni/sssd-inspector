// analyzer_hosts_test.go
//
// Regression tests for /etc/hosts handling:
//   - missing from supportconfig (collection gap: no claim of "malformed")
//   - missing from analysed host ("# /etc/hosts - File not found" recorded)
//   - present and well-formed (inside network.txt or standalone hosts)
//   - present but malformed (missing loopback)
package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestSupportconfig(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		target := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			t.Fatalf("failed to mkdir: %v", err)
		}
		if err := os.WriteFile(target, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}
	return dir
}

// Case 1: When /etc/hosts is omitted entirely from the supportconfig,
// the tool must NOT report "Malformed /etc/hosts file."
func TestAnalyzeHosts_NotCollectedDoesNotReportMalformed(t *testing.T) {
	dir := writeTestSupportconfig(t, map[string]string{
		"network.txt": "#==[ Network ]=====================================#\n# config\n",
	})
	report := &ReportData{}
	analyzeHosts(dir, report)

	if report.HostsFileStatus != HostsStatusNotCollected {
		t.Errorf("HostsFileStatus = %q, want %q", report.HostsFileStatus, HostsStatusNotCollected)
	}
	if len(report.HostsIssues) != 0 {
		t.Errorf("expected no HostsIssues when not collected, got %v", report.HostsIssues)
	}
	for _, p := range report.Problems {
		if p == "Malformed /etc/hosts file." {
			t.Errorf("unexpected 'Malformed /etc/hosts file.' problem when file was simply not collected")
		}
	}
}

// Case 2: When supportconfig records "# /etc/hosts - File not found",
// it must be explicitly reported as missing from the host, NOT malformed.
func TestAnalyzeHosts_MissingOnHostExplicitProblem(t *testing.T) {
	dir := writeTestSupportconfig(t, map[string]string{
		"network.txt": "#==[ Configuration File ]===========================#\n# /etc/hosts - File not found\n\n#==[ Next File ]=====================================#\n",
	})
	report := &ReportData{}
	analyzeHosts(dir, report)

	if report.HostsFileStatus != HostsStatusMissingHost {
		t.Errorf("HostsFileStatus = %q, want %q", report.HostsFileStatus, HostsStatusMissingHost)
	}
	foundMissing := false
	for _, p := range report.Problems {
		if p == "Malformed /etc/hosts file." {
			t.Errorf("should NOT report 'Malformed /etc/hosts file.' when the file is missing from host")
		}
		if p == "[HOSTS] /etc/hosts is missing from the system. Without /etc/hosts, localhost and the local hostname rely entirely on DNS resolution." {
			foundMissing = true
		}
	}
	if !foundMissing {
		t.Errorf("expected explicit missing hosts problem, got %v", report.Problems)
	}
}

// Case 3: Present in network.txt with proper loopback -> clean, no problems.
func TestAnalyzeHosts_PresentInNetworkTxt(t *testing.T) {
	dir := writeTestSupportconfig(t, map[string]string{
		"network.txt": "#==[ Configuration File ]===========================#\n" +
			"# /etc/hosts\n" +
			"127.0.0.1 localhost\n" +
			"::1       localhost ipv6-localhost\n" +
			"192.0.2.10 testhost01.example.test testhost01\n" +
			"\n#==[ Next File ]=====================================#\n",
	})
	report := &ReportData{Hostname: "testhost01.example.test"}
	analyzeHosts(dir, report)

	if report.HostsFileStatus != HostsStatusPresent {
		t.Errorf("HostsFileStatus = %q, want %q", report.HostsFileStatus, HostsStatusPresent)
	}
	if len(report.HostsIssues) != 0 {
		t.Errorf("unexpected HostsIssues: %v", report.HostsIssues)
	}
	for _, p := range report.Problems {
		if p == "Malformed /etc/hosts file." {
			t.Errorf("unexpected malformed problem on clean hosts file")
		}
	}
}

// Case 4: Standalone hosts file (fallback for tests and raw directories)
func TestAnalyzeHosts_StandaloneFallback(t *testing.T) {
	dir := writeTestSupportconfig(t, map[string]string{
		"hosts": "127.0.0.1 localhost\n192.0.2.10 testhost01.example.test\n",
	})
	report := &ReportData{Hostname: "testhost01.example.test"}
	analyzeHosts(dir, report)

	if report.HostsFileStatus != HostsStatusPresent {
		t.Errorf("HostsFileStatus = %q, want %q", report.HostsFileStatus, HostsStatusPresent)
	}
	if len(report.HostsIssues) != 0 {
		t.Errorf("unexpected HostsIssues: %v", report.HostsIssues)
	}
}

// Case 5: Present but missing 127.0.0.1 / localhost -> reported as malformed
func TestAnalyzeHosts_MissingLoopbackIsMalformed(t *testing.T) {
	dir := writeTestSupportconfig(t, map[string]string{
		"network.txt": "# /etc/hosts\n192.0.2.10 testhost01.example.test\n",
	})
	report := &ReportData{}
	analyzeHosts(dir, report)

	if report.HostsFileStatus != HostsStatusPresent {
		t.Errorf("HostsFileStatus = %q, want %q", report.HostsFileStatus, HostsStatusPresent)
	}
	hasMalformed := false
	for _, p := range report.Problems {
		if p == "Malformed /etc/hosts file." {
			hasMalformed = true
		}
	}
	if !hasMalformed {
		t.Errorf("expected 'Malformed /etc/hosts file.' problem, got %v", report.Problems)
	}
}

// Case 6: Commented-out loopback line must not satisfy the check
func TestAnalyzeHosts_CommentedLoopbackNotAccepted(t *testing.T) {
	dir := writeTestSupportconfig(t, map[string]string{
		"network.txt": "# /etc/hosts\n# 127.0.0.1 localhost\n192.0.2.10 testhost01.example.test\n",
	})
	report := &ReportData{}
	analyzeHosts(dir, report)

	hasMalformed := false
	for _, p := range report.Problems {
		if p == "Malformed /etc/hosts file." {
			hasMalformed = true
		}
	}
	if !hasMalformed {
		t.Errorf("commented-out 127.0.0.1 must still flag malformed /etc/hosts, got %v", report.Problems)
	}
}

// Case 7: Hostname warning when host has an FQDN but hosts file omits it
func TestAnalyzeHosts_WarnsWhenHostNotInHosts(t *testing.T) {
	dir := writeTestSupportconfig(t, map[string]string{
		"network.txt": "# /etc/hosts\n127.0.0.1 localhost\n",
	})
	report := &ReportData{Hostname: "testhost01.example.test"}
	analyzeHosts(dir, report)

	foundWarning := false
	for _, w := range report.Warnings {
		if w == "[HOSTS] System hostname 'testhost01.example.test' is not mapped to an IP in /etc/hosts. Kerberos ticket issuance and AD join verification work best when the local hostname resolves consistently." {
			foundWarning = true
		}
	}
	if !foundWarning {
		t.Errorf("expected hostname warning in %v", report.Warnings)
	}
}
