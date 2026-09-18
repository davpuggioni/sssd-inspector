//go:build !cli

// Package main is the entry point for SSSD Inspector GUI mode
// This file contains the main function that handles the hybrid CLI/GUI application
// with Wails framework integration for the GUI.
package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"sssd-inspector/constants"
)

//go:embed all:frontend/dist
var assets embed.FS

// main is the application entry point
// It handles command-line argument parsing and routes to either CLI or GUI mode
func main() {
	// Setup CLI Flags from the shared registry (cli_flags.go). The hybrid binary
	// keeps its historical surface: differential analysis (-compare) is CLI-only.
	opts := registerCLIFlags(flag.CommandLine, false)
	flag.Parse()

	if *opts.Version {
		fmt.Printf("%s version %s (Hybrid)\n", constants.AppName, constants.AppVersion)
		os.Exit(0)
	}

	// Traffic Cop Logic (If they used the strict -analyze flag)
	if *opts.Analyze != "" {
		if err := runCLI(*opts.Analyze, *opts.TXT, *opts.HTML, *opts.Anonymize, *opts.JSON); err != nil {
			log.Fatalf("CLI execution failed: %v", err)
		}
		os.Exit(0) // Exit immediately. Do not load the GUI.
	}

	// Fallback: If they provided a path WITHOUT -analyze, flag.Parse() stops parsing.
	// We must manually scan the remaining arguments so that flags like -html or -anonymize
	// placed AFTER the path still work perfectly.
	if flag.NArg() > 0 {
		path := flag.Arg(0)
		isAnonymize := *opts.Anonymize
		isTxt := *opts.TXT
		isHtml := *opts.HTML
		isJSON := *opts.JSON
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
			if strings.Contains(arg, "-json") || strings.Contains(arg, "--json") {
				isJSON = true
				hasExplicitFormat = true
			}
		}

		// If they just passed the path and NO format flags, default to both to match previous behavior
		if !hasExplicitFormat && !*opts.TXT && !*opts.HTML {
			if appConfig.CLI.DefaultGenerateBothFormats {
				isTxt = true
				isHtml = true
			}
		}

		if err := runCLI(path, isTxt, isHtml, isAnonymize, isJSON); err != nil {
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
