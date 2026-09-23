// main_cover_measure_test.go
//
// Measures how much of the *entry point* code is actually executed.
//
// main_coverage_test.go asserts behaviour (exit codes, stdout, stderr) of the
// compiled binaries, but behaviour assertions cannot tell whether part of
// main() is dead or unreachable. This file closes that gap: it builds the
// binaries with `-cover`, runs them with GOCOVERDIR set, and reads the
// collected profile back with `go tool covdata func`.
//
// The guards that matter:
//   - main() of the CLI binary must be 100%: it is expected to be a single
//     "hand over to the dispatcher" statement. If someone moves logic back into
//     main(), its coverage drops below 100% and this test fails, forcing the
//     logic into the tested dispatchers.
//   - runCLIEntry / runHybridCLI must stay fully exercised by the invocations
//     below, so a newly added dispatch branch cannot go untested silently.
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// covFuncLine matches lines of `go tool covdata func` output:
//
//	sssd-inspector/main_cli.go:25:	main	100.0%
var covFuncLine = regexp.MustCompile(`^(\S+):(\d+):\s+(\S+)\s+([0-9]+(?:\.[0-9]+)?)%$`)

// buildEntryPointCovered compiles the package with -cover, adding the given
// build tags, and returns the binary path.
func buildEntryPointCovered(t *testing.T, tags string) string {
	t.Helper()
	if _, err := exec.LookPath("go"); err != nil {
		t.Skipf("go toolchain not available: %v", err)
	}
	bin := filepath.Join(t.TempDir(), "sssd-inspector-covered")
	args := []string{"build", "-cover", "-o", bin}
	if tags != "" {
		args = append(args, "-tags", tags)
	}
	args = append(args, ".")
	cmd := exec.Command("go", args...)
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build -cover (tags=%q) failed: %v\n%s", tags, err, out)
	}
	return bin
}

// runCoveredInvocation executes bin with a GOCOVERDIR pointing at covDir.
// Exit codes are not asserted: the callers check coverage, not behaviour
// (main_coverage_test.go owns behaviour).
func runCoveredInvocation(t *testing.T, bin, covDir string, args ...string) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = t.TempDir() // isolate generated report files
	cmd.Env = append(os.Environ(), "GOCOVERDIR="+covDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			t.Fatalf("running %s %v failed: %v\n%s", bin, args, err, out)
		}
	}
}

// entryPointCoverage returns function-name -> statement coverage percentage for
// the measurement directory, keeping only functions defined in the given file
// (e.g. "main_cli.go").
func entryPointCoverage(t *testing.T, covDir, file string) map[string]float64 {
	t.Helper()
	cmd := exec.Command("go", "tool", "covdata", "func", "-i="+covDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("go tool covdata unavailable (%v): %s", err, out)
	}
	coverage := map[string]float64{}
	for _, line := range strings.Split(string(out), "\n") {
		m := covFuncLine.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil || !strings.HasSuffix(m[1], file) {
			continue
		}
		pct, err := strconv.ParseFloat(m[4], 64)
		if err != nil {
			t.Fatalf("cannot parse coverage %% in %q: %v", line, err)
		}
		coverage[m[3]] = pct
	}
	if len(coverage) == 0 {
		t.Fatalf("no coverage data for %s in %s — is the binary built with -cover?", file, covDir)
	}
	return coverage
}

// TestMainCoverage_CLIEntryPointIsFullyCovered: the CLI main() must stay a
// one-liner delegating to runCLIEntry, and the dispatchers must be exercised by
// the invocations below.
func TestMainCoverage_CLIEntryPointIsFullyCovered(t *testing.T) {
	bin := buildEntryPointCovered(t, "cli")

	covDir := t.TempDir()
	runCoveredInvocation(t, bin, covDir, "-v")
	runCoveredInvocation(t, bin, covDir, "-h")
	runCoveredInvocation(t, bin, covDir, "-bogus") // usage error -> exit code 2
	runCoveredInvocation(t, bin, covDir, "-analyze", filepath.Join(t.TempDir(), "missing"))
	runCoveredInvocation(t, bin, covDir, "-compare", "onlyone") // usage error
	// Valid -compare needs two real directories.
	runCoveredInvocation(t, bin, covDir, "-compare",
		writeSupportconfigFixture(t)+":"+writeSupportconfigFixture(t))
	// A well-formed -compare whose paths do not exist must fail gracefully.
	runCoveredInvocation(t, bin, covDir, "-compare",
		filepath.Join(t.TempDir(), "missing")+":"+filepath.Join(t.TempDir(), "missing"))
	runCoveredInvocation(t, bin, covDir, "-analyze", writeSupportconfigFixture(t), "-json")
	// Positional path with every format/anonymize spelling (short and long) ...
	runCoveredInvocation(t, bin, covDir, writeSupportconfigFixture(t),
		"-txt", "--txt", "-html", "--html", "-json", "--json", "-anonymize", "--anonymize")
	// ... a failing positional path, which must exit 1 ...
	runCoveredInvocation(t, bin, covDir, filepath.Join(t.TempDir(), "missing"))
	// ... the configured default (TXT + HTML) when no format flag is given ...
	runCoveredInvocation(t, bin, covDir, writeSupportconfigFixture(t))
	// ... and no arguments at all (the CLI has nothing to do, but must not hang).
	runCoveredInvocation(t, bin, covDir)
	// Raw-log mode: the -logdir dispatch branch, both outcomes (a directory of
	// raw logs that succeeds, and a path that fails with exit code 1).
	runCoveredInvocation(t, bin, covDir, "-logdir", writeRawLogFixture(t))
	runCoveredInvocation(t, bin, covDir, "-logdir", filepath.Join(t.TempDir(), "missing"))

	coverage := entryPointCoverage(t, covDir, "main_cli.go")

	if got := coverage["main"]; got < 100 {
		t.Errorf("main() coverage = %.1f%%, want 100%%. main() must only delegate to runCLIEntry(); "+
			"move new logic into the tested dispatchers instead of main()", got)
	}
	if got := coverage["runCLIEntry"]; got < 100 {
		t.Errorf("runCLIEntry coverage = %.1f%%, want 100%% (flag parsing, help and error paths are all invoked)", got)
	}
	if got := coverage["dispatchCLI"]; got < 98 {
		t.Errorf("dispatchCLI coverage = %.1f%%, want >= 98%%: a dispatch branch was added without a test run here", got)
	}
}

// TestMainCoverage_HybridEntryPoint: same guard for the hybrid binary. Its
// main() cannot reach 100% because the final "start the GUI" statement must not
// be executed in a headless test (it would block on the display server), so the
// threshold pins the achievable maximum instead.
func TestMainCoverage_HybridEntryPoint(t *testing.T) {
	bin := buildEntryPointCovered(t, "")

	covDir := t.TempDir()
	runCoveredInvocation(t, bin, covDir, "-v")
	runCoveredInvocation(t, bin, covDir, "-h")
	runCoveredInvocation(t, bin, covDir, "-bogus") // usage error -> exit code 2
	runCoveredInvocation(t, bin, covDir, "-analyze", filepath.Join(t.TempDir(), "missing"))
	runCoveredInvocation(t, bin, covDir, "-analyze", writeSupportconfigFixture(t), "-json")
	// Positional path with every format/anonymize spelling ...
	runCoveredInvocation(t, bin, covDir, writeSupportconfigFixture(t),
		"-txt", "-html", "-json", "-anonymize", "--json")
	// ... a failing positional path, which must exit 1 ...
	runCoveredInvocation(t, bin, covDir, filepath.Join(t.TempDir(), "missing"))
	// ... and positional path alone, which must apply the configured defaults.
	runCoveredInvocation(t, bin, covDir, writeSupportconfigFixture(t))
	// Raw-log mode: the -logdir dispatch branch must be exercised here too,
	// including its failure path (a missing path exits 1).
	runCoveredInvocation(t, bin, covDir, "-logdir", writeRawLogFixture(t))
	runCoveredInvocation(t, bin, covDir, "-logdir", filepath.Join(t.TempDir(), "missing"))

	coverage := entryPointCoverage(t, covDir, "main_gui.go")

	// main() = "CLI handled? exit : launchGUI()". The CLI branch is taken, so
	// two of the three statements run; launchGUI() is intentionally never called.
	if got := coverage["main"]; got < 66 {
		t.Errorf("hybrid main() coverage = %.1f%%, want >= 66%% (only the launchGUI() statement is unreachable in tests)", got)
	}
	if got := coverage["launchGUI"]; got != 0 {
		t.Errorf("launchGUI coverage = %.1f%%, want 0%%: the GUI must never start during tests", got)
	}
	// runHybridCLI keeps exactly one unreachable statement: "no arguments ->
	// let the caller launch the GUI", which by definition cannot be executed here.
	if got := coverage["runHybridCLI"]; got < 96 {
		t.Errorf("runHybridCLI coverage = %.1f%%, want >= 96%%: a dispatch branch is no longer exercised", got)
	}
}
