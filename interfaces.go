// interfaces.go - Core interfaces for dependency injection and testability

package main

// FileScanner defines the interface for file scanning operations
// This interface enables dependency injection and makes the code more testable
type FileScanner interface {
	// ScanFiles streams files line-by-line and executes a callback for each line
	ScanFiles(dirPath string, files []string, lineFunc func(line string)) error

	// AnyFileContains checks if any file contains a specific pattern (early exit)
	AnyFileContains(dirPath string, files []string, search string) bool

	// ReadFileSafe reads a small file directly into a string
	ReadFileSafe(dirPath string, fileName string) string

	// ExtractSection extracts a specific section from a file
	ExtractSection(dirPath string, fileName string, header string) string
}

// ArchiveExtractor defines the interface for archive extraction operations
type ArchiveExtractor interface {
	// ExtractArchiveToTemp safely extracts an archive to a temporary directory
	ExtractArchiveToTemp(archivePath string, progressFunc func(string, int)) (string, error)
}

// ReportGenerator defines the interface for report generation
type ReportGenerator interface {
	// BuildTextReport generates a text report from analysis data
	BuildTextReport(report ReportData) string

	// WriteHTMLReportFile generates an HTML report file
	WriteHTMLReportFile(report ReportData, filename string)
}

// KBArticleMatcher defines the interface for knowledge base matching
type KBArticleMatcher interface {
	// MatchKBArticles matches KB articles against analysis data
	MatchKBArticles(dirPath string, report *ReportData)
}

// ProgressReporter defines the interface for progress reporting
type ProgressReporter interface {
	// ReportProgress reports analysis progress
	ReportProgress(message string, percentage int)
}

// AnalysisEngine defines the main analysis interface
type AnalysisEngine interface {
	// Analyze performs complete analysis on a directory
	Analyze(dirPath string, anonymize bool, progress ProgressReporter) ReportData
}

// Default implementations for backward compatibility

// defaultFileScanner provides default file scanning implementation
var defaultFileScanner FileScanner = &defaultFileScannerImpl{}

type defaultFileScannerImpl struct{}

func (d *defaultFileScannerImpl) ScanFiles(dirPath string, files []string, lineFunc func(line string)) error {
	return defaultFileProcessor.ScanFiles(dirPath, files, lineFunc)
}

func (d *defaultFileScannerImpl) AnyFileContains(dirPath string, files []string, search string) bool {
	return defaultFileProcessor.AnyFileContains(dirPath, files, search)
}

func (d *defaultFileScannerImpl) ReadFileSafe(dirPath string, fileName string) string {
	return defaultSafeFileReader.ReadFileSafe(dirPath, fileName)
}

func (d *defaultFileScannerImpl) ExtractSection(dirPath string, fileName string, header string) string {
	return defaultSectionExtractor.ExtractSection(dirPath, fileName, header)
}

// defaultArchiveExtractor provides default archive extraction implementation
var defaultArchiveExtractor ArchiveExtractor = &defaultArchiveExtractorImpl{}

type defaultArchiveExtractorImpl struct{}

func (d *defaultArchiveExtractorImpl) ExtractArchiveToTemp(archivePath string, progressFunc func(string, int)) (string, error) {
	return extractArchiveToTemp(archivePath, progressFunc)
}

// defaultReportGenerator provides default report generation implementation
var defaultReportGenerator ReportGenerator = &defaultReportGeneratorImpl{}

type defaultReportGeneratorImpl struct{}

func (d *defaultReportGeneratorImpl) BuildTextReport(report ReportData) string {
	return buildTextReport(report)
}

func (d *defaultReportGeneratorImpl) WriteHTMLReportFile(report ReportData, filename string) {
	writeHTMLReportFile(report, filename)
}

// defaultKBArticleMatcher provides default KB matching implementation
var defaultKBArticleMatcher KBArticleMatcher = &defaultKBArticleMatcherImpl{}

type defaultKBArticleMatcherImpl struct{}

func (d *defaultKBArticleMatcherImpl) MatchKBArticles(dirPath string, report *ReportData) {
	matchKBArticles(dirPath, report)
}

// defaultAnalysisEngine provides default analysis implementation
var defaultAnalysisEngine AnalysisEngine = &defaultAnalysisEngineImpl{}

type defaultAnalysisEngineImpl struct {
	scanner   FileScanner
	extractor ArchiveExtractor
	generator ReportGenerator
	kbMatcher KBArticleMatcher
}

func NewDefaultAnalysisEngine() AnalysisEngine {
	return &defaultAnalysisEngineImpl{
		scanner:   defaultFileScanner,
		extractor: defaultArchiveExtractor,
		generator: defaultReportGenerator,
		kbMatcher: defaultKBArticleMatcher,
	}
}

func (d *defaultAnalysisEngineImpl) Analyze(dirPath string, anonymize bool, progress ProgressReporter) ReportData {
	// Convert ProgressReporter to progress function
	var progressFunc func(string, int)
	if progress != nil {
		progressFunc = progress.ReportProgress
	}

	// Use existing analyzeData function for backward compatibility
	return analyzeData(dirPath, anonymize, progressFunc)
}
