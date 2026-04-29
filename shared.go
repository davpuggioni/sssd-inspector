// Package main provides shared functionality for both CLI and GUI modes
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"sssd-inspector/config"
	"sssd-inspector/constants"
)

// Global configuration instance
var appConfig *config.Config

// init initializes the application configuration
func init() {
	var err error
	appConfig, err = config.LoadConfig("")
	if err != nil {
		log.Printf("Warning: %v, using defaults", err)
		appConfig = config.DefaultConfig()
	}

	// Validate configuration
	if err := appConfig.Validate(); err != nil {
		log.Printf("Configuration validation error: %v", err)
	}
}

// runCLI executes the application in command-line mode
// It handles both directory and archive inputs, performs analysis, and generates reports.
// Returns an error if any step of the CLI execution fails.
func runCLI(path string, genTxt bool, genHtml bool, anonymize bool) error {
	fmt.Printf("Running in CLI mode analyzing: %s\n", path)
	if anonymize {
		fmt.Println("[!] Anonymization mode enabled. PII will be redacted.")
	}

	// Mock progress func for CLI output so the streaming engine doesn't panic
	progressFunc := func(msg string, pct int) {
		fmt.Printf("[Progress %d%%] %s\n", pct, msg)
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("error accessing path: %w", err)
	}

	var dirPath string
	if info.IsDir() {
		dirPath = path
	} else {
		// Use the new secure archive extractor
		dirPath, err = extractArchiveToTemp(path, progressFunc)
		if err != nil {
			return fmt.Errorf("extract error: %w", err)
		}
		defer os.RemoveAll(dirPath) // Clean up temp files when done
	}

	// Use the updated analyzeData signature
	report := analyzeData(dirPath, anonymize, progressFunc)

	report.Timestamp = time.Now().Format(constants.TimestampFormat)
	report.AppVersion = constants.AppVersion

	reportText := buildTextReport(report)
	fmt.Println("\n" + reportText)

	baseName := filepath.Base(path)

	if genTxt {
		txtReportFile := baseName + constants.DefaultOutputSuffix + "." + constants.TXTFormat
		err = os.WriteFile(txtReportFile, []byte(reportText), 0644)
		if err != nil {
			return fmt.Errorf("error writing txt report: %w", err)
		}
		fmt.Printf("Text report saved to: %s\n", txtReportFile)
	}

	if genHtml {
		htmlReportFile := baseName + constants.DefaultOutputSuffix + "." + constants.HTMLFormat
		writeHTMLReportFile(report, htmlReportFile)
		fmt.Printf("HTML report saved to: %s\n", htmlReportFile)
	}

	return nil
}
