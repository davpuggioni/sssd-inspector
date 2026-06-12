// utils.go - Legacy wrapper functions for backward compatibility
// All actual implementations have moved to pkg/ subpackages.
package main

import (
	"context"

	"sssd-inspector/pkg/analysis"
	"sssd-inspector/pkg/extract"
	"sssd-inspector/pkg/fileutil"
	"sssd-inspector/pkg/report"
)

// anyFileContains streams files line-by-line, instantly stopping if a match is found
func anyFileContains(dirPath string, files []string, search string) bool {
	return fileutil.DefaultFileProcessor.AnyFileContains(dirPath, files, search)
}

// scanFiles streams files line-by-line and executes a callback
func scanFiles(dirPath string, files []string, lineFunc func(line string)) {
	_ = fileutil.DefaultFileProcessor.ScanFiles(dirPath, files, lineFunc)
}

// extractSection streams a file and extracts a specific command block
func extractSection(dirPath string, fileName string, header string) string {
	return fileutil.DefaultSectionExtractor.ExtractSection(dirPath, fileName, header)
}

// readFileSafe reads a small file directly into a string
func readFileSafe(dirPath string, fileName string) string {
	lines := fileutil.GlobalFileCache.GetLines(dirPath, fileName)
	if lines == nil {
		return ""
	}
	return fileutil.DefaultSafeFileReader.ReadFileSafe(lines)
}

// isRelevantFile checks if a file is relevant for analysis
func isRelevantFile(name string) bool {
	return fileutil.DefaultFileFilter.IsRelevantFile(name)
}

// analyzeData delegates to the new package using appConfig for buffer settings
func analyzeData(dirPath string, anonymize bool, progressFunc func(string, int)) ReportData {
	ctx := analysis.NewAnalyzerContextWithBuffer(
		appConfig.GetBufferSizeBytes(),
		appConfig.GetMaxLineLengthBytes(),
	)
	return ctx.AnalyzeData(context.Background(), dirPath, anonymize, progressFunc)
}

// analyzeLogsOnly delegates to the new package using appConfig for buffer settings
func analyzeLogsOnly(dirPath string, logFiles []string, anonymize bool, progressFunc func(string, int)) ReportData {
	ctx := analysis.NewAnalyzerContextWithBuffer(
		appConfig.GetBufferSizeBytes(),
		appConfig.GetMaxLineLengthBytes(),
	)
	return ctx.AnalyzeLogsOnly(context.Background(), dirPath, logFiles, anonymize, progressFunc)
}

// extractArchiveToTemp extracts an XZ archive to a temporary directory
func extractArchiveToTemp(archivePath string, progressFunc func(string, int)) (string, error) {
	return extract.ExtractArchiveToTemp(archivePath, progressFunc)
}

// buildTextReport generates a text report from the analysis data
func buildTextReport(data ReportData) string {
	return report.BuildTextReport(data)
}

// writeHTMLReportFile generates an HTML report file
func writeHTMLReportFile(data ReportData, filename string) {
	report.WriteHTMLReportFile(data, filename)
}
