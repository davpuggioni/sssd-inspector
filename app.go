package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"sssd-inspector/internal/constants"
	"sssd-inspector/internal/analyzer" // Imported to access core schemas and engines
)

// App represents the main application structure for SSSD Inspector.
type App struct {
	ctx context.Context // Wails application context for runtime operations
}

// NewApp creates and returns a new App instance.
func NewApp() *App {
	return &App{}
}

// startup is called by the Wails framework when the application starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// emitEvent safely emits a Wails runtime event, silently skipping if no
// valid context is available (e.g., during unit tests or CLI mode).
func (a *App) emitEvent(event string, args ...interface{}) {
	if a.ctx == nil {
		return
	}
	runtime.EventsEmit(a.ctx, event, args...)
}

// OpenFileBrowser opens a native file browser dialog for selecting supportconfig archives.
func (a *App) OpenFileBrowser() (string, error) {
	options := runtime.OpenDialogOptions{
		Title: constants.TitleSelectArchive,
		Filters: []runtime.FileFilter{
			{DisplayName: constants.FilterSupportconfig, Pattern: constants.PatternSupportconfig},
			{DisplayName: constants.FilterAllFiles, Pattern: constants.PatternAllFiles},
		},
	}
	return runtime.OpenFileDialog(a.ctx, options)
}

// Analyze handles the secure extraction, routing, and cleanup of the target logs.
// Exposes the core analysis routine to the JavaScript frontend.
func (a *App) Analyze(targetPath string, anonymize bool) (analyzer.ReportData, error) {
	// Emit progress directly to the Javascript UI
	progressFunc := func(msg string, pct int) {
		a.emitEvent(constants.EventAnalyzeProgress, msg, pct)
	}

	progressFunc(constants.MsgInitializing, constants.ProgressStart)

	info, err := os.Stat(targetPath)
	if err != nil {
		return analyzer.ReportData{}, fmt.Errorf("error accessing path: %w", err)
	}

	var dirPath string

	if info.IsDir() {
		dirPath = targetPath
	} else {
		// NOTE: Ensure ExtractArchiveToTemp is capitalized (exported) inside internal/analyzer
		dirPath, err = analyzer.ExtractArchiveToTemp(targetPath, progressFunc)
		if err != nil {
			return analyzer.ReportData{}, err
		}
		defer os.RemoveAll(dirPath)
	}

	// NOTE: Ensure AnalyzeData is capitalized (exported) inside internal/analyzer
	report := analyzer.AnalyzeData(dirPath, anonymize, progressFunc)
	progressFunc(constants.MsgAnalysisComplete, constants.ProgressComplete)
	return report, nil
}

// SavePDF saves a base64-encoded PDF report to a user-selected location.
func (a *App) SavePDF(b64 string) (string, error) {
	parts := strings.SplitN(b64, "base64,", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid PDF data received from frontend")
	}

	pdfBytes, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	options := runtime.SaveDialogOptions{
		DefaultFilename: constants.DefaultPDFName,
		Title:           constants.TitleSavePDF,
		Filters: []runtime.FileFilter{
			{DisplayName: constants.FilterPDF, Pattern: constants.PatternPDF},
		},
	}
	filePath, err := runtime.SaveFileDialog(a.ctx, options)
	if err != nil {
		return "", err
	}

	if filePath == "" {
		return constants.ErrCancelled, nil
	}

	err = os.WriteFile(filePath, pdfBytes, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filePath, nil
}

// SaveTXT saves a text report based on the provided ReportData structure.
func (a *App) SaveTXT(report analyzer.ReportData) (string, error) {
	options := runtime.SaveDialogOptions{
		DefaultFilename: constants.DefaultTXTName,
		Title:           constants.TitleSaveTXT,
		Filters: []runtime.FileFilter{
			{DisplayName: constants.FilterText, Pattern: constants.PatternText},
		},
	}
	filePath, err := runtime.SaveFileDialog(a.ctx, options)
	if err != nil {
		return "", err
	}
	if filePath == "" {
		return constants.ErrCancelled, nil
	}

	// NOTE: Ensure BuildTextReport is capitalized (exported) inside internal/analyzer
	txtContent := analyzer.BuildTextReport(report)
	err = os.WriteFile(filePath, []byte(txtContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	return filePath, nil
}
