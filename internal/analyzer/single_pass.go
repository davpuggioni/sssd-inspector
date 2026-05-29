// Package analyzer provides comprehensive single-pass file scraping pipelines
// to match multiple overlapping pattern sequences simultaneously.
package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// matchCategory identifies what type of pattern was matched
type matchCategory int

const (
	mcError matchCategory = iota
	mcKeytabPrincipal
	mcKeytabNotFound
	mcNoPrincipal
	mcWatchdog
	mcCryptoBug
	mcAccountExpired
	mcKB
)

// patternInfo holds metadata about a pattern for single-pass matching
type patternInfo struct {
	category    matchCategory
	description string      // for errors: human-readable description; for KB: pattern text
	kbArticle   *TIDArticle // non-nil only for KB patterns
}

// singlePassResult holds ALL results extracted from log files in a single scan.
type singlePassResult struct {
	SSSDLogErrors []SSSDLogError
	Timeline      []TimelineEvent

	// Quick pattern results
	KeytabFound       bool
	KeytabNotFound    bool
	NoKeytabPrincipal bool
	WatchdogFound     bool
	CryptoBugFound    bool
	AccountExpired    bool

	// KB article evidence (TIDID -> evidence lines)
	KBEvidence map[string][]string

	// Additional problems/warnings from log patterns
	Problems []string
	Warnings []string
}

// performSinglePassScan does ONE scan of all log files (sssd.txt, messages, messages.txt)
// and extracts ALL information: error patterns, keytab info, watchdog, crypto bugs,
// account status, KB article evidence, and timeline events.
func performSinglePassScan(dirPath string, macType string, kbArticles []TIDArticle) *singlePassResult {
	return performSinglePassScanOnFiles(dirPath, []string{"sssd.txt", "messages", "messages.txt"}, macType, kbArticles)
}

// performSinglePassScanOnFiles does ONE scan of the specified log files
// and extracts ALL information: error patterns, keytab info, watchdog, crypto bugs,
// account status, KB article evidence, and timeline events.
func performSinglePassScanOnFiles(dirPath string, logFiles []string, macType string, kbArticles []TIDArticle) *singlePassResult {
	result := &singlePassResult{
		KBEvidence: make(map[string][]string),
	}

	// Build error patterns using the YAML engine in logs.go
	errorPatterns := buildErrorPatterns(macType)

	// STEP 1: Build combined pattern lookup
	type combinedPattern struct {
		pattern string
		info    patternInfo
	}

	var allPatterns []combinedPattern

	// Error patterns
	for pattern, desc := range errorPatterns {
		allPatterns = append(allPatterns, combinedPattern{
			pattern: pattern,
			info: patternInfo{
				category:    mcError,
				description: desc,
			},
		})
	}

	// Quick patterns from analyzeSSSDConfigAndLogs
	quickPatterns := []struct {
		pattern  string
		category matchCategory
		desc     string
	}{
		{"User account has expired", mcAccountExpired, ""},
		{"Clients credentials have been revoked", mcAccountExpired, ""},
		{"terminated by own WATCHDOG", mcWatchdog, ""},
		{"service key not available", mcCryptoBug, ""},
		{"TGT failed verification", mcCryptoBug, ""},
		{"KDC has no support for encryption type", mcCryptoBug, ""},
		{"KVNO Principal", mcKeytabPrincipal, ""},
		{"Default principal:", mcKeytabPrincipal, ""},
		{"Key table file '/etc/krb5.keytab' not found", mcKeytabNotFound, ""},
		{"No suitable principal found in keytab", mcNoPrincipal, ""},
	}

	for _, qp := range quickPatterns {
		allPatterns = append(allPatterns, combinedPattern{
			pattern: qp.pattern,
			info: patternInfo{
				category: qp.category,
			},
		})
	}

	// KB article patterns
	for i := range kbArticles {
		article := kbArticles[i]
		for _, pat := range article.LogPatterns {
			allPatterns = append(allPatterns, combinedPattern{
				pattern: pat,
				info: patternInfo{
					category:  mcKB,
					kbArticle: &article,
				},
			})
		}
	}

	if len(allPatterns) == 0 {
		return result
	}

	// STEP 2: Build the mega-regex from all patterns
	var quotedPatterns []string
	patternLookup := make(map[string]patternInfo)

	for _, cp := range allPatterns {
		qp := regexp.QuoteMeta(cp.pattern)
		quotedPatterns = append(quotedPatterns, qp)
		patternLookup[strings.ToLower(cp.pattern)] = cp.info
	}

	combinedRegex := globalRegexCache.Get("(?i)(" + strings.Join(quotedPatterns, "|") + ")")
	timeRegex := globalRegexCache.Get(`(?:\((\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})\)|([A-Z][a-z]{2}\s+\d+\s+\d{2}:\d{2}:\d{2})|(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})?))`)

	errorExamples := make(map[string][]string)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultFileScanTimeout)
	defer cancel()

	timelineCapacity := 1000
	if len(errorPatterns) > 0 {
		timelineCapacity = len(errorPatterns) * 2
		if timelineCapacity > 5000 {
			timelineCapacity = 5000
		}
	}
	result.Timeline = make([]TimelineEvent, 0, timelineCapacity)

	// STEP 3: SINGLE PASS through all specified log files
	scanFilesWithContext(ctx, dirPath, logFiles, func(line string) {
		lineTrimmed := strings.TrimSpace(line)
		if lineTrimmed == "" {
			return
		}

		lowered := strings.ToLower(lineTrimmed)
		if !strings.Contains(lowered, "sssd") &&
			!strings.Contains(lowered, "krb5") &&
			!strings.Contains(lowered, "ldap") &&
			!strings.Contains(lowered, "keytab") &&
			!strings.Contains(lowered, "winbind") &&
			!strings.Contains(lowered, "ad ") &&
			!strings.Contains(lowered, "gpo") &&
			!strings.Contains(lowered, "pam") &&
			!strings.Contains(lowered, "nss") &&
			!strings.Contains(lowered, "hbac") &&
			!strings.Contains(lowered, "ipa") &&
			!strings.Contains(lowered, "kdc") &&
			!strings.Contains(lowered, "tgt ") &&
			!strings.Contains(lowered, "tls") &&
			!strings.Contains(lowered, "gssapi") &&
			!strings.Contains(lowered, "library") &&
			!strings.Contains(lowered, "dlopen") &&
			!strings.Contains(lowered, "shared") {
			return
		}

		if strings.Contains(lowered, "winbindd") {
			return
		}

		matches := combinedRegex.FindAllString(lineTrimmed, -1)
		for _, match := range matches {
			info, ok := patternLookup[strings.ToLower(match)]
			if !ok {
				continue
			}

			switch info.category {
			case mcError:
				examples := errorExamples[info.description]
				if len(examples) < 3 {
					isDupe := false
					for _, ex := range examples {
						if ex == lineTrimmed {
							isDupe = true
							break
						}
					}
					if !isDupe {
						errorExamples[info.description] = append(examples, lineTrimmed)
					}
				}
				ts := extractTimestamp(lineTrimmed, timeRegex)
				result.Timeline = append(result.Timeline, TimelineEvent{
					Timestamp: ts,
					Message:   info.description,
					RawLog:    lineTrimmed,
				})

			case mcKeytabPrincipal:
				result.KeytabFound = true

			case mcKeytabNotFound:
				result.KeytabNotFound = true

			case mcNoPrincipal:
				result.NoKeytabPrincipal = true

			case mcWatchdog:
				result.WatchdogFound = true

			case mcCryptoBug:
				result.CryptoBugFound = true

			case mcAccountExpired:
				result.AccountExpired = true

			case mcKB:
				if info.kbArticle != nil {
					evidence := result.KBEvidence[info.kbArticle.TIDID]
					if len(evidence) < 3 {
						cleanLine := strings.TrimSpace(lineTrimmed)
						isDupe := false
						for _, ev := range evidence {
							if ev == cleanLine {
								isDupe = true
								break
							}
						}
						if !isDupe {
							result.KBEvidence[info.kbArticle.TIDID] = append(evidence, cleanLine)
						}
					}
				}
			}
		}
	})

	// STEP 4: Build SSSDLogErrors from collected examples
	for _, desc := range sortedErrorDescriptions(errorExamples) {
		result.SSSDLogErrors = append(result.SSSDLogErrors, SSSDLogError{
			Description: desc,
			Examples:    errorExamples[desc],
		})
	}

	sort.SliceStable(result.Timeline, func(i, j int) bool {
		ti := normalizeTimestamp(result.Timeline[i].Timestamp)
		tj := normalizeTimestamp(result.Timeline[j].Timestamp)
		if ti == "" && tj == "" {
			return false
		}
		if ti == "" {
			return false
		}
		if tj == "" {
			return true
		}
		return ti < tj
	})

	if result.AccountExpired {
		result.Problems = append(result.Problems, "[AUTHENTICATION] Logs indicate an Active Directory user account is expired, locked, or credentials have been revoked.")
	}
	if result.WatchdogFound {
		result.Warnings = append(result.Warnings, "[TUNING] Since a WATCHDOG termination was found, consider setting 'ignore_group_members = true' in sssd.conf to speed up ssh/sudo initial lookups.")
	}
	if result.CryptoBugFound {
		result.Warnings = append(result.Warnings, "[AD CRYPTO BUG] Crypto mismatch or 'service key not available' detected. Microsoft AD forces deprecated RC4 encryption if the 'operatingSystemVersion' attribute in AD starts with a number less than 6 (e.g., '5.14.21'). If your Linux crypto-policy disables RC4, authentication will fail. Fix: Prepend the AD attribute with 'Linux ' (e.g., 'Linux 5.14'), OR re-enable RC4 on this host using 'update-crypto-policies --set DEFAULT:AD-SUPPORT' and reboot.")
	}
	if result.KeytabNotFound {
		result.Problems = append(result.Problems, "[KERBEROS] /etc/krb5.keytab file is missing, preventing SSSD from authenticating.")
	} else if result.NoKeytabPrincipal {
		result.Problems = append(result.Problems, "[KERBEROS] No suitable principal found in keytab. The machine password may have been changed externally.")
		result.Warnings = append(result.Warnings, "[DIAGNOSTIC HINT] Run 'kvno HOSTNAME$' and compare the output to 'klist -k'. If kvno is higher, the keytab is outdated and needs to be refreshed.")
	}

	return result
}

// extractTimestamp extracts a timestamp from a log line using the provided time regex
func extractTimestamp(line string, timeRegex *regexp.Regexp) string {
	tsMatch := timeRegex.FindStringSubmatch(line)
	if len(tsMatch) > 1 && tsMatch[1] != "" {
		return tsMatch[1]
	} else if len(tsMatch) > 2 && tsMatch[2] != "" {
		return tsMatch[2]
	} else if len(tsMatch) > 3 && tsMatch[3] != "" {
		return tsMatch[3]
	}
	return "Unknown Time"
}

// sortedErrorDescriptions returns the error descriptions in alphabetical order
func sortedErrorDescriptions(errorMap map[string][]string) []string {
	descs := make([]string, 0, len(errorMap))
	for desc := range errorMap {
		descs = append(descs, desc)
	}
	sort.Strings(descs)
	return descs
}

// analyzeKerberosConfig analyzes krb5.conf without scanning log files.
func analyzeKerberosConfig(dirPath string, report *ReportData) {
	krb5Content := extractSection(dirPath, "etc.txt", "# /etc/krb5.conf")

	if krb5Content != "" {
		inDomainRealm := false
		for _, line := range strings.Split(krb5Content, "\n") {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "rc4-hmac") {
				report.Warnings = append(report.Warnings, "[SECURITY] Legacy 'rc4-hmac' encryption found in /etc/krb5.conf. Modern Active Directory domains will reject this, causing silent authentication failures.")
			}
			if strings.HasPrefix(line, "default_realm") {
				parts := strings.Split(line, "=")
				if len(parts) >= 2 {
					report.KerberosRealm = strings.TrimSpace(parts[1])
				}
			}
			if strings.HasPrefix(line, "[domain_realm]") {
				inDomainRealm = true
				continue
			} else if strings.HasPrefix(line, "[") {
				inDomainRealm = false
			}

			if inDomainRealm && strings.Contains(line, "=") && !strings.HasPrefix(line, "#") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) >= 2 {
					domainPart := strings.TrimSpace(parts[0])
					if !strings.HasPrefix(domainPart, ".") && !strings.HasPrefix(domainPart, "*") {
						report.Warnings = append(report.Warnings, fmt.Sprintf("[KERBEROS] The [domain_realm] mapping '%s' lacks a leading dot. Consider changing it to '.%s' to properly map subdomains.", domainPart, domainPart))
					}
				}
			}
		}
	}
}

// matchKBArticlesWithEvidence matches KB articles using pre-collected evidence from single-pass.
func matchKBArticlesWithEvidence(dirPath string, report *ReportData, kbArticles []TIDArticle, evidence map[string][]string) {
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

// loadKBArticles loads all KB article JSON files for single-pass scanning
func loadKBArticles(dirPath string) []TIDArticle {
	var kbArticles []TIDArticle
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
		var article TIDArticle
		if err := json.Unmarshal(content, &article); err != nil {
			continue
		}
		kbArticles = append(kbArticles, article)
	}

	return kbArticles
}
