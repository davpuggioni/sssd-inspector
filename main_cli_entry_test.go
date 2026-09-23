//go:build cli

// main_cli_entry_test.go
//
// Unit tests for the CLI entry point dispatcher (runCLIEntry/dispatchCLI).
//
// main_coverage_test.go covers main() end-to-end by spawning the compiled
// binary; this file pins every individual routing decision in-process, so a
// broken dispatch order is reported as a precise assertion failure instead of
// as a mysterious process-level failure.
package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sssd-inspector/constants"
)

// newCLIOptions parses args through the real CLI registry, mirroring
// runCLIEntry without performing the dispatch.
func newCLIOptions(t *testing.T, args ...string) (*cliOptions, []string, error) {
	t.Helper()
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	opts := registerCLIFlags(fs, true)
	err := fs.Parse(args)
	return opts, fs.Args(), err
}

// TestDispatchCLI_VersionFlag: -v prints the version on stdout and never runs
// an analysis, even when a path is also given.
func TestDispatchCLI_VersionFlag(t *testing.T) {
	opts, positional, err := newCLIOptions(t, "-v", "/nonexistent-path")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var stdout, stderr bytes.Buffer
	code := dispatchCLI([]string{"-v", "/nonexistent-path"}, positional, opts, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), constants.AppVersion) {
		t.Errorf("stdout must contain the version %q, got %q", constants.AppVersion, stdout.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr must stay empty, got %q", stderr.String())
	}
}

// TestDispatchCLI_NoArguments: neither -v, -compare, -analyze nor a path.
func TestDispatchCLI_NoArguments(t *testing.T) {
	opts, positional, err := newCLIOptions(t)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var stdout, stderr bytes.Buffer
	if code := dispatchCLI(nil, positional, opts, &stdout, &stderr); code != 0 {
		t.Errorf("exit code = %d, want 0 (historical flag.ExitOnError behaviour)", code)
	}
	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Errorf("no output expected, got stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

// TestDispatchCLI_CompareUsageErrors: malformed -compare specs must exit 1 with
// a message explaining the expected format.
func TestDispatchCLI_CompareUsageErrors(t *testing.T) {
	for _, spec := range []string{"only-one", ":b", "a:", ":"} {
		t.Run(spec, func(t *testing.T) {
			opts, positional, err := newCLIOptions(t, "-compare", spec)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			var stdout, stderr bytes.Buffer
			code := dispatchCLI([]string{"-compare", spec}, positional, opts, &stdout, &stderr)
			if code != 1 {
				t.Errorf("exit code = %d, want 1", code)
			}
			if !strings.Contains(stderr.String(), "requires two paths") {
				t.Errorf("stderr must explain the -compare format, got %q", stderr.String())
			}
		})
	}
}

// TestDispatchCLI_CompareWinsOverAnalyze pins the dispatch order: when both
// -compare and -analyze are supplied, -compare is executed.
func TestDispatchCLI_CompareWinsOverAnalyze(t *testing.T) {
	dir := writeSupportconfigFixture(t)
	spec := dir + ":" + filepath.Join(t.TempDir(), "missing-b")
	args := []string{"-compare", spec, "-analyze", dir}
	opts, positional, err := newCLIOptions(t, args...)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var stdout, stderr bytes.Buffer
	code := dispatchCLI(args, positional, opts, &stdout, &stderr)

	if strings.Contains(stderr.String(), "CLI execution failed") {
		t.Errorf("the analysis path ran instead of -compare: %q", stderr.String())
	}
	if code == 0 {
		t.Errorf("expected -compare to run and fail on the missing second path")
	}
}

// TestDispatchCLI_AnalyzeErrorIsExitCode1: a bad -analyze path must produce
// exit code 1 and a diagnostic on stderr.
func TestDispatchCLI_AnalyzeErrorIsExitCode1(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope")
	args := []string{"-analyze", missing, "-txt"}
	opts, positional, err := newCLIOptions(t, args...)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var stdout, stderr bytes.Buffer
	if code := dispatchCLI(args, positional, opts, &stdout, &stderr); code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "CLI execution failed") {
		t.Errorf("stderr must report the failure, got %q", stderr.String())
	}
}

// TestDispatchCLI_PositionalPathWithTrailingFormatFlags covers the legacy
// convention where format flags follow the path ("path -json"): flag.Parse
// stops at the first non-flag argument, so the dispatcher rescans raw args.
func TestDispatchCLI_PositionalPathWithTrailingFormatFlags(t *testing.T) {
	dir := writeSupportconfigFixture(t)
	work := t.TempDir()
	chdir(t, work)

	args := []string{dir, "-json"}
	opts, positional, err := newCLIOptions(t, args...)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(positional) == 0 || positional[0] != dir {
		t.Fatalf("positional args = %v, want the directory as first argument", positional)
	}
	var stdout, stderr bytes.Buffer
	code := dispatchCLI(args, positional, opts, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %q)", code, stderr.String())
	}
	if _, err := os.Stat(findReport(t, work, "_report.json")); err != nil {
		t.Errorf("JSON report not written: %v", err)
	}
}

// TestDispatchCLI_PositionalPathOnlyDefaultsToBothFormats: with no format flag
// at all the configured default applies.
func TestDispatchCLI_PositionalPathOnlyDefaultsToBothFormats(t *testing.T) {
	dir := writeSupportconfigFixture(t)
	work := t.TempDir()
	chdir(t, work)

	opts, positional, err := newCLIOptions(t, dir)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var stdout, stderr bytes.Buffer
	if code := dispatchCLI([]string{dir}, positional, opts, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %q)", code, stderr.String())
	}

	entries, err := os.ReadDir(work)
	if err != nil {
		t.Fatal(err)
	}
	formats := map[string]bool{}
	for _, e := range entries {
		switch {
		case strings.HasSuffix(e.Name(), ".txt"):
			formats["txt"] = true
		case strings.HasSuffix(e.Name(), ".html"):
			formats["html"] = true
		}
	}
	if !formats["txt"] || !formats["html"] {
		t.Errorf("default behaviour must generate both TXT and HTML reports, got %v", entries)
	}
}

// TestRunCLIEntry_HelpExitsZeroAndShowsFlags: "-h" must not be treated as a
// usage error (exit 2) and must list the flags.
func TestRunCLIEntry_HelpExitsZeroAndShowsFlags(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := runCLIEntry([]string{"-h"}, &stdout, &stderr); code != 0 {
		t.Errorf("exit code for -h = %d, want 0", code)
	}
	if !strings.Contains(stderr.String(), "-analyze") {
		t.Errorf("usage output missing from stderr: %q", stderr.String())
	}
}

// TestRunCLIEntry_UnknownFlagExitsTwo pins the usage-error exit code.
func TestRunCLIEntry_UnknownFlagExitsTwo(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := runCLIEntry([]string{"--nope"}, &stdout, &stderr); code != 2 {
		t.Errorf("exit code for an unknown flag = %d, want 2", code)
	}
}

// TestRunCLIEntry_EndToEnd analyzes a directory through the entry point.
// runCLI writes the JSON report to the working directory (its human-readable
// summary goes straight to the process stdout), so the report file is the
// observable here.
func TestRunCLIEntry_EndToEnd(t *testing.T) {
	dir := writeSupportconfigFixture(t)
	work := t.TempDir()
	chdir(t, work)

	var stdout, stderr bytes.Buffer
	if code := runCLIEntry([]string{"-analyze", dir, "-json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %q)", code, stderr.String())
	}
	data, err := os.ReadFile(findReport(t, work, "_report.json"))
	if err != nil {
		t.Fatalf("JSON report not written: %v", err)
	}
	if !strings.Contains(string(data), "{") {
		t.Errorf("expected a JSON document in the report file, got %q", data)
	}
}

// TestDispatchCLI_LogDirPrecedence pins the documented dispatch order: with
// both -logdir and -analyze, raw-log mode wins and the supportconfig path is
// never attempted (its path is deliberately nonexistent here).
func TestDispatchCLI_LogDirPrecedence(t *testing.T) {
	args := []string{
		"-logdir", writeRawLogFixture(t),
		"-analyze", filepath.Join(t.TempDir(), "no-such-supportconfig"),
	}
	opts, positional, err := newCLIOptions(t, args...)
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := dispatchCLI(args, positional, opts, &stdout, &stderr); code != 0 {
		t.Errorf("exit code = %d, want 0 (-logdir must win over -analyze; stderr: %q)",
			code, stderr.String())
	}
	if strings.Contains(stderr.String(), "CLI execution failed") {
		t.Errorf("-analyze must not run when -logdir is set, got stderr %q", stderr.String())
	}
}

// TestDispatchCLI_LogDirMissingPathFails: a bad -logdir path exits 1 with an
// explanation instead of a silent success.
func TestDispatchCLI_LogDirMissingPathFails(t *testing.T) {
	args := []string{"-logdir", filepath.Join(t.TempDir(), "missing")}
	opts, positional, err := newCLIOptions(t, args...)
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := dispatchCLI(args, positional, opts, &stdout, &stderr); code != 1 {
		t.Errorf("exit code = %d, want 1 for a missing -logdir path", code)
	}
	if !strings.Contains(stderr.String(), "LogDir execution failed") {
		t.Errorf("stderr = %q, want an explanation of the failure", stderr.String())
	}
}

// TestDispatchCLI_LogDirEndToEnd: full raw-log run through the dispatcher,
// producing a JSON report — raw-log mode must expose the same outputs as
// -analyze.
func TestDispatchCLI_LogDirEndToEnd(t *testing.T) {
	work := t.TempDir()
	chdir(t, work)

	args := []string{"-logdir", writeRawLogFixture(t), "-json"}
	opts, positional, err := newCLIOptions(t, args...)
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := dispatchCLI(args, positional, opts, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %q)", code, stderr.String())
	}
	data, err := os.ReadFile(findReport(t, work, "_report.json"))
	if err != nil {
		t.Fatalf("JSON report not written: %v", err)
	}
	if !strings.Contains(string(data), "summary") {
		t.Errorf("expected a JSON report with a summary section, got %q", truncateForLog(string(data)))
	}
}
