// analyzer_core.go
package main

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"sssd-inspector/constants"
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
		report.SssdConfigFound = true
		report.SSSDConfigSnippet = sssdConfContent
		// Phase A: real INI parser + typed AD option validator (replaces the fragile
		// substring-based legacy scanning in analyzeSSSDConfig.
		cfg := parseSssdConfig(sssdConfContent)
		validateDuplicateKeys(cfg, &report)   // duplicate keys apply to the whole config
		validateDomainStructure(cfg, &report) // mandatory id_provider + no inherit_from in domains
		validateADConfig(cfg, &report)        // AD-specific typed-option checks
		validateConfigStructure(cfg, &report) // Phase 2: domains + responders
		validateIDMapRanges(cfg, &report)     // Phase 2: idmap range overlaps
	} else {
		report.Problems = append(report.Problems, "sssd.conf or SSSD configuration block not found in the supportconfig.")
	}

	if report.SssdService != "Running" {
		report.Problems = append(report.Problems, "sssd.service is not actively running.")
	}

	// Phase B: cross-source correlation/reconciliation. Runs after config, DNS,
	// Kerberos,and hostname analysis have populated the report (see runCorrelation).
	runCorrelation(dirPath, &report)

	// MAC denials (from security-*.txt files, not from log scanning)
	analyzeMACDenials(dirPath, &report)

	// ---- Phase 6: KB Article matching (using single-pass evidence) ----
	if progressFunc != nil {
		progressFunc("Matching Knowledge Base Articles...", 80)
	}
	// Match KB using single-pass evidence and config-only patterns
	matchKBArticlesWithEvidence(dirPath, &report, kbArticles, singlePassResult.KBEvidence)

	// ---- Phase 6b: Advanced correlation (temporal bursts + fuzzy KB) ----
	if progressFunc != nil {
		progressFunc("Running advanced correlation analysis...", 85)
	}
	// Sliding-window bursts: repeated occurrences of the same diagnostic
	// event within a bounded window (retry loops, timeouts, flapping).
	report.TemporalClusters = analyzeTemporalClusters(report.Timeline)
	// TF-IDF fuzzy suggestions for log lines the deterministic engines
	// could not classify, against the loaded KB corpus.
	report.KBSuggestions = analyzeKBSuggestions(report.Timeline, kbArticles, report.MatchedTIDs)

	// ---- Phase 6c: data-driven YAML rules (optional, additive) ----
	if customRules := loadAnalysisRules(); len(customRules) > 0 {
		applyAnalysisRules(dirPath, &report, customRules)
	}

	// ---- Phase 6d: Sequence-based root-cause correlation ----
	// Collapses characteristic cause->effect chains in the timeline (DNS/SRV
	// failover, Kerberos realm+clock skew, watchdog/overload, backend offline)
	// into consolidated root-cause ConfigFindings, so the report stops listing
	// N separate symptoms of M underlying failures.
	if progressFunc != nil {
		progressFunc("Correlating root-cause sequences...", 88)
	}
	correlateSequences(report.Timeline, &report)

	// ---- Phase 7: Final checks ----
	// Clean up duplicate entries mapped during distributed analysis
	report.Problems = deduplicateProblems(report.Problems)
	report.Warnings = deduplicateProblems(report.Warnings)

	// Executive summary: aggregate all signals into a health score and a
	// dominant root-cause headline (computed after dedup so the counts are
	// final, and before anonymization so the headline can be scrubbed too).
	computeExecutiveSummary(&report)

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

	// ---- Phase 5: build correlation graph (after summary, before anonymization) ----
	// The graph is built on the raw (non-anonymized) report so that entity
	// extraction can see the real domain/realm/hostname values; anonymizeReport
	// will scrub the graph's Label/Value/LineText fields afterwards.
	report.Graph = buildCorrelationGraph(&report)

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

	// Snapshot the original domain/realm BEFORE any field is overwritten with
	// its placeholder. maskString must be able to match the real values even
	// after SearchDomain/KerberosRealm themselves are mutated below, otherwise
	// any string scrubbed later (SSSDConfigSnippet, Problems, Warnings, ...)
	// would no longer match the original domain.
	origDomain := r.SearchDomain
	origRealm := r.KerberosRealm

	maskString := func(s string) string {
		s = ipRegex.ReplaceAllString(s, "XXX.XXX.XXX.XXX")
		s = ipv6Regex.ReplaceAllString(s, "XXXX:XXXX::XXXX")
		s = macRegex.ReplaceAllString(s, "XX:XX:XX:XX:XX:XX")
		s = emailRegex.ReplaceAllString(s, "[REDACTED_USER]@example.com")

		if origDomain != "" && origDomain != "None" {
			s = strings.ReplaceAll(s, origDomain, "example.com")
			s = strings.ReplaceAll(s, strings.ToUpper(origDomain), "EXAMPLE.COM")
		}
		if origRealm != "" && origRealm != "Not configured" {
			s = strings.ReplaceAll(s, origRealm, "EXAMPLE.COM")
			s = strings.ReplaceAll(s, strings.ToLower(origRealm), "example.com")
		}
		// AdDomain (ad_domain from sssd.conf) and Hostname are not covered by
		// the SearchDomain/KerberosRealm handling above, so replace them here to
		// guarantee the domain and server name never leak into any field.
		if r.AdDomain != "" {
			s = strings.ReplaceAll(s, r.AdDomain, "example.com")
			s = strings.ReplaceAll(s, strings.ToUpper(r.AdDomain), "EXAMPLE.COM")
		}
		if r.Hostname != "" {
			s = strings.ReplaceAll(s, r.Hostname, "redacted-host")
		}
		// uname -a starts with "Linux <nodename> <release> <machine> ... <os>".
		// The nodename is often a SHORT name (e.g. "srv123") that differs from
		// report.Hostname (e.g. the FQDN "srv123.example.com"), so the exact
		// r.Hostname substitution above cannot match it. Redact the nodename
		// token directly (the second whitespace-separated field) to harden the
		// kernel version (and any copied uname line) against hostname leaks.
		if strings.HasPrefix(s, "Linux ") {
			parts := strings.Fields(s)
			if len(parts) > 2 {
				s = "Linux redacted-host " + strings.Join(parts[2:], " ")
			}
		}
		return s
	}

	r.HardwareManufacturer = "[REDACTED]"
	r.HardwareModel = "[REDACTED]"
	r.VirtualIdentity = "[REDACTED]"
	r.SCCStatus = "[REDACTED]"

	// The kernel version is the uname -a line, which embeds the server's
	// hostname (e.g. "Linux prod-server-01 5.14.21-...#1 SMP ... x86_64").
	// maskString now redacts the hostname/domain/IP parts while preserving the
	// actual kernel-release information.
	r.KernelVersion = maskString(r.KernelVersion)

	for i, ns := range r.Nameservers {
		r.Nameservers[i] = maskString(ns)
	}
	// Configuration findings embed the offending config lines (domains, IPs):
	// scrub message and evidence BEFORE r.SearchDomain is overwritten with
	// the placeholder, so the original domain can still be matched.
	for i := range r.ConfigFindings {
		r.ConfigFindings[i].Message = maskString(r.ConfigFindings[i].Message)
		r.ConfigFindings[i].Evidence = maskString(r.ConfigFindings[i].Evidence)
	}

	// The executive-summary headline embeds finding text (domains, realms):
	// scrub it BEFORE r.SearchDomain is overwritten with the placeholder,
	// otherwise the original domain could no longer be matched.
	// (Same ordering constraint as the Problems/Warnings scrubbing above.)
	r.Summary.Headline = maskString(r.Summary.Headline)

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

	// Scrub the correlation graph's entity labels/values and source line text
	// so the frontend never sees raw PII (domains, IPs, hostnames).
	// AdDomain and Hostname are not covered by maskString (which only handles
	// SearchDomain/KerberosRealm), so we explicitly replace them here.
	for i := range r.Graph.Entities {
		r.Graph.Entities[i].Label = maskString(r.Graph.Entities[i].Label)
		r.Graph.Entities[i].Value = maskString(r.Graph.Entities[i].Value)
		if r.AdDomain != "" {
			r.Graph.Entities[i].Label = strings.ReplaceAll(r.Graph.Entities[i].Label, r.AdDomain, "example.com")
			r.Graph.Entities[i].Value = strings.ReplaceAll(r.Graph.Entities[i].Value, r.AdDomain, "example.com")
		}
		if r.Hostname != "" {
			r.Graph.Entities[i].Label = strings.ReplaceAll(r.Graph.Entities[i].Label, r.Hostname, "redacted-host")
			r.Graph.Entities[i].Value = strings.ReplaceAll(r.Graph.Entities[i].Value, r.Hostname, "redacted-host")
		}
	}
	for i := range r.Graph.Findings {
		r.Graph.Findings[i].Message = maskString(r.Graph.Findings[i].Message)
		r.Graph.Findings[i].Evidence = maskString(r.Graph.Findings[i].Evidence)
		if r.AdDomain != "" {
			r.Graph.Findings[i].Evidence = strings.ReplaceAll(r.Graph.Findings[i].Evidence, r.AdDomain, "example.com")
		}
		if r.Hostname != "" {
			r.Graph.Findings[i].Evidence = strings.ReplaceAll(r.Graph.Findings[i].Evidence, r.Hostname, "redacted-host")
		}
	}
	for i := range r.Graph.Sources {
		r.Graph.Sources[i].LineText = maskString(r.Graph.Sources[i].LineText)
		if r.AdDomain != "" {
			r.Graph.Sources[i].LineText = strings.ReplaceAll(r.Graph.Sources[i].LineText, r.AdDomain, "example.com")
		}
		if r.Hostname != "" {
			r.Graph.Sources[i].LineText = strings.ReplaceAll(r.Graph.Sources[i].LineText, r.Hostname, "redacted-host")
		}
	}
}

// CompareConfigs runs the full analysis pipeline on two supportconfig
// directories and returns a ComparisonReport describing the delta between
// them: which findings are common, which are unique to A or B, and the
// health-score difference (A − B). A positive delta means B is healthier.
//
// Both sides are analyzed with anonymization enabled so the caller can
// freely forward the report to a UI or external system.
func CompareConfigs(dirA, dirB string, progressFunc func(string, int)) ComparisonReport {
	progressFunc(constants.MsgInitializing, constants.ProgressStart)

	progressFunc("Analyzing first supportconfig…", 10)
	reportA := analyzeData(dirA, true, func(msg string, pct int) {
		// Map the sub-analysis progress 0..100 onto the global 10..45 band.
		scaled := 10 + int(float64(pct)*0.35)
		progressFunc(msg, scaled)
	})

	progressFunc("Analyzing second supportconfig…", 50)
	reportB := analyzeData(dirB, true, func(msg string, pct int) {
		scaled := 50 + int(float64(pct)*0.35)
		progressFunc(msg, scaled)
	})

	progressFunc("Computing delta…", 90)

	keySet := func(r ReportData) map[string]int {
		m := make(map[string]int)
		for _, f := range r.ConfigFindings {
			k := f.Category + "||" + f.Message
			m[k]++
		}
		for _, c := range r.TemporalClusters {
			k := "cluster||" + c.Description
			m[k] += c.EventCount
		}
		return m
	}

	setA := keySet(reportA)
	setB := keySet(reportB)

	var common, onlyInA, onlyInB []string
	for k := range setA {
		if _, ok := setB[k]; ok {
			common = append(common, k)
		} else {
			onlyInA = append(onlyInA, k)
		}
	}
	for k := range setB {
		if _, ok := setA[k]; !ok {
			onlyInB = append(onlyInB, k)
		}
	}

	progressFunc(constants.MsgAnalysisComplete, constants.ProgressComplete)

	return ComparisonReport{
		A:          reportA,
		B:          reportB,
		Common:     common,
		OnlyInA:    onlyInA,
		OnlyInB:    onlyInB,
		ScoreDelta: reportB.Summary.HealthScore - reportA.Summary.HealthScore,
	}
}

// findingKeys extracts a canonical "category|message" key from every ConfigFinding.
func findingKeys(r *ReportData) []string {
	seen := make(map[string]struct{})
	var keys []string
	for _, f := range r.ConfigFindings {
		key := f.Category + "|" + f.Message
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	return keys
}
