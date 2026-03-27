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
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 1. Setup CLI Flags
	versionShort := flag.Bool("v", false, "Print program version")
	cliPath := flag.String("analyze", "", "Path to the supportconfig directory or log file")
	txtReport := flag.Bool("txt", false, "Generate a TXT report")
	htmlReport := flag.Bool("html", false, "Generate an HTML report")
	anonymize := flag.Bool("anonymize", false, "Redact PII (IPs, Domains) from the report")
	flag.Parse()

	if *versionShort {
		fmt.Printf("sssd-analyzer-gui version 0.1.5 (Hybrid)\n")
		os.Exit(0)
	}

	// 2. Traffic Cop Logic (If they used the strict -analyze flag)
	if *cliPath != "" {
		runCLI(*cliPath, *txtReport, *htmlReport, *anonymize)
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
			if strings.Contains(arg, "-anonymize") {
				isAnonymize = true
			}
			if strings.Contains(arg, "-txt") {
				isTxt = true
				hasExplicitFormat = true
			}
			if strings.Contains(arg, "-html") {
				isHtml = true
				hasExplicitFormat = true
			}
		}

		// If they just passed the path and NO format flags, default to both to match previous behavior
		if !hasExplicitFormat && !*txtReport && !*htmlReport {
			isTxt = true
			isHtml = true
		}

		runCLI(path, isTxt, isHtml, isAnonymize)
		os.Exit(0)
	}

	// 3. Launch the Wails GUI
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "SSSD Inspector",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 244, G: 244, B: 249, A: 255},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop: true,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func runCLI(path string, genTxt bool, genHtml bool, anonymize bool) {
	fmt.Printf("Running in CLI mode analyzing: %s\n", path)
	if anonymize {
		fmt.Println("[!] Anonymization mode enabled. PII will be redacted.")
	}
	fileMap := make(map[string]string)

	info, err := os.Stat(path)
	if err != nil {
		log.Fatalf("Error accessing path: %v", err)
	}

	if info.IsDir() {
		loadFromDir(path, fileMap)
	} else {
		loadFromArchive(path, fileMap)
	}

	report := analyzeData(fileMap, anonymize)

	now := time.Now()
	report.Timestamp = now.Format("02:01:2006 15:04:05")

	reportText := buildTextReport(report)
	fmt.Println(reportText)

	baseName := filepath.Base(path)

	if genTxt {
		txtReportFile := baseName + "_report.txt"
		err = os.WriteFile(txtReportFile, []byte(reportText), 0644)
		if err != nil {
			fmt.Printf("Error writing txt report: %v\n", err)
		} else {
			fmt.Printf("Text report saved to: %s\n", txtReportFile)
		}
	}

	if genHtml {
		htmlReportFile := baseName + "_report.html"
		writeHTMLReportFile(report, htmlReportFile)
		fmt.Printf("HTML report saved to: %s\n", htmlReportFile)
	}
}
