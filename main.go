package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	txtReport := flag.Bool("txt", false, "Generate a TXT report (requires -analyze)")
	htmlReport := flag.Bool("html", false, "Generate an HTML report (requires -analyze)")
	anonymize := flag.Bool("anonymize", false, "Redact PII (IPs, Domains) from the report")
	flag.Parse()

	if *versionShort {
		fmt.Printf("sssd-analyzer-gui version 0.1.4 (Hybrid)\n")
		os.Exit(0)
	}

	// 2. Traffic Cop Logic
	if *cliPath != "" {
		runCLI(*cliPath, *txtReport, *htmlReport, *anonymize)
		os.Exit(0) // Exit immediately. Do not load the GUI.
	}

	if flag.NArg() > 0 {
		runCLI(flag.Arg(0), true, true, *anonymize)
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
