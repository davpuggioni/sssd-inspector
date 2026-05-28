package main

import (
	"archive/tar"
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
)

// DefaultExtractionTimeout is the maximum time allowed for extracting an archive.
const DefaultExtractionTimeout = 30 * time.Minute

// extractionBufferSize is the chunk size used when copying extracted file data.
const extractionBufferSize = 64 * 1024 // 64KB

// extractArchiveToTemp safely extracts relevant files to a temporary directory.
// It uses a context with timeout to prevent hanging on corrupted archives,
// provides real-time progress updates per extracted file,
// and handles OS interrupt signals (Ctrl+C) for graceful cleanup.
func extractArchiveToTemp(archivePath string, progressFunc func(string, int)) (string, error) {
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

	// --- Create context with timeout ---
	// Use the configured extraction timeout from config.yaml; fallback to 30 min default.
	extractionTimeout := DefaultExtractionTimeout
	if appConfig != nil && appConfig.Analysis.ExtractionTimeout != "" {
		if d, err := time.ParseDuration(appConfig.Analysis.ExtractionTimeout); err == nil && d > 0 {
			extractionTimeout = d
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), extractionTimeout)
	defer cancel()

	// --- Set up OS signal handling for graceful interruption ---
	// Signal goroutine: listen for Ctrl+C and cancel the context
	signalCtx, signalStop := signal.NotifyContext(ctx, os.Interrupt)
	defer signalStop()

	// Wrap the file with a context-aware reader that can be interrupted
	// even during XZ decompression initialization.
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
	const maxFileSize = 2 * 1024 * 1024 * 1024 // 2GB max per file safety limit

	// --- First pass: count relevant files for progress calculation ---
	// We need to do this by estimating from archive size vs file count
	relevantCount := 0
	totalFilesInArchive := 0
	if progressFunc != nil {
		// Estimate: scan tar headers to count total files (non-recursive, just headers)
		type scanResult struct {
			relevant int
			total    int
		}
		scanCh := make(chan scanResult, 1)
		go func() {
			rel := 0
			tot := 0
			// Reset the tar reader by creating a new one from a fresh XZ reader
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
				if hdr.Typeflag == tar.TypeReg && isRelevantFile(filepath.Base(hdr.Name)) {
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
			// Pre-scan timed out; fall back to size-based estimation
			relevantCount = 0
			totalFilesInArchive = 0
		}
	}

	// --- Second pass: actual extraction ---
	// Re-open the archive file for the actual extraction pass
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

	// Pool for extraction write buffers to reduce allocations
	bufPool := sync.Pool{
		New: func() interface{} {
			b := make([]byte, extractionBufferSize)
			return &b
		},
	}

	for {
		// Check for cancellation between each tar entry
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
			// Log the error but try to continue with remaining files instead of aborting
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
			if hdr.Size > maxFileSize {
				log.Printf("Warning: skipping oversized file %s (%d bytes)", fileName, hdr.Size)
				continue
			}

			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				log.Printf("Warning: failed to create output file %s: %v", targetPath, err)
				continue
			}

			// Copy with buffered I/O for better performance on large files
			written, err := copyWithBuffer(outFile, io.LimitReader(tr, hdr.Size), bufPool)
			if err != nil {
				outFile.Close()
				log.Printf("Warning: failed to write file %s: %v", fileName, err)
				continue
			}
			outFile.Close()

			extractedCount++
			totalExtractedBytes += written

			// Update progress after each relevant file
			if progressFunc != nil {
				// Progress range: 5% to 95% spanning the extraction phase
				var pct int
				if relevantCount > 0 {
					// File-count based progress (when pre-scan succeeded)
					pct = 5 + (extractedCount * 90 / relevantCount)
					if pct > 95 {
						pct = 95
					}
					progressFunc(fmt.Sprintf("Extracting: %s (%d/%d relevant files)", fileName, extractedCount, relevantCount), pct)
				} else if totalArchiveSize > 0 {
					// Size-based fallback progress
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

// copyWithBuffer copies from src to dst using a reusable buffer from the pool,
// returning the number of bytes written. This is more efficient than io.CopyN
// for large files because it reuses buffers and reduces GC pressure.
func copyWithBuffer(dst io.Writer, src io.Reader, pool sync.Pool) (int64, error) {
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
