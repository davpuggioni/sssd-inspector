package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

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

// Analyze handles the secure extraction, routing, and cleanup of the target logs
func (a *App) Analyze(targetPath string, anonymize bool) (ReportData, error) {
	// Emit progress directly to the Javascript UI
	progressFunc := func(msg string, pct int) {
		runtime.EventsEmit(a.ctx, "analyze-progress", msg, pct)
	}

	progressFunc("Initializing streaming engine...", 0)

	info, err := os.Stat(targetPath)
	if err != nil {
		return ReportData{}, fmt.Errorf("error accessing path: %v", err)
	}

	var dirPath string

	if info.IsDir() {
		dirPath = targetPath
	} else {
		dirPath, err = extractArchiveToTemp(targetPath, progressFunc)
		if err != nil {
			return ReportData{}, err
		}
		// ALWAYS clean up the massive logs off the disk after analysis!
		defer os.RemoveAll(dirPath)
	}

	report := analyzeData(dirPath, anonymize, progressFunc)
	progressFunc("Analysis Complete!", 100)
	return report, nil
}

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

func (a *App) SaveTXT(report ReportData) (string, error) {
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
