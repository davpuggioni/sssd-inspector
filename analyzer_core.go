// analyzer_core.go
package main

import (
	"regexp"
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
	report.Timestamp = time.Now().Format("02:01:2006 15:04:05")
	report.AppVersion = "0.2.0" // Major bump for Streaming Engine & PII Redaction
	report.SssdService = "Not Running / Unknown"
	report.WinbindService = "Not Running / Unknown"
	report.NscdStatus = "Not Running / Unknown"
	report.TimeService = "Not Running / Unknown"
	report.KerberosRealm = "Not configured"
	report.HardwareManufacturer = "Unknown"
	report.HardwareModel = "Unknown"
	report.Hypervisor = "Unknown"
	report.VirtualIdentity = "Unknown"

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

	if report.SssdService != "Running" {
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

	maskString := func(s string) string {
		s = ipRegex.ReplaceAllString(s, "XXX.XXX.XXX.XXX")
		s = ipv6Regex.ReplaceAllString(s, "XXXX:XXXX::XXXX")
		s = macRegex.ReplaceAllString(s, "XX:XX:XX:XX:XX:XX")
		s = emailRegex.ReplaceAllString(s, "[REDACTED_USER]@example.com")

		if r.SearchDomain != "" && r.SearchDomain != "None" {
			s = strings.ReplaceAll(s, r.SearchDomain, "example.com")
			s = strings.ReplaceAll(s, strings.ToUpper(r.SearchDomain), "EXAMPLE.COM")
		}
		if r.KerberosRealm != "" && r.KerberosRealm != "Not configured" {
			s = strings.ReplaceAll(s, r.KerberosRealm, "EXAMPLE.COM")
			s = strings.ReplaceAll(s, strings.ToLower(r.KerberosRealm), "example.com")
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
