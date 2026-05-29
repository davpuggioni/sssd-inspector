// Package analyzer handles knowledge base pattern matching, data correlations, and reporting.
package analyzer

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// matchKBArticles loads JSON files from kb_articles/ and correlates them with supportconfig data via streaming.
// Since it resides inside package analyzer, it can directly read globalRegexCache and scanFiles.
func matchKBArticles(dirPath string, report *ReportData) {
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

	activeSecModule := report.MACType

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

		isSELinuxArticle := strings.Contains(strings.ToLower(article.Title), "selinux") ||
			strings.Contains(strings.ToLower(article.Description), "selinux")

		for _, pat := range article.LogPatterns {
			if strings.Contains(strings.ToLower(pat), "selinux") {
				isSELinuxArticle = true
				break
			}
		}

		// Skip SELinux TIDs if we are on an AppArmor system to prevent false positives
		if activeSecModule == "AppArmor" && isSELinuxArticle {
			continue
		}

		article.Evidence = []string{}

		logMatched := len(article.LogPatterns) == 0
		if !logMatched {
			for _, pattern := range article.LogPatterns {
				// Use global regex cache to avoid recompiling the same pattern for every KB article
				matcher := globalRegexCache.Get("(?i)" + regexp.QuoteMeta(pattern))
				scanFiles(dirPath, logFiles, func(line string) {
					if matcher.MatchString(line) {
						logMatched = true
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
				// Use global regex cache for compiled regex reusability
				matcher := globalRegexCache.Get("(?i)" + regexp.QuoteMeta(pattern))
				scanFiles(dirPath, configFiles, func(line string) {
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
