// analyzer_singlepass.go - Single-pass log scanner for SSSD Inspector
// Performs ONE scan of all log files extracting ALL information simultaneously,
// instead of scanning the same files ~29 times during analysis.

package main

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

// SinglePassResult holds ALL results extracted from log files in a single scan.
// These results are passed to individual analyzer functions to avoid re-scanning.
type SinglePassResult struct {
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
// This eliminates the ~29 separate scans that were previously performed.
func performSinglePassScan(dirPath string, macType string, kbArticles []TIDArticle) *SinglePassResult {
	result := &SinglePassResult{
		KBEvidence: make(map[string][]string),
	}

	// Build error patterns
	errorPatterns := buildErrorPatterns(macType)

	// STEP 1: Build combined pattern lookup
	// Each matched string maps to its category and metadata
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
		// Account expired / revoked credentials
		{"User account has expired", mcAccountExpired, ""},
		{"Clients credentials have been revoked", mcAccountExpired, ""},
		// Watchdog
		{"terminated by own WATCHDOG", mcWatchdog, ""},
		// Crypto bug
		{"service key not available", mcCryptoBug, ""},
		{"TGT failed verification", mcCryptoBug, ""},
		{"KDC has no support for encryption type", mcCryptoBug, ""},
		// Keytab
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

	// No patterns to match? Return empty results
	if len(allPatterns) == 0 {
		return result
	}

	// STEP 2: Build the Aho-Corasick automaton from all literal patterns.
	// All patterns are literal substrings (previously escaped with
	// regexp.QuoteMeta and joined into one giant regex alternation); the
	// trie matches them all in a single O(n) pass per line, avoiding RE2's
	// alternation-size limits and per-branch exploration cost.
	//
	// The automaton is case-insensitive and cached per process like the
	// regex cache, so it is compiled exactly once.
	literalPatterns := make([]string, len(allPatterns))
	patternInfos := make([]patternInfo, len(allPatterns))
	for i, cp := range allPatterns {
		literalPatterns[i] = cp.pattern
		patternInfos[i] = cp.info
	}
	acMatcher := globalACCache.Get(literalPatterns)

	// Time regex for timeline extraction (also cached)
	timeRegex := globalRegexCache.Get(`(?:\((\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})\)|([A-Z][a-z]{2}\s+\d+\s+\d{2}:\d{2}:\d{2})|(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})?))`)

	// Per-description error example storage (max 3 per description)
	errorExamples := make(map[string][]string)

	// Context with timeout for safety
	ctx, cancel := context.WithTimeout(context.Background(), DefaultFileScanTimeout)
	defer cancel()

	// Pre-allocate timeline with reasonable capacity to reduce reallocations
	// Most supportconfig files have < 1000 error lines
	timelineCapacity := 1000
	if len(errorPatterns) > 0 {
		timelineCapacity = len(errorPatterns) * 2 // upper bound estimate
		if timelineCapacity > 5000 {
			timelineCapacity = 5000
		}
	}
	result.Timeline = make([]TimelineEvent, 0, timelineCapacity)

	// STEP 3: SINGLE PASS through all log files
	logFileNames := []string{"sssd.txt", "messages", "messages.txt"}
	scanFilesWithContext(ctx, dirPath, logFileNames, func(line string) {
		lineTrimmed := strings.TrimSpace(line)
		if lineTrimmed == "" {
			return
		}

		// Quick pre-filter: skip lines that don't contain SSSD-related keywords.
		// This avoids running the expensive mega-regex on ~90% of syslog lines
		// that are unrelated (kernel messages, sshd, cron, etc.)
		// The ToLower call here is a bottleneck but avoids regex on 90% of lines.
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

		// Ignore winbindd lines to prevent false positives
		if strings.Contains(lowered, "winbindd") {
			return
		}

		// Single O(n) pass: match ALL literal patterns via the
		// Aho-Corasick automaton (case-insensitive, overlapping matches
		// included). `lowered` is already lowercased above, matching the
		// automaton's folded patterns.
		acMatcher.Match(lowered, func(idx int) {
			info := patternInfos[idx]

			switch info.category {
			case mcError:
				// Collect error examples (max 3 per description)
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
				// Build timeline event
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
					// Only take first 3 lines of evidence per article
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
		})
	})

	// STEP 4: Build SSSDLogErrors from collected examples
	for _, desc := range sortedErrorDescriptions(errorExamples) {
		result.SSSDLogErrors = append(result.SSSDLogErrors, SSSDLogError{
			Description: desc,
			Examples:    errorExamples[desc],
		})
	}

	// Sort timeline chronologically
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

	// Build problem/warning strings from quick pattern results
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
		return tsMatch[1] // SSSD native format
	} else if len(tsMatch) > 2 && tsMatch[2] != "" {
		return tsMatch[2] // Syslog format
	} else if len(tsMatch) > 3 && tsMatch[3] != "" {
		return tsMatch[3] // ISO 8601 format
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
// This replaces the log-scanning portion of analyzeKerberosAndKeytab.
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
// It only scans for config patterns (not log patterns, which are already done).
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

		// Skip SELinux TIDs on AppArmor systems
		if activeSecModule == "AppArmor" && isSELinuxArticle {
			continue
		}

		article.Evidence = evidence[article.TIDID]

		// Check if log pattern matched (evidence exists) or no log patterns required
		logMatched := len(article.LogPatterns) == 0 || len(article.Evidence) > 0

		// Check config patterns (still need to scan config files)
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

	// Try executable-relative path first
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
