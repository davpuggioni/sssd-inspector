// interfaces_test.go
//
// Tests for the DI wiring layer (interfaces.go, 0% coverage): the default
// implementations must delegate to the engine functions, and the default
// analysis engine must run the full pipeline with and without a progress
// reporter.
package main

import (
	"os"
	"strings"
	"testing"
)

// stubProgress is a ProgressReporter implementation for tests.
type stubProgress struct {
	calls [][2]interface{}
}

func (s *stubProgress) ReportProgress(message string, percentage int) {
	s.calls = append(s.calls, [2]interface{}{message, percentage})
}

func TestDefaultFileScanner_Delegates(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"probe.txt": "# /etc/krb5.conf\n default_realm = X\n",
	})
	defer os.RemoveAll(dir)

	if !defaultFileScanner.AnyFileContains(dir, []string{"probe.txt"}, "default_realm") {
		t.Errorf("AnyFileContains delegation failed")
	}
	if got := defaultFileScanner.ReadFileSafe(dir, "probe.txt"); !strings.Contains(got, "default_realm") {
		t.Errorf("ReadFileSafe delegation failed: %q", got)
	}
	if got := defaultFileScanner.ExtractSection(dir, "probe.txt", "# /etc/krb5.conf"); !strings.Contains(got, "default_realm") {
		t.Errorf("ExtractSection delegation failed: %q", got)
	}
	called := false
	if err := defaultFileScanner.ScanFiles(dir, []string{"probe.txt"}, func(string) { called = true }); err != nil || !called {
		t.Errorf("ScanFiles delegation failed: err=%v called=%v", err, called)
	}
}

func TestDefaultArchiveExtractor_Delegates(t *testing.T) {
	archive := t.TempDir() + "/probe.tar.xz"
	writeTXZ(t, archive, supportconfigEntries())
	dir, err := defaultArchiveExtractor.ExtractArchiveToTemp(archive, nil)
	if err != nil {
		t.Fatalf("extraction delegation failed: %v", err)
	}
	defer os.RemoveAll(dir)
	if _, err := os.Stat(dir + "/sssd.txt"); err != nil {
		t.Errorf("expected sssd.txt after extraction: %v", err)
	}
}

func TestDefaultReportGenerator_Delegates(t *testing.T) {
	report := ReportData{AppVersion: "test", Timestamp: "now"}
	if txt := defaultReportGenerator.BuildTextReport(report); !strings.Contains(txt, "test") {
		t.Errorf("BuildTextReport delegation produced no app version")
	}
	out := t.TempDir() + "/r.html"
	defaultReportGenerator.WriteHTMLReportFile(report, out)
	if _, err := os.Stat(out); err != nil {
		t.Errorf("WriteHTMLReportFile delegation failed: %v", err)
	}
}

func TestDefaultKBMatcher_Delegates(t *testing.T) {
	dir := setupMockDir(t, map[string]string{"messages": "nothing relevant\n"})
	defer os.RemoveAll(dir)
	var report ReportData
	defaultKBArticleMatcher.MatchKBArticles(dir, &report) // must not panic on empty KB dir
}

func TestNewDefaultAnalysisEngine_FullPipeline(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"sssd.conf":             "[domain/example.com]\nid_provider = ad\nad_domain = example.com\n",
		"sssd.txt":              "Dec 10 12:00:00 server sssd: ordinary line\n",
		"basic-environment.txt": "Hostname: host.example.com\n",
		"nsswitch.conf":         "passwd: files sss\ngroup: files sss\n",
		"rpm.txt":               "sssd-2.9.4-150500.x86_64\n",
	})
	defer os.RemoveAll(dir)

	engine := NewDefaultAnalysisEngine()
	if engine == nil {
		t.Fatalf("NewDefaultAnalysisEngine returned nil")
	}
	prog := &stubProgress{}
	report := engine.Analyze(dir, false, prog)
	if report.Hostname != "host.example.com" {
		t.Errorf("engine pipeline did not populate hostname: %q", report.Hostname)
	}
	if len(prog.calls) == 0 {
		t.Errorf("engine must forward progress reports")
	}

	// Nil progress reporter must be tolerated (used by headless callers).
	report2 := engine.Analyze(dir, true, nil)
	if report2.Hostname != "redacted-host" {
		t.Errorf("expected anonymized hostname, got %q", report2.Hostname)
	}
}
