// kb.go
package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// matchKBArticles loads JSON files from kb_articles/ and correlates them with supportconfig data
func matchKBArticles(fileMap map[string]string, report *ReportData) {
	exePath, err := os.Executable()
	kbDir := "kb_articles"

	if err == nil {
		exeDir := filepath.Join(filepath.Dir(exePath), "kb_articles")
		if _, err := os.Stat(exeDir); !os.IsNotExist(err) {
			kbDir = exeDir
		}
	}

	if _, err := os.Stat(kbDir); os.IsNotExist(err) {
		return
	}

	files, err := filepath.Glob(filepath.Join(kbDir, "*.json"))
	if err != nil || len(files) == 0 {
		return
	}

	logFiles := []string{"sssd.txt", "messages", "messages.txt"}
	configFiles := []string{"sssd.conf"}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			log.Printf("Warning: Failed to read KB article %s: %v", file, err)
			continue
		}

		var article TIDArticle
		if err := json.Unmarshal(content, &article); err != nil {
			log.Printf("Warning: Invalid JSON in KB article %s: %v", file, err)
			continue
		}

		// Initialize evidence slice
		article.Evidence = []string{}

		logMatched := len(article.LogPatterns) == 0
		if !logMatched {
			for _, pattern := range article.LogPatterns {
				matcher := regexp.MustCompile("(?i)" + regexp.QuoteMeta(pattern))
				scanFiles(fileMap, logFiles, func(line string) {
					if matcher.MatchString(line) {
						logMatched = true
						// Capture up to 3 examples
						if len(article.Evidence) < 3 {
							cleanLine := strings.TrimSpace(line)
							isDupe := false
							for _, ev := range article.Evidence {
								if ev == cleanLine {
									isDupe = true
									break
								}
							}
							if !isDupe {
								article.Evidence = append(article.Evidence, cleanLine)
							}
						}
					}
				})
			}
		}

		configMatched := len(article.ConfigPatterns) == 0
		if !configMatched {
			for _, pattern := range article.ConfigPatterns {
				matcher := regexp.MustCompile("(?i)" + regexp.QuoteMeta(pattern))
				scanFiles(fileMap, configFiles, func(line string) {
					if matcher.MatchString(line) {
						configMatched = true
					}
				})
			}
		}

		if logMatched && configMatched {
			if len(article.LogPatterns) > 0 || len(article.ConfigPatterns) > 0 {
				report.MatchedTIDs = append(report.MatchedTIDs, article)
			}
		}
	}
}
