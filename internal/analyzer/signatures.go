// Package analyzer provides comprehensive log pattern dictionaries, timestamp normalizers,
// and single-pass text file query engines.
package analyzer

import (
	"bufio"
	"embed"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// 💡 Tell the Go compiler to bake the file straight into the executable binary
//go:embed sssd_error_patterns.yaml
var embeddedAssets embed.FS

// logFileNames lists the files to scan for SSSD log messages
var logFileNames = []string{"sssd.txt", "messages", "messages.txt"}

// analyzeSSSDConfigAndLogs orchestrates the full analysis of SSSD configuration and logs.
func analyzeSSSDConfigAndLogs(dirPath string, report *ReportData) {
	sssdConfContent := readFileSafe(dirPath, "sssd.conf")
	if sssdConfContent == "" {
		sssdConfContent = extractSection(dirPath, "sssd.txt", "# /etc/sssd/sssd.conf")
	}

	// Quick checks for known issues that map to report.Problems/Warnings directly
	if anyFileContains(dirPath, logFileNames, "User account has expired") || anyFileContains(dirPath, logFileNames, "Clients credentials have been revoked") {
		report.Problems = append(report.Problems, "[AUTHENTICATION] Logs indicate an Active Directory user account is expired, locked, or credentials have been revoked.")
	}
	if anyFileContains(dirPath, logFileNames, "terminated by own WATCHDOG") {
		report.Warnings = append(report.Warnings, "[TUNING] Since a WATCHDOG termination was found, consider setting 'ignore_group_members = true' in sssd.conf to speed up ssh/sudo initial lookups.")
	}
	if anyFileContains(dirPath, logFileNames, "service key not available") || anyFileContains(dirPath, logFileNames, "TGT failed verification") || anyFileContains(dirPath, logFileNames, "KDC has no support for encryption type") {
		report.Warnings = append(report.Warnings, "[AD CRYPTO BUG] Crypto mismatch or 'service key not available' detected. Microsoft AD forces deprecated RC4 encryption if the 'operatingSystemVersion' attribute in AD starts with a number less than 6 (e.g., '5.14.21'). If your Linux crypto-policy disables RC4, authentication will fail. Fix: Prepend the AD attribute with 'Linux ' (e.g., 'Linux 5.14'), OR re-enable RC4 on this host using 'update-crypto-policies --set DEFAULT:AD-SUPPORT' and reboot.")
	}

	errorPatterns := buildErrorPatterns(report.MACType)
	detectedLogErrors, timelineEvents := scanAndCollectErrors(dirPath, errorPatterns)

	report.Timeline = timelineEvents
	report.SSSDLogErrors = buildSortedLogErrors(detectedLogErrors)

	if sssdConfContent != "" {
		analyzeSSSDConfig(sssdConfContent, report)
	} else {
		report.Problems = append(report.Problems, "sssd.conf or SSSD configuration block not found in the supportconfig.")
	}

	if report.SssdService != "Running" {
		report.Problems = append(report.Problems, "sssd.service is not actively running.")
	}
}

// buildErrorPatterns embeds the diagnostic signature maps directly from the external yaml file,
// performing contextual filtering for SELinux and AppArmor environments dynamically.
func buildErrorPatterns(macType string) map[string]string {
	errorPatterns := make(map[string]string)

	// Local Go structure matching your nested array YAML layout schema
	var config struct {
		SSSDErrorPatterns []struct {
			Pattern     string `yaml:"pattern"`
			Description string `yaml:"description"`
		} `yaml:"SSSD_ERROR_PATTERNS"`
	}

	// 💡 Read the configuration straight out of the binary's memory!
	data, err := embeddedAssets.ReadFile("sssd_error_patterns.yaml")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Fatal structural error: Embedded sssd_error_patterns.yaml could not be read from binary memory.")
		return errorPatterns
	}

	// Safely decode the structural list payload
	if unmarshalErr := yaml.Unmarshal(data, &config); unmarshalErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to parse sssd_error_patterns.yaml structural payload: %v\n", unmarshalErr)
		return errorPatterns
	}

	// Transform slice of elements back into a flat key-value map for O(1) processing lookups
	for _, item := range config.SSSDErrorPatterns {
		if item.Pattern != "" && item.Description != "" {
			errorPatterns[item.Pattern] = item.Description
		}
	}

	// Filter out cross-platform structural target noise
	if macType != "SELinux" {
		for key, desc := range errorPatterns {
			if strings.Contains(desc, "SELinux") && !strings.Contains(desc, "AppArmor") {
				delete(errorPatterns, key)
			}
		}
	}

	return errorPatterns
}

// normalizeTimestamp converts a timestamp string from any supported format into a canonical string.
func normalizeTimestamp(ts string) string {
	if t, err := time.Parse("2006-01-02 15:04:05", ts); err == nil {
		return t.Format("2006-01-02 15:04:05")
	}

	isoFormats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.999999999Z",
		"2006-01-02T15:04:05.999Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05.999999999-07:00",
		"2006-01-02T15:04:05.999-07:00",
		"2006-01-02T15:04:05+07:00",
		"2006-01-02T15:04:05.999999999+07:00",
		"2006-01-02T15:04:05.999+07:00",
	}
	for _, isoFmt := range isoFormats {
		if t, err := time.Parse(isoFmt, ts); err == nil {
			return t.Format("2006-01-02 15:04:05")
		}
	}

	syslogFormats := []string{
		"Jan  2 15:04:05",
		"Jan 2 15:04:05",
		"Jan  2 15:04",
		"Jan 2 15:04",
	}
	for _, syslogFmt := range syslogFormats {
		if t, err := time.Parse(syslogFmt, ts); err == nil {
			currentYear := time.Now().Year()
			t = t.AddDate(currentYear-t.Year(), 0, 0)
			return t.Format("2006-01-02 15:04:05")
		}
	}

	syslogFormatsWithYear := []string{
		"Jan  2 15:04:05 2006",
		"Jan 2 15:04:05 2006",
	}
	for _, syslogFmt := range syslogFormatsWithYear {
		if t, err := time.Parse(syslogFmt, ts); err == nil {
			return t.Format("2006-01-02 15:04:05")
		}
	}

	return ""
}

// scanAndCollectErrors scans the log files for known SSSD error patterns.
func scanAndCollectErrors(dirPath string, errorPatterns map[string]string) (map[string][]string, []TimelineEvent) {
	detectedLogErrors := make(map[string][]string)
	var timelineEvents []TimelineEvent

	if len(errorPatterns) == 0 {
		return detectedLogErrors, timelineEvents
	}

	var errorKeys []string
	fastLookup := make(map[string]string)
	for pattern, desc := range errorPatterns {
		errorKeys = append(errorKeys, regexp.QuoteMeta(pattern))
		fastLookup[strings.ToLower(pattern)] = desc
	}
	errorRegex := globalRegexCache.Get("(?i)(" + strings.Join(errorKeys, "|") + ")")
	timeRegex := globalRegexCache.Get(`(?:\((\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})\)|([A-Z][a-z]{2}\s+\d+\s+\d{2}:\d{2}:\d{2})|(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})?))`)

	scanFiles(dirPath, logFileNames, func(line string) {
		lineTrimmed := strings.TrimSpace(line)

		if strings.Contains(strings.ToLower(lineTrimmed), "winbindd") {
			return
		}

		matches := errorRegex.FindAllString(lineTrimmed, -1)
		for _, match := range matches {
			description, ok := fastLookup[strings.ToLower(match)]
			if !ok {
				continue
			}

			examples := detectedLogErrors[description]
			if len(examples) < 3 {
				isDupe := false
				for _, ex := range examples {
					if ex == lineTrimmed {
						isDupe = true
						break
					}
				}
				if !isDupe {
					detectedLogErrors[description] = append(detectedLogErrors[description], lineTrimmed)

					tsMatch := timeRegex.FindStringSubmatch(lineTrimmed)
					ts := "Unknown Time"
					if len(tsMatch) > 1 && tsMatch[1] != "" {
						ts = tsMatch[1]
					} else if len(tsMatch) > 2 && tsMatch[2] != "" {
						ts = tsMatch[2]
					} else if len(tsMatch) > 3 && tsMatch[3] != "" {
						ts = tsMatch[3]
					}

					timelineEvents = append(timelineEvents, TimelineEvent{
						Timestamp: ts,
						Message:   description,
						RawLog:    lineTrimmed,
					})
				}
			}
		}
	})

	sort.SliceStable(timelineEvents, func(i, j int) bool {
		ti := normalizeTimestamp(timelineEvents[i].Timestamp)
		tj := normalizeTimestamp(timelineEvents[j].Timestamp)
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

	return detectedLogErrors, timelineEvents
}

// buildSortedLogErrors converts the detected log errors map into a sorted slice of SSSDLogError.
func buildSortedLogErrors(detectedLogErrors map[string][]string) []SSSDLogError {
	var errors []SSSDLogError
	for desc, lines := range detectedLogErrors {
		errors = append(errors, SSSDLogError{Description: desc, Examples: lines})
	}
	sort.Slice(errors, func(i, j int) bool {
		return errors[i].Description < errors[j].Description
	})
	return errors
}

// analyzeSSSDConfig parses the sssd.conf content and detects common configuration issues.
func analyzeSSSDConfig(sssdConfContent string, report *ReportData) {
	report.SssdConfigFound = true
	report.SSSDConfigSnippet = sssdConfContent

	scanner := bufio.NewScanner(strings.NewReader(sssdConfContent))
	hasSimpleAllowGroups, hasSimpleAllow, accessProviderSimple, idMappingFalse := false, false, false, false

	confLowerStr := strings.ToLower(sssdConfContent)
	if !strings.Contains(confLowerStr, "ldap_use_tokengroups = false") {
		report.Warnings = append(report.Warnings, "[TUNING] If AD users authenticate but fail authorization (missing groups), consider setting 'ldap_use_tokengroups = False'.")
	}
	if !strings.Contains(confLowerStr, "timeout =") {
		report.Warnings = append(report.Warnings, "[TUNING] No LDAP timeout specified. Adding 'timeout = 30' can help stabilize slow Active Directory connections.")
	}

	currentSection := ""
	seenKeys := make(map[string]map[string]bool)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line
			if seenKeys[currentSection] == nil {
				seenKeys[currentSection] = make(map[string]bool)
			}
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.ToLower(strings.TrimSpace(parts[0]))
			if currentSection != "" && key != "debug_level" && key != "" {
				if seenKeys[currentSection][key] {
					report.Problems = append(report.Problems, fmt.Sprintf("CONFIGURATION ERROR: Duplicate parameter '%s' found in section %s of sssd.conf. SSSD may behave unpredictably.", key, currentSection))
				} else {
					seenKeys[currentSection][key] = true
				}
			}
		}

		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "id_provider") && strings.Contains(lowerLine, "ad") {
			report.ADProviderMode = true
		}
		if strings.Contains(lowerLine, "enumerate") && strings.Contains(lowerLine, "true") {
			if !report.EnumerateIssue {
				report.EnumerateIssue = true
				report.Problems = append(report.Problems, "[DEPRECATION] 'enumerate = true' is set in sssd.conf. This causes severe performance issues, is deprecated for AD/IPA, and is unsupported in SSSD 2.10+.")
			}
		}
		if strings.Contains(lowerLine, "use_fully_qualified_names") && strings.Contains(lowerLine, "true") {
			report.UseFQDNSet = true
		}
		if strings.HasPrefix(lowerLine, "simple_allow_groups") {
			hasSimpleAllowGroups, hasSimpleAllow = true, true
		} else if strings.HasPrefix(lowerLine, "simple_allow_users") {
			hasSimpleAllow = true
		}
		if strings.HasPrefix(lowerLine, "access_provider") && strings.Contains(lowerLine, "simple") {
			accessProviderSimple = true
		}
		if strings.HasPrefix(lowerLine, "ldap_id_mapping") && strings.Contains(lowerLine, "false") {
			idMappingFalse = true
		}
		if strings.HasPrefix(lowerLine, "krb5_validate") && strings.Contains(lowerLine, "false") {
			report.Problems = append(report.Problems, "[SECURITY RISK] 'krb5_validate = false' is set. This disables KDC spoofing protection. If used to bypass the AD RC4 bug, remove this and fix the AD operatingSystemVersion attribute or update local crypto policies instead.")
		}
	}

	if err := scanner.Err(); err != nil {
		report.Problems = append(report.Problems, fmt.Sprintf("Error scanning sssd.conf: %v", err))
	}

	if hasSimpleAllow && !accessProviderSimple {
		report.Problems = append(report.Problems, "CONFIGURATION ERROR: 'simple_allow_users' or 'simple_allow_groups' is used in sssd.conf, but 'access_provider = simple' is not set (e.g., using 'ad'). These parameters will be ignored. Use ad_access_filter instead.")
	}
	if hasSimpleAllowGroups {
		report.Warnings = append(report.Warnings, "[DIAGNOSTIC HINT] 'simple_allow_groups' is active. If users authenticate but fail authorization, test by commenting it out and using 'simple_allow_users = <username>' to isolate group resolution issues.")
	}
	if idMappingFalse {
		report.Problems = append(report.Problems, "[WARNING] 'ldap_id_mapping = False' is set. AD logins will fail silently unless UNIX attributes (uidNumber, gidNumber) are manually populated in Active Directory (RFC2307).")
	}
}
