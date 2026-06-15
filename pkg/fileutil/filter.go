package fileutil

import "sssd-inspector/constants"

// FileFilter determines which files are relevant for analysis
type FileFilter struct {
	relevantFiles map[string]bool
}

// NewFileFilter creates a new FileFilter with default relevant files
func NewFileFilter() *FileFilter {
	ff := &FileFilter{
		relevantFiles: make(map[string]bool),
	}
	for _, file := range constants.RelevantFiles() {
		ff.relevantFiles[file] = true
	}
	return ff
}

// IsRelevantFile checks if a file is relevant for analysis
func (ff *FileFilter) IsRelevantFile(name string) bool {
	return ff.relevantFiles[name]
}

// AddRelevantFile adds a file to the relevant files list
func (ff *FileFilter) AddRelevantFile(filename string) {
	ff.relevantFiles[filename] = true
}

// RemoveRelevantFile removes a file from the relevant files list
func (ff *FileFilter) RemoveRelevantFile(filename string) {
	delete(ff.relevantFiles, filename)
}

// GetRelevantFiles returns the list of all relevant files
func (ff *FileFilter) GetRelevantFiles() []string {
	files := make([]string, 0, len(ff.relevantFiles))
	for file := range ff.relevantFiles {
		files = append(files, file)
	}
	return files
}

// DefaultFileFilter is the default filter instance
var DefaultFileFilter = NewFileFilter()
