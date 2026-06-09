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
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/ulikunitz/xz"

	"sssd-inspector/constants"
)

// extractArchiveToTemp safely extracts relevant files to a temporary directory.
// It uses a context with timeout to prevent hanging on corrupted archives,
// provides real-time progress updates per extracted file,
// and handles OS interrupt signals (Ctrl+C) for graceful cleanup.
//
// Memory optimizations:
//   - Single-pass extraction: no pre-scan phase (eliminates double decompression)
//   - Size-based progress: uses compressed file size × estimated ratio instead
//     of a full tar header pre-scan, avoiding a second decompression pass
//   - Large archives (>50MB compressed) get extra GC and buffer adjustments
//   - Buffer pooling reduces allocation pressure during file writes
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

	// For very large archives, adjust GC to handle the memory spike better
	isLargeArchive := totalArchiveSize > constants.LargeArchiveThreshold

	// --- Create context with timeout ---
	extractionTimeout := constants.DefaultExtractionTimeout
	if appConfig != nil && appConfig.Analysis.ExtractionTimeout != "" {
		if d, err := time.ParseDuration(appConfig.Analysis.ExtractionTimeout); err == nil && d > 0 {
			extractionTimeout = d
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), extractionTimeout)
	defer cancel()

	// --- Set up OS signal handling for graceful interruption ---
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

	// --- Single-pass extraction (no pre-scan) ---
	// Instead of doing a pre-scan to count files, we use size-based progress.
	// The estimated decompressed size = compressed_size × compression_ratio.
	// This eliminates the expensive second decompression pass.
	estimatedDecompressedSize := totalArchiveSize * constants.XZEstimatedCompressionRatio

	if progressFunc != nil {
		if isLargeArchive {
			progressFunc(fmt.Sprintf("Decompressing large archive (%.0f MB compressed)...", float64(totalArchiveSize)/(1024*1024)), 3)
		} else {
			progressFunc("Decompressing archive...", 3)
		}
	}

	// Pool for extraction write buffers to reduce allocations
	bufPool := sync.Pool{
		New: func() interface{} {
			b := make([]byte, constants.ExtractionBufferSize)
			return &b
		},
	}

	extractedCount := 0
	totalExtractedBytes := int64(0)

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
			if hdr.Size > constants.MaxArchiveFileSize {
				log.Printf("Warning: skipping oversized file %s (%d bytes)", fileName, hdr.Size)
				continue
			}

			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				log.Printf("Warning: failed to create output file %s: %v", targetPath, err)
				continue
			}

			// Copy with buffered I/O for better performance on large files
			written, err := copyWithBuffer(outFile, io.LimitReader(tr, hdr.Size), &bufPool)
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
				// Size-based progress: use bytes decompressed vs estimated total
				// Progress range: 5% to 95% spanning the extraction phase
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

	// For large archives, help the GC reclaim XZ decoder buffers and temp allocations
	// that may still be referenced after the extraction loop completes.
	if isLargeArchive {
		runtime.GC()
	}

	return tempDir, nil
}

// copyWithBuffer copies from src to dst using a reusable buffer from the pool,
// returning the number of bytes written. This is more efficient than io.CopyN
// for large files because it reuses buffers and reduces GC pressure.
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
