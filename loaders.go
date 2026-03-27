// loaders.go
package main

import (
	"archive/tar"
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/ulikunitz/xz"
)

func loadFromArchive(path string, fileMap map[string]string) {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("Error opening archive: %v", err)
		return
	}
	defer f.Close()

	r, err := xz.NewReader(f)
	if err != nil {
		log.Printf("Error creating xz reader: %v", err)
		return
	}

	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Tar extraction error: %v", err)
			return
		}

		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		fileName := filepath.Base(hdr.Name)
		if isRelevantFile(fileName) {
			buf := new(bytes.Buffer)
			if _, err := io.Copy(buf, tr); err != nil {
				log.Printf("Failed to read %s: %v", hdr.Name, err)
			}
			fileMap[fileName] = buf.String()
		}
	}
}

func loadFromDir(root string, fileMap map[string]string) {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && isRelevantFile(filepath.Base(path)) {
			content, err := os.ReadFile(path)
			if err == nil {
				fileMap[filepath.Base(path)] = string(content)
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("Error reading directory: %v", err)
	}
}
