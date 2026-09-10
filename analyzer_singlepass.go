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

// prefilterKeywords are the case-insensitive literals used to quickly
// discard lines that cannot be SSSD-related. They are matched through a
// dedicated Aho-Corasick automaton so the pre-filter itself is O(n) per
// line instead of N sequential strings.Contains scans.
var prefilterKeywords = []string{
	"sssd", "krb5", "ldap", "keytab", "winbind", "ad ", "gpo", "pam",
	"nss", "hbac", "ipa", "kdc", "tgt ", "tls", "gssapi", "library",
	"dlopen", "shared",
}

// regexMetaTokens identify patterns written with regular-expression syntax
// rather than as plain literals. Such patterns cannot be matched by the
// Aho-Corasick automaton (which treats every byte literally) and are routed
// to RE2 instead. Without this split they previously failed SILENTLY: e.g.
// "Attribute .* not allowed for user" never matched any log line.
var regexMetaTokens = []string{`.*`, `\d`, `\s`, `\w`, `+?`, `(?i)`}

// isRegexPattern reports whether a pattern uses regular-expression syntax.
func isRegexPattern(pattern string) bool {
	for _, tok := range regexMetaTokens {
		if strings.Contains(pattern, tok) {
			return true
		}
	}
	return false
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

	// Patterns routed to the Aho-Corasick automaton (plain literals) and
	// patterns routed to RE2 (true regular expressions).
	var literalPatterns []combinedPattern
	var regexPatterns []combinedPattern

	classify := func(pattern string, info patternInfo) {
		if isRegexPattern(pattern) {
			regexPatterns = append(regexPatterns, combinedPattern{pattern: pattern, info: info})
		} else {
			literalPatterns = append(literalPatterns, combinedPattern{pattern: pattern, info: info})
		}
	}

	// Error patterns
	for pattern, desc := range errorPatterns {
		classify(pattern, patternInfo{
			category:    mcError,
			description: desc,
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
		classify(qp.pattern, patternInfo{
			category: qp.category,
		})
	}

	// KB article patterns
	for i := range kbArticles {
		article := kbArticles[i]
		for _, pat := range article.LogPatterns {
			classify(pat, patternInfo{
				category:  mcKB,
				kbArticle: &article,
			})
		}
	}

	// No patterns to match? Return empty results
	if len(literalPatterns) == 0 && len(regexPatterns) == 0 {
		return result
	}

	// STEP 2: Build the Aho-Corasick automaton from all literal patterns.
	// All literal patterns are matched in a single O(n) pass per line,
	// avoiding RE2's alternation-size limits and per-branch exploration cost.
	//
	// The automaton is case-insensitive and cached per process like the
	// regex cache, so it is compiled exactly once.
	//
	// Patterns that use regular-expression syntax (e.g. "Attribute .* not
	// allowed") cannot be matched by the trie and are evaluated with RE2
	// on the pre-filtered lines only, so their cost stays negligible.
	acLiterals := make([]string, len(literalPatterns))
	patternInfos := make([]patternInfo, len(literalPatterns))
	for i, cp := range literalPatterns {
		acLiterals[i] = cp.pattern
		patternInfos[i] = cp.info
	}
	acMatcher := globalACCache.Get(acLiterals)

	// Pre-compile the regex-routed patterns (cached, case-insensitive like AC).
	type compiledRegex struct {
		re   *regexp.Regexp
		info patternInfo
	}
	regexMatchers := make([]compiledRegex, 0, len(regexPatterns))
	for _, cp := range regexPatterns {
		regexMatchers = append(regexMatchers, compiledRegex{
			re:   globalRegexCache.Get("(?i)" + cp.pattern),
			info: cp.info,
		})
	}

	// Pre-filter automaton: matches ALL SSSD-related keywords in one O(n)
	// pass per line, replacing ~18 sequential strings.Contains scans.
	prefilterAC := globalACCache.Get(prefilterKeywords)
	prefilterHits := make([]bool, prefilterAC.Len())

	// Time regex for timeline extraction (also cached)
	timeRegex := globalRegexCache.Get(`(?:\((\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})\)|([A-Z][a-z]{2}\s+\d+\s+\d{2}:\d{2}:\d{2})|(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})?))`)

	// Per-description error example storage (max 3 per description)
	errorExamples := make(map[string][]string)

	// Bounded timeline aggregation (P7): collapse repeats of the same
	// diagnostic event into one row with a count instead of one row per
	// match (retry-loop logs would otherwise grow the timeline unbounded).
	tlAgg := newTimelineAggregator()

	// Context with timeout for safety
	ctx, cancel := context.WithTimeout(context.Background(), DefaultFileScanTimeout)
	defer cancel()

	// (Capacity is managed by the timeline aggregator itself.)

	// STEP 3: SINGLE PASS through all log files
	logFileNames := []string{"sssd.txt", "messages", "messages.txt"}
	scanFilesWithContext(ctx, dirPath, logFileNames, func(line string) {
		lineTrimmed := strings.TrimSpace(line)
		if lineTrimmed == "" {
			return
		}

		// Quick pre-filter: skip lines that don't contain SSSD-related keywords.
		// This avoids running the expensive pattern matching on ~90% of syslog
		// lines that are unrelated (kernel messages, sshd, cron, etc.).
		// The pre-filter itself is an Aho-Corasick automaton, so it costs one
		// O(n) pass per line regardless of the number of keywords.
		lowered := strings.ToLower(lineTrimmed)
		hitCount := 0
		prefilterAC.Match(lowered, func(idx int) {
			if !prefilterHits[idx] {
				prefilterHits[idx] = true
				hitCount++
			}
		})
		matched := hitCount > 0
		if matched {
			// Reset the per-line hit flags for the next line.
			for idx := range prefilterHits {
				prefilterHits[idx] = false
			}
		}
		if !matched {
			return
		}

		// Ignore winbindd lines to prevent false positives
		if strings.Contains(lowered, "winbindd") {
			return
		}

		// handlePattern centralizes the per-match processing so it can be
		// invoked both by the Aho-Corasick automaton (literals) and by the
		// RE2 fallback (regex-routed patterns).
		handlePattern := func(info patternInfo) {
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
				// Build timeline event (aggregated: repeats bump a counter).
				ts := extractTimestamp(lineTrimmed, timeRegex)
				tlAgg.add(ts, info.description, lineTrimmed)

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
		}

		// Single O(n) pass: match ALL literal patterns via the
		// Aho-Corasick automaton (case-insensitive, overlapping matches
		// included). `lowered` is already lowercased above, matching the
		// automaton's folded patterns.
		acMatcher.Match(lowered, func(idx int) {
			handlePattern(patternInfos[idx])
		})

		// Regex-routed patterns: evaluated only on pre-filtered lines.
		for _, rm := range regexMatchers {
			if rm.re.MatchString(lineTrimmed) {
				handlePattern(rm.info)
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

	// Sort timeline chronologically
	result.Timeline = tlAgg.events
	if tlAgg.dropped > 0 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("[TIMELINE] Timeline truncated: %d distinct event(s) beyond the %d-row cap were dropped (counts preserved for retained rows).", tlAgg.dropped, maxTimelineEvents))
	}
	sortTimelineChronological(result.Timeline)

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
		inLibdefaults := false
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
			// Phase 2: [libdefaults] encryption-type analysis. Restricting the
			// ticket enctypes to RC4-only (or enabling weak crypto) breaks
			// authentication against modern AD domains that disable RC4.
			if strings.HasPrefix(line, "[libdefaults]") {
				inLibdefaults = true
				continue
			} else if strings.HasPrefix(line, "[") {
				inLibdefaults = false
			}
			if inLibdefaults && strings.Contains(line, "=") && !strings.HasPrefix(line, "#") {
				parts := strings.SplitN(line, "=", 2)
				key := strings.ToLower(strings.TrimSpace(parts[0]))
				value := strings.ToLower(strings.TrimSpace(parts[1]))
				if key == "allow_weak_crypto" && (value == "true" || value == "yes") {
					report.Warnings = append(report.Warnings, "[SECURITY] 'allow_weak_crypto = true' is set in /etc/krb5.conf. This enables deprecated single-DES/RC4 encryption types; modern AD domains with RC4 disabled will reject tickets.")
				}
				if (key == "default_tgs_enctypes" || key == "default_tkt_enctypes") && value != "" &&
					strings.Contains(value, "rc4") && !strings.Contains(value, "aes") {
					report.Warnings = append(report.Warnings, fmt.Sprintf("[AD CRYPTO] '%s' in /etc/krb5.conf restricts tickets to RC4 only ('%s'). If the AD domain or the local crypto-policy disables RC4, authentication fails. Prefer 'aes256-cts-hmac-sha1-96 aes128-cts-hmac-sha1-96'.", key, value))
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

// matchKBArticlesWithEvidence matches KB articles using pre-collected evidence
// from single-pass (log patterns) plus a SINGLE scan of sssd.conf for ALL
// config patterns.
//
// Previously every article × every config pattern triggered its own regex
// scan of sssd.conf (O(articles × patterns) file passes). Now all config
// patterns are compiled into one Aho-Corasick automaton and the file is
// read once, making the config-matching cost O(config size).
func matchKBArticlesWithEvidence(dirPath string, report *ReportData, kbArticles []TIDArticle, evidence map[string][]string) {
	activeSecModule := report.MACType

	// Select the applicable articles (same gating as the legacy matcher).
	type candidate struct {
		article    *TIDArticle
		hasConfig  bool
		configHits int
	}
	candidates := make([]*candidate, 0, len(kbArticles))
	// Map config pattern -> indices of candidates requiring it.
	patternOwners := make(map[string][]int)

	for i := range kbArticles {
		article := &kbArticles[i]
		title := strings.ToLower(article.Title)
		desc := strings.ToLower(article.Description)
		isSELinuxArticle := strings.Contains(title, "selinux") || strings.Contains(desc, "selinux")
		if !isSELinuxArticle {
			for _, pat := range article.LogPatterns {
				if strings.Contains(strings.ToLower(pat), "selinux") {
					isSELinuxArticle = true
					break
				}
			}
		}
		// Skip SELinux TIDs on AppArmor systems
		if activeSecModule == "AppArmor" && isSELinuxArticle {
			continue
		}

		article.Evidence = evidence[article.TIDID]
		// Articles with no patterns at all are not reportable.
		if len(article.LogPatterns) == 0 && len(article.ConfigPatterns) == 0 {
			continue
		}

		candidates = append(candidates, &candidate{article: article, hasConfig: len(article.ConfigPatterns) > 0})
		idx := len(candidates) - 1
		for _, pat := range article.ConfigPatterns {
			patternOwners[strings.ToLower(pat)] = append(patternOwners[strings.ToLower(pat)], idx)
		}
	}

	// Gather all unique config patterns and match them with ONE automaton
	// over ONE scan of sssd.conf.
	if len(patternOwners) > 0 {
		allPatterns := make([]string, 0, len(patternOwners))
		for pat := range patternOwners {
			allPatterns = append(allPatterns, pat)
		}
		ac := globalACCache.Get(allPatterns)
		hits := make([]bool, len(allPatterns))
		patternIdx := make(map[string]int, len(allPatterns))
		for i, pat := range allPatterns {
			patternIdx[pat] = i
		}

		markHit := func(canonicalIdx int) {
			hits[canonicalIdx] = true
		}

		scanFiles(dirPath, []string{"sssd.conf"}, func(line string) {
			ac.Match(strings.ToLower(line), markHit)
		})

		for pat, idx := range patternIdx {
			if !hits[idx] {
				continue
			}
			for _, ci := range patternOwners[pat] {
				candidates[ci].configHits++
			}
		}
	}

	// Assemble results: an article matches if its log patterns produced
	// evidence (or it has none) AND at least one of its config patterns
	// matched (or it has none) — preserving the legacy OR semantics.
	for _, c := range candidates {
		logMatched := len(c.article.LogPatterns) == 0 || len(c.article.Evidence) > 0
		configMatched := !c.hasConfig || c.configHits > 0
		if logMatched && configMatched {
			report.MatchedTIDs = append(report.MatchedTIDs, *c.article)
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
		// Bridge both supported schemas (curated + scraper) into a
		// uniform TIDArticle.
		normalizeTIDArticle(&article)
		kbArticles = append(kbArticles, article)
	}

	return kbArticles
}
