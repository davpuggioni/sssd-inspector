package analysis

import (
	"bufio"
	_ "embed"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"sssd-inspector/constants"
	"sssd-inspector/pkg/types"
)

//go:embed sssd_error_patterns.yaml
var embeddedPatternsYAML []byte

//go:embed sssd_config_checks.yaml
var embeddedConfigChecksYAML []byte

type errorPatternEntry struct {
	Pattern     string `yaml:"pattern"`
	Description string `yaml:"description"`
	Category    string `yaml:"category,omitempty"`
}

type errorPatternConfig struct {
	SSSDErrorPatterns []errorPatternEntry `yaml:"SSSD_ERROR_PATTERNS"`
}

type configCheck struct {
	Type      string `yaml:"type"`
	Pattern   string `yaml:"pattern"`
	Value     string `yaml:"value,omitempty"`
	Flag      string `yaml:"flag,omitempty"`
	ExtraFlag string `yaml:"extra_flag,omitempty"`
	Dedup     bool   `yaml:"dedup,omitempty"`
	Message   string `yaml:"message,omitempty"`
}

type configChecksYAML struct {
	SSSDConfigChecks []configCheck `yaml:"SSSD_CONFIG_CHECKS"`
}

// Package-level cached YAML data, parsed once at init.
// These are immutable after init and safe for concurrent read access.
var (
	cachedErrorPatterns []errorPatternEntry
	cachedConfigChecks  []configCheck
)

func init() {
	var errCfg errorPatternConfig
	if err := yaml.Unmarshal(embeddedPatternsYAML, &errCfg); err == nil {
		cachedErrorPatterns = errCfg.SSSDErrorPatterns
	}
	var cfg configChecksYAML
	if err := yaml.Unmarshal(embeddedConfigChecksYAML, &cfg); err == nil {
		cachedConfigChecks = cfg.SSSDConfigChecks
	}
}

// loadErrorPatternEntries returns error pattern entries filtered by MAC type.
// The underlying data is parsed once at package init and cached.
func (ctx *AnalyzerContext) loadErrorPatternEntries(macType string) []errorPatternEntry {
	var entries []errorPatternEntry
	for _, entry := range cachedErrorPatterns {
		if macType != "SELinux" && strings.Contains(entry.Description, "SELinux") && !strings.Contains(entry.Description, "AppArmor") {
			continue
		}
		entries = append(entries, entry)
	}
	return entries
}

// normalizeTimestamp converts a timestamp string to canonical YYYY-MM-DD HH:MM:SS
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

// analyzeSSSDConfig parses sssd.conf content and detects common issues
func (ctx *AnalyzerContext) analyzeSSSDConfig(sssdConfContent string, report *types.ReportData) {
	report.SssdConfigFound = true
	report.SSSDConfigSnippet = sssdConfContent

	scanner := bufio.NewScanner(strings.NewReader(sssdConfContent))
	hasSimpleAllowGroups, hasSimpleAllow, accessProviderSimple, idMappingFalse := false, false, false, false

	configChecks := ctx.loadConfigChecks()
	problemReported := make(map[string]bool)
	foundPatterns := make(map[string]bool)

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

		for _, check := range configChecks {
			if strings.Contains(lowerLine, check.Pattern) {
				foundPatterns[check.Pattern] = true
			}
		}

		for _, check := range configChecks {
			keyMatches := strings.Contains(lowerLine, check.Pattern)
			if !keyMatches {
				continue
			}
			valueMatches := check.Value == "" || strings.Contains(lowerLine, check.Value)
			if !valueMatches {
				continue
			}

			switch check.Type {
			case "flag":
				switch check.Flag {
				case "ad_provider":
					report.ADProviderMode = true
				case "use_fqdn":
					report.UseFQDNSet = true
				case "simple_allow_groups":
					hasSimpleAllowGroups = true
					hasSimpleAllow = true
				case "simple_allow":
					hasSimpleAllow = true
				case "access_provider_simple":
					accessProviderSimple = true
				case "id_mapping_false":
					idMappingFalse = true
				}

			case "problem":
				msgKey := check.Message
				if check.Dedup {
					if problemReported[msgKey] {
						continue
					}
					problemReported[msgKey] = true
					if check.Pattern == "enumerate" {
						report.EnumerateIssue = true
					}
				}
				report.Problems = append(report.Problems, check.Message)

			case "warning":
				msgKey := check.Message
				if check.Dedup {
					if problemReported[msgKey] {
						continue
					}
					problemReported[msgKey] = true
				}
				report.Warnings = append(report.Warnings, check.Message)
			}
		}
	}

	for _, check := range configChecks {
		if check.Type == "missing" {
			if !foundPatterns[check.Pattern] {
				msgKey := check.Message
				if check.Dedup {
					if problemReported[msgKey] {
						continue
					}
					problemReported[msgKey] = true
				}
				report.Warnings = append(report.Warnings, check.Message)
			}
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

// loadConfigChecks returns SSSD config check rules from cached parsed YAML.
// The underlying data is parsed once at package init and cached.
func (ctx *AnalyzerContext) loadConfigChecks() []configCheck {
	return cachedConfigChecks
}

// analyzeKerberosConfig analyzes krb5.conf without scanning log files
func (ctx *AnalyzerContext) analyzeKerberosConfig(dirPath string, report *types.ReportData) {
	krb5Content := ctx.ExtractSection(dirPath, "etc.txt", "# /etc/krb5.conf")

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

// LogFileNames returns the list of log files to scan
func LogFileNames() []string {
	return constants.LogFileNames()
}

// scanAndCollectErrors scans log files for SSSD error patterns
func (ctx *AnalyzerContext) scanAndCollectErrors(dirPath string, errorPatterns map[string]string) (map[string][]string, []types.TimelineEvent) {
	detectedLogErrors := make(map[string][]string)
	var timelineEvents []types.TimelineEvent

	var errorKeys []string
	fastLookup := make(map[string]string)
	for pattern, desc := range errorPatterns {
		errorKeys = append(errorKeys, regexp.QuoteMeta(pattern))
		fastLookup[strings.ToLower(pattern)] = desc
	}
	errorRegex := ctx.RegexCache.Get("(?i)(" + strings.Join(errorKeys, "|") + ")")

	timeRegex := ctx.RegexCache.Get(`(?:\((\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})\)|([A-Z][a-z]{2}\s+\d+\s+\d{2}:\d{2}:\d{2})|(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})?))`)

	ctx.ScanFiles(dirPath, constants.LogFileNames(), func(line string) {
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

					timelineEvents = append(timelineEvents, types.TimelineEvent{
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

// buildSortedLogErrors converts the detected log errors map into a sorted slice
func buildSortedLogErrors(detectedLogErrors map[string][]string) []types.SSSDLogError {
	var errors []types.SSSDLogError
	for desc, lines := range detectedLogErrors {
		errors = append(errors, types.SSSDLogError{Description: desc, Examples: lines})
	}
	sort.Slice(errors, func(i, j int) bool {
		return errors[i].Description < errors[j].Description
	})
	return errors
}
