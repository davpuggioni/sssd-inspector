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
	flag.Parse()

	if *versionShort {
		fmt.Printf("sssd-analyzer-gui version 0.1.1 (Hybrid)\n")
		os.Exit(0)
	}

	// 2. Traffic Cop Logic: Did they pass a file argument in the terminal?
	if flag.NArg() > 0 {
		runCLI(flag.Arg(0))
		os.Exit(0) // Exit immediately. Do not load the GUI.
	}

	// 3. No arguments? Launch the Wails GUI!
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
			app, // This binds our app.go functions to Javascript
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop: true,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

// runCLI contains your old main.go terminal logic
func runCLI(path string) {
	fmt.Printf("Running in CLI mode analyzing: %s\n", path)
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

	report := analyzeData(fileMap)

	now := time.Now()
	report.Timestamp = now.Format("02:01:2006 15:04:05")

	// Re-use your text and HTML generation logic
	reportText := buildTextReport(report)
	fmt.Println(reportText)

	baseName := filepath.Base(path)
	txtReportFile := baseName + "_report.txt"
	htmlReportFile := baseName + "_report.html"

	err = os.WriteFile(txtReportFile, []byte(reportText), 0644)
	if err != nil {
		fmt.Printf("Error writing txt report: %v\n", err)
	} else {
		fmt.Printf("Text report saved to: %s\n", txtReportFile)
	}

	writeHTMLReportFile(report, htmlReportFile)
	fmt.Printf("HTML report saved to: %s\n", htmlReportFile)
}
