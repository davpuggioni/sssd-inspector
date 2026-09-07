//go:build !cli

// Package main provides the main application structure for SSSD Inspector
// This file contains the App struct and methods that handle the Wails framework integration,
// providing the bridge between the Go backend and the JavaScript frontend.
package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"sssd-inspector/constants"
)

// App represents the main application structure for SSSD Inspector.
// It holds the application context and provides methods for file operations,
// analysis, and report generation that are exposed to the frontend.
type App struct {
	ctx context.Context // Wails application context for runtime operations
}

// NewApp creates and returns a new App instance.
// This function is called by the Wails framework during application initialization.
func NewApp() *App {
	return &App{}
}

// startup is called by the Wails framework when the application starts.
// It initializes the application context and performs any necessary setup.
// This method is automatically called by Wails and should not be called manually.
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
// It returns the selected file path or an empty string if the user cancels the dialog.
// The dialog filters for supportconfig archive files (.txz, .tar.xz) but also allows
// selection of all file types.
//
// Returns:
//   - string: The selected file path, or empty string if cancelled
//   - error: Any error that occurred while opening the dialog
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

// extractIfArchive returns the path to a supportconfig directory.
// If path points to a directory it is returned unchanged; if it points to
// an archive it is extracted to a temporary directory (whose path is
// returned). The caller is responsible for cleaning up the temp dir.
func extractIfArchive(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return path, nil
	}
	// Archive: extract to temp (progress callback is a no-op; the
	// higher-level caller drives its own progress reporting).
	return extractArchiveToTemp(path, func(msg string, pct int) {})
}

// Analyze handles the secure extraction, routing, and cleanup of the target logs.
// This is the main analysis method that orchestrates the entire diagnostic process.
// It supports both directory and archive file inputs, provides real-time progress
// updates via events, and automatically cleans up temporary files.
//
// Parameters:
//   - targetPath: Path to the supportconfig directory or archive file
//   - anonymize: Whether to redact PII (Personally Identifiable Information) from the report
//
// Returns:
//   - ReportData: Comprehensive analysis results containing system information,
//     detected problems, warnings, and knowledge base matches
//   - error: Any error that occurred during the analysis process
//
// Events:
//   - Emits "analyze-progress" events with (message, percentage) tuples during processing
func (a *App) Analyze(targetPath string, anonymize bool) (ReportData, error) {
	// Emit progress directly to the Javascript UI
	progressFunc := func(msg string, pct int) {
		a.emitEvent(constants.EventAnalyzeProgress, msg, pct)
	}

	progressFunc(constants.MsgInitializing, constants.ProgressStart)

	info, err := os.Stat(targetPath)
	if err != nil {
		return ReportData{}, fmt.Errorf("error accessing path: %w", err)
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
	progressFunc(constants.MsgAnalysisComplete, constants.ProgressComplete)
	return report, nil
}

// SavePDF saves a base64-encoded PDF report to a user-selected location.
// This method handles the decoding of base64 PDF data and provides a native
// save dialog for the user to choose the output location.
//
// Parameters:
//   - b64: Base64-encoded PDF data with data URL prefix (e.g., "data:application/pdf;base64,...")
//
// Returns:
//   - string: Path where the file was saved, or "cancelled" if the user cancelled
//   - error: Any error that occurred during the save operation
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

// SaveJSON saves a JSON-structured report to a user-selected location.
// Unlike the TXT/PDF exports it contains the full machine-readable data:
// executive summary, config findings with provenance, timeline and KB
// evidence — suitable for tooling and long-term case archiving.
func (a *App) SaveJSON(report ReportData) (string, error) {
	options := runtime.SaveDialogOptions{
		DefaultFilename: constants.DefaultJSONName,
		Title:           "Save JSON Report",
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON Report", Pattern: "*.json"},
		},
	}
	filePath, err := runtime.SaveFileDialog(a.ctx, options)
	if err != nil {
		return "", err
	}
	if filePath == "" {
		return constants.ErrCancelled, nil
	}
	data, err := buildJSONReport(report)
	if err != nil {
		return "", fmt.Errorf("failed to serialize JSON report: %w", err)
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	return filePath, nil
}

// SaveTXT saves a text report based on the provided ReportData structure.
// This method converts the analysis results to a formatted text report and
// provides a native save dialog for the user to choose the output location.
//
// Parameters:
//   - report: ReportData structure containing the analysis results
//
// Returns:
//   - string: Path where the file was saved, or "cancelled" if the user cancelled
//   - error: Any error that occurred during the save operation
func (a *App) SaveTXT(report ReportData) (string, error) {
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
	txtContent := buildTextReport(report)
	err = os.WriteFile(filePath, []byte(txtContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	return filePath, nil
}
