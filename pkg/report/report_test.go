package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sssd-inspector/pkg/types"
)

// TestBuildTextReport_Empty verifies text report generation with empty data
func TestBuildTextReport_Empty(t *testing.T) {
	data := types.ReportData{}
	result := BuildTextReport(data)

	if result == "" {
		t.Error("BuildTextReport should not return empty string")
	}

	if !strings.Contains(result, "SUPPORTCONFIG SSSD ANALYSIS REPORT") {
		t.Error("report should contain title")
	}
}

// TestBuildTextReport_WithProblems verifies problems are included
func TestBuildTextReport_WithProblems(t *testing.T) {
	data := types.ReportData{
		Timestamp:     "01:01:2026 12:00:00",
		AppVersion:    "1.0.0",
		SssdInstalled: true,
		Problems:      []string{"Problem A", "Problem B"},
		Warnings:      []string{"Warning X"},
	}

	result := BuildTextReport(data)
	if !strings.Contains(result, "Problem A") {
		t.Error("report should contain 'Problem A'")
	}
	if !strings.Contains(result, "Warning X") {
		t.Error("report should contain 'Warning X'")
	}
}

// TestBuildTextReport_WithErrors verifies SSSD log errors are rendered
func TestBuildTextReport_WithErrors(t *testing.T) {
	data := types.ReportData{
		SSSDLogErrors: []types.SSSDLogError{
			{Description: "Connection failed", Examples: []string{"error line 1"}},
		},
	}

	result := BuildTextReport(data)
	if !strings.Contains(result, "Connection failed") {
		t.Error("report should contain error description")
	}
}

// TestWriteHTMLReportFile verifies HTML report file generation
func TestWriteHTMLReportFile(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "report.html")

	data := types.ReportData{
		Timestamp:  "01:01:2026 12:00:00",
		AppVersion: "1.0.0",
		Problems:   []string{"Test problem"},
	}

	WriteHTMLReportFile(data, filename)

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatalf("HTML file was not created: %v", err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read HTML file: %v", err)
	}

	if !strings.Contains(string(content), "Test problem") {
		t.Error("HTML should contain the problem")
	}
	if !strings.Contains(string(content), "SSSD Supportconfig Analysis") {
		t.Error("HTML should contain report title")
	}
}

// TestWriteHTMLReportFile_Empty verifies HTML generation with empty data
func TestWriteHTMLReportFile_Empty(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "empty_report.html")

	data := types.ReportData{}
	WriteHTMLReportFile(data, filename)

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatalf("HTML file was not created: %v", err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read HTML file: %v", err)
	}

	if len(content) == 0 {
		t.Error("HTML file should not be empty")
	}
}
