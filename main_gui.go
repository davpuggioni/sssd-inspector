//go:build !cli

// Package main is the entry point for SSSD Inspector GUI mode
// This file contains the main function that handles the hybrid CLI/GUI application
// with Wails framework integration for the GUI.
package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"io"
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

// main is the application entry point of the hybrid binary.
//
// It decides whether this invocation is a CLI one (handled by runHybridCLI)
// or a GUI launch. All decidable routing lives in runHybridCLI so that it is
// unit-testable (main_gui_entry_test.go); main itself is exercised end-to-end
// against the real binary by main_coverage_test.go.
func main() {
	if code, handled := runHybridCLI(os.Args[1:], os.Stdout, os.Stderr); handled {
		os.Exit(code)
	}
	launchGUI()
}

// runHybridCLI handles the command-line surface of the hybrid binary and
// reports whether the invocation was fully handled. When it returns
// handled == false the caller must start the GUI.
//
// Exit codes: 0 success, 1 runtime error, 2 usage error.
func runHybridCLI(args []string, stdout, stderr io.Writer) (int, bool) {
	// Setup CLI Flags from the shared registry (cli_flags.go). The hybrid binary
	// keeps its historical surface: differential analysis (-compare) is CLI-only.
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(stderr)
	opts := registerCLIFlags(fs, false)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0, true
		}
		return 2, true
	}

	if *opts.Version {
		fmt.Fprintf(stdout, "%s version %s (Hybrid)\n", constants.AppName, constants.AppVersion)
		return 0, true
	}

	// Traffic Cop Logic (If they used the strict -analyze flag)
	if *opts.Analyze != "" {
		if err := runCLI(*opts.Analyze, *opts.TXT, *opts.HTML, *opts.Anonymize, *opts.JSON); err != nil {
			fmt.Fprintf(stderr, "CLI execution failed: %v\n", err)
			return 1, true
		}
		return 0, true // Exit immediately. Do not load the GUI.
	}

	// Fallback: If they provided a path WITHOUT -analyze, flag.Parse() stops parsing.
	// We must manually scan the remaining arguments so that flags like -html or -anonymize
	// placed AFTER the path still work perfectly.
	if len(fs.Args()) > 0 {
		path := fs.Args()[0]
		isAnonymize := *opts.Anonymize
		isTxt := *opts.TXT
		isHtml := *opts.HTML
		isJSON := *opts.JSON
		hasExplicitFormat := false

		// Manually scan remaining arguments for all our flags
		for _, arg := range args {
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
			if strings.Contains(arg, "-"+constants.FlagJSON) {
				isJSON = true
				hasExplicitFormat = true
			}
		}

		// If they just passed the path and NO format flags, default to both to match previous behavior
		if !hasExplicitFormat && !*opts.TXT && !*opts.HTML {
			if appConfig != nil && appConfig.CLI.DefaultGenerateBothFormats {
				isTxt = true
				isHtml = true
			}
		}

		if err := runCLI(path, isTxt, isHtml, isAnonymize, isJSON); err != nil {
			fmt.Fprintf(stderr, "CLI execution failed: %v\n", err)
			return 1, true
		}
		return 0, true
	}

	// No arguments: the caller launches the GUI.
	return 0, false
}

// launchGUI starts the Wails desktop application with configuration-based settings.
//
// It is deliberately not unit-tested: it requires a display server and the
// Wails/WebKit runtime. main_coverage_test.go guards the entry point itself,
// while everything the GUI calls into (App.Analyze, the dialogs) is covered by
// the app tests.
func launchGUI() {
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
