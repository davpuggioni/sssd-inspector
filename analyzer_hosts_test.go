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
	"strings"
	"testing"
)

// containsSubstring reports whether any element of list contains sub. The
// hosts tests assert on substrings because the exact wording of an issue is
// meant to change over time; what must not change is that the issue is raised.
func containsSubstring(list []string, sub string) bool {
	for _, s := range list {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// hostsBundle wraps the given /etc/hosts body in the section framing a real
// supportconfig uses. Without the "#==[ Configuration File ]" header the
// locator finds no marker and reports the file as not collected, which would
// make a hosts-content test pass or fail for the wrong reason.
func hostsBundle(body string) string {
	return "#==[ Configuration File ]===========================#\n" +
		"# /etc/hosts\n" +
		body +
		"\n#==[ Next File ]=====================================#\n"
}

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

// Case 5: a file that parses but has no loopback line is a configuration gap,
// not corruption. It must NOT be called malformed: the operator would go
// looking for a broken file when the fix is a single added line.
func TestAnalyzeHosts_MissingLoopbackIsNotMalformed(t *testing.T) {
	dir := writeTestSupportconfig(t, map[string]string{
		"network.txt": "# /etc/hosts\n192.0.2.10 testhost01.example.test\n",
	})
	report := &ReportData{}
	analyzeHosts(dir, report)

	if report.HostsFileStatus != HostsStatusPresent {
		t.Errorf("HostsFileStatus = %q, want %q (the file parses fine)", report.HostsFileStatus, HostsStatusPresent)
	}
	for _, p := range report.Problems {
		if strings.Contains(p, "Malformed") {
			t.Errorf("a parsable file without a loopback must not be called malformed, got %q", p)
		}
	}
	if len(report.Problems) != 0 {
		t.Errorf("a parsable file is not a breakage: expected no Problems, got %v", report.Problems)
	}
	if !containsSubstring(report.HostsIssues, "Missing loopback entry") {
		t.Errorf("expected the loopback gap in HostsIssues, got %v", report.HostsIssues)
	}
	// The gap must still reach the operator, as a warning rather than a
	// silently dropped observation.
	found := false
	for _, w := range report.Warnings {
		if strings.Contains(w, "no loopback entry") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a loopback warning, got %v", report.Warnings)
	}
}

// Case 6: a commented-out loopback line does not count as an entry, so the
// same gap is reported — still as a gap, not as corruption.
func TestAnalyzeHosts_CommentedLoopbackNotAccepted(t *testing.T) {
	dir := writeTestSupportconfig(t, map[string]string{
		"network.txt": "# /etc/hosts\n# 127.0.0.1 localhost\n192.0.2.10 testhost01.example.test\n",
	})
	report := &ReportData{}
	analyzeHosts(dir, report)

	if !containsSubstring(report.HostsIssues, "Missing loopback entry") {
		t.Errorf("commented-out 127.0.0.1 must still be flagged as a missing loopback, got %v", report.HostsIssues)
	}
	if report.HostsFileStatus != HostsStatusPresent {
		t.Errorf("HostsFileStatus = %q, want %q", report.HostsFileStatus, HostsStatusPresent)
	}
}

// Case 6b: content that cannot be read as /etc/hosts at all IS malformed.
// This is the only case that earns the word, so it is pinned explicitly.
func TestAnalyzeHosts_UnparsableContentIsMalformed(t *testing.T) {
	dir := writeTestSupportconfig(t, map[string]string{
		"network.txt": "# /etc/hosts\nthis is not a hosts file\nnor is this line\n",
	})
	report := &ReportData{}
	analyzeHosts(dir, report)

	if report.HostsFileStatus != HostsStatusMalformed {
		t.Errorf("HostsFileStatus = %q, want %q", report.HostsFileStatus, HostsStatusMalformed)
	}
	found := false
	for _, p := range report.Problems {
		if strings.Contains(p, "Malformed /etc/hosts file") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a malformed problem for unparsable content, got %v", report.Problems)
	}
}

// Case 6c: a line whose first field is not an address is malformed input, but
// one valid entry elsewhere keeps the file usable — so the corruption is
// reported as an issue without condemning the whole file.
func TestAnalyzeHosts_BadEntryAmongGoodOnesIsNotFatal(t *testing.T) {
	dir := writeTestSupportconfig(t, map[string]string{
		"network.txt": "# /etc/hosts\nnot-an-ip hostbadexample\n127.0.0.1 localhost\n",
	})
	report := &ReportData{}
	analyzeHosts(dir, report)

	if report.HostsFileStatus != HostsStatusPresent {
		t.Errorf("HostsFileStatus = %q, want %q (one good entry keeps the file usable)", report.HostsFileStatus, HostsStatusPresent)
	}
	if !containsSubstring(report.HostsIssues, "not-an-ip hostbadexample") {
		t.Errorf("the unusable line should be named in HostsIssues, got %v", report.HostsIssues)
	}
	// The loopback was found, so no gap is reported.
	for _, issue := range report.HostsIssues {
		if strings.Contains(issue, "Missing loopback entry") {
			t.Errorf("unexpected loopback gap: %v", report.HostsIssues)
		}
	}
}

// Case 6d: a loopback written in any of its equivalent textual forms must be
// recognised. The expanded IPv6 form and the rest of 127.0.0.0/8 are valid
// loopback entries that a string comparison would reject, and the operator
// would be told to fix a file that is already correct.
func TestAnalyzeHosts_LoopbackInAnyEquivalentFormIsAccepted(t *testing.T) {
	// The alias deliberately does not contain "localhost": an alias match would
	// set hasLocalback on its own and make the test pass regardless of whether
	// the address was recognised, which is exactly the bug being pinned.
	cases := []struct {
		name         string
		entry        string
		wantLoopback bool
	}{
		{"expanded IPv6", "0:0:0:0:0:0:0:1 ip6-localhost\n", true},
		{"short IPv6", "::1 ip6-localhost\n", true},
		{"upper 127.0.0.0/8", "127.0.0.2 loopback-host\n", true},
		{"IPv4 loopback", "127.0.0.1 loopback-host\n", true},
		// Control: a routable address must NOT be mistaken for a loopback.
		// Without this, a check that accepted everything would also pass.
		{"non-loopback control", "10.0.0.1 somehost\n", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := writeTestSupportconfig(t, map[string]string{
				"network.txt": hostsBundle(tc.entry),
			})
			report := &ReportData{}
			analyzeHosts(dir, report)

			if report.HostsFileStatus != HostsStatusPresent {
				t.Errorf("HostsFileStatus = %q, want %q", report.HostsFileStatus, HostsStatusPresent)
			}
			gotGap := containsSubstring(report.HostsIssues, "Missing loopback entry")
			if gotGap == tc.wantLoopback {
				t.Errorf("%q: loopback gap reported = %v, want %v (issues: %v)",
					tc.entry, gotGap, !tc.wantLoopback, report.HostsIssues)
			}
			for _, w := range report.Warnings {
				if strings.Contains(w, "no loopback entry") != !tc.wantLoopback {
					t.Errorf("loopback warning mismatch for %q: %v", tc.entry, w)
				}
			}
		})
	}
}

// Case 6e: a present but empty file cannot resolve anything, so it is
// malformed. There are no offending lines to quote here, which is a different
// message from the "unparsable lines" case and was previously untested.
func TestAnalyzeHosts_EmptyFileIsMalformed(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"completely empty", ""},
		{"only comments", "# nothing collected here\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := writeTestSupportconfig(t, map[string]string{
				"network.txt": hostsBundle(tc.body),
			})
			report := &ReportData{}
			analyzeHosts(dir, report)

			if report.HostsFileStatus != HostsStatusMalformed {
				t.Errorf("HostsFileStatus = %q, want %q", report.HostsFileStatus, HostsStatusMalformed)
			}
			// With no offending line to name, the message must say so rather
			// than quoting nothing.
			if !containsSubstring(report.HostsIssues, "no address entries could be parsed") {
				t.Errorf("expected the no-entries message, got %v", report.HostsIssues)
			}
		})
	}
}

// Case 6f: host names are case-insensitive, so an upper-case alias satisfies
// the loopback check just as a lower-case one does.
func TestAnalyzeHosts_LoopbackAliasIsCaseInsensitive(t *testing.T) {
	dir := writeTestSupportconfig(t, map[string]string{
		"network.txt": hostsBundle("10.0.0.5 HOSTNAME.EXAMPLE.COM\n192.168.1.5 LOCALHOST.LOCALDOMAIN\n"),
	})
	report := &ReportData{}
	analyzeHosts(dir, report)

	if report.HostsFileStatus != HostsStatusPresent {
		t.Errorf("HostsFileStatus = %q, want %q", report.HostsFileStatus, HostsStatusPresent)
	}
	if containsSubstring(report.HostsIssues, "Missing loopback entry") {
		t.Errorf("upper-case LOCALHOST.LOCALDOMAIN is a loopback alias, got %v", report.HostsIssues)
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
