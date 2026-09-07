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

// normalizeTIDArticle bridges the two supported TID JSON schemas so the rest
// of the pipeline sees a uniform struct:
//
//  1. Curated format (kb_articles/TID-*.json): tid_id + conditions with
//     hand-crafted log/config patterns.
//  2. Scraper format (kbscraper TIDs/*.json): kb_id + full article text
//     (plain_text, situation, resolution, cause, ...), no patterns.
//
// For scraped articles it derives TIDID ("TID-<kb_id>") and a Description
// for rendering (situation → cause → truncated plain_text). Matching
// patterns are intentionally left empty for scraped articles.
func normalizeTIDArticle(article *TIDArticle) {
	// Promote nested "conditions" patterns (alternate curated schema) into
	// the top-level fields when they are absent.
	if article.Conditions != nil {
		if len(article.LogPatterns) == 0 && len(article.Conditions.LogPatterns) > 0 {
			article.LogPatterns = article.Conditions.LogPatterns
		}
		if len(article.ConfigPatterns) == 0 && len(article.Conditions.ConfigPatterns) > 0 {
			article.ConfigPatterns = article.Conditions.ConfigPatterns
		}
	}

	if article.TIDID == "" && article.KbID != "" {
		article.TIDID = "TID-" + strings.TrimSpace(article.KbID)
	}

	if article.Description == "" {
		switch {
		case strings.TrimSpace(article.Situation) != "":
			article.Description = firstParagraph(article.Situation)
		case strings.TrimSpace(article.Cause) != "":
			article.Description = firstParagraph(article.Cause)
		case strings.TrimSpace(article.PlainText) != "":
			article.Description = truncateText(firstParagraph(article.PlainText), 300)
		}
	}
}

// firstParagraph returns the first non-empty paragraph (up to a blank line)
// of s, with surrounding whitespace trimmed.
func firstParagraph(s string) string {
	for _, para := range strings.Split(s, "\n\n") {
		if p := strings.TrimSpace(para); p != "" {
			// Collapse newlines inside the paragraph for readability.
			return strings.Join(strings.Fields(p), " ")
		}
	}
	return strings.TrimSpace(s)
}

// truncateText shortens s to at most max runes, appending an ellipsis.
func truncateText(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "…"
}

// matchKBArticles loads JSON files from kb_articles/ and correlates them with supportconfig data via streaming
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
		normalizeTIDArticle(&article)

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
