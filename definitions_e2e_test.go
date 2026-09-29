// definitions_e2e_test.go
//
// End-to-end guard for the definitions CLI surface: the flags are invoked on
// the REAL compiled binary, because their failure mode is process-level. A gate
// that exits 0 while printing problems would be worse than no gate at all, and
// the hybrid binary must never try to open a window for a command-line
// invocation.
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runBinaryIn executes bin in dir and returns exit code, stdout and stderr. It
// differs from runBinary (main_coverage_test.go) only by the working directory:
// definitions discovery reads ./rules.yaml, so the fixture must BE the process
// working directory.
func runBinaryIn(t *testing.T, dir, bin string, args ...string) (int, string, string) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("running %s %v failed: %v", bin, args, err)
		}
		code = exitErr.ExitCode()
	}
	return code, stdout.String(), stderr.String()
}

// writeRulesFixture creates a directory containing rules.yaml and returns it.
func writeRulesFixture(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "rules.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestDefinitionsFlagsEndToEnd exercises the three definitions commands against
// the CLI binary and pins their exit codes.
func TestDefinitionsFlagsEndToEnd(t *testing.T) {
	bin := buildEntryPoint(t, "cli")
	fixture := writeRulesFixture(t, validRulesYAML)

	// -definitions-info: the search paths plus the fixture's own rules.yaml.
	code, stdout, stderr := runBinaryIn(t, fixture, bin, "-definitions-info")
	if code != 0 {
		t.Fatalf("-definitions-info exit code = %d, want 0 (stderr: %s)", code, stderr)
	}
	for _, want := range []string{"Definition search paths", filepath.Join(fixture, "rules.yaml"), "legacy-rc4"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout is missing %q:\n%s", want, stdout)
		}
	}

	// -validate-rules on valid definitions: exit 0, no error output.
	code, stdout, stderr = runBinaryIn(t, fixture, bin, "-validate-rules")
	if code != 0 {
		t.Fatalf("-validate-rules exit code = %d, want 0 (stdout: %s)", code, stdout)
	}
	if !strings.Contains(stdout, "[ok]") || stderr != "" {
		t.Errorf("stdout = %q, stderr = %q, want an [ok] verdict and no error output", stdout, stderr)
	}

	// -validate-rules on a broken file: exit 1, with the reason on stdout.
	broken := writeRulesFixture(t, "rules:\n  - name: \"x\"\n    severity: \"shout\"\n    message: \"m\"\n    patterns: [\"p\"]\n")
	code, stdout, stderr = runBinaryIn(t, broken, bin, "-validate-rules")
	if code != 1 {
		t.Fatalf("-validate-rules on a broken file: exit code = %d, want 1 (stdout: %s)", code, stdout)
	}
	if !strings.Contains(stdout, "invalid severity") || !strings.Contains(stderr, "would skip") {
		t.Errorf("missing the failure explanation (stdout: %s / stderr: %s)", stdout, stderr)
	}

	// -rules-test against a supportconfig directory: the dry-run must report the
	// match and its evidence.
	supportconfig := t.TempDir()
	if err := os.WriteFile(filepath.Join(supportconfig, "sssd.conf"), []byte("[domain/x]\nkrb5_enctype = rc4-hmac\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr = runBinaryIn(t, fixture, bin, "-rules-test", supportconfig)
	if code != 0 {
		t.Fatalf("-rules-test exit code = %d, want 0 (stderr: %s)", code, stderr)
	}
	if !strings.Contains(stdout, "[YES] legacy-rc4") || !strings.Contains(stdout, "evidence:") {
		t.Errorf("-rules-test stdout does not report the match:\n%s", stdout)
	}

	// -rules-test without a path is a usage error, not a silent pass.
	code, _, stderr = runBinaryIn(t, fixture, bin, "-rules-test", "")
	if code != 2 {
		t.Errorf("-rules-test with an empty path: exit code = %d, want 2 (stderr: %s)", code, stderr)
	}
}

// TestDefinitionsFlagsHybridBinary: the hybrid binary handles the same commands
// without starting the GUI (handled == true), which is the only way a
// definitions command can work on a headless machine.
func TestDefinitionsFlagsHybridBinary(t *testing.T) {
	bin := buildEntryPoint(t, "")
	fixture := writeRulesFixture(t, validRulesYAML)

	code, stdout, stderr := runBinaryIn(t, fixture, bin, "-definitions-info")
	if code != 0 {
		t.Fatalf("hybrid -definitions-info exit code = %d, want 0 (stderr: %s)", code, stderr)
	}
	if !strings.Contains(stdout, "Definition search paths") {
		t.Errorf("hybrid -definitions-info stdout = %q", stdout)
	}

	code, stdout, stderr = runBinaryIn(t, fixture, bin, "-rules-test", "")
	if code != 2 {
		t.Errorf("hybrid -rules-test with an empty path: exit code = %d, want 2 (stdout: %s / stderr: %s)", code, stdout, stderr)
	}
}
