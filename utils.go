package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// anyFileContains streams files line-by-line, instantly stopping if a match is found (Near-zero RAM usage)
func anyFileContains(dirPath string, files []string, search string) bool {
	for _, name := range files {
		f, err := os.Open(filepath.Join(dirPath, name))
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024) // Support long lines up to 1MB
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), search) {
				f.Close()
				return true
			}
		}
		f.Close()
	}
	return false
}

// scanFiles streams files line-by-line and executes a callback, discarding the bytes immediately
func scanFiles(dirPath string, files []string, lineFunc func(line string)) {
	for _, name := range files {
		f, err := os.Open(filepath.Join(dirPath, name))
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)
		for scanner.Scan() {
			lineFunc(scanner.Text())
		}
		f.Close()
	}
}

// extractSection streams a file and extracts a specific command block
func extractSection(dirPath string, fileName string, header string) string {
	f, err := os.Open(filepath.Join(dirPath, fileName))
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

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

// readFileSafe reads a small file directly into a string (Used only for tiny configs like sssd.conf)
func readFileSafe(dirPath string, fileName string) string {
	b, err := os.ReadFile(filepath.Join(dirPath, fileName))
	if err != nil {
		return ""
	}
	return string(b)
}

func isRelevantFile(name string) bool {
	relevant := []string{
		"nsswitch.conf", "hosts", "nscd.conf", "sssd.conf",
		"systemd.txt", "basic-environment.txt", "updates.txt",
		"y2log.txt", "sssd.txt", "rpm.txt", "etc.txt",
		"network.txt", "ntp.txt", "pam.txt", "fs-diskio.txt",
		"storage.txt", "security-apparmor.txt", "security-selinux.txt",
		"memory.txt", "sar.txt", "messages", "messages.txt", "boot.txt",
	}
	for _, r := range relevant {
		if name == r {
			return true
		}
	}
	return false
}
