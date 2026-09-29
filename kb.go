// kb.go
package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
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

// jsonErrorLine converts a json.Unmarshal byte-offset error into a 1-based
// source line, so a malformed KB article points the user at the exact line.
func jsonErrorLine(data []byte, err error) int {
	var synErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	offset := int64(-1)
	if errors.As(err, &synErr) {
		offset = synErr.Offset
	} else if errors.As(err, &typeErr) {
		offset = typeErr.Offset
	}
	if offset <= 0 || offset > int64(len(data)) {
		return 0
	}
	return strings.Count(string(data[:offset]), "\n") + 1
}

// loadEmbeddedKBArticles reads all knowledge base articles compiled into the
// binary. A malformed embedded file is a build defect but is still reported
// as a diagnostic instead of vanishing silently.
func loadEmbeddedKBArticles() ([]TIDArticle, []Diagnostic) {
	entries, err := fs.ReadDir(embeddedKBFS, "kb_articles")
	if err != nil {
		return nil, nil
	}
	var articles []TIDArticle
	var diags []Diagnostic
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		name := "kb_articles/" + entry.Name()
		data, err := embeddedKBFS.ReadFile(name)
		if err != nil {
			diags = append(diags, Diagnostic{
				File:     name,
				Message:  fmt.Sprintf("cannot read embedded KB article: %v", err),
				Severity: SevWarning,
			})
			continue
		}
		var article TIDArticle
		if err := json.Unmarshal(data, &article); err != nil {
			diags = append(diags, Diagnostic{
				File:     name,
				Line:     jsonErrorLine(data, err),
				Message:  fmt.Sprintf("invalid embedded KB article JSON: %v", err),
				Severity: SevWarning,
			})
			continue
		}
		normalizeTIDArticle(&article)
		articles = append(articles, article)
	}
	return articles, diags
}

// externalKBDirs returns every directory scanned for external TID JSONs, in
// ascending precedence (later entries override earlier ones by article key):
// the working directory (historical), next to the executable (historical,
// which used to win over the working directory), then the system-wide and
// per-user definition roots that match the config.yaml search convention.
// Duplicate locations are collapsed by absolute path so the same directory
// is never scanned twice when the executable IS the working directory.
func externalKBDirs() []string {
	var raw []string
	raw = append(raw, "kb_articles")
	if exePath, err := os.Executable(); err == nil {
		raw = append(raw, filepath.Join(filepath.Dir(exePath), "kb_articles"))
	}
	for _, root := range []string{systemDefinitionsRoot(), userDefinitionsRoot()} {
		if root == "" {
			continue
		}
		raw = append(raw, filepath.Join(root, "kb_articles"))
	}

	seen := make(map[string]bool, len(raw))
	dirs := make([]string, 0, len(raw))
	for _, d := range raw {
		key := d
		if abs, err := filepath.Abs(d); err == nil {
			key = abs
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		dirs = append(dirs, d)
	}
	return dirs
}

// loadExternalKBArticles searches every externalKBDirs() location for TID
// JSONs, allowing users to drop custom articles without recompiling the
// inspector. Articles are merged across directories (highest-precedence
// directory wins per article key); malformed files are skipped and reported
// as diagnostics instead of disappearing silently.
func loadExternalKBArticles() ([]TIDArticle, []Diagnostic) {
	var diags []Diagnostic
	articles := []TIDArticle{}
	index := map[string]int{}

	for _, kbDir := range externalKBDirs() {
		dirArticles, dirDiags := loadKBArticlesFromDir(kbDir)
		diags = append(diags, dirDiags...)
		for _, article := range dirArticles {
			key := article.TIDID
			if key == "" {
				key = article.KbID
			}
			if key == "" {
				articles = append(articles, article)
				continue
			}
			if pos, ok := index[key]; ok {
				articles[pos] = article // higher-precedence directory overrides
			} else {
				index[key] = len(articles)
				articles = append(articles, article)
			}
		}
	}
	return articles, diags
}

// loadKBArticlesFromDir parses every *.json of ONE directory of curated KB
// articles. Malformed or unreadable files are skipped and reported as
// diagnostics instead of disappearing silently; a missing directory is not an
// error (articles are optional). Splitting this out of
// loadExternalKBArticles lets the Definitions Studio report per-location
// counts through exactly the same parser.
func loadKBArticlesFromDir(kbDir string) ([]TIDArticle, []Diagnostic) {
	info, err := os.Stat(kbDir)
	if err != nil || !info.IsDir() {
		return nil, nil
	}
	files, err := filepath.Glob(filepath.Join(kbDir, "*.json"))
	if err != nil || len(files) == 0 {
		return nil, nil
	}

	var diags []Diagnostic
	articles := make([]TIDArticle, 0, len(files))
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			diags = append(diags, Diagnostic{
				File:     absPath(file),
				Message:  fmt.Sprintf("cannot read KB article: %v", err),
				Severity: SevWarning,
			})
			continue
		}
		var article TIDArticle
		if err := json.Unmarshal(data, &article); err != nil {
			diags = append(diags, Diagnostic{
				File:     absPath(file),
				Line:     jsonErrorLine(data, err),
				Message:  fmt.Sprintf("invalid KB article JSON: %v", err),
				Severity: SevWarning,
			})
			continue
		}
		normalizeTIDArticle(&article)
		articles = append(articles, article)
	}
	return articles, diags
}

// loadKBArticlesDiag returns the combined knowledge base corpus plus any
// loading diagnostics. Embedded articles load first, then external articles
// from every externalKBDirs() location are merged (external articles with
// the same TIDID or KbID take precedence over the embedded copy).
//
// Variadic args are ignored for backward-compatibility with callers passing dirPath.
func loadKBArticlesDiag(args ...string) ([]TIDArticle, []Diagnostic) {
	embedded, embDiags := loadEmbeddedKBArticles()
	external, extDiags := loadExternalKBArticles()
	diags := append(embDiags, extDiags...)
	if len(external) == 0 {
		return embedded, diags
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
	return result, diags
}

// loadKBArticles is the diagnostics-free convenience wrapper over
// loadKBArticlesDiag for callers that only need the corpus.
func loadKBArticles(args ...string) []TIDArticle {
	articles, _ := loadKBArticlesDiag(args...)
	return articles
}

// matchKBArticles loads KB articles (embedded + external overrides) and correlates
// them with supportconfig data via streaming.
func matchKBArticles(dirPath string, report *ReportData) {
	articles, diags := loadKBArticlesDiag()
	report.AddDiagnostics(diags)
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
