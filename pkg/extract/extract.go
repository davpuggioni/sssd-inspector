// Package extract provides archive extraction functionality for SSSD Inspector
package extract

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"sssd-inspector/constants"
	sssderrors "sssd-inspector/errors"
	"sssd-inspector/logger"

	"github.com/ulikunitz/xz"
)

// XZEstimatedCompressionRatio for size-based progress estimation
const XZEstimatedCompressionRatio = 12

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
		return "", sssderrors.Wrap(err, sssderrors.ErrSystemError, "failed to create temp dir")
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
		return "", sssderrors.NewFileNotFound(archivePath)
	}
	totalArchiveSize := fileInfo.Size()

	// For very large archives, adjust GC to handle the memory spike better
	isLargeArchive := totalArchiveSize > constants.LargeArchiveThreshold

	// --- Create context with timeout ---
	ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultExtractionTimeout)
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
			return "", sssderrors.Wrap(res.err, sssderrors.ErrInvalidArchive, "xz decompression init failed")
		}
		r = res.r
	case <-signalCtx.Done():
		os.RemoveAll(tempDir)
		return "", sssderrors.NewTimeout("archive decompression")
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
			b := make([]byte, constants.ExtractionBufferSize)
			return &b
		},
	}

	extractedCount := 0
	totalExtractedBytes := int64(0)

	for {
		select {
		case <-signalCtx.Done():
			os.RemoveAll(tempDir)
			return "", sssderrors.NewTimeout("archive decompression")
		default:
		}

		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Warn("tar entry error, continuing", logger.Fields{"error": err.Error()})
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
				logger.Warn("skipping oversized file", logger.Fields{"file": fileName, "size": hdr.Size})
				continue
			}

			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				logger.Warn("failed to create output file", logger.Fields{"path": targetPath, "error": err.Error()})
				continue
			}

			written, err := copyWithBuffer(outFile, io.LimitReader(tr, hdr.Size), &bufPool)
			if err != nil {
				outFile.Close()
				logger.Warn("failed to write file", logger.Fields{"file": fileName, "error": err.Error()})
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
