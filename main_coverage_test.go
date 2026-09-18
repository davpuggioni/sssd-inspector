// main_coverage_test.go
//
// Exercises the *real* compiled entry point (main()).
//
// main() itself cannot be called from a unit test: it terminates the process
// through os.Exit. Everything decidable was therefore moved into runCLIEntry /
// runHybridCLI (see main_cli_entry_test.go and main_gui_entry_test.go), but the
// statements left inside main() — "hand control to the dispatcher" and "start
// the GUI when nothing was handled" — still need a guard, because they are the
// wiring that has no other protection.
//
// Strategy: `go build` twice (cli + hybrid), then execute the produced binaries
// with the CLI-only argument combinations and assert on exit code, stdout and
// stderr. This also catches process-level regressions such as turning an
// os.Exit(1) into log.Fatalf.
package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// buildEntryPoint compiles the package with the given build tags into a
// temporary binary and returns its path.
func buildEntryPoint(t *testing.T, tags string) string {
	t.Helper()
	if _, err := exec.LookPath("go"); err != nil {
		t.Skipf("go toolchain not available: %v", err)
	}
	bin := filepath.Join(t.TempDir(), "sssd-inspector-test")
	args := []string{"build", "-o", bin}
	if tags != "" {
		args = append(args, "-tags", tags)
	}
	args = append(args, ".")
	cmd := exec.Command("go", args...)
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build %s failed: %v\n%s", tags, err, out)
	}
	return bin
}

// runBinary executes the binary with args and returns exit code, stdout, stderr.
func runBinary(t *testing.T, bin string, args ...string) (int, string, string) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = t.TempDir() // isolate generated report files
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

// chdir changes directory for the duration of the test (inlined instead of
// testing.T.Chdir to keep the tests independent of the toolchain version).
// It lives here, in a build-tag-free file, because both flavours need it.
func chdir(t *testing.T, dir string) {
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

// findReport returns the path of the single file ending with suffix inside dir,
// failing the test when zero or more than one is found. runCLI writes its
// reports next to the current working directory, hence the directory argument.
func findReport(t *testing.T, dir, suffix string) string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var found []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), suffix) {
			found = append(found, filepath.Join(dir, e.Name()))
		}
	}
	if len(found) != 1 {
		t.Fatalf("expected exactly one %q report in %s, got %v", suffix, dir, found)
	}
	return found[0]
}

// writeSupportconfigFixture creates a minimal supportconfig directory that
// produces at least one Kerberos/keytab finding.
func writeSupportconfigFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"sssd.txt":              "Failed to initialize credentials using keytab [MEMORY:/etc/krb5.keytab]: Preauthentication failed\n",
		"rpm.txt":               "sssd-2.9.4-150500.x86_64\n",
		"basic-environment.txt": "Hostname: host.example.com\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// TestMainEntry_CLIVersionFlag covers main() -> runCLIEntry() -> dispatchCLI()
// for -v, the cheapest complete invocation of the CLI entry point.
func TestMainEntry_CLIVersionFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping binary build in short mode")
	}
	bin := buildEntryPoint(t, "cli")

	code, stdout, stderr := runBinary(t, bin, "-v")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %q)", code, stderr)
	}
	if !strings.Contains(stdout, "version") {
		t.Errorf("stdout does not report a version: %q", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got %q", stderr)
	}
}

// TestMainEntry_CLIUsageFlag covers the -h path: exit code 0 and usage output
// on stderr, which is where flag.ContinueOnError writes it.
func TestMainEntry_CLIUsageFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping binary build in short mode")
	}
	bin := buildEntryPoint(t, "cli")

	code, _, stderr := runBinary(t, bin, "-h")
	if code != 0 {
		t.Fatalf("exit code for -h = %d, want 0", code)
	}
	for _, want := range []string{"-analyze", "-anonymize", "-json", "-txt", "-html", "-compare"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("usage output does not mention %s:\n%s", want, stderr)
		}
	}
}

// TestMainEntry_CLIUnknownFlag: an unknown flag must not fall through to a
// successful exit, otherwise a typo'd flag silently does nothing.
func TestMainEntry_CLIUnknownFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping binary build in short mode")
	}
	bin := buildEntryPoint(t, "cli")

	code, _, stderr := runBinary(t, bin, "-definitely-not-a-flag")
	if code == 0 {
		t.Errorf("unknown flag exited 0; want non-zero. stderr=%q", stderr)
	}
	if !strings.Contains(stderr, "flag provided but not defined") {
		t.Errorf("expected flag parse error on stderr, got %q", stderr)
	}
}

// TestMainEntry_CLICompareInvalidSpec covers dispatchCLI's -compare validation.
func TestMainEntry_CLICompareInvalidSpec(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping binary build in short mode")
	}
	bin := buildEntryPoint(t, "cli")

	code, _, stderr := runBinary(t, bin, "-compare", "only-one-path")
	if code == 0 {
		t.Errorf("-compare with a single path exited 0; want a failure")
	}
	if !strings.Contains(stderr, "requires two paths") {
		t.Errorf("expected the -compare usage error, got %q", stderr)
	}
}

// TestMainEntry_CLIAnalyzeMissingPath: dispatchCLI routes -analyze to runCLI,
// whose error must surface as exit code 1.
func TestMainEntry_CLIAnalyzeMissingPath(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping binary build in short mode")
	}
	bin := buildEntryPoint(t, "cli")

	missing := filepath.Join(t.TempDir(), "does-not-exist")
	code, _, stderr := runBinary(t, bin, "-analyze", missing, "-txt")
	if code != 1 {
		t.Errorf("exit code for a missing -analyze path = %d, want 1 (stderr: %q)", code, stderr)
	}
	if !strings.Contains(stderr, "CLI execution failed") {
		t.Errorf("expected 'CLI execution failed' on stderr, got %q", stderr)
	}
}

// TestMainEntry_HybridVersionFlag covers main() -> runHybridCLI() -> handled.
func TestMainEntry_HybridVersionFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping binary build in short mode")
	}
	bin := buildEntryPoint(t, "")

	code, stdout, stderr := runBinary(t, bin, "-v")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %q)", code, stderr)
	}
	if !strings.Contains(stdout, "Hybrid") {
		t.Errorf("hybrid binary should report the Hybrid flavour, got %q", stdout)
	}
}

// TestMainEntry_HybridRejectsCompare pins the documented CLI/hybrid split at
// process level: -compare is not part of the hybrid surface.
func TestMainEntry_HybridRejectsCompare(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping binary build in short mode")
	}
	bin := buildEntryPoint(t, "")

	code, _, stderr := runBinary(t, bin, "-compare", "a:b")
	if code == 0 {
		t.Errorf("hybrid binary accepted -compare; it must be CLI-only")
	}
	if !strings.Contains(stderr, "flag provided but not defined") {
		t.Errorf("expected the flag parse error, got %q", stderr)
	}
}

// TestMainEntry_HybridAnalyzeDoesNotStartGUI is the guard for the most
// dangerous hybrid regression: a CLI invocation that falls through to
// launchGUI() would block forever (or fail without an X11/Wayland server) in
// headless environments. The command must terminate by itself.
func TestMainEntry_HybridAnalyzeDoesNotStartGUI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping binary build in short mode")
	}
	bin := buildEntryPoint(t, "")

	// Run with a hard timeout instead of a goroutine so no testing.T method is
	// called off the test goroutine.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "-analyze", writeSupportconfigFixture(t), "-txt")
	cmd.Dir = t.TempDir()
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("hybrid binary with -analyze did not terminate within 60s: it probably reached launchGUI()")
	}
	if err != nil {
		t.Fatalf("hybrid binary with -analyze failed: %v (stderr: %q)", err, stderr.String())
	}
}

// TestMainEntry_CLIAnalyzeEndToEnd runs a real supportconfig directory through
// the binary: main() -> dispatchCLI() -> runCLI() -> report on stdout.
func TestMainEntry_CLIAnalyzeEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping binary build in short mode")
	}
	bin := buildEntryPoint(t, "cli")

	code, stdout, stderr := runBinary(t, bin, "-analyze", writeSupportconfigFixture(t), "-json")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %q)", code, stderr)
	}
	lower := strings.ToLower(stdout)
	if !strings.Contains(lower, "krb5") && !strings.Contains(lower, "keytab") {
		t.Errorf("expected keytab/krb5 findings in the JSON report, got:\n%s", stdout)
	}
}

// TestMainEntry_CLIPositionalPathEndToEnd covers the last dispatching branch of
// main(): a path passed without -analyze, with the format flag after it.
func TestMainEntry_CLIPositionalPathEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping binary build in short mode")
	}
	bin := buildEntryPoint(t, "cli")

	code, stdout, stderr := runBinary(t, bin, writeSupportconfigFixture(t), "-json")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %q)", code, stderr)
	}
	if !strings.Contains(stdout, "SSSD ANALYSIS REPORT") {
		t.Errorf("expected the report banner on stdout, got:\n%s", stdout)
	}
}

// TestMainEntry_CLICompareEndToEnd covers -compare through the real binary:
// main() -> dispatchCLI() -> runCompare() -> CompareConfigs().
func TestMainEntry_CLICompareEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping binary build in short mode")
	}
	bin := buildEntryPoint(t, "cli")

	before := writeSupportconfigFixture(t)
	after := writeSupportconfigFixture(t)
	// "Fix" B by removing the failing log line, so the diff has content.
	if err := os.WriteFile(filepath.Join(after, "sssd.txt"), []byte("started\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	code, stdout, stderr := runBinary(t, bin, "-compare", before+":"+after, "-json")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %q)", code, stderr)
	}
	if !strings.Contains(stdout, "COMPARISON REPORT") {
		t.Errorf("expected the comparison banner on stdout, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "compare_report.json") {
		t.Errorf("expected the JSON comparison notice on stdout, got:\n%s", stdout)
	}
}
