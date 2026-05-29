package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"sssd-inspector/internal/analyzer"
	"sssd-inspector/internal/config"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Pass 'false' so it doesn't show a terminal help screen if clicked normally via GUI
	if analyzer.HandleCommandLineArgs(false) {
		return // A CLI operation was handled, shut down gracefully!
	}

	// Default behavior: Launch the full desktop GUI
	app := NewApp()

	if config.Global == nil {
		log.Fatalf("Fatal: Global configuration instance was not initialized.")
	}

	err := wails.Run(&options.App{
		Title:            config.Global.GUI.Window.Title,
		Width:            config.Global.GUI.Window.Width,
		Height:           config.Global.GUI.Window.Height,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: config.Global.GUI.Colors.Background.R, G: config.Global.GUI.Colors.Background.G, B: config.Global.GUI.Colors.Background.B, A: config.Global.GUI.Colors.Background.A},
		OnStartup:        app.startup,
		Bind:             []interface{}{app},
		DragAndDrop:      &options.DragAndDrop{EnableFileDrop: true},
	})

	if err != nil {
		log.Fatalf("Failed to start GUI application: %v", err)
	}
}
