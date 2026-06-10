// loaders.go - Legacy wrapper for archive extraction
// All actual implementations have moved to pkg/extract.
package main

import "sssd-inspector/pkg/extract"

// extractArchiveToTemp delegates to the new package
func extractArchiveToTemp(archivePath string, progressFunc func(string, int)) (string, error) {
	return extract.ExtractArchiveToTemp(archivePath, progressFunc)
}
