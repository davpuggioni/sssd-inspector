// utils.go - Legacy wrapper functions
// All actual implementations have moved to pkg/fileutil.
// These functions maintain backward compatibility.
package main

import (
	"context"
	"fmt"

	"sssd-inspector/pkg/analysis"
	"sssd-inspector/pkg/fileutil"
)

// Global analysis context for backward compatibility
var globalAnalyzer = analysis.NewAnalyzerContext()

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

// Context-aware scanning functions
func scanFilesWithContext(ctx context.Context, dirPath string, files []string, lineFunc func(line string)) {
	for _, name := range files {
		select {
		case <-ctx.Done():
			fmt.Printf("Warning: scan of %s cancelled: %v\n", name, ctx.Err())
			return
		default:
		}
		_ = fileutil.DefaultFileProcessor.ScanFiles(dirPath, []string{name}, lineFunc)
	}
}

// analyzeData delegates to the new package
func analyzeData(dirPath string, anonymize bool, progressFunc func(string, int)) ReportData {
	return globalAnalyzer.AnalyzeData(dirPath, anonymize, progressFunc)
}

// analyzeLogsOnly delegates to the new package
func analyzeLogsOnly(dirPath string, logFiles []string, anonymize bool, progressFunc func(string, int)) ReportData {
	return globalAnalyzer.AnalyzeLogsOnly(dirPath, logFiles, anonymize, progressFunc)
}
