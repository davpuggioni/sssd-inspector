// analyzer_core.go - Legacy wrapper for core analysis functions
// All actual implementations have moved to pkg/analysis.
package main

import (
	"sssd-inspector/pkg/analysis"
	"sssd-inspector/pkg/types"
)

// deduplicateProblems delegates to the new package
func deduplicateProblems(problems []string) []string {
	return analysis.DeduplicateProblems(problems)
}

// getSSSDVersion delegates to the new package
func getSSSDVersion(packages []string) (int, int) {
	return analysis.GetSSSDVersion(packages)
}

// anonymizeReport delegates to the new package
// ReportData is a type alias for types.ReportData, so the cast is safe
func anonymizeReport(r *ReportData) {
	analysis.AnonymizeReport((*types.ReportData)(r))
}
