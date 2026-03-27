// utils.go
package main

import (
	"bufio"
	"strings"
)

// anyFileContains checks if any of the specified files contains the substring.
// This is highly memory-efficient as it avoids concatenating files.
func anyFileContains(fileMap map[string]string, files []string, search string) bool {
	for _, name := range files {
		if content, ok := fileMap[name]; ok {
			if strings.Contains(content, search) {
				return true
			}
		}
	}
	return false
}

// scanFiles lazily scans lines across multiple files without concatenating them.
// This allows the garbage collector to immediately clean up processed lines.
func scanFiles(fileMap map[string]string, files []string, lineFunc func(line string)) {
	for _, name := range files {
		if content, ok := fileMap[name]; ok {
			scanner := bufio.NewScanner(strings.NewReader(content))
			for scanner.Scan() {
				lineFunc(scanner.Text())
			}
		}
	}
}

// isRelevantFile checks if the extracted file is one we care about parsing
func isRelevantFile(name string) bool {
	relevant := []string{
		"nsswitch.conf", "hosts", "nscd.conf", "sssd.conf",
		"systemd.txt", "basic-environment.txt", "updates.txt",
		"y2log.txt", "sssd.txt", "rpm.txt", "etc.txt",
		"network.txt", "ntp.txt", "pam.txt", "fs-diskio.txt",
		"storage.txt", "security-apparmor.txt", "security-selinux.txt",
		"memory.txt", "sar.txt", "messages", "messages.txt",
	}
	for _, r := range relevant {
		if name == r {
			return true
		}
	}
	return false
}

// extractSection extracts a specific output block from a supportconfig text file
func extractSection(text, header string) string {
	scanner := bufio.NewScanner(strings.NewReader(text))
	var sb strings.Builder
	inSection := false

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, header) {
			inSection = true
			continue
		}
		if inSection {
			// Stop extraction if we hit the next command block's header
			if strings.HasPrefix(line, "#==[") {
				break
			}
			sb.WriteString(line + "\n")
		}
	}
	return sb.String()
}
