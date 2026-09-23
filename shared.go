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
func runCLI(path string, genTxt bool, genHtml bool, anonymize bool, genJSON bool) error {
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

	return writeReports(report, baseName, reportText, genTxt, genHtml, genJSON)
}

// writeReports renders the requested report formats as <baseName>_report.* in
// the working directory and prints where each file went. It is shared by the
// supportconfig path (runCLI) and raw-log mode (runLogDirAnalyze) so both
// cannot drift apart. reportText is the pre-rendered text report, so callers
// that already printed it do not pay for a second rendering.
func writeReports(report ReportData, baseName, reportText string, genTxt, genHtml, genJSON bool) error {
	if genTxt {
		txtReportFile := baseName + constants.DefaultOutputSuffix + "." + constants.TXTFormat
		if err := os.WriteFile(txtReportFile, []byte(reportText), 0644); err != nil {
			return fmt.Errorf("error writing txt report: %w", err)
		}
		fmt.Printf("Text report saved to: %s\n", txtReportFile)
	}

	if genHtml {
		htmlReportFile := baseName + constants.DefaultOutputSuffix + "." + constants.HTMLFormat
		writeHTMLReportFile(report, htmlReportFile)
		fmt.Printf("HTML report saved to: %s\n", htmlReportFile)
	}

	if genJSON {
		jsonReportFile := baseName + constants.DefaultOutputSuffix + ".json"
		data, err := buildJSONReport(report)
		if err != nil {
			return fmt.Errorf("error serializing JSON report: %w", err)
		}
		if err := os.WriteFile(jsonReportFile, data, 0644); err != nil {
			return fmt.Errorf("error writing JSON report: %w", err)
		}
		fmt.Printf("JSON report saved to: %s\n", jsonReportFile)
	}

	return nil
}

// runLogDirAnalyze analyzes raw SSSD log files directly (-logdir): it scans
// the *.log files (plus rotated variants) of the given directory or a single
// log file and generates the requested reports. Unlike runCLI it does not
// expect a supportconfig archive or directory layout.
//
// Format semantics mirror -analyze: only explicitly requested formats are
// written (-txt/-html/-json); with none of them the report is printed to
// stdout only.
func runLogDirAnalyze(path string, genTxt bool, genHtml bool, anonymize bool, genJSON bool) error {
	fmt.Printf("Running in CLI mode analyzing raw SSSD logs: %s\n", path)
	if anonymize {
		fmt.Println("[!] Anonymization mode enabled. PII will be redacted.")
	}

	dirPath, logFiles, err := collectLogFiles(path)
	if err != nil {
		return err
	}
	fmt.Printf("Found %d SSSD log files\n", len(logFiles))

	progressFunc := func(msg string, pct int) {
		fmt.Printf("[Progress %d%%] %s\n", pct, msg)
	}

	report := analyzeLogsOnly(dirPath, logFiles, anonymize, progressFunc)

	report.Timestamp = time.Now().Format(constants.TimestampFormat)
	report.AppVersion = constants.AppVersion

	reportText := buildTextReport(report)
	fmt.Println("\n" + reportText)

	// For a directory "/var/log/sssd" the report is "sssd_report.*"; for a
	// single file "sssd_example.com.log" it is "sssd_example.com.log_report.*".
	baseName := filepath.Base(path)

	return writeReports(report, baseName, reportText, genTxt, genHtml, genJSON)
}
