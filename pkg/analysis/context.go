// Package analysis provides SSSD supportconfig analysis functionality
package analysis

import (
	"context"
	"strings"

	"sssd-inspector/pkg/fileutil"
	"sssd-inspector/pkg/types"
)

// AnalyzerContext provides dependency injection for analysis functions.
type AnalyzerContext struct {
	Scanner    *fileutil.FileProcessor
	SecExtract *fileutil.SectionExtractor
	FileCache  *fileutil.FileCache
	RegexCache *fileutil.RegexCache
	ctx        context.Context // current analysis context, set before each run
}

// NewAnalyzerContext creates a new AnalyzerContext with default dependencies.
func NewAnalyzerContext() *AnalyzerContext {
	return &AnalyzerContext{
		Scanner:    fileutil.DefaultFileProcessor,
		SecExtract: fileutil.DefaultSectionExtractor,
		FileCache:  fileutil.GlobalFileCache,
		RegexCache: fileutil.GlobalRegexCache,
		ctx:        context.Background(),
	}
}

// NewAnalyzerContextWithBuffer creates a new AnalyzerContext with custom buffer sizes.
func NewAnalyzerContextWithBuffer(bufferSize, maxLineLength int) *AnalyzerContext {
	return &AnalyzerContext{
		Scanner:    fileutil.NewFileProcessorWithConfig(bufferSize, maxLineLength),
		SecExtract: fileutil.NewSectionExtractor(),
		FileCache:  fileutil.GlobalFileCache,
		RegexCache: fileutil.GlobalRegexCache,
		ctx:        context.Background(),
	}
}

// CTX returns the current analysis context (set before each AnalyzeData/AnalyzeLogsOnly call).
func (ctx *AnalyzerContext) CTX() context.Context {
	return ctx.ctx
}

// --- Public API methods ---

// AnalyzeData is the main orchestrator for full supportconfig analysis.
// Pass context.Background() for normal operation, or a cancellable context for early termination.
func (ctx *AnalyzerContext) AnalyzeData(c context.Context, dirPath string, anonymize bool, progressFunc func(string, int)) types.ReportData {
	ctx.ctx = c
	return ctx.analyzeData(dirPath, anonymize, progressFunc)
}

// AnalyzeLogsOnly performs lightweight analysis on raw SSSD log files.
// Pass context.Background() for normal operation, or a cancellable context for early termination.
func (ctx *AnalyzerContext) AnalyzeLogsOnly(c context.Context, dirPath string, logFiles []string, anonymize bool, progressFunc func(string, int)) types.ReportData {
	ctx.ctx = c
	return ctx.analyzeLogsOnly(dirPath, logFiles, anonymize, progressFunc)
}

// --- Test support methods (exported for use by tests) ---

// AnalyzeMACStatus delegates to private method
func (ctx *AnalyzerContext) AnalyzeMACStatus(dirPath string, report *types.ReportData) {
	ctx.analyzeMACStatus(dirPath, report)
}

// AnalyzeSSSDConfigAndLogs delegates to private method
func (ctx *AnalyzerContext) AnalyzeSSSDConfigAndLogs(dirPath string, report *types.ReportData) {
	ctx.analyzeSSSDConfigAndLogs(dirPath, report)
}

// AnalyzeTime delegates to private method
func (ctx *AnalyzerContext) AnalyzeTime(dirPath string, report *types.ReportData) {
	ctx.analyzeTime(dirPath, report)
}

// AnalyzeNSSwitch delegates to private method
func (ctx *AnalyzerContext) AnalyzeNSSwitch(dirPath string, report *types.ReportData) {
	ctx.analyzeNSSwitch(dirPath, report)
}

// AnalyzePerformance delegates to private method
func (ctx *AnalyzerContext) AnalyzePerformance(dirPath string, report *types.ReportData) {
	ctx.analyzePerformance(dirPath, report)
}

// AnalyzePackages delegates to private method
func (ctx *AnalyzerContext) AnalyzePackages(dirPath string, report *types.ReportData) {
	ctx.analyzePackages(dirPath, report)
}

// AnalyzeSSSDFilePermissions delegates to private method
func (ctx *AnalyzerContext) AnalyzeSSSDFilePermissions(dirPath string, report *types.ReportData) {
	ctx.analyzeSSSDFilePermissions(dirPath, report)
}

// AnalyzeDiskSpace delegates to private method
func (ctx *AnalyzerContext) AnalyzeDiskSpace(dirPath string, report *types.ReportData) {
	ctx.analyzeDiskSpace(dirPath, report)
}

// AnalyzeServices delegates to private method
func (ctx *AnalyzerContext) AnalyzeServices(dirPath string, report *types.ReportData) {
	ctx.analyzeServices(dirPath, report)
}

// AnalyzeHostnameAndFQDN delegates to private method
func (ctx *AnalyzerContext) AnalyzeHostnameAndFQDN(dirPath string, report *types.ReportData) {
	ctx.analyzeHostnameAndFQDN(dirPath, report)
}

// AnalyzeOSAndHardware delegates to private method
func (ctx *AnalyzerContext) AnalyzeOSAndHardware(dirPath string, report *types.ReportData) {
	ctx.analyzeOSAndHardware(dirPath, report)
}

// MatchKBArticles delegates to the KB matching logic (used by tests)
func (ctx *AnalyzerContext) MatchKBArticles(dirPath string, report *types.ReportData) {
	ctx.matchKBArticles(dirPath, report)
}

// buildErrorPatterns loads the SSSD error pattern dictionary from embedded YAML
func (ctx *AnalyzerContext) buildErrorPatterns(macType string) map[string]string {
	entries := ctx.loadErrorPatternEntries(macType)
	errorPatterns := make(map[string]string, len(entries))
	for _, entry := range entries {
		errorPatterns[entry.Pattern] = entry.Description
	}
	return errorPatterns
}

// analyzeSSSDConfigAndLogs performs combined config and log analysis (used by tests)
func (ctx *AnalyzerContext) analyzeSSSDConfigAndLogs(dirPath string, report *types.ReportData) {
	sssdConfContent := ctx.ReadFileSafe(dirPath, "sssd.conf")
	if sssdConfContent == "" {
		sssdConfContent = ctx.ExtractSection(dirPath, "sssd.txt", "# /etc/sssd/sssd.conf")
	}

	logFiles := []string{"sssd.txt", "messages", "messages.txt"}

	// Quick checks for known issues
	if ctx.AnyFileContains(dirPath, logFiles, "User account has expired") || ctx.AnyFileContains(dirPath, logFiles, "Clients credentials have been revoked") {
		report.Problems = append(report.Problems, "[AUTHENTICATION] Logs indicate an Active Directory user account is expired, locked, or credentials have been revoked.")
	}
	if ctx.AnyFileContains(dirPath, logFiles, "terminated by own WATCHDOG") {
		report.Warnings = append(report.Warnings, "[TUNING] Since a WATCHDOG termination was found, consider setting 'ignore_group_members = true' in sssd.conf to speed up ssh/sudo initial lookups.")
	}
	if ctx.AnyFileContains(dirPath, logFiles, "service key not available") || ctx.AnyFileContains(dirPath, logFiles, "TGT failed verification") || ctx.AnyFileContains(dirPath, logFiles, "KDC has no support for encryption type") {
		report.Warnings = append(report.Warnings, "[AD CRYPTO BUG] Crypto mismatch or 'service key not available' detected.")
	}

	errorPatterns := ctx.buildErrorPatterns(report.MACType)
	detectedLogErrors, timelineEvents := ctx.scanAndCollectErrors(dirPath, errorPatterns)

	report.Timeline = timelineEvents
	report.SSSDLogErrors = buildSortedLogErrors(detectedLogErrors)

	if sssdConfContent != "" {
		ctx.analyzeSSSDConfig(sssdConfContent, report)
	} else {
		report.Problems = append(report.Problems, "sssd.conf or SSSD configuration block not found in the supportconfig.")
	}

	if report.SssdService != "Running" {
		report.Problems = append(report.Problems, "sssd.service is not actively running.")
	}
}

// matchKBArticles performs KB matching with inline evidence scanning (used by tests)
func (ctx *AnalyzerContext) matchKBArticles(dirPath string, report *types.ReportData) {
	kbArticles := ctx.loadKBArticles(dirPath)
	if len(kbArticles) == 0 {
		return
	}

	logFiles := []string{"sssd.txt", "messages", "messages.txt", "sssd_pam.log"}
	evidence := make(map[string][]string)

	for _, article := range kbArticles {
		for _, pattern := range article.LogPatterns {
			ctx.ScanFiles(dirPath, logFiles, func(line string) {
				if strings.Contains(line, pattern) {
					evidence[article.TIDID] = append(evidence[article.TIDID], line)
				}
			})
		}
	}

	ctx.matchKBArticlesWithEvidence(dirPath, report, kbArticles, evidence)
}

// --- Convenience wrappers ---

// ScanFiles scans files line-by-line, checking context for cancellation between files.
func (ctx *AnalyzerContext) ScanFiles(dirPath string, files []string, lineFunc func(line string)) {
	_ = ctx.Scanner.ScanFilesContext(ctx.ctx, dirPath, files, lineFunc)
}

// AnyFileContains checks if any file contains a pattern, respecting context cancellation.
func (ctx *AnalyzerContext) AnyFileContains(dirPath string, files []string, search string) bool {
	return ctx.Scanner.AnyFileContainsContext(ctx.ctx, dirPath, files, search)
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
