//go:build cli

// Package main is the entry point for SSSD Inspector CLI mode
// This file contains the main function for the static CLI binary
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"sssd-inspector/constants"
)

// main is the application entry point for CLI mode
// It handles command-line argument parsing and executes CLI-only operations
func main() {
	// Setup CLI Flags using configuration constants
	versionShort := flag.Bool(constants.FlagVersion, false, constants.DescVersion)
	cliPath := flag.String(constants.FlagAnalyze, "", constants.DescAnalyze)
	txtReport := flag.Bool(constants.FlagTXT, false, constants.DescTXT)
	htmlReport := flag.Bool(constants.FlagHTML, false, constants.DescHTML)
	// JSON report: full machine-readable export (summary + findings + evidence).
	jsonReport := flag.Bool("json", false, "Generate a JSON report (structured, machine-readable)")
	compareFlag := flag.String("compare", "", "Compare two supportconfig paths (format: pathA:pathB)")
	anonymize := flag.Bool(constants.FlagAnonymize, false, constants.DescAnonymize)
	flag.Parse()

	if *versionShort {
		fmt.Printf("%s version %s (CLI)\n", constants.AppName, constants.AppVersion)
		os.Exit(0)
	}

	// Compare mode: differential analysis between two supportconfigs
	if *compareFlag != "" {
		parts := strings.SplitN(*compareFlag, ":", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			fmt.Fprintln(os.Stderr, "Error: -compare requires two paths separated by ':' (e.g., -compare /path/A:/path/B)")
			os.Exit(1)
		}
		if err := runCompare(parts[0], parts[1], *anonymize, *jsonReport); err != nil {
			fmt.Fprintf(os.Stderr, "Compare execution failed: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Traffic Cop Logic (If they used the strict -analyze flag)
	if *cliPath != "" {
		if err := runCLI(*cliPath, *txtReport, *htmlReport, *anonymize, *jsonReport); err != nil {
			fmt.Fprintf(os.Stderr, "CLI execution failed: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Fallback logic for positional arguments
	if flag.NArg() > 0 {
		path := flag.Arg(0)
		genTxt := *txtReport
		genHtml := *htmlReport
		genJSON := *jsonReport
		genAnonymize := *anonymize

		// Manual flag scanning for format flags
		for _, arg := range os.Args[1:] {
			if arg == "-txt" || arg == "--txt" {
				genTxt = true
			}
			if arg == "-html" || arg == "--html" {
				genHtml = true
			}
			if arg == "-json" || arg == "--json" {
				genJSON = true
			}
			if arg == "-anonymize" || arg == "--anonymize" {
				genAnonymize = true
			}
		}

		// Apply defaults from configuration if no format specified
		if !genTxt && !genHtml && !genJSON && appConfig != nil && appConfig.CLI.DefaultGenerateBothFormats {
			genTxt = true
			genHtml = true
		}

		if err := runCLI(path, genTxt, genHtml, genAnonymize, genJSON); err != nil {
			fmt.Fprintf(os.Stderr, "CLI execution failed: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
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
