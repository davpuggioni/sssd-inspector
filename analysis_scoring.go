// analysis_scoring.go
// Executive-summary engine: aggregates every diagnostic signal produced by
// the analysis pipeline (structured config findings, log errors, problems,
// warnings) into a weighted health score plus a dominant root-cause
// inference. The score is a linear penalty model: each severity class has a
// fixed weight calibrated so that a single critical misconfiguration (e.g. a
// broken krb5_realm) can never be hidden by many benign tuning hints.
package main

import (
	"fmt"
	"strings"
)

// Severity penalties used by computeExecutiveSummary.
const (
	penaltyCritical = 25
	penaltyError    = 12
	penaltyProblem  = 4
	penaltyWarning  = 2
	penaltyLogError = 3
)

// categoryHeadlines maps a dominant finding category to a triage headline
// plus the first remediation command to try.
var categoryHeadlines = map[string]string{
	"krb5_realm":      "Kerberos realm misconfiguration: authentication against AD will fail until the realm is corrected.",
	"join":            "The host appears not to be joined to Active Directory (no default_realm / keytab).",
	"dns":             "DNS resolution does not match the AD domain: service-location (SRV) lookups will fail.",
	"hostname":        "Hostname/FQDN is inconsistent with the AD domain: the join and SPNs are unreliable.",
	"ad_server":       "AD Domain Controller configuration issue: GSSAPI/Kerberos requires hostnames, not IPs.",
	"access":          "Access-control configuration is inconsistent: simple_allow_* rules are being ignored.",
	"ldap_sasl_mech":  "LDAP SASL mechanism is incompatible with the AD provider (GSSAPI required).",
	"kerberos_method": "kerberos_method is invalid or unusual: verify keytab-based machine credentials.",
	"ldap_id_mapping": "ID mapping disabled without RFC2307 attributes: logins will fail silently.",
	"case_sensitive":  "case_sensitive = True on an AD domain causes spurious user-not-found errors.",
	"krb5_validate":   "Kerberos ticket validation is disabled: security risk (KDC spoofing).",
	"duplicate":       "Duplicate parameters in sssd.conf make SSSD behaviour order-dependent.",
	"ad_crypto":       "Encryption-type mismatch between AD and the local crypto policy (RC4 issue).",
}

// computeExecutiveSummary aggregates all diagnostic signals of a completed
// analysis into the report summary. It must be called after all findings
// have been collected and deduplicated (i.e. at the end of analyzeData).
func computeExecutiveSummary(r *ReportData) {
	s := &r.Summary

	for _, f := range r.ConfigFindings {
		switch f.Severity {
		case SevCritical:
			s.CriticalCount++
		case SevError:
			s.ErrorCount++
		case SevWarning:
			s.WarningCount++
		}
	}
	s.WarningCount += len(r.Warnings)
	s.ProblemCount = len(r.Problems)
	s.LogErrorCount = len(r.SSSDLogErrors)

	// Weight log-error *categories* into the dominant root-cause so a burst of,
	// say, DNS/SRV failover failures is surfaced as the dominant category, not
	// drowned out by a long tail of config hints.
	logCat := logCategoryHistogram(r)
	weightedLog := make(map[string]int)
	for cat, n := range logCat {
		if sev, ok := logCategorySeverity[cat]; ok {
			weightedLog[cat] = n * sev
		} else {
			weightedLog[cat] = n
		}
	}

	// Weighted linear penalty model, floored at 0.
	score := 100
	score -= penaltyCritical * s.CriticalCount
	score -= penaltyError * s.ErrorCount
	score -= penaltyProblem * s.ProblemCount
	score -= penaltyWarning * s.WarningCount
	score -= penaltyLogError * s.LogErrorCount
	if score < 0 {
		score = 0
	}
	s.HealthScore = score

	s.TopCategory, s.TopCategoryHits = dominantCategory(r, weightedLog)
	s.Headline = buildHeadline(r, s, weightedLog)
}

// dominantCategory returns the most weighted finding category, breaking ties in
// favour of the higher severity. The weightedLog histogram (from log-error
// categories) is merged into the config-finding weights so the dominant root
// cause reflects both config and symptom signals.
func dominantCategory(r *ReportData, weightedLog map[string]int) (string, int) {
	hits := make(map[string]int)
	weight := make(map[string]int)
	for _, f := range r.ConfigFindings {
		hits[f.Category]++
		switch f.Severity {
		case SevCritical:
			weight[f.Category] += 3
		case SevError:
			weight[f.Category] += 2
		default:
			weight[f.Category]++
		}
	}
	for cat, w := range weightedLog {
		weight[cat] += w
		hits[cat] += w
	}
	best, bestW := "", -1
	for cat, w := range weight {
		if w > bestW {
			best, bestW = cat, w
		}
	}
	return best, hits[best]
}

// buildHeadline produces the one-line triage statement shown at the top of
// the report. It prefers an explicit root cause from the category map, then
// falls back to symptom-driven headlines. The optional rootCauseLabel (e.g. from
// a detected sequence) is appended when present to explain the dominant cause.
func buildHeadline(r *ReportData, s *ExecutiveSummary, weightedLog map[string]int) string {
	// A detected sequence finding overrides the generic headline so the triage
	// immediately explains the cause->effect chain.
	for _, f := range r.ConfigFindings {
		if strings.HasPrefix(f.Message, "[CORRELATED ROOT CAUSE]") {
			return "[ROOT CAUSE] " + strings.TrimPrefix(f.Message, "[CORRELATED ROOT CAUSE] ")
		}
	}
	if s.CriticalCount > 0 {
		if h, ok := categoryHeadlines[s.TopCategory]; ok {
			return "[CRITICAL] " + h
		}
		return "[CRITICAL] One or more critical AD-integration misconfigurations were detected; authentication is expected to fail."
	}
	if s.ErrorCount > 0 {
		if h, ok := categoryHeadlines[s.TopCategory]; ok {
			return "[ERROR] " + h
		}
		return "[ERROR] Configuration errors detected: review the findings below before deeper log analysis."
	}
	if len(r.Problems) > 0 {
		return "[PROBLEMS] The SSSD service or its environment shows faults; see the problem list for details."
	}
	if s.WarningCount > 0 || len(r.Warnings) > 0 {
		return "[HEALTHY-WITH-HINTS] No blocking misconfiguration detected; tuning and diagnostic hints are listed below."
	}
	return "[HEALTHY] No SSSD/AD misconfiguration or related log errors were detected."
}

// healthScoreClass returns a CSS class for the score badge.
func healthScoreClass(score int) string {
	switch {
	case score >= 80:
		return "success"
	case score >= 50:
		return "warn"
	default:
		return "fail"
	}
}

// summaryLines renders the executive summary for the plain-text report.
func summaryLines(r ReportData) []string {
	s := r.Summary
	lines := []string{
		fmt.Sprintf(" Health Score:        %d/100 (%s)", s.HealthScore, healthScoreClass(s.HealthScore)),
		fmt.Sprintf(" Critical Findings:   %d", s.CriticalCount),
		fmt.Sprintf(" Error Findings:      %d", s.ErrorCount),
		fmt.Sprintf(" Warnings/Hints:      %d", s.WarningCount),
		fmt.Sprintf(" Log Error Patterns:  %d", s.LogErrorCount),
	}
	if s.TopCategory != "" {
		lines = append(lines, fmt.Sprintf(" Dominant Category:   %s (%d findings)", s.TopCategory, s.TopCategoryHits))
	}
	if s.Headline != "" {
		lines = append(lines, " Triage:              "+strings.TrimSpace(s.Headline))
	}
	return lines
}

// logCategorySeverity assigns an ordinal "severity credit" to each coarse log
// category so the dominant root-cause selection weights severe classes higher.
var logCategorySeverity = map[string]int{
	"offline": 3, "krb5": 3, "tls": 3, "keytab": 3, "crypto": 3,
	"dns": 2, "srv": 2, "net": 2, "gpo": 2, "idmap": 2, "access": 2,
	"join": 3,
}

// logCategoryHistogram buckets every detected SSSD log error into its coarse
// category (dns/net/time/...). Used by computeExecutiveSummary to bias the
// dominant-category toward the root cause behind the symptom cluster.
func logCategoryHistogram(r *ReportData) map[string]int {
	out := make(map[string]int)
	for _, e := range r.SSSDLogErrors {
		out[categoryFor(e.Description)]++
	}
	// Temporal-cluster descriptions also carry a category signal.
	for _, c := range r.TemporalClusters {
		out[categoryFor(c.Description)]++
	}
	return out
}
