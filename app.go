package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// OpenFileBrowser opens the native OS file browser
func (a *App) OpenFileBrowser() (string, error) {
	options := runtime.OpenDialogOptions{
		Title: "Select Supportconfig Archive",
		Filters: []runtime.FileFilter{
			{DisplayName: "Supportconfig Archives (*.txz, *.tar.xz)", Pattern: "*.txz;*.tar.xz"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	}
	return runtime.OpenFileDialog(a.ctx, options)
}

// Analyze parses the archive
func (a *App) Analyze(targetPath string) (ReportData, error) {
	fileMap := make(map[string]string)
	info, err := os.Stat(targetPath)
	if err != nil {
		return ReportData{}, fmt.Errorf("error accessing path: %v", err)
	}
	if info.IsDir() {
		loadFromDir(targetPath, fileMap)
	} else {
		loadFromArchive(targetPath, fileMap)
	}
	return analyzeData(fileMap), nil
}

// SaveTXT opens a native Save dialog and writes the report to a TXT file
func (a *App) SaveTXT(report ReportData) (string, error) {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	
	// Open the native Save dialog
	options := runtime.SaveDialogOptions{
		DefaultFilename: fmt.Sprintf("SSSD_Analysis_Report_%s.txt", timestamp),
		Title:           "Save TXT Report",
		Filters: []runtime.FileFilter{
			{DisplayName: "Text Document (*.txt)", Pattern: "*.txt"},
		},
	}
	filePath, err := runtime.SaveFileDialog(a.ctx, options)
	if err != nil {
		return "", err
	}

	// If user cancelled the dialog, filePath will be empty
	if filePath == "" {
		return "cancelled", nil
	}

	txtContent := buildTextReport(report)

	// Save the file exactly where the user requested
	err = os.WriteFile(filePath, []byte(txtContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %v", err)
	}

	return filePath, nil
}
