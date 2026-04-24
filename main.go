// Package main is the entry point for SSSD Inspector, a hybrid CLI/GUI application
// for analyzing SSSD configurations and logs from SUSE supportconfig archives.
// This file contains the main function that handles command-line argument parsing,
// configuration loading, and routes execution between CLI and GUI modes.
package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"sssd-inspector/config"
	"sssd-inspector/constants"
)

//go:embed all:frontend/dist
var assets embed.FS

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

// main is the application entry point
// It handles command-line argument parsing and routes to either CLI or GUI mode
func main() {
	// Setup CLI Flags using configuration constants
	versionShort := flag.Bool(constants.FlagVersion, false, constants.DescVersion)
	cliPath := flag.String(constants.FlagAnalyze, "", constants.DescAnalyze)
	txtReport := flag.Bool(constants.FlagTXT, false, constants.DescTXT)
	htmlReport := flag.Bool(constants.FlagHTML, false, constants.DescHTML)
	anonymize := flag.Bool(constants.FlagAnonymize, false, constants.DescAnonymize)
	flag.Parse()

	if *versionShort {
		fmt.Printf("%s version %s (Hybrid)\n", constants.AppName, constants.AppVersion)
		os.Exit(0)
	}

	// Traffic Cop Logic (If they used the strict -analyze flag)
	if *cliPath != "" {
		if err := runCLI(*cliPath, *txtReport, *htmlReport, *anonymize); err != nil {
			log.Fatalf("CLI execution failed: %v", err)
		}
		os.Exit(0) // Exit immediately. Do not load the GUI.
	}

	// Fallback: If they provided a path WITHOUT -analyze, flag.Parse() stops parsing.
	// We must manually scan the remaining arguments so that flags like -html or -anonymize
	// placed AFTER the path still work perfectly.
	if flag.NArg() > 0 {
		path := flag.Arg(0)
		isAnonymize := *anonymize
		isTxt := *txtReport
		isHtml := *htmlReport
		hasExplicitFormat := false

		// Manually scan remaining arguments for all our flags
		for _, arg := range os.Args[1:] {
			if strings.Contains(arg, "-"+constants.FlagAnonymize) {
				isAnonymize = true
			}
			if strings.Contains(arg, "-"+constants.FlagTXT) {
				isTxt = true
				hasExplicitFormat = true
			}
			if strings.Contains(arg, "-"+constants.FlagHTML) {
				isHtml = true
				hasExplicitFormat = true
			}
		}

		// If they just passed the path and NO format flags, default to both to match previous behavior
		if !hasExplicitFormat && !*txtReport && !*htmlReport {
			if appConfig.CLI.DefaultGenerateBothFormats {
				isTxt = true
				isHtml = true
			}
		}

		if err := runCLI(path, isTxt, isHtml, isAnonymize); err != nil {
			log.Fatalf("CLI execution failed: %v", err)
		}
		os.Exit(0)
	}

	// Launch the Wails GUI with configuration-based settings
	app := NewApp()

	// Get window settings from configuration
	windowWidth := appConfig.GUI.Window.Width
	windowHeight := appConfig.GUI.Window.Height
	windowTitle := appConfig.GUI.Window.Title
	bgColor := appConfig.GUI.Colors.Background

	err := wails.Run(&options.App{
		Title:  windowTitle,
		Width:  windowWidth,
		Height: windowHeight,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{
			R: bgColor.R,
			G: bgColor.G,
			B: bgColor.B,
			A: bgColor.A,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop: true,
		},
	})

	if err != nil {
		log.Fatalf("Failed to start GUI application: %v", err)
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
