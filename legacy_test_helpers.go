// legacy_test_helpers.go - Bridge functions for existing tests
// These functions delegate to the new pkg/analysis package.
// They exist only to support existing tests without modification.
// When tests are migrated to the new package, this file can be removed.
package main

import "sssd-inspector/pkg/analysis"

// testAnalyzer returns a shared AnalyzerContext for use in test helpers
func testAnalyzer() *analysis.AnalyzerContext {
	return analysis.NewAnalyzerContext()
}

// analyzeMACStatus delegates to pkg/analysis
func analyzeMACStatus(dirPath string, report *ReportData) {
	testAnalyzer().AnalyzeMACStatus(dirPath, report)
}

// analyzeSSSDConfigAndLogs delegates to pkg/analysis
func analyzeSSSDConfigAndLogs(dirPath string, report *ReportData) {
	testAnalyzer().AnalyzeSSSDConfigAndLogs(dirPath, report)
}

// analyzeTime delegates to pkg/analysis
func analyzeTime(dirPath string, report *ReportData) {
	testAnalyzer().AnalyzeTime(dirPath, report)
}

// analyzeNSSwitch delegates to pkg/analysis
func analyzeNSSwitch(dirPath string, report *ReportData) {
	testAnalyzer().AnalyzeNSSwitch(dirPath, report)
}

// analyzePerformance delegates to pkg/analysis
func analyzePerformance(dirPath string, report *ReportData) {
	testAnalyzer().AnalyzePerformance(dirPath, report)
}

// analyzePackages delegates to pkg/analysis
func analyzePackages(dirPath string, report *ReportData) {
	testAnalyzer().AnalyzePackages(dirPath, report)
}

// analyzeSSSDFilePermissions delegates to pkg/analysis
func analyzeSSSDFilePermissions(dirPath string, report *ReportData) {
	testAnalyzer().AnalyzeSSSDFilePermissions(dirPath, report)
}

// analyzeDiskSpace delegates to pkg/analysis
func analyzeDiskSpace(dirPath string, report *ReportData) {
	testAnalyzer().AnalyzeDiskSpace(dirPath, report)
}

// analyzeServices delegates to pkg/analysis
func analyzeServices(dirPath string, report *ReportData) {
	testAnalyzer().AnalyzeServices(dirPath, report)
}

// analyzeHostnameAndFQDN delegates to pkg/analysis
func analyzeHostnameAndFQDN(dirPath string, report *ReportData) {
	testAnalyzer().AnalyzeHostnameAndFQDN(dirPath, report)
}

// analyzeOSAndHardware delegates to pkg/analysis
func analyzeOSAndHardware(dirPath string, report *ReportData) {
	testAnalyzer().AnalyzeOSAndHardware(dirPath, report)
}

// matchKBArticles delegates to pkg/analysis
func matchKBArticles(dirPath string, report *ReportData) {
	testAnalyzer().MatchKBArticles(dirPath, report)
}
