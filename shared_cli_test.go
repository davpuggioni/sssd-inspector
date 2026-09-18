// shared_cli_test.go
//
// End-to-end regression tests for runCLI (shared.go), which had 0% coverage.
// They exercise the path every user-visible flag combination flows through:
// directory input, report file generation, and the machine-readable export.
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sssd-inspector/constants"
)

// chdirInto switches the process working directory for the test duration.
// runCLI writes its report files next to the input base name in CWD, so each
// test gets an isolated temp working directory.
func chdirInto(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
}

// captureStdout redirects os.Stdout through a pipe while fn runs and returns
// everything printed. Keeps the test log clean and lets us assert that the
// CLI actually prints the report (its primary UX contract).
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	done := make(chan string, 1)
	go func() {
		var sb strings.Builder
		buf := make([]byte, 4096)
		for {
			n, err := r.Read(buf)
			if n > 0 {
				sb.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
		done <- sb.String()
	}()
	fn()
	_ = w.Close()
	os.Stdout = orig
	out := <-done
	_ = r.Close()
	return out
}

// cliFixture is a minimal supportconfig directory that triggers a known
// Kerberos log finding.
func cliFixture(t *testing.T) string {
	t.Helper()
	dir := setupMockDir(t, map[string]string{
		"sssd.conf":             "[domain/example.com]\nid_provider = ad\nad_domain = example.com\n",
		"sssd.txt":              "Dec 10 12:00:00 server sssd: Failed to initialize credentials using keytab [MEMORY:/etc/krb5.keytab]: Preauthentication failed\n",
		"basic-environment.txt": "Hostname: host.example.com\n",
		"nsswitch.conf":         "passwd: files sss\ngroup: files sss\n",
		"rpm.txt":               "sssd-2.9.4-150500.x86_64\n",
	})
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

// reportOutputNames returns the expected on-disk report file names for input.
func reportOutputNames(path string) map[string]string {
	base := filepath.Base(path)
	return map[string]string{
		"txt":  base + constants.DefaultOutputSuffix + "." + constants.TXTFormat,
		"html": base + constants.DefaultOutputSuffix + "." + constants.HTMLFormat,
		"json": base + constants.DefaultOutputSuffix + ".json",
	}
}

// TestRunCLI_SupportconfigDir_GeneratesAllReports is the core CLI regression
// test: analyze a directory, generate every format, verify files and the
// machine-readable export.
func TestRunCLI_SupportconfigDir_GeneratesAllReports(t *testing.T) {
	dir := cliFixture(t)
	work := t.TempDir()
	chdirInto(t, work)

	stdout := captureStdout(t, func() {
		if err := runCLI(dir, true, true, false, true); err != nil {
			t.Fatalf("runCLI failed: %v", err)
		}
	})

	for kind, name := range reportOutputNames(dir) {
		info, err := os.Stat(filepath.Join(work, name))
		if err != nil {
			t.Errorf("expected %s report %s to be written: %v", kind, name, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("report %s is empty", name)
		}
	}
	if !strings.Contains(stdout, "Preauthentication") {
		t.Errorf("expected the report to be printed to stdout, got %d bytes", len(stdout))
	}

	data, err := os.ReadFile(filepath.Join(work, reportOutputNames(dir)["json"]))
	if err != nil {
		t.Fatalf("cannot read JSON report: %v", err)
	}
	var report ReportData
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("JSON report is not valid JSON: %v", err)
	}
	if report.AppVersion != constants.AppVersion {
		t.Errorf("expected AppVersion %q, got %q", constants.AppVersion, report.AppVersion)
	}
	found := false
	for _, e := range report.SSSDLogErrors {
		if strings.Contains(e.Description, "Preauthentication") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected the keytab/preauth log error in the JSON export, got %d log errors", len(report.SSSDLogErrors))
	}
}
