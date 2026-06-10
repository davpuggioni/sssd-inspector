// report.go - Legacy wrapper for report generation
// All actual implementations have moved to pkg/report.
package main

import "sssd-inspector/pkg/report"

// buildTextReport delegates to the new package
func buildTextReport(data ReportData) string {
	return report.BuildTextReport(data)
}

// writeHTMLReportFile delegates to the new package
func writeHTMLReportFile(data ReportData, filename string) {
	report.WriteHTMLReportFile(data, filename)
}
