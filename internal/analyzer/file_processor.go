package analyzer

import (
	"archive/tar"
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ulikunitz/xz"

	"sssd-inspector/internal/constants"
	"sssd-inspector/internal/config"
)

// Extraction tuning parameters
const (
	DefaultFileScanTimeout   = 30 * time.Second
	DefaultExtractionTimeout = 30 * time.Minute
	extractionBufferSize     = 32 * 1024 // 32KB buffer chunks
)

// FileProcessor provides streaming file processing capabilities
type FileProcessor struct {
	bufferSize    int
	maxLineLength int
}

func NewFileProcessor() *FileProcessor {
	return &FileProcessor{
		bufferSize:    64 * 1024,
		maxLineLength: 1024 * 1024,
	}
}

func (fp *FileProcessor) AnyFileContains(dirPath string, files []string, search string) bool {
	for _, name := range files {
		if fp.fileContains(filepath.Join(dirPath, name), search) {
			return true
		}
	}
	return false
}

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

func (fp *FileProcessor) ScanFiles(dirPath string, files []string, lineFunc func(line string)) error {
	for _, name := range files {
		if err := fp.scanFile(filepath.Join(dirPath, name), lineFunc); err != nil {
			fmt.Printf("Warning: failed to scan file %s: %v\n", name, err)
		}
	}
	return nil
}

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

func NewSectionExtractor() *SectionExtractor {
	return &SectionExtractor{
		processor: NewFileProcessor(),
	}
}

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

func NewSafeFileReader() *SafeFileReader {
	return &SafeFileReader{}
}

func (sfr *SafeFileReader) ReadFileSafe(dirPath string, fileName string) string {
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

func NewFileFilter() *FileFilter {
	ff := &FileFilter{
		relevantFiles: make(map[string]bool),
	}
	for _, file := range constants.RelevantFiles() {
		ff.relevantFiles[file] = true
	}
	return ff
}

func (ff *FileFilter) IsRelevantFile(name string) bool {
	return ff.relevantFiles[name]
}

// Default internal instances
var (
	defaultFileProcessor    = NewFileProcessor()
	defaultSectionExtractor = NewSectionExtractor()
	defaultSafeFileReader   = NewSafeFileReader()
	defaultFileFilter       = NewFileFilter()
)

func anyFileContains(dirPath string, files []string, search string) bool {
	return defaultFileProcessor.AnyFileContains(dirPath, files, search)
}

func scanFiles(dirPath string, files []string, lineFunc func(line string)) {
	_ = defaultFileProcessor.ScanFiles(dirPath, files, lineFunc)
}

func extractSection(dirPath string, fileName string, header string) string {
	return defaultSectionExtractor.ExtractSection(dirPath, fileName, header)
}

func readFileSafe(dirPath string, fileName string) string {
	return defaultSafeFileReader.ReadFileSafe(dirPath, fileName)
}

func isRelevantFile(name string) bool {
	return defaultFileFilter.IsRelevantFile(name)
}

// Context-aware scanners
func scanFilesWithContext(ctx context.Context, dirPath string, files []string, lineFunc func(line string)) {
	for _, name := range files {
		select {
		case <-ctx.Done():
			return
		default:
		}
		_ = scanFileWithContext(ctx, filepath.Join(dirPath, name), lineFunc)
	}
}

func scanFileWithContext(ctx context.Context, filePath string, lineFunc func(line string)) error {
	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	scanner := createPooledScanner(f)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		lineFunc(scanner.Text())
	}
	return scanner.Err()
}

// ============================================================================
// SECTION 2: Restored High-Performance XZ Tar Extractor Engine
// ============================================================================

func ExtractArchiveToTemp(archivePath string, progressFunc func(string, int)) (string, error) {
	tempDir, err := os.MkdirTemp("", "sssd-inspector-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %v", err)
	}

	f, err := os.Open(archivePath)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}
	defer f.Close()

	fileInfo, err := f.Stat()
	if err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to stat archive: %v", err)
	}
	totalArchiveSize := fileInfo.Size()

	// Wires up to your clean new configuration environment
	extractionTimeout := DefaultExtractionTimeout
	if config.Global != nil && config.Global.Analysis.ExtractionTimeout != "" {
		if d, err := time.ParseDuration(config.Global.Analysis.ExtractionTimeout); err == nil && d > 0 {
			extractionTimeout = d
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), extractionTimeout)
	defer cancel()

	signalCtx, signalStop := signal.NotifyContext(ctx, os.Interrupt)
	defer signalStop()

	type readResult struct {
		r   *xz.Reader
		err error
	}

	xzCh := make(chan readResult, 1)
	go func() {
		r, err := xz.NewReader(f)
		xzCh <- readResult{r, err}
	}()

	var r *xz.Reader
	select {
	case res := <-xzCh:
		if res.err != nil {
			os.RemoveAll(tempDir)
			return "", fmt.Errorf("xz decompression init failed: %v", res.err)
		}
		r = res.r
	case <-signalCtx.Done():
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("decompression cancelled: %v", signalCtx.Err())
	}

	tr := tar.NewReader(r)
	const maxFileSize = 2 * 1024 * 1024 * 1024 // 2GB safety limit

	relevantCount := 0
	totalFilesInArchive := 0
	if progressFunc != nil {
		type scanResult struct {
			relevant int
			total    int
		}
		scanCh := make(chan scanResult, 1)
		go func() {
			rel := 0
			tot := 0
			f2, err2 := os.Open(archivePath)
			if err2 != nil {
				scanCh <- scanResult{0, 0}
				return
			}
			defer f2.Close()
			r2, err2 := xz.NewReader(f2)
			if err2 != nil {
				scanCh <- scanResult{0, 0}
				return
			}
			tr2 := tar.NewReader(r2)
			for {
				hdr, err3 := tr2.Next()
				if err3 == io.EOF {
					break
				}
				if err3 != nil {
					break
				}
				tot++
				// Redirects filter tracking directly to local internal instances
				if hdr.Typeflag == tar.TypeReg && defaultFileFilter.IsRelevantFile(filepath.Base(hdr.Name)) {
					rel++
				}
			}
			scanCh <- scanResult{rel, tot}
		}()

		select {
		case sr := <-scanCh:
			relevantCount = sr.relevant
			totalFilesInArchive = sr.total
		case <-signalCtx.Done():
			os.RemoveAll(tempDir)
			return "", fmt.Errorf("decompression cancelled during pre-scan: %v", signalCtx.Err())
		case <-time.After(10 * time.Second):
			relevantCount = 0
			totalFilesInArchive = 0
		}
	}

	f.Close()
	f, err = os.Open(archivePath)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}
	defer f.Close()

	r2, err := xz.NewReader(f)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("xz decompression init failed: %v", err)
	}
	tr = tar.NewReader(r2)

	extractedCount := 0
	totalExtractedBytes := int64(0)

	if progressFunc != nil {
		if totalFilesInArchive > 0 {
			progressFunc(fmt.Sprintf("Pre-scan complete: %d relevant files out of %d total", relevantCount, totalFilesInArchive), 3)
		} else {
			progressFunc("Decompressing archive...", 3)
		}
	}

	bufPool := sync.Pool{
		New: func() interface{} {
			b := make([]byte, extractionBufferSize)
			return &b
		},
	}

	for {
		select {
		case <-signalCtx.Done():
			os.RemoveAll(tempDir)
			return "", fmt.Errorf("decompression cancelled: %v", signalCtx.Err())
		default:
		}

		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Warning: tar entry error (continuing): %v", err)
			continue
		}

		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		fileName := filepath.Base(hdr.Name)
		if defaultFileFilter.IsRelevantFile(fileName) {
			targetPath := filepath.Join(tempDir, fileName)
			if !strings.HasPrefix(targetPath, filepath.Clean(tempDir)+string(os.PathSeparator)) {
				continue
			}

			if hdr.Size > maxFileSize {
				log.Printf("Warning: skipping oversized file %s (%d bytes)", fileName, hdr.Size)
				continue
			}

			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				log.Printf("Warning: failed to create output file %s: %v", targetPath, err)
				continue
			}

			written, err := copyWithBuffer(outFile, io.LimitReader(tr, hdr.Size), bufPool)
			if err != nil {
				outFile.Close()
				log.Printf("Warning: failed to write file %s: %v", fileName, err)
				continue
			}
			outFile.Close()

			extractedCount++
			totalExtractedBytes += written

			if progressFunc != nil {
				var pct int
				if relevantCount > 0 {
					pct = 5 + (extractedCount * 90 / relevantCount)
					if pct > 95 {
						pct = 95
					}
					progressFunc(fmt.Sprintf("Extracting: %s (%d/%d relevant files)", fileName, extractedCount, relevantCount), pct)
				} else if totalArchiveSize > 0 {
					pct = 5 + int(50*totalExtractedBytes/totalArchiveSize)
					if pct > 55 {
						pct = 55
					}
					progressFunc(fmt.Sprintf("Extracting: %s (%.0f KB decompressed)", fileName, float64(totalExtractedBytes)/1024), pct)
				} else {
					progressFunc(fmt.Sprintf("Extracting: %s (%d files)", fileName, extractedCount), 50)
				}
			}
		}
	}

	return tempDir, nil
}

// copyWithBuffer serves as the missing streaming translation utility 
// to write pooled allocation chunks straight to disk.
func copyWithBuffer(dst io.Writer, src io.Reader, pool sync.Pool) (int64, error) {
	bufPtr := pool.Get().(*[]byte)
	defer pool.Put(bufPtr)
	return io.CopyBuffer(dst, src, *bufPtr)
}
