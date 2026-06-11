// legacy_test_helpers.go - Bridge functions for existing tests
// These functions delegate to the new pkg/analysis package.
// They exist only to support existing tests without modification.
// When tests are migrated to the new package, this file can be removed.
package main

// analyzeMACStatus delegates to pkg/analysis
func analyzeMACStatus(dirPath string, report *ReportData) {
	globalAnalyzer.AnalyzeMACStatus(dirPath, report)
}

// analyzeSSSDConfigAndLogs delegates to pkg/analysis
func analyzeSSSDConfigAndLogs(dirPath string, report *ReportData) {
	globalAnalyzer.AnalyzeSSSDConfigAndLogs(dirPath, report)
}

// analyzeTime delegates to pkg/analysis
func analyzeTime(dirPath string, report *ReportData) {
	globalAnalyzer.AnalyzeTime(dirPath, report)
}

// analyzeNSSwitch delegates to pkg/analysis
func analyzeNSSwitch(dirPath string, report *ReportData) {
	globalAnalyzer.AnalyzeNSSwitch(dirPath, report)
}

// analyzePerformance delegates to pkg/analysis
func analyzePerformance(dirPath string, report *ReportData) {
	globalAnalyzer.AnalyzePerformance(dirPath, report)
}

// analyzePackages delegates to pkg/analysis
func analyzePackages(dirPath string, report *ReportData) {
	globalAnalyzer.AnalyzePackages(dirPath, report)
}

// analyzeSSSDFilePermissions delegates to pkg/analysis
func analyzeSSSDFilePermissions(dirPath string, report *ReportData) {
	globalAnalyzer.AnalyzeSSSDFilePermissions(dirPath, report)
}

// analyzeDiskSpace delegates to pkg/analysis
func analyzeDiskSpace(dirPath string, report *ReportData) {
	globalAnalyzer.AnalyzeDiskSpace(dirPath, report)
}

// analyzeServices delegates to pkg/analysis
func analyzeServices(dirPath string, report *ReportData) {
	globalAnalyzer.AnalyzeServices(dirPath, report)
}

// analyzeHostnameAndFQDN delegates to pkg/analysis
func analyzeHostnameAndFQDN(dirPath string, report *ReportData) {
	globalAnalyzer.AnalyzeHostnameAndFQDN(dirPath, report)
}

// analyzeOSAndHardware delegates to pkg/analysis
func analyzeOSAndHardware(dirPath string, report *ReportData) {
	globalAnalyzer.AnalyzeOSAndHardware(dirPath, report)
}

// matchKBArticles delegates to pkg/analysis
func matchKBArticles(dirPath string, report *ReportData) {
	globalAnalyzer.MatchKBArticles(dirPath, report)
}
