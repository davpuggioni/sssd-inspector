// analyzer_core.go
package main

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
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

// analyzeData is now delegated to utils.go which calls analysis.NewAnalyzerContext().AnalyzeData
// This file is retained only for its deduplicateProblems function (used by tests)
// and getSSSDVersion function.

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
