// analyzer_core.go
package main

import (
	"regexp"
	"sort"
	"sssd-inspector/constants"
	"strconv"
	"strings"
	"time"
)

// deduplicateProblems removes duplicate strings from a slice efficiently
func deduplicateProblems(problems []string) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, p := range problems {
		if _, exists := seen[p]; !exists {
			seen[p] = struct{}{}
			result = append(result, p)
		}
	}
	return result
}

// getSSSDVersion safely extracts the major and minor version numbers from the SSSD rpm package string
func getSSSDVersion(packages []string) (int, int) {
	for _, pkg := range packages {
		parts := strings.Fields(pkg)
		for _, p := range parts {
			// FIXED S1017: Unconditional TrimPrefix
			p = strings.TrimPrefix(p, "sssd-")

			if len(p) > 0 && p[0] >= '0' && p[0] <= '9' && strings.Contains(p, ".") {
				vParts := strings.SplitN(p, ".", 3)
				if len(vParts) >= 2 {
					major, err1 := strconv.Atoi(vParts[0])
					minor, err2 := strconv.Atoi(vParts[1])
					if err1 == nil && err2 == nil {
						return major, minor
					}
				}
			}
		}
	}
	return 0, 0
}

// analyzeData is the main orchestrator. It routes the streaming directory path to specialized analyzers.
// Now uses Single-Pass scanning for log files to eliminate ~29 redundant scans.
func analyzeData(dirPath string, anonymize bool, progressFunc func(string, int)) ReportData {
	// Clear file cache from previous analysis to ensure fresh data
	globalFileCache.Clear()

	var report ReportData
	report.Timestamp = time.Now().Format(constants.TimestampFormat)
	report.AppVersion = constants.AppVersion
	report.SssdService = constants.StatusNotRunning
	report.WinbindService = constants.StatusNotRunning
	report.NscdStatus = constants.StatusNotRunning
	report.TimeService = constants.StatusNotRunning
	report.KerberosRealm = constants.StatusNotConfigured
	report.HardwareManufacturer = constants.StatusUnknown
	report.HardwareModel = constants.StatusUnknown
	report.Hypervisor = constants.StatusUnknown
	report.VirtualIdentity = constants.StatusUnknown

	// ---- Phase 1: System analysis (no log scanning) ----
	if progressFunc != nil {
		progressFunc("Scanning Hardware & OS Data...", 5)
	}
	analyzeBasicHealth(dirPath, &report)
	analyzeOSAndHardware(dirPath, &report)
	analyzeHostnameAndFQDN(dirPath, &report)
	analyzeSCC(dirPath, &report)

	if progressFunc != nil {
		progressFunc("Analyzing Network & Kerberos state...", 15)
	}
	analyzeDNS(dirPath, &report)
	analyzeTime(dirPath, &report)
	analyzePerformance(dirPath, &report)

	// ---- Phase 2: Config analysis (also no log scanning) ----
	if progressFunc != nil {
		progressFunc("Evaluating SSSD Configurations...", 25)
	}
	analyzePAM(dirPath, &report)
	analyzeNSSwitch(dirPath, &report)
	analyzeHosts(dirPath, &report)
	analyzeNSCD(dirPath, &report)
	analyzePackages(dirPath, &report)
	analyzeSSSDVersionAge(&report)
	analyzeServices(dirPath, &report)
	analyzeMACStatus(dirPath, &report)
	analyzeSSSDFilePermissions(dirPath, &report)
	analyzeDiskSpace(dirPath, &report)

	// ---- Phase 3: Single-Pass Log Scanning ----
	// ONE scan of all log files extracts ALL information simultaneously,
	// replacing ~29 separate scans that were previously performed.
	if progressFunc != nil {
		progressFunc("Single-pass scanning and parsing SSSD Logs...", 40)
	}

	// Load KB articles for combined pattern matching
	kbArticles := loadKBArticles(dirPath)
	singlePassResult := performSinglePassScan(dirPath, report.MACType, kbArticles)

	// Apply single-pass results to report
	report.SSSDLogErrors = singlePassResult.SSSDLogErrors
	report.Timeline = singlePassResult.Timeline
	report.Problems = append(report.Problems, singlePassResult.Problems...)
	report.Warnings = append(report.Warnings, singlePassResult.Warnings...)

	// Keytab results from single-pass
	if singlePassResult.KeytabFound {
		report.KeytabFound = true
	}

	// ---- Phase 4: Kerberos config analysis (krb5.conf only, no log scanning) ----
	if progressFunc != nil {
		progressFunc("Analyzing Kerberos configuration...", 55)
	}
	// Analyze krb5.conf without scanning logs (already done in single-pass)
	analyzeKerberosConfig(dirPath, &report)

	if !singlePassResult.KeytabFound && !singlePassResult.KeytabNotFound && !singlePassResult.NoKeytabPrincipal {
		report.Problems = append(report.Problems, "No Kerberos Keytab (Machine Account) principal found. AD join might be broken.")
	}

	// ---- Phase 5: SSSD Config & remaining log-based analysis ----
	if progressFunc != nil {
		progressFunc("Evaluating SSSD Configurations...", 65)
	}
	// Read sssd.conf (uses file cache)
	sssdConfContent := readFileSafe(dirPath, "sssd.conf")
	if sssdConfContent == "" {
		sssdConfContent = extractSection(dirPath, "sssd.txt", "# /etc/sssd/sssd.conf")
	}
	if sssdConfContent != "" {
		analyzeSSSDConfig(sssdConfContent, &report)
	} else {
		report.Problems = append(report.Problems, "sssd.conf or SSSD configuration block not found in the supportconfig.")
	}

	if report.SssdService != constants.StatusRunning {
		report.Problems = append(report.Problems, "sssd.service is not actively running.")
	}

	// MAC denials (from security-*.txt files, not from log scanning)
	analyzeMACDenials(dirPath, &report)

	// ---- Phase 6: KB Article matching (using single-pass evidence) ----
	if progressFunc != nil {
		progressFunc("Matching Knowledge Base Articles...", 80)
	}
	// Match KB using single-pass evidence and config-only patterns
	matchKBArticlesWithEvidence(dirPath, &report, kbArticles, singlePassResult.KBEvidence)

	// ---- Phase 7: Final checks ----
	// Clean up duplicate entries mapped during distributed analysis
	report.Problems = deduplicateProblems(report.Problems)
	report.Warnings = deduplicateProblems(report.Warnings)

	// Final check: Validate if debug_level=9 was set during problems
	if len(report.Problems) > 0 || len(report.SSSDLogErrors) > 0 {
		hasDebug9 := false
		checkDebug := func(line string) {
			if hasDebug9 {
				return
			}
			if strings.Contains(line, "debug_level") && strings.Contains(line, "9") {
				trimmed := strings.ReplaceAll(line, " ", "")
				if strings.Contains(trimmed, "debug_level=9") {
					hasDebug9 = true
				}
			}
		}
		scanFiles(dirPath, []string{"sssd.conf", "sssd.txt"}, checkDebug)

		if report.SssdConfigFound && !hasDebug9 {
			report.Warnings = append(report.Warnings, "[DIAGNOSTIC HINT] SSSD debug level is low. To get better logs, set debug_level=9 in sssd.conf, restart sssd, reproduce the error, and generate a new supportconfig.")
		}
	}

	if anonymize {
		if progressFunc != nil {
			progressFunc("Sanitizing PII data...", 95)
		}
		anonymizeReport(&report)
	}

	return report
}

// anonymizeReport heavily scrubs PII from the report to ensure safe sharing
func anonymizeReport(r *ReportData) {
	ipRegex := regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
	macRegex := regexp.MustCompile(`(?i)\b(?:[0-9a-f]{2}[:-]){5}[0-9a-f]{2}\b`)
	ipv6Regex := regexp.MustCompile(`(?i)\b(?:[a-f0-9]{1,4}:){7}[a-f0-9]{1,4}\b|\b(?:[a-f0-9]{1,4}:){1,7}:|\b:(?::[a-f0-9]{1,4}){1,7}\b`)
	emailRegex := regexp.MustCompile(`(?i)\b[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}\b`)

	// Capture domain and realm before any mutation, since r.SearchDomain and
	// r.KerberosRealm get overwritten to their replacement values later in this
	// function (lines 320-321), and subsequent maskString calls would lose the
	// original values needed for redaction.
	origDomain := r.SearchDomain
	origRealm := r.KerberosRealm

	// Build a list of all known domain values to redact, including:
	//   - SearchDomain (e.g., "corp.company.com")
	//   - KerberosRealm (e.g., "CORP.COMPANY.COM")
	//   - Domains found in sssd.conf snippet via a generic FQDN pattern
	//     that catches all occurrences, including section headers like
	//     [domain/sub.corp.company.com] and all key=value pairs.
	//
	// Instead of trying to parse specific config keys, we use a generic
	// domain/FQDN regex that matches any string like "sub.example.com"
	// or "SUB.EXAMPLE.COM" anywhere in the report text.
	var domainValues []string
	if origDomain != "" && origDomain != "None" {
		domainValues = append(domainValues, origDomain)
	}
	if origRealm != "" && origRealm != "Not configured" {
		domainValues = append(domainValues, origRealm)
	}

	// Generic FQDN regex: matches domain names like "sub.corp.company.com",
	// "COMPANY.COM", or even "sub.corp.company" (NetBIOS or short AD name).
	// It matches any string with at least one dot, avoiding matches inside
	// IP addresses or email addresses (those have their own dedicated regex).
	// Pattern breakdown:
	//   (?i)              - case-insensitive
	//   [a-z0-9]+         - label starting with alphanumeric
	//   (?:[-][a-z0-9]+)* - optional hyphenated part in middle
	//   (?:[.][a-z0-9]+(?:[-][a-z0-9]+)*)+ - one or more dot-separated labels
	//   \b                - word boundary at end
	fqdnRegex := regexp.MustCompile(`(?i)\b[a-z0-9]+(?:[-][a-z0-9]+)*(?:\.[a-z0-9]+(?:[-][a-z0-9]+)*)+\b`)

	// Extract all domain strings from the sssd.conf snippet using the FQDN regex.
	// This catches domains in any context: section headers, key values, comments, etc.
	if r.SSSDConfigSnippet != "" {
		matches := fqdnRegex.FindAllString(r.SSSDConfigSnippet, -1)
		domainValues = append(domainValues, matches...)
	}

	// Deduplicate domain values. If both a lowercase variant (e.g., "company.com")
	// and its uppercase equivalent (e.g., "COMPANY.COM") exist, keep only the
	// uppercase one since its case-insensitive regex already matches both forms
	// and produces the correct uppercase replacement ("EXAMPLE.COM").
	seenDomains := make(map[string]struct{})
	var uniqueDomains []string
	hasUpperVariant := make(map[string]bool)
	for _, d := range domainValues {
		if d == strings.ToUpper(d) {
			hasUpperVariant[strings.ToLower(d)] = true
		}
	}
	for _, d := range domainValues {
		if _, ok := seenDomains[d]; !ok {
			seenDomains[d] = struct{}{}
			if d != strings.ToUpper(d) && hasUpperVariant[strings.ToLower(d)] {
				continue
			}
			uniqueDomains = append(uniqueDomains, d)
		}
	}
	// Sort by length descending so longer/more specific domains are matched first.
	sort.Slice(uniqueDomains, func(i, j int) bool {
		return len(uniqueDomains[i]) > len(uniqueDomains[j])
	})

	maskString := func(s string) string {
		s = ipRegex.ReplaceAllString(s, "XXX.XXX.XXX.XXX")
		s = ipv6Regex.ReplaceAllString(s, "XXXX:XXXX::XXXX")
		s = macRegex.ReplaceAllString(s, "XX:XX:XX:XX:XX:XX")
		s = emailRegex.ReplaceAllString(s, "[REDACTED_USER]@example.com")

		for _, d := range uniqueDomains {
			dRegex := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(d))
			replacement := "example.com"
			if d == strings.ToUpper(d) {
				replacement = "EXAMPLE.COM"
			}
			s = dRegex.ReplaceAllString(s, replacement)
		}
		return s
	}

	r.HardwareManufacturer = "[REDACTED]"
	r.HardwareModel = "[REDACTED]"
	r.VirtualIdentity = "[REDACTED]"
	r.SCCStatus = "[REDACTED]"

	for i, ns := range r.Nameservers {
		r.Nameservers[i] = maskString(ns)
	}
	r.SearchDomain = maskString(r.SearchDomain)
	r.KerberosRealm = maskString(r.KerberosRealm)

	for i, p := range r.Problems {
		r.Problems[i] = maskString(p)
	}
	for i, w := range r.Warnings {
		r.Warnings[i] = maskString(w)
	}
	for i, mac := range r.MACDenialExamples {
		r.MACDenialExamples[i] = maskString(mac)
	}
	for i := range r.SSSDLogErrors {
		r.SSSDLogErrors[i].Description = maskString(r.SSSDLogErrors[i].Description)
		for j, ex := range r.SSSDLogErrors[i].Examples {
			r.SSSDLogErrors[i].Examples[j] = maskString(ex)
		}
	}
	for i := range r.MatchedTIDs {
		for j, ev := range r.MatchedTIDs[i].Evidence {
			r.MatchedTIDs[i].Evidence[j] = maskString(ev)
		}
	}
	r.SSSDConfigSnippet = maskString(r.SSSDConfigSnippet)
}

// analyzeLogsOnly performs a lightweight analysis on raw SSSD log files only.
// This is used by -logdir mode where no supportconfig metadata is available.
// It skips all system/config analysis and only scans the provided log files
// for SSSD error patterns, building a timeline and collecting problems/warnings.
func analyzeLogsOnly(dirPath string, logFiles []string, anonymize bool, progressFunc func(string, int)) ReportData {
	globalFileCache.Clear()

	var report ReportData
	report.Timestamp = time.Now().Format("02:01:2006 15:04:05")
	report.AppVersion = constants.AppVersion

	// Mark all system-level info as N/A since we have no supportconfig
	report.SssdService = constants.StatusNARawLogMode
	report.WinbindService = constants.StatusNARawLogMode
	report.NscdStatus = constants.StatusNARawLogMode
	report.TimeService = constants.StatusNARawLogMode
	report.KerberosRealm = constants.StatusNARawLogMode
	report.HardwareManufacturer = constants.StatusNARawLogMode
	report.HardwareModel = constants.StatusNARawLogMode
	report.Hypervisor = constants.StatusNARawLogMode
	report.VirtualIdentity = constants.StatusNARawLogMode
	report.MACType = constants.StatusUnknown
	report.SssdInstalled = true

	if progressFunc != nil {
		progressFunc("Scanning raw SSSD log files...", 10)
	}

	// Use an empty KB article list since we can't match KB without config files
	var kbArticles []TIDArticle

	// Perform single-pass scan on the provided log files
	singlePassResult := performSinglePassScanOnFiles(dirPath, logFiles, report.MACType, kbArticles)

	// Apply single-pass results to report
	report.SSSDLogErrors = singlePassResult.SSSDLogErrors
	report.Timeline = singlePassResult.Timeline
	report.Problems = append(report.Problems, singlePassResult.Problems...)
	report.Warnings = append(report.Warnings, singlePassResult.Warnings...)

	// Keytab results from single-pass
	if singlePassResult.KeytabFound {
		report.KeytabFound = true
	}

	if progressFunc != nil {
		progressFunc("Compiling log analysis results...", 80)
	}

	// Deduplicate
	report.Problems = deduplicateProblems(report.Problems)
	report.Warnings = deduplicateProblems(report.Warnings)

	if anonymize {
		if progressFunc != nil {
			progressFunc("Sanitizing PII data...", 95)
		}
		anonymizeReport(&report)
	}

	return report
}
