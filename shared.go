// Package main provides shared functionality for both CLI and GUI modes
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sssd-inspector/config"
	"sssd-inspector/constants"
	sssderrors "sssd-inspector/errors"
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
		return sssderrors.NewFileAccess(path, err)
	}

	var dirPath string
	if info.IsDir() {
		dirPath = path
	} else {
		// Use the new secure archive extractor
		dirPath, err = extractArchiveToTemp(path, progressFunc)
		if err != nil {
			return sssderrors.Wrap(err, sssderrors.ErrInvalidArchive, "extract error")
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
			return sssderrors.Wrap(err, sssderrors.ErrSystemError, "error writing txt report")
		}
		fmt.Printf("Text report saved to: %s\n", txtReportFile)
	}

	if genHtml {
		htmlReportFile := baseName + constants.DefaultOutputSuffix + "." + constants.HTMLFormat
		writeHTMLReportFile(report, htmlReportFile)
		fmt.Printf("HTML report saved: %s\n", htmlReportFile)
	}

	return nil
}

// runLogDirAnalyze analyzes raw SSSD log files directly from a directory (e.g., /var/log/sssd/).
// Unlike runCLI, it does not expect a supportconfig archive or directory layout.
// It scans *.log files for SSSD error patterns and generates a report with log-only findings.
func runLogDirAnalyze(dirPath string, genTxt bool, genHtml bool, anonymize bool) error {
	fmt.Printf("Analyzing raw SSSD log directory: %s\n", dirPath)
	if anonymize {
		fmt.Println("[!] Anonymization mode enabled. PII will be redacted.")
	}

	// Verify the directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		return sssderrors.NewFileAccess(dirPath, err)
	}
	if !info.IsDir() {
		return sssderrors.NewConfigInvalid(fmt.Sprintf("path is not a directory: %s", dirPath))
	}

	// Progress func for CLI output
	progressFunc := func(msg string, pct int) {
		fmt.Printf("[Progress %d%%] %s\n", pct, msg)
	}

	// Find all .log files in the directory
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return sssderrors.NewFileAccess(dirPath, err)
	}

	var logFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".log") {
			logFiles = append(logFiles, entry.Name())
		}
	}

	if len(logFiles) == 0 {
		return sssderrors.NewConfigNotFound(dirPath)
	}

	fmt.Printf("Found %d SSSD log files\n", len(logFiles))

	report := analyzeLogsOnly(dirPath, logFiles, anonymize, progressFunc)

	report.Timestamp = time.Now().Format(constants.TimestampFormat)
	report.AppVersion = constants.AppVersion

	reportText := buildTextReport(report)
	fmt.Println("\n" + reportText)

	baseName := filepath.Base(dirPath)

	if genTxt {
		txtReportFile := baseName + constants.DefaultOutputSuffix + "." + constants.TXTFormat
		err = os.WriteFile(txtReportFile, []byte(reportText), 0644)
		if err != nil {
			return sssderrors.Wrap(err, sssderrors.ErrSystemError, "error writing txt report")
		}
		fmt.Printf("Text report saved to: %s\n", txtReportFile)
	}

	if genHtml {
		htmlReportFile := baseName + constants.DefaultOutputSuffix + "." + constants.HTMLFormat
		writeHTMLReportFile(report, htmlReportFile)
		fmt.Printf("HTML report saved: %s\n", htmlReportFile)
	}

	return nil
}
