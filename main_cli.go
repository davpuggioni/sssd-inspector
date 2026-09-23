//go:build cli

// Package main is the entry point for SSSD Inspector CLI mode
// This file contains the main function for the static CLI binary
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"sssd-inspector/constants"
)

// main is the application entry point for CLI mode.
//
// It is intentionally reduced to a single statement: everything decidable
// lives in runCLIEntry/dispatchCLI, which are unit-testable (main_cli_entry_test.go)
// and are also exercised end-to-end against the real binary by
// main_coverage_test.go.
func main() {
	os.Exit(runCLIEntry(os.Args[1:], os.Stdout, os.Stderr))
}

// runCLIEntry parses args and runs the CLI-only dispatch, returning the
// process exit code: 0 on success, 1 on a runtime error, 2 on a usage error.
//
// Keeping the routing here (instead of inside main) matters: this is the code
// path where the -logdir flag was silently lost during a project restructure,
// so every dispatch decision is pinned by tests.
func runCLIEntry(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(stderr)
	opts := registerCLIFlags(fs, true)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	return dispatchCLI(args, fs.Args(), opts, stdout, stderr)
}

// dispatchCLI applies the documented dispatch order:
// -v (version) -> -compare -> -analyze -> positional path.
func dispatchCLI(rawArgs, positional []string, opts *cliOptions, stdout, stderr io.Writer) int {
	if *opts.Version {
		fmt.Fprintf(stdout, "%s version %s (CLI)\n", constants.AppName, constants.AppVersion)
		return 0
	}

	// Compare mode: differential analysis between two supportconfigs
	if *opts.Compare != "" {
		parts := strings.SplitN(*opts.Compare, ":", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			fmt.Fprintln(stderr, "Error: -compare requires two paths separated by ':' (e.g., -compare /path/A:/path/B)")
			return 1
		}
		if err := runCompare(parts[0], parts[1], *opts.Anonymize, *opts.JSON); err != nil {
			fmt.Fprintf(stderr, "Compare execution failed: %v\n", err)
			return 1
		}
		return 0
	}

	// Raw SSSD log mode: analyze *.log files directly (e.g. /var/log/sssd)
	// without any supportconfig. Takes precedence over -analyze, matching the
	// historical dispatch order (Version -> Compare -> LogDir -> Analyze -> path).
	if *opts.LogDir != "" {
		if err := runLogDirAnalyze(*opts.LogDir, *opts.TXT, *opts.HTML, *opts.Anonymize, *opts.JSON); err != nil {
			fmt.Fprintf(stderr, "LogDir execution failed: %v\n", err)
			return 1
		}
		return 0
	}

	// Traffic Cop Logic (If they used the strict -analyze flag)
	if *opts.Analyze != "" {
		if err := runCLI(*opts.Analyze, *opts.TXT, *opts.HTML, *opts.Anonymize, *opts.JSON); err != nil {
			fmt.Fprintf(stderr, "CLI execution failed: %v\n", err)
			return 1
		}
		return 0
	}

	// Fallback logic for positional arguments
	if len(positional) > 0 {
		path := positional[0]
		genTxt := *opts.TXT
		genHtml := *opts.HTML
		genJSON := *opts.JSON
		genAnonymize := *opts.Anonymize

		// Manual flag scanning for format flags placed after the path.
		for _, arg := range rawArgs {
			switch arg {
			case "-" + constants.FlagTXT, "--" + constants.FlagTXT:
				genTxt = true
			case "-" + constants.FlagHTML, "--" + constants.FlagHTML:
				genHtml = true
			case "-" + constants.FlagJSON, "--" + constants.FlagJSON:
				genJSON = true
			case "-" + constants.FlagAnonymize, "--" + constants.FlagAnonymize:
				genAnonymize = true
			}
		}

		// Apply defaults from configuration if no format specified
		if !genTxt && !genHtml && !genJSON && appConfig != nil && appConfig.CLI.DefaultGenerateBothFormats {
			genTxt = true
			genHtml = true
		}

		if err := runCLI(path, genTxt, genHtml, genAnonymize, genJSON); err != nil {
			fmt.Fprintf(stderr, "CLI execution failed: %v\n", err)
			return 1
		}
		return 0
	}

	// No -v/-compare/-analyze and no positional path: nothing to do, exactly
	// like the historical flag.ExitOnError implementation (silent exit 0).
	return 0
}

// runCompare executes a differential analysis between two supportconfig
// directories and writes a comparison report (TXT and/or JSON).
func runCompare(pathA, pathB string, anonymize, genJSON bool) error {
	resolveDir := func(p string) (string, bool, error) {
		info, err := os.Stat(p)
		if err != nil {
			return p, false, err
		}
		if info.IsDir() {
			return p, false, nil
		}
		// Assume archive — extract to temp dir
		extractor := &defaultArchiveExtractorImpl{}
		dir, err := extractor.ExtractArchiveToTemp(p, func(msg string, pct int) {
			fmt.Fprintf(os.Stderr, "[extract %3d%%] %s\n", pct, msg)
		})
		if err != nil {
			return p, false, err
		}
		return dir, true, nil
	}

	extractDirA, isArchiveA, err := resolveDir(pathA)
	if err != nil {
		return fmt.Errorf("failed to resolve A: %w", err)
	}
	extractDirB, isArchiveB, err := resolveDir(pathB)
	if err != nil {
		return fmt.Errorf("failed to resolve B: %w", err)
	}
	defer func() {
		if isArchiveA {
			os.RemoveAll(extractDirA)
		}
		if isArchiveB {
			os.RemoveAll(extractDirB)
		}
	}()

	progressFunc := func(msg string, pct int) {
		fmt.Fprintf(os.Stderr, "[Compare %3d%%] %s\n", pct, msg)
	}

	result := CompareConfigs(extractDirA, extractDirB, progressFunc)

	// Build comparison TXT report
	var sb strings.Builder
	sb.WriteString("============================================================\n")
	sb.WriteString("       SUPPORTCONFIG SSSD COMPARISON REPORT (DIFF)\n")
	sb.WriteString("============================================================\n")
	fmt.Fprintf(&sb, " Supportconfig A: %s\n", pathA)
	fmt.Fprintf(&sb, " Supportconfig B: %s\n", pathB)
	fmt.Fprintf(&sb, " Health Score A:  %d/100\n", result.A.Summary.HealthScore)
	fmt.Fprintf(&sb, " Health Score B:  %d/100\n", result.B.Summary.HealthScore)
	fmt.Fprintf(&sb, " Score Delta:     %+d\n", result.ScoreDelta)
	sb.WriteString("------------------------------------------------------------\n")
	fmt.Fprintf(&sb, " Common findings:    %d\n", len(result.Common))
	fmt.Fprintf(&sb, " Only in A:          %d\n", len(result.OnlyInA))
	fmt.Fprintf(&sb, " Only in B:          %d\n", len(result.OnlyInB))
	sb.WriteString("============================================================\n\n")

	if len(result.OnlyInA) > 0 {
		sb.WriteString("--- RESOLVED (present in A, absent in B) ---\n")
		for _, f := range result.OnlyInA {
			fmt.Fprintf(&sb, "  [✓] %s\n", f)
		}
		sb.WriteString("\n")
	}
	if len(result.OnlyInB) > 0 {
		sb.WriteString("--- NEW (present in B, absent in A) ---\n")
		for _, f := range result.OnlyInB {
			fmt.Fprintf(&sb, "  [✗] %s\n", f)
		}
		sb.WriteString("\n")
	}
	if len(result.Common) > 0 {
		sb.WriteString("--- UNCHANGED (present in both) ---\n")
		for _, f := range result.Common {
			fmt.Fprintf(&sb, "  [=] %s\n", f)
		}
		sb.WriteString("\n")
	}

	txtContent := sb.String()
	fmt.Print(txtContent)

	if genJSON {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to serialize comparison JSON: %w", err)
		}
		if err := os.WriteFile("compare_report.json", data, 0644); err != nil {
			return fmt.Errorf("failed to write compare_report.json: %w", err)
		}
		fmt.Println("JSON comparison report saved to: compare_report.json")
	}

	return nil
}
