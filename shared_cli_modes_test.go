// shared_cli_modes_test.go
//
// runCLI mode matrix: archive input, PII redaction, error handling and the
// no-output case. Companion of shared_cli_test.go.
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunCLI_Archive_EndToEnd exercises the tar.xz extraction branch of runCLI.
func TestRunCLI_Archive_EndToEnd(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "supportconfig-cli.tar.xz")
	writeTXZ(t, archive, supportconfigEntries())
	work := t.TempDir()
	chdirInto(t, work)

	captureStdout(t, func() {
		if err := runCLI(archive, true, false, false, true); err != nil {
			t.Fatalf("runCLI on archive failed: %v", err)
		}
	})

	names := reportOutputNames(archive)
	for _, name := range []string{names["txt"], names["json"]} {
		if _, err := os.Stat(filepath.Join(work, name)); err != nil {
			t.Errorf("expected report %s from archive input: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(work, names["html"])); !os.IsNotExist(err) {
		t.Errorf("HTML report must not be generated when genHtml=false")
	}
}

// TestRunCLI_AnonymizeRedacts pins the PII contract of the CLI path: with
// -anonymize, no raw IP or internal domain may survive in the JSON export.
func TestRunCLI_AnonymizeRedacts(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"sssd.conf":             "[domain/corp.internal]\nid_provider = ad\nad_domain = corp.internal\n",
		"sssd.txt":              "Dec 10 12:00:00 web01 conn 10.20.30.40: LDAP server corp.internal unreachable\n",
		"basic-environment.txt": "Hostname: web01.corp.internal\n",
		"nsswitch.conf":         "passwd: files sss\ngroup: files sss\n",
		"rpm.txt":               "sssd-2.9.4-150500.x86_64\n",
	})
	t.Cleanup(func() { os.RemoveAll(dir) })
	work := t.TempDir()
	chdirInto(t, work)

	captureStdout(t, func() {
		if err := runCLI(dir, false, false, true, true); err != nil {
			t.Fatalf("runCLI with anonymize failed: %v", err)
		}
	})

	data, err := os.ReadFile(filepath.Join(work, reportOutputNames(dir)["json"]))
	if err != nil {
		t.Fatalf("cannot read JSON report: %v", err)
	}
	if strings.Contains(string(data), "10.20.30.40") {
		t.Errorf("anonymized JSON export leaks the raw IP address")
	}
	if strings.Contains(string(data), "corp.internal") {
		t.Errorf("anonymized JSON export leaks the internal domain")
	}
}

// TestRunCLI_MissingPath_ReturnsError: a nonexistent input must surface an
// error (never a silent success or a panic).
func TestRunCLI_MissingPath_ReturnsError(t *testing.T) {
	chdirInto(t, t.TempDir())
	captureStdout(t, func() {
		err := runCLI(filepath.Join(t.TempDir(), "does-not-exist"), true, true, false, true)
		if err == nil {
			t.Errorf("expected an error for a missing input path, got nil")
		}
	})
}

// TestRunCLI_NoReportsRequested only prints to stdout — nothing is written.
func TestRunCLI_NoReportsRequested(t *testing.T) {
	dir := cliFixture(t)
	work := t.TempDir()
	chdirInto(t, work)

	out := captureStdout(t, func() {
		if err := runCLI(dir, false, false, false, false); err != nil {
			t.Fatalf("runCLI failed: %v", err)
		}
	})
	if len(out) == 0 {
		t.Errorf("expected the report on stdout even when no files are requested")
	}
	entries, err := os.ReadDir(work)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("expected no report files when all generators are off, found %v", entries)
	}
}
