package main

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"
)

// extractArchiveToTemp safely extracts relevant files to a temporary directory.
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

	r, err := xz.NewReader(f)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	tr := tar.NewReader(r)
	const maxFileSize = 2 * 1024 * 1024 * 1024 // 2GB max per file safety limit

	if progressFunc != nil {
		progressFunc("Decompressing archive securely to disk...", 5)
	}

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			os.RemoveAll(tempDir)
			return "", fmt.Errorf("tar extraction error: %v", err)
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
				continue
			}

			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				continue
			}

			_, err = io.CopyN(outFile, tr, hdr.Size)
			if err != nil {
				outFile.Close()
				continue
			}
			outFile.Close()
		}
	}
	return tempDir, nil
}
