// Package extract provides archive extraction functionality for SSSD Inspector
package extract

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/ulikunitz/xz"
)

// ExtractionBufferSize is the buffer size for archive extraction
const ExtractionBufferSize = 32 * 1024

// largeArchiveThreshold for triggering extra GC after extraction
const largeArchiveThreshold = 50 * 1024 * 1024 // 50MB

// DefaultExtractionTimeout for archive extraction
const DefaultExtractionTimeout = 5 * time.Minute

// XZEstimatedCompressionRatio for size-based progress estimation
const XZEstimatedCompressionRatio = 12

// MaxArchiveFileSize prevents zip/tar bombs
const MaxArchiveFileSize = 100 * 1024 * 1024 // 100MB

// IsRelevantFile determines if a file should be extracted
var IsRelevantFile func(name string) bool

func init() {
	// Default: extract everything (can be overridden by main package)
	IsRelevantFile = func(name string) bool { return true }
}

// isRelevantFile checks if a file name matches relevant files for analysis
func isRelevantFile(name string) bool {
	if IsRelevantFile != nil {
		return IsRelevantFile(name)
	}
	return true
}

// ExtractArchiveToTemp safely extracts relevant files to a temporary directory.
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

	// --- Get file size for progress calculation ---
	fileInfo, err := f.Stat()
	if err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to stat archive: %v", err)
	}
	totalArchiveSize := fileInfo.Size()

	// For very large archives, adjust GC to handle the memory spike better
	isLargeArchive := totalArchiveSize > largeArchiveThreshold

	// --- Create context with timeout ---
	extractionTimeout := DefaultExtractionTimeout

	ctx, cancel := context.WithTimeout(context.Background(), extractionTimeout)
	defer cancel()

	// --- Set up OS signal handling for graceful interruption ---
	signalCtx, signalStop := signal.NotifyContext(ctx, os.Interrupt)
	defer signalStop()

	// Wrap the file with a context-aware reader that can be interrupted
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

	// Single-pass extraction with size-based progress estimation
	estimatedDecompressedSize := totalArchiveSize * XZEstimatedCompressionRatio

	if progressFunc != nil {
		if isLargeArchive {
			progressFunc(fmt.Sprintf("Decompressing large archive (%.0f MB compressed)...", float64(totalArchiveSize)/(1024*1024)), 3)
		} else {
			progressFunc("Decompressing archive...", 3)
		}
	}

	// Pool for extraction write buffers
	bufPool := sync.Pool{
		New: func() interface{} {
			b := make([]byte, ExtractionBufferSize)
			return &b
		},
	}

	extractedCount := 0
	totalExtractedBytes := int64(0)

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
		if isRelevantFile(fileName) {
			// Prevent path traversal
			targetPath := filepath.Join(tempDir, fileName)
			if !strings.HasPrefix(targetPath, filepath.Clean(tempDir)+string(os.PathSeparator)) {
				continue
			}

			// Prevent Zip/Tar Bombs
			if hdr.Size > MaxArchiveFileSize {
				log.Printf("Warning: skipping oversized file %s (%d bytes)", fileName, hdr.Size)
				continue
			}

			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				log.Printf("Warning: failed to create output file %s: %v", targetPath, err)
				continue
			}

			written, err := copyWithBuffer(outFile, io.LimitReader(tr, hdr.Size), &bufPool)
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
				if estimatedDecompressedSize > 0 {
					pct = 5 + int(85*totalExtractedBytes/estimatedDecompressedSize)
					if pct > 95 {
						pct = 95
					}
					progressFunc(fmt.Sprintf("Extracting: %s (%.0f KB decompressed, %d files)", fileName, float64(totalExtractedBytes)/1024, extractedCount), pct)
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

	if isLargeArchive {
		runtime.GC()
	}

	return tempDir, nil
}

// copyWithBuffer copies from src to dst using a reusable buffer from the pool.
func copyWithBuffer(dst io.Writer, src io.Reader, pool *sync.Pool) (int64, error) {
	bufPtr := pool.Get().(*[]byte)
	defer pool.Put(bufPtr)
	buf := *bufPtr

	var written int64
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				return written, ew
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			return written, er
		}
	}
	return written, nil
}
