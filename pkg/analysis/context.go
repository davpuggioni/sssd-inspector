// Package analysis provides SSSD supportconfig analysis functionality
package analysis

import (
	"sssd-inspector/constants"
	"sssd-inspector/pkg/fileutil"
	"sssd-inspector/pkg/types"
)

// AnalyzerContext provides dependency injection for analysis functions.
// It encapsulates all dependencies needed by the analyzers, making them
// testable without global state.
type AnalyzerContext struct {
	Scanner    *fileutil.FileProcessor
	SecExtract *fileutil.SectionExtractor
	FileCache  *fileutil.FileCache
	RegexCache *fileutil.RegexCache
}

// NewAnalyzerContext creates a new AnalyzerContext with default dependencies.
func NewAnalyzerContext() *AnalyzerContext {
	return &AnalyzerContext{
		Scanner:    fileutil.DefaultFileProcessor,
		SecExtract: fileutil.DefaultSectionExtractor,
		FileCache:  fileutil.GlobalFileCache,
		RegexCache: fileutil.GlobalRegexCache,
	}
}

// AnalyzeData is the main orchestrator. It routes the streaming directory path
// to specialized analyzers. Now uses Single-Pass scanning for log files.
func (ctx *AnalyzerContext) AnalyzeData(dirPath string, anonymize bool, progressFunc func(string, int)) types.ReportData {
	return ctx.analyzeData(dirPath, anonymize, progressFunc)
}

// AnalyzeLogsOnly performs a lightweight analysis on raw SSSD log files only.
func (ctx *AnalyzerContext) AnalyzeLogsOnly(dirPath string, logFiles []string, anonymize bool, progressFunc func(string, int)) types.ReportData {
	return ctx.analyzeLogsOnly(dirPath, logFiles, anonymize, progressFunc)
}

// ScanFiles is a convenience wrapper for scanning files
func (ctx *AnalyzerContext) ScanFiles(dirPath string, files []string, lineFunc func(line string)) {
	_ = ctx.Scanner.ScanFiles(dirPath, files, lineFunc)
}

// AnyFileContains checks if any file contains a pattern
func (ctx *AnalyzerContext) AnyFileContains(dirPath string, files []string, search string) bool {
	return ctx.Scanner.AnyFileContains(dirPath, files, search)
}

// ExtractSection extracts a specific section from a config file
func (ctx *AnalyzerContext) ExtractSection(dirPath string, fileName string, header string) string {
	return ctx.SecExtract.ExtractSection(dirPath, fileName, header)
}

// ReadFileSafe reads a config file using the file cache
func (ctx *AnalyzerContext) ReadFileSafe(dirPath string, fileName string) string {
	lines := ctx.FileCache.GetLines(dirPath, fileName)
	if lines == nil {
		return ""
	}
	return fileutil.DefaultSafeFileReader.ReadFileSafe(lines)
}

// getDecade checks the first digit of a version number for MAC type checks
func getDecade(v string) int {
	if len(v) == 0 {
		return 0
	}
	return int(v[0] - '0')
}

// Constants used by analyzers
func unknown() string       { return constants.StatusUnknown }
func notRunning() string    { return constants.StatusNotRunning }
func notConfigured() string { return constants.StatusNotConfigured }
