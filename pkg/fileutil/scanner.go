// Package fileutil provides file scanning and processing utilities
package fileutil

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileProcessor provides streaming file processing capabilities
type FileProcessor struct {
	BufferSize    int
	MaxLineLength int
}

// NewFileProcessor creates a new FileProcessor with default settings
func NewFileProcessor() *FileProcessor {
	return &FileProcessor{
		BufferSize:    64 * 1024,   // 64KB default buffer
		MaxLineLength: 1024 * 1024, // 1MB max line length
	}
}

// NewFileProcessorWithConfig creates a new FileProcessor with custom settings
func NewFileProcessorWithConfig(bufferSize, maxLineLength int) *FileProcessor {
	return &FileProcessor{
		BufferSize:    bufferSize,
		MaxLineLength: maxLineLength,
	}
}

// AnyFileContains streams files line-by-line, instantly stopping if a match is found
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

// ScanFiles streams files line-by-line and executes a callback
func (fp *FileProcessor) ScanFiles(dirPath string, files []string, lineFunc func(line string)) error {
	for _, name := range files {
		if err := fp.scanFile(filepath.Join(dirPath, name), lineFunc); err != nil {
			fmt.Printf("Warning: failed to scan file %s: %v\n", name, err)
		}
	}
	return nil
}

// scanFile processes a single file and applies the line function
func (fp *FileProcessor) scanFile(filePath string, lineFunc func(line string)) error {
	f, err := os.Open(filePath)
	if err != nil {
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
	buf := make([]byte, 0, fp.BufferSize)
	scanner.Buffer(buf, fp.MaxLineLength)
	return scanner
}

// SectionExtractor handles extraction of specific sections from files
type SectionExtractor struct {
	Processor *FileProcessor
}

// NewSectionExtractor creates a new SectionExtractor
func NewSectionExtractor() *SectionExtractor {
	return &SectionExtractor{
		Processor: NewFileProcessor(),
	}
}

// ExtractSection streams a file and extracts a specific command block
func (se *SectionExtractor) ExtractSection(dirPath string, fileName string, header string) string {
	filePath := filepath.Join(dirPath, fileName)
	f, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := se.Processor.createScanner(f)
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
			sb.WriteString(line)
			sb.WriteString("\n")
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

// ReadFileSafe reads a small file directly into a string.
// For large files, use the streaming methods instead.
func (sfr *SafeFileReader) ReadFileSafe(lines []string) string {
	return strings.Join(lines, "\n")
}

// Default instances
var (
	DefaultFileProcessor    = NewFileProcessor()
	DefaultSectionExtractor = NewSectionExtractor()
	DefaultSafeFileReader   = NewSafeFileReader()
)
