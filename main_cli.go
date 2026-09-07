//go:build cli

// Package main is the entry point for SSSD Inspector CLI mode
// This file contains the main function for the static CLI binary
package main

import (
	"flag"
	"fmt"
	"os"

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
	anonymize := flag.Bool(constants.FlagAnonymize, false, constants.DescAnonymize)
	flag.Parse()

	if *versionShort {
		fmt.Printf("%s version %s (CLI)\n", constants.AppName, constants.AppVersion)
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

	// No arguments provided - show usage
	fmt.Printf("SSSD Inspector version %s (CLI)\n", constants.AppVersion)
	fmt.Println("Usage: sssd-inspector [options] <path>")
	fmt.Println("\nOptions:")
	fmt.Printf("  -%s, --%s\t%s\n", constants.FlagVersion, "version", constants.DescVersion)
	fmt.Printf("  -%s, --%s\t%s\n", constants.FlagAnalyze, "analyze", constants.DescAnalyze)
	fmt.Printf("  -%s, --%s\t%s\n", constants.FlagTXT, "txt", constants.DescTXT)
	fmt.Printf("  -%s, --%s\t%s\n", constants.FlagHTML, "html", constants.DescHTML)
	fmt.Println("  -json, --json\tGenerate a JSON report (structured, machine-readable)")
	fmt.Printf("  -%s, --%s\t%s\n", constants.FlagAnonymize, "anonymize", constants.DescAnonymize)
	fmt.Println("\nExamples:")
	fmt.Println("  sssd-inspector -analyze /path/to/supportconfig.txz")
	fmt.Println("  sssd-inspector /path/to/supportconfig.txz -txt -html")
	fmt.Println("  sssd-inspector /path/to/supportconfig.txz -anonymize")
	os.Exit(1)
}
