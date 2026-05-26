// Package utils provides utility functions for file processing and analysis
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sssd-inspector/constants"
)

// FileProcessor provides streaming file processing capabilities
type FileProcessor struct {
	bufferSize    int
	maxLineLength int
}

// NewFileProcessor creates a new FileProcessor with default settings
func NewFileProcessor() *FileProcessor {
	return &FileProcessor{
		bufferSize:    64 * 1024,   // 64KB default buffer
		maxLineLength: 1024 * 1024, // 1MB max line length
	}
}

// NewFileProcessorWithConfig creates a new FileProcessor with custom settings
func NewFileProcessorWithConfig(bufferSize, maxLineLength int) *FileProcessor {
	return &FileProcessor{
		bufferSize:    bufferSize,
		maxLineLength: maxLineLength,
	}
}

// anyFileContains streams files line-by-line, instantly stopping if a match is found (Near-zero RAM usage)
// It searches through multiple files for a specific pattern and returns true if found.
func (fp *FileProcessor) AnyFileContains(dirPath string, files []string, search string) bool {
	for _, name := range files {
		if fp.fileContains(filepath.Join(dirPath, name), search) {
			return true
		}
	}
	return false
}

// fileContains checks if a single file contains the search pattern
func (fp *FileProcessor) fileContains(filePath, search string) bool {
	f, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := fp.createScanner(f)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), search) {
			return true
		}
	}
	return false
}

// ScanFiles streams files line-by-line and executes a callback, discarding the bytes immediately
// This method processes multiple files and applies the provided function to each line.
func (fp *FileProcessor) ScanFiles(dirPath string, files []string, lineFunc func(line string)) error {
	for _, name := range files {
		if err := fp.scanFile(filepath.Join(dirPath, name), lineFunc); err != nil {
			// Log error but continue processing other files
			fmt.Printf("Warning: failed to scan file %s: %v\n", name, err)
		}
	}
	return nil
}

// scanFile processes a single file and applies the line function
func (fp *FileProcessor) scanFile(filePath string, lineFunc func(line string)) error {
	f, err := os.Open(filePath)
	if err != nil {
		// If the file simply doesn't exist in the supportconfig, silently ignore it.
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer f.Close()

	scanner := fp.createScanner(f)
	for scanner.Scan() {
		lineFunc(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error for file %s: %w", filePath, err)
	}
	return nil
}

// createScanner creates a buffered scanner with configured limits
func (fp *FileProcessor) createScanner(f *os.File) *bufio.Scanner {
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, fp.bufferSize)
	scanner.Buffer(buf, fp.maxLineLength)
	return scanner
}

// SectionExtractor handles extraction of specific sections from files
type SectionExtractor struct {
	processor *FileProcessor
}

// NewSectionExtractor creates a new SectionExtractor
func NewSectionExtractor() *SectionExtractor {
	return &SectionExtractor{
		processor: NewFileProcessor(),
	}
}

// ExtractSection streams a file and extracts a specific command block
// It searches for a header line and extracts all content until a terminator is found.
func (se *SectionExtractor) ExtractSection(dirPath string, fileName string, header string) string {
	filePath := filepath.Join(dirPath, fileName)
	f, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := se.processor.createScanner(f)
	var sb strings.Builder
	inSection := false

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, header) {
			inSection = true
			continue
		}
		if inSection {
			if strings.HasPrefix(line, "#==[") {
				break
			}
			sb.WriteString(line + "\n")
		}
	}
	return sb.String()
}

// SafeFileReader handles safe reading of small configuration files
type SafeFileReader struct{}

// NewSafeFileReader creates a new SafeFileReader
func NewSafeFileReader() *SafeFileReader {
	return &SafeFileReader{}
}

// ReadFileSafe reads a small file directly into a string
// This method is designed for tiny configuration files like sssd.conf.
// For large files, use the streaming methods instead.
// Uses the global FileCache to avoid reading the same file multiple times
// during a single analysis pass.
func (sfr *SafeFileReader) ReadFileSafe(dirPath string, fileName string) string {
	// Use FileCache to avoid reading the same config file multiple times
	lines := globalFileCache.GetLines(dirPath, fileName)
	if lines == nil {
		return ""
	}
	return strings.Join(lines, "\n")
}

// FileFilter determines which files are relevant for analysis
type FileFilter struct {
	relevantFiles map[string]bool
}

// NewFileFilter creates a new FileFilter with default relevant files
func NewFileFilter() *FileFilter {
	ff := &FileFilter{
		relevantFiles: make(map[string]bool),
	}

	// Initialize with default relevant files
	for _, file := range constants.RelevantFiles() {
		ff.relevantFiles[file] = true
	}

	return ff
}

// IsRelevantFile checks if a file is relevant for analysis
// Returns true if the file should be included in the analysis process.
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

// Global instances for backward compatibility
var (
	defaultFileProcessor    = NewFileProcessor()
	defaultSectionExtractor = NewSectionExtractor()
	defaultSafeFileReader   = NewSafeFileReader()
	defaultFileFilter       = NewFileFilter()
)

// Backward compatibility functions - these maintain the original function signatures
// while using the new structured approach internally

// anyFileContains streams files line-by-line, instantly stopping if a match is found (Near-zero RAM usage)
// Legacy function maintained for backward compatibility
func anyFileContains(dirPath string, files []string, search string) bool {
	return defaultFileProcessor.AnyFileContains(dirPath, files, search)
}

// scanFiles streams files line-by-line and executes a callback, discarding the bytes immediately
// Legacy function maintained for backward compatibility
func scanFiles(dirPath string, files []string, lineFunc func(line string)) {
	defaultFileProcessor.ScanFiles(dirPath, files, lineFunc)
}

// extractSection streams a file and extracts a specific command block
// Legacy function maintained for backward compatibility
func extractSection(dirPath string, fileName string, header string) string {
	return defaultSectionExtractor.ExtractSection(dirPath, fileName, header)
}

// readFileSafe reads a small file directly into a string (Used only for tiny configs like sssd.conf)
// Legacy function maintained for backward compatibility
func readFileSafe(dirPath string, fileName string) string {
	return defaultSafeFileReader.ReadFileSafe(dirPath, fileName)
}

// isRelevantFile checks if a file is relevant for analysis
// Legacy function maintained for backward compatibility
func isRelevantFile(name string) bool {
	return defaultFileFilter.IsRelevantFile(name)
}
