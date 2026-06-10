package analysis

import (
	"context"
	"regexp"
	"sort"
	"strings"

	"sssd-inspector/constants"
	"sssd-inspector/pkg/types"
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
	description string
	kbArticle   *types.TIDArticle
}

// SinglePassResult holds ALL results extracted from log files in a single scan.
type SinglePassResult struct {
	SSSDLogErrors     []types.SSSDLogError
	Timeline          []types.TimelineEvent
	KeytabFound       bool
	KeytabNotFound    bool
	NoKeytabPrincipal bool
	WatchdogFound     bool
	CryptoBugFound    bool
	AccountExpired    bool
	KBEvidence        map[string][]string
	Problems          []string
	Warnings          []string
}

// performSinglePassScan does ONE scan of all log files
func (ctx *AnalyzerContext) performSinglePassScan(dirPath string, macType string, kbArticles []types.TIDArticle) *SinglePassResult {
	return ctx.performSinglePassScanOnFiles(dirPath, []string{"sssd.txt", "messages", "messages.txt"}, macType, kbArticles)
}

// performSinglePassScanOnFiles does ONE scan of the specified log files
func (ctx *AnalyzerContext) performSinglePassScanOnFiles(dirPath string, logFiles []string, macType string, kbArticles []types.TIDArticle) *SinglePassResult {
	result := &SinglePassResult{
		KBEvidence: make(map[string][]string),
	}

	allEntries := ctx.loadErrorPatternEntries(macType)

	type combinedPattern struct {
		pattern string
		info    patternInfo
	}

	var allPatterns []combinedPattern

	categoryMap := map[string]matchCategory{
		"account_expired":     mcAccountExpired,
		"watchdog":            mcWatchdog,
		"crypto_bug":          mcCryptoBug,
		"keytab_principal":    mcKeytabPrincipal,
		"keytab_not_found":    mcKeytabNotFound,
		"no_keytab_principal": mcNoPrincipal,
	}

	for _, entry := range allEntries {
		cat, ok := categoryMap[entry.Category]
		if !ok {
			cat = mcError
		}
		allPatterns = append(allPatterns, combinedPattern{
			pattern: entry.Pattern,
			info: patternInfo{
				category:    cat,
				description: entry.Description,
			},
		})
	}

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

	// Build the combined mega-regex
	var quotedPatterns []string
	patternLookup := make(map[string]patternInfo)

	for _, cp := range allPatterns {
		qp := regexp.QuoteMeta(cp.pattern)
		quotedPatterns = append(quotedPatterns, qp)
		patternLookup[strings.ToLower(cp.pattern)] = cp.info
	}

	combinedRegex := ctx.RegexCache.Get("(?i)(" + strings.Join(quotedPatterns, "|") + ")")

	timeRegex := ctx.RegexCache.Get(`(?:\((\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})\)|([A-Z][a-z]{2}\s+\d+\s+\d{2}:\d{2}:\d{2})|(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})?))`)

	errorExamples := make(map[string][]string)

	ctx_, cancel := context.WithTimeout(context.Background(), constants.DefaultFileScanTimeout)
	defer cancel()

	timelineCapacity := 1000
	if len(allEntries) > 0 {
		timelineCapacity = len(allEntries) * 2
		if timelineCapacity > 5000 {
			timelineCapacity = 5000
		}
	}
	result.Timeline = make([]types.TimelineEvent, 0, timelineCapacity)

	// Scan files with context
	for _, name := range logFiles {
		select {
		case <-ctx_.Done():
			return result
		default:
		}
		// Use the scanner directly
		ctx.ScanFiles(dirPath, []string{name}, func(line string) {
			lineTrimmed := strings.TrimSpace(line)
			if lineTrimmed == "" {
				return
			}

			select {
			case <-ctx_.Done():
				return
			default:
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
					result.Timeline = append(result.Timeline, types.TimelineEvent{
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
	}

	// Build SSSDLogErrors from collected examples
	for _, desc := range sortedErrorDescriptions(errorExamples) {
		result.SSSDLogErrors = append(result.SSSDLogErrors, types.SSSDLogError{
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
