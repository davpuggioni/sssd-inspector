//go:build !cli

// app_export_test.go
//
// Tests for the GUI export paths (SavePDF, SaveJSON, SaveTXT) and the file
// browser. These exercise the native dialogs through the seams declared in
// app.go (openFileDialogFn / saveFileDialogFn), so the Wails runtime — which
// would call log.Fatalf and kill the test process without a frontend context —
// is never invoked.
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"sssd-inspector/constants"
)

// errFakeDialog is the sentinel returned by the stubbed dialog functions.
var errFakeDialog = errors.New("fake dialog failure")

// withSaveDialog points the save-dialog seam at path for the duration of the
// test, restoring the real implementation afterwards.
func withSaveDialog(t *testing.T, path string) {
	t.Helper()
	old := saveFileDialogFn
	t.Cleanup(func() { saveFileDialogFn = old })
	saveFileDialogFn = func(context.Context, runtime.SaveDialogOptions) (string, error) {
		return path, nil
	}
}

// withSaveDialogError makes the save dialog fail, covering the error branch.
func withSaveDialogError(t *testing.T) {
	t.Helper()
	old := saveFileDialogFn
	t.Cleanup(func() { saveFileDialogFn = old })
	saveFileDialogFn = func(context.Context, runtime.SaveDialogOptions) (string, error) {
		return "", errFakeDialog
	}
}

// analyzedApp returns an App with a completed analysis of a fixture.
func analyzedApp(t *testing.T, anonymize bool) (*App, ReportData) {
	t.Helper()
	dir := writeSupportconfigFixture(t)
	app := NewApp()
	report, err := app.Analyze(dir, anonymize)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	return app, report
}

// TestAppSaveJSON_WritesValidReport drives SaveJSON end-to-end against a
// temporary destination file, including the Wails-style JSON round trip.
func TestAppSaveJSON_WritesValidReport(t *testing.T) {
	app, report := analyzedApp(t, false)

	out := filepath.Join(t.TempDir(), "report.json")
	withSaveDialog(t, out)

	saved, err := app.SaveJSON(roundTripReport(t, report))
	if err != nil {
		t.Fatalf("SaveJSON: %v", err)
	}
	if saved != out {
		t.Errorf("SaveJSON returned %q, want %q", saved, out)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("report not written: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Errorf("written file is not valid JSON: %v", err)
	}
	if len(decoded) == 0 {
		t.Errorf("written JSON report is empty")
	}
}

// TestAppSaveTXT_WritesReport drives SaveTXT end-to-end.
func TestAppSaveTXT_WritesReport(t *testing.T) {
	app, report := analyzedApp(t, false)

	out := filepath.Join(t.TempDir(), "report.txt")
	withSaveDialog(t, out)

	saved, err := app.SaveTXT(report)
	if err != nil {
		t.Fatalf("SaveTXT: %v", err)
	}
	if saved != out {
		t.Errorf("SaveTXT returned %q, want %q", saved, out)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("report not written: %v", err)
	}
	if len(data) == 0 {
		t.Errorf("written TXT report is empty")
	}
}

// TestAppSavePDF_WritesDecodedPayload drives SavePDF end-to-end: the base64
// data URL sent by the frontend must be decoded and written verbatim.
func TestAppSavePDF_WritesDecodedPayload(t *testing.T) {
	app, _ := analyzedApp(t, false)

	out := filepath.Join(t.TempDir(), "report.pdf")
	withSaveDialog(t, out)

	payload := []byte("%PDF-1.4 fake pdf body\n")
	dataURL := "data:application/pdf;base64," + base64.StdEncoding.EncodeToString(payload)

	saved, err := app.SavePDF(dataURL)
	if err != nil {
		t.Fatalf("SavePDF: %v", err)
	}
	if saved != out {
		t.Errorf("SavePDF returned %q, want %q", saved, out)
	}
	written, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("file not written: %v", err)
	}
	if string(written) != string(payload) {
		t.Errorf("written payload = %q, want %q", written, payload)
	}
}

// TestAppSavePDF_RejectsMalformedDataURL: a payload without the base64 marker
// must be rejected before any dialog interaction.
func TestAppSavePDF_RejectsMalformedDataURL(t *testing.T) {
	app, _ := analyzedApp(t, false)
	withSaveDialogError(t) // must not be reached

	if _, err := app.SavePDF("not-a-data-url"); err == nil {
		t.Errorf("expected an error for a malformed PDF data URL")
	}
}

// TestAppSave_AnonymizedReportHasNoPII: the export path must write the
// anonymized report produced by Analyze.
func TestAppSave_AnonymizedReportHasNoPII(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "sssd.txt", "Failed to connect to 10.11.12.13 for realm corp.example.com\n")
	writeFile(t, dir, "rpm.txt", "sssd-2.9.4-150500.x86_64\n")

	app := NewApp()
	report, err := app.Analyze(dir, true)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	out := filepath.Join(t.TempDir(), "report.json")
	withSaveDialog(t, out)
	if _, err := app.SaveJSON(report); err != nil {
		t.Fatalf("SaveJSON: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "10.11.12.13") {
		t.Errorf("anonymized export leaked the raw IP")
	}
}

// TestAppSave_DialogCancelled: an empty destination means "user cancelled" and
// must be reported as such instead of being treated as an error.
func TestAppSave_DialogCancelled(t *testing.T) {
	app, report := analyzedApp(t, false)
	withSaveDialog(t, "")

	saved, err := app.SaveTXT(report)
	if err != nil {
		t.Fatalf("cancellation must not be an error: %v", err)
	}
	if saved != constants.ErrCancelled {
		t.Errorf("result = %q, want %q", saved, constants.ErrCancelled)
	}
}

// TestAppSave_DialogError: a failing dialog surfaces its error.
func TestAppSave_DialogError(t *testing.T) {
	app, report := analyzedApp(t, false)
	withSaveDialogError(t)

	if _, err := app.SaveJSON(report); err == nil {
		t.Errorf("expected the dialog error to be returned")
	}
	if _, err := app.SaveTXT(report); err == nil {
		t.Errorf("expected the dialog error to be returned by SaveTXT")
	}
	if _, err := app.SavePDF("data:application/pdf;base64,AAAA"); err == nil {
		t.Errorf("expected the dialog error to be returned by SavePDF")
	}
}

// TestAppSave_WriteErrorIsReported: an unwritable destination must surface an
// error rather than pretending the export succeeded.
func TestAppSave_WriteErrorIsReported(t *testing.T) {
	app, report := analyzedApp(t, false)
	withSaveDialog(t, filepath.Join(t.TempDir(), "missing-dir", "report.json"))

	if _, err := app.SaveJSON(report); err == nil {
		t.Errorf("expected a write error for a non-existent directory")
	}
}

// TestAppOpenFileBrowser_ReturnsSelectedPath.
func TestAppOpenFileBrowser_ReturnsSelectedPath(t *testing.T) {
	old := openFileDialogFn
	t.Cleanup(func() { openFileDialogFn = old })
	openFileDialogFn = func(context.Context, runtime.OpenDialogOptions) (string, error) {
		return "/tmp/selected.txz", nil
	}

	app := NewApp()
	got, err := app.OpenFileBrowser()
	if err != nil {
		t.Fatalf("OpenFileBrowser: %v", err)
	}
	if got != "/tmp/selected.txz" {
		t.Errorf("path = %q, want /tmp/selected.txz", got)
	}
}

// TestAppOpenFileBrowser_PropagatesError.
func TestAppOpenFileBrowser_PropagatesError(t *testing.T) {
	old := openFileDialogFn
	t.Cleanup(func() { openFileDialogFn = old })
	openFileDialogFn = func(context.Context, runtime.OpenDialogOptions) (string, error) {
		return "", errFakeDialog
	}

	app := NewApp()
	if _, err := app.OpenFileBrowser(); err == nil {
		t.Errorf("expected the dialog error to be returned")
	}
}

// TestAppStartup_StoresContext: startup() is the Wails lifecycle hook that
// hands the frontend context to the App; every dialog depends on it.
func TestAppStartup_StoresContext(t *testing.T) {
	app := NewApp()
	ctx := context.Background()
	app.startup(ctx)
	if app.ctx != ctx {
		t.Errorf("startup must store the provided context")
	}
}
