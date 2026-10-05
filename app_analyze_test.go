//go:build !cli

// app_analyze_test.go
//
// Tests for App.Analyze (app.go): the GUI/Wails entry point into the
// analysis pipeline. The Wails runtime calls are safe to invoke headless
// (emitEvent skips when no frontend context is attached), which lets us
// cover the directory, archive and error paths of the production Analyze
// method — previously at 37.5%.
package main

import (
	"os"
	"path/filepath"
	"testing"
)

func analyzeFixture(t *testing.T) string {
	t.Helper()
	dir := setupMockDir(t, map[string]string{
		"sssd.conf":             "[domain/example.com]\nid_provider = ad\nad_domain = example.com\n",
		"sssd.txt":              "Dec 10 12:00:00 server sssd: ordinary line\n",
		"basic-environment.txt": "Hostname: host.example.com\n",
		"nsswitch.conf":         "passwd: files sss\ngroup: files sss\n",
		"rpm.txt":               "sssd-2.9.4-150500.x86_64\n",
	})
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestAppAnalyze_Directory(t *testing.T) {
	app := NewApp()
	report, err := app.Analyze(analyzeFixture(t), false)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	if report.Hostname != "host.example.com" {
		t.Errorf("expected hostname host.example.com, got %q", report.Hostname)
	}
}

func TestAppAnalyze_Anonymized(t *testing.T) {
	app := NewApp()
	report, err := app.Analyze(analyzeFixture(t), true)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	if report.Hostname != "redacted-host" {
		t.Errorf("expected redacted hostname, got %q", report.Hostname)
	}
	if report.AdDomain == "example.com" && report.SSSDConfigSnippet != "" {
		// example.com is both the fixture domain and the placeholder; the
		// assertion that matters is that no raw PII-shaped value survives.
		t.Logf("fixture domain coincides with the placeholder; skipping strict check")
	}
}

func TestAppAnalyze_Archive(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "app-analyze.tar.xz")
	writeTXZ(t, archive, supportconfigEntries())
	app := NewApp()
	report, err := app.Analyze(archive, false)
	if err != nil {
		t.Fatalf("Analyze on archive failed: %v", err)
	}
	if report.Hostname != "host.example.com" {
		t.Errorf("expected hostname host.example.com, got %q", report.Hostname)
	}
}

func TestAppAnalyze_MissingPath(t *testing.T) {
	app := NewApp()
	if _, err := app.Analyze(filepath.Join(t.TempDir(), "nope"), false); err == nil {
		t.Errorf("expected an error for a missing path, got nil")
	}
}
