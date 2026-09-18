//go:build !cli

// main_gui_entry_test.go
//
// Unit tests for the hybrid binary entry point (runHybridCLI).
//
// The critical property under test is the handled/not-handled contract: an
// invocation that was handled must never reach launchGUI() (which needs a
// display server), and an invocation with nothing to do must fall through to
// the GUI. main() itself is covered end-to-end by main_coverage_test.go.
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sssd-inspector/constants"
)

// TestRunHybridCLI_NoArgumentsLaunchesGUI: handled == false is what makes
// main() call launchGUI(). If this ever returns true, the GUI would never
// start when the user double-clicks the binary.
func TestRunHybridCLI_NoArgumentsLaunchesGUI(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code, handled := runHybridCLI(nil, &stdout, &stderr)
	if handled {
		t.Errorf("a bare invocation must not be handled by the CLI dispatcher (handled=%v)", handled)
	}
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Errorf("no output expected before the GUI starts, got stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

// TestRunHybridCLI_VersionFlagReportsHybridFlavour.
func TestRunHybridCLI_VersionFlagReportsHybridFlavour(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code, handled := runHybridCLI([]string{"-v"}, &stdout, &stderr)
	if !handled {
		t.Fatalf("-v must be handled by the CLI dispatcher")
	}
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), constants.AppVersion) {
		t.Errorf("stdout must report the version, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Hybrid") {
		t.Errorf("the hybrid binary must identify itself as Hybrid, got %q", stdout.String())
	}
}

// TestRunHybridCLI_AnalyzeIsHandledWithoutGUI is the regression guard for the
// "CLI invocation opens a window" class of bug.
func TestRunHybridCLI_AnalyzeIsHandledWithoutGUI(t *testing.T) {
	dir := writeSupportconfigFixture(t)
	work := t.TempDir()
	chdir(t, work)

	var stdout, stderr bytes.Buffer
	code, handled := runHybridCLI([]string{"-analyze", dir, "-json"}, &stdout, &stderr)
	if !handled {
		t.Fatalf("-analyze must be handled by the CLI dispatcher, not by launchGUI()")
	}
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %q)", code, stderr.String())
	}
	if _, err := os.Stat(findReport(t, work, "_report.json")); err != nil {
		t.Errorf("JSON report not written: %v", err)
	}
}

// TestRunHybridCLI_AnalyzeFailureIsExitCode1: the historical implementation used
// log.Fatalf here (exit code 1); the refactor must keep that observable contract.
func TestRunHybridCLI_AnalyzeFailureIsExitCode1(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")
	var stdout, stderr bytes.Buffer
	code, handled := runHybridCLI([]string{"-analyze", missing, "-txt"}, &stdout, &stderr)
	if !handled {
		t.Fatalf("-analyze with a bad path must be handled")
	}
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "CLI execution failed") {
		t.Errorf("stderr must report the failure, got %q", stderr.String())
	}
}

// TestRunHybridCLI_PositionalPathIsHandled covers the fallback branch used when
// the path is passed without -analyze.
func TestRunHybridCLI_PositionalPathIsHandled(t *testing.T) {
	dir := writeSupportconfigFixture(t)
	work := t.TempDir()
	chdir(t, work)

	var stdout, stderr bytes.Buffer
	code, handled := runHybridCLI([]string{dir, "-json"}, &stdout, &stderr)
	if !handled {
		t.Fatalf("a positional supportconfig path must be handled, not passed to the GUI")
	}
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %q)", code, stderr.String())
	}
	if _, err := os.Stat(findReport(t, work, "_report.json")); err != nil {
		t.Errorf("JSON report not written: %v", err)
	}
}

// TestRunHybridCLI_TrailingAnonymizeFlagIsHonoured pins the manual argument
// rescan: "-anonymize" placed after the path must still redact PII.
func TestRunHybridCLI_TrailingAnonymizeFlagIsHonoured(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "sssd.txt", "Failed to connect to 192.168.55.44 for domain evil.example.com\n")
	writeFile(t, dir, "rpm.txt", "sssd-2.9.4-150500.x86_64\n")

	work := t.TempDir()
	chdir(t, work)

	var stdout, stderr bytes.Buffer
	code, handled := runHybridCLI([]string{dir, "-json", "-anonymize"}, &stdout, &stderr)
	if !handled || code != 0 {
		t.Fatalf("handled=%v code=%d stderr=%q", handled, code, stderr.String())
	}
	data, err := os.ReadFile(findReport(t, work, "_report.json"))
	if err != nil {
		t.Fatalf("JSON report not written: %v", err)
	}
	if strings.Contains(string(data), "192.168.55.44") {
		t.Errorf("-anonymize after the path was ignored: raw IP leaked into the report")
	}
}

// TestRunHybridCLI_PositionalPathDefaultsToBothFormats: no format flag means the
// configured default (TXT + HTML) is applied.
func TestRunHybridCLI_PositionalPathDefaultsToBothFormats(t *testing.T) {
	dir := writeSupportconfigFixture(t)
	work := t.TempDir()
	chdir(t, work)

	var stdout, stderr bytes.Buffer
	code, handled := runHybridCLI([]string{dir}, &stdout, &stderr)
	if !handled || code != 0 {
		t.Fatalf("handled=%v code=%d stderr=%q", handled, code, stderr.String())
	}
	entries, err := os.ReadDir(work)
	if err != nil {
		t.Fatal(err)
	}
	formats := map[string]bool{}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".txt") {
			formats["txt"] = true
		}
		if strings.HasSuffix(e.Name(), ".html") {
			formats["html"] = true
		}
	}
	if !formats["txt"] || !formats["html"] {
		t.Errorf("default behaviour must generate both TXT and HTML reports, got %v", entries)
	}
}

// TestRunHybridCLI_CompareFlagIsNotRegistered pins the documented split between
// the two binaries at dispatcher level.
func TestRunHybridCLI_CompareFlagIsNotRegistered(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code, handled := runHybridCLI([]string{"-compare", "a:b"}, &stdout, &stderr)
	if !handled {
		t.Fatalf("a flag parse error must be handled, not passed to the GUI")
	}
	if code == 0 {
		t.Errorf("-compare must not be accepted by the hybrid binary")
	}
	if !strings.Contains(stderr.String(), "flag provided but not defined") {
		t.Errorf("expected the flag parse error, got %q", stderr.String())
	}
}

// TestRunHybridCLI_HelpExitsZero.
func TestRunHybridCLI_HelpExitsZero(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code, handled := runHybridCLI([]string{"-h"}, &stdout, &stderr)
	if !handled {
		t.Fatalf("-h must be handled")
	}
	if code != 0 {
		t.Errorf("exit code for -h = %d, want 0", code)
	}
	if !strings.Contains(stderr.String(), "-analyze") {
		t.Errorf("usage output missing from stderr: %q", stderr.String())
	}
}
