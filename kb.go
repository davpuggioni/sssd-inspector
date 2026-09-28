// kb.go
package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"sssd-inspector/constants"
)

// embeddedKBFS contains all knowledge base articles compiled into the binary.
// This allows the inspector to work anywhere as a single standalone executable
// without having to distribute or copy a separate kb_articles/ directory.
//
//go:embed kb_articles/*.json
var embeddedKBFS embed.FS

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

// loadEmbeddedKBArticles reads all knowledge base articles compiled into the binary.
func loadEmbeddedKBArticles() []TIDArticle {
	entries, err := fs.ReadDir(embeddedKBFS, "kb_articles")
	if err != nil {
		return nil
	}
	var articles []TIDArticle
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := embeddedKBFS.ReadFile("kb_articles/" + entry.Name())
		if err != nil {
			continue
		}
		var article TIDArticle
		if err := json.Unmarshal(data, &article); err != nil {
			continue
		}
		normalizeTIDArticle(&article)
		articles = append(articles, article)
	}
	return articles
}

// loadExternalKBArticles searches for external KB articles next to the executable
// or in the current working directory, allowing users to drop custom TID JSONs
// without recompiling the inspector.
func loadExternalKBArticles() []TIDArticle {
	kbDir := "kb_articles"
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Join(filepath.Dir(exePath), "kb_articles")
		if info, err := os.Stat(exeDir); err == nil && info.IsDir() {
			kbDir = exeDir
		}
	}
	if info, err := os.Stat(kbDir); err != nil || !info.IsDir() {
		return nil
	}
	files, err := filepath.Glob(filepath.Join(kbDir, "*.json"))
	if err != nil || len(files) == 0 {
		return nil
	}
	var articles []TIDArticle
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		var article TIDArticle
		if err := json.Unmarshal(data, &article); err != nil {
			continue
		}
		normalizeTIDArticle(&article)
		articles = append(articles, article)
	}
	return articles
}

// loadKBArticles returns the combined knowledge base corpus: embedded articles
// are loaded first, then any external kb_articles/*.json found next to the
// binary or in the current directory are merged (external articles with the same
// TIDID or KbID take precedence over the embedded copy).
//
// Variadic args are ignored for backward-compatibility with callers passing dirPath.
func loadKBArticles(args ...string) []TIDArticle {
	embedded := loadEmbeddedKBArticles()
	external := loadExternalKBArticles()
	if len(external) == 0 {
		return embedded
	}

	index := make(map[string]int, len(embedded)+len(external))
	result := make([]TIDArticle, 0, len(embedded)+len(external))

	for _, a := range embedded {
		key := a.TIDID
		if key == "" {
			key = a.KbID
		}
		index[key] = len(result)
		result = append(result, a)
	}

	for _, a := range external {
		key := a.TIDID
		if key == "" {
			key = a.KbID
		}
		if idx, ok := index[key]; ok {
			result[idx] = a
		} else {
			index[key] = len(result)
			result = append(result, a)
		}
	}
	return result
}

// matchKBArticles loads KB articles (embedded + external overrides) and correlates
// them with supportconfig data via streaming.
func matchKBArticles(dirPath string, report *ReportData) {
	articles := loadKBArticles()
	if len(articles) == 0 {
		return
	}

	logFiles := constants.SupportconfigLogFiles()
	configFiles := []string{"sssd.conf"}
	activeSecModule := report.MACType

	for _, article := range articles {
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
