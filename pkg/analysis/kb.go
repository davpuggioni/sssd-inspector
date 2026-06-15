package analysis

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"sssd-inspector/pkg/fileutil"
	"sssd-inspector/pkg/types"
)

// loadKBArticles loads all KB article JSON files for single-pass scanning
func (ctx *AnalyzerContext) loadKBArticles(_ string) []types.TIDArticle {
	var kbArticles []types.TIDArticle
	kbDir := "kb_articles"

	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Join(filepath.Dir(exePath), "kb_articles")
		if _, err := os.Stat(exeDir); !os.IsNotExist(err) {
			kbDir = exeDir
		}
	}

	if _, err := os.Stat(kbDir); os.IsNotExist(err) {
		return nil
	}

	files, err := filepath.Glob(filepath.Join(kbDir, "*.json"))
	if err != nil || len(files) == 0 {
		return nil
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		var article types.TIDArticle
		if err := json.Unmarshal(content, &article); err != nil {
			continue
		}
		kbArticles = append(kbArticles, article)
	}

	return kbArticles
}

// matchKBArticlesWithEvidence matches KB articles using pre-collected evidence from single-pass.
func (ctx *AnalyzerContext) matchKBArticlesWithEvidence(dirPath string, report *types.ReportData, kbArticles []types.TIDArticle, evidence map[string][]string) {
	configFiles := []string{"sssd.conf"}
	activeSecModule := report.MACType

	for _, article := range kbArticles {
		isSELinuxArticle := strings.Contains(strings.ToLower(article.Title), "selinux") ||
			strings.Contains(strings.ToLower(article.Description), "selinux")

		for _, pat := range article.LogPatterns {
			if strings.Contains(strings.ToLower(pat), "selinux") {
				isSELinuxArticle = true
				break
			}
		}

		if activeSecModule == "AppArmor" && isSELinuxArticle {
			continue
		}

		article.Evidence = evidence[article.TIDID]

		logMatched := len(article.LogPatterns) == 0 || len(article.Evidence) > 0

		configMatched := len(article.ConfigPatterns) == 0
		if !configMatched {
			for _, pattern := range article.ConfigPatterns {
				matcher := fileutil.GlobalRegexCache.Get("(?i)" + pattern)
				if matcher == nil {
					continue
				}
				ctx.ScanFiles(dirPath, configFiles, func(line string) {
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
