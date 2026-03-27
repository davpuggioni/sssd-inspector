package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

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

// Analyze parses the archive and optionally anonymizes PII
func (a *App) Analyze(targetPath string, anonymize bool) (ReportData, error) {
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
	return analyzeData(fileMap, anonymize), nil
}

// SavePDF opens a native Save dialog and writes the PDF to disk
func (a *App) SavePDF(b64 string) (string, error) {
	parts := strings.SplitN(b64, "base64,", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid PDF data received from frontend")
	}

	pdfBytes, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %v", err)
	}

	options := runtime.SaveDialogOptions{
		DefaultFilename: "SSSD_Analysis_Report.pdf",
		Title:           "Save PDF Report",
		Filters: []runtime.FileFilter{
			{DisplayName: "PDF Document (*.pdf)", Pattern: "*.pdf"},
		},
	}
	filePath, err := runtime.SaveFileDialog(a.ctx, options)
	if err != nil {
		return "", err
	}

	if filePath == "" {
		return "cancelled", nil
	}

	err = os.WriteFile(filePath, pdfBytes, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %v", err)
	}

	return filePath, nil
}

// SaveTXT opens a native Save dialog and writes the report to a TXT file
func (a *App) SaveTXT(report ReportData) (string, error) {
	// ... Re-use your existing SaveTXT logic ...
	options := runtime.SaveDialogOptions{
		DefaultFilename: "SSSD_Analysis_Report.txt",
		Title:           "Save TXT Report",
		Filters: []runtime.FileFilter{
			{DisplayName: "Text Document (*.txt)", Pattern: "*.txt"},
		},
	}
	filePath, err := runtime.SaveFileDialog(a.ctx, options)
	if err != nil {
		return "", err
	}
	if filePath == "" {
		return "cancelled", nil
	}
	txtContent := buildTextReport(report)
	err = os.WriteFile(filePath, []byte(txtContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %v", err)
	}
	return filePath, nil
}
