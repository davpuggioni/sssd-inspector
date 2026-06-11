package analysis

import (
	"regexp"
	"sort"
	"sssd-inspector/constants"
	"sssd-inspector/pkg/types"
	"strconv"
	"strings"
	"time"
)

// Package-level compiled regex patterns for PII redaction.
// Compiling once improves performance and avoids repeated allocations.
var (
	anonymizeIPRegex    = regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
	anonymizeMACRegex   = regexp.MustCompile(`(?i)\b(?:[0-9a-f]{2}[:-]){5}[0-9a-f]{2}\b`)
	anonymizeIPv6Regex  = regexp.MustCompile(`(?i)\b(?:[a-f0-9]{1,4}:){7}[a-f0-9]{1,4}\b|\b(?:[a-f0-9]{1,4}:){1,7}:|\b:(?::[a-f0-9]{1,4}){1,7}\b`)
	anonymizeEmailRegex = regexp.MustCompile(`(?i)\b[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}\b`)
	anonymizeFQDNRegex  = regexp.MustCompile(`(?i)\b[a-z0-9]+(?:[-][a-z0-9]+)*(?:\.[a-z0-9]+(?:[-][a-z0-9]+)*)+\b`)
)

// DeduplicateProblems removes duplicate strings from a slice
func DeduplicateProblems(problems []string) []string {
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

// GetSSSDVersion extracts major and minor version numbers from package string
func GetSSSDVersion(packages []string) (int, int) {
	for _, pkg := range packages {
		parts := strings.Fields(pkg)
		for _, p := range parts {
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

// analyzeData is the main orchestrator for full supportconfig analysis
func (ctx *AnalyzerContext) analyzeData(dirPath string, anonymize bool, progressFunc func(string, int)) types.ReportData {
	ctx.FileCache.Clear()

	var report types.ReportData
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
	ctx.analyzeBasicHealth(dirPath, &report)
	ctx.analyzeOSAndHardware(dirPath, &report)
	ctx.analyzeHostnameAndFQDN(dirPath, &report)
	ctx.analyzeSCC(dirPath, &report)

	if progressFunc != nil {
		progressFunc("Analyzing Network & Kerberos state...", 15)
	}
	ctx.analyzeDNS(dirPath, &report)
	ctx.analyzeTime(dirPath, &report)
	ctx.analyzePerformance(dirPath, &report)

	// ---- Phase 2: Config analysis ----
	if progressFunc != nil {
		progressFunc("Evaluating SSSD Configurations...", 25)
	}
	ctx.analyzePAM(dirPath, &report)
	ctx.analyzeNSSwitch(dirPath, &report)
	ctx.analyzeHosts(dirPath, &report)
	ctx.analyzeNSCD(dirPath, &report)
	ctx.analyzePackages(dirPath, &report)
	ctx.analyzeSSSDVersionAge(&report)
	ctx.analyzeServices(dirPath, &report)
	ctx.analyzeMACStatus(dirPath, &report)
	ctx.analyzeSSSDFilePermissions(dirPath, &report)
	ctx.analyzeDiskSpace(dirPath, &report)

	// ---- Phase 3: Single-Pass Log Scanning ----
	if progressFunc != nil {
		progressFunc("Single-pass scanning and parsing SSSD Logs...", 40)
	}

	kbArticles := ctx.loadKBArticles(dirPath)
	singlePassResult := ctx.performSinglePassScan(dirPath, report.MACType, kbArticles)

	report.SSSDLogErrors = singlePassResult.SSSDLogErrors
	report.Timeline = singlePassResult.Timeline
	report.Problems = append(report.Problems, singlePassResult.Problems...)
	report.Warnings = append(report.Warnings, singlePassResult.Warnings...)

	if singlePassResult.KeytabFound {
		report.KeytabFound = true
	}

	// ---- Phase 4: Kerberos config analysis ----
	if progressFunc != nil {
		progressFunc("Analyzing Kerberos configuration...", 55)
	}
	ctx.analyzeKerberosConfig(dirPath, &report)

	if !singlePassResult.KeytabFound && !singlePassResult.KeytabNotFound && !singlePassResult.NoKeytabPrincipal {
		report.Problems = append(report.Problems, "No Kerberos Keytab (Machine Account) principal found. AD join might be broken.")
	}

	// ---- Phase 5: SSSD Config & remaining log-based analysis ----
	if progressFunc != nil {
		progressFunc("Evaluating SSSD Configurations...", 65)
	}
	sssdConfContent := ctx.ReadFileSafe(dirPath, "sssd.conf")
	if sssdConfContent == "" {
		sssdConfContent = ctx.ExtractSection(dirPath, "sssd.txt", "# /etc/sssd/sssd.conf")
	}
	if sssdConfContent != "" {
		ctx.analyzeSSSDConfig(sssdConfContent, &report)
	} else {
		report.Problems = append(report.Problems, "sssd.conf or SSSD configuration block not found in the supportconfig.")
	}

	if report.SssdService != constants.StatusRunning {
		report.Problems = append(report.Problems, "sssd.service is not actively running.")
	}

	// MAC denials
	ctx.analyzeMACDenials(dirPath, &report)

	// ---- Phase 6: KB Article matching ----
	if progressFunc != nil {
		progressFunc("Matching Knowledge Base Articles...", 80)
	}
	ctx.matchKBArticlesWithEvidence(dirPath, &report, kbArticles, singlePassResult.KBEvidence)

	// ---- Phase 7: Final checks ----
	report.Problems = DeduplicateProblems(report.Problems)
	report.Warnings = DeduplicateProblems(report.Warnings)

	// Check debug_level
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
		ctx.ScanFiles(dirPath, []string{"sssd.conf", "sssd.txt"}, checkDebug)

		if report.SssdConfigFound && !hasDebug9 {
			report.Warnings = append(report.Warnings, "[DIAGNOSTIC HINT] SSSD debug level is low. To get better logs, set debug_level=9 in sssd.conf, restart sssd, reproduce the error, and generate a new supportconfig.")
		}
	}

	if anonymize {
		if progressFunc != nil {
			progressFunc("Sanitizing PII data...", 95)
		}
		AnonymizeReport(&report)
	}

	return report
}

// analyzeLogsOnly performs lightweight analysis on raw SSSD log files
func (ctx *AnalyzerContext) analyzeLogsOnly(dirPath string, logFiles []string, anonymize bool, progressFunc func(string, int)) types.ReportData {
	ctx.FileCache.Clear()

	var report types.ReportData
	report.Timestamp = time.Now().Format("02:01:2006 15:04:05")
	report.AppVersion = constants.AppVersion

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

	var kbArticles []types.TIDArticle

	singlePassResult := ctx.performSinglePassScanOnFiles(dirPath, logFiles, report.MACType, kbArticles)

	report.SSSDLogErrors = singlePassResult.SSSDLogErrors
	report.Timeline = singlePassResult.Timeline
	report.Problems = append(report.Problems, singlePassResult.Problems...)
	report.Warnings = append(report.Warnings, singlePassResult.Warnings...)

	if singlePassResult.KeytabFound {
		report.KeytabFound = true
	}

	if progressFunc != nil {
		progressFunc("Compiling log analysis results...", 80)
	}

	report.Problems = DeduplicateProblems(report.Problems)
	report.Warnings = DeduplicateProblems(report.Warnings)

	if anonymize {
		if progressFunc != nil {
			progressFunc("Sanitizing PII data...", 95)
		}
		AnonymizeReport(&report)
	}

	return report
}

// AnonymizeReport scrubs PII from the report
func AnonymizeReport(r *types.ReportData) {
	origDomain := r.SearchDomain
	origRealm := r.KerberosRealm

	var domainValues []string
	if origDomain != "" && origDomain != "None" {
		domainValues = append(domainValues, origDomain)
	}
	if origRealm != "" && origRealm != "Not configured" {
		domainValues = append(domainValues, origRealm)
	}

	if r.SSSDConfigSnippet != "" {
		matches := anonymizeFQDNRegex.FindAllString(r.SSSDConfigSnippet, -1)
		domainValues = append(domainValues, matches...)
	}

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
	sort.Slice(uniqueDomains, func(i, j int) bool {
		return len(uniqueDomains[i]) > len(uniqueDomains[j])
	})

	maskString := func(s string) string {
		s = anonymizeIPRegex.ReplaceAllString(s, "XXX.XXX.XXX.XXX")
		s = anonymizeIPv6Regex.ReplaceAllString(s, "XXXX:XXXX::XXXX")
		s = anonymizeMACRegex.ReplaceAllString(s, "XX:XX:XX:XX:XX:XX")
		s = anonymizeEmailRegex.ReplaceAllString(s, "[REDACTED_USER]@example.com")

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
	for i := range r.Timeline {
		r.Timeline[i].Message = maskString(r.Timeline[i].Message)
		r.Timeline[i].RawLog = maskString(r.Timeline[i].RawLog)
	}
	r.SSSDConfigSnippet = maskString(r.SSSDConfigSnippet)
}
