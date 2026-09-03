// analyzer_correlate.go
// Phase B: A cross-source correlation ("reconciliation") engine. It compares
// independently gathered facts — sssd.conf AD settings, DNS/search domain,
// krb5.conf default_realm, hostname — and flags contradictions, which are the
// most common silent AD integration breakers.
package main

import "strings"

// canonicalADDomain derives the most reliable lowercased AD DNS domain name.
// Precedence: explicit ad_domain in sssd.conf, else the resolv.conf search
// domain, else the hostname FQDN suffix.
func canonicalADDomain(report *ReportData) string {
	if report.AdDomain != "" {
		return strings.ToLower(strings.TrimPrefix(report.AdDomain, "."))
	}
	if report.SearchDomain != "" {
		d := firstSearchToken(report.SearchDomain)
		if d != "" {
			return strings.ToLower(strings.TrimPrefix(d, "."))
		}
	}
	if report.Hostname != "" && strings.Contains(report.Hostname, ".") {
		if suffix := hostnameSuffix(report.Hostname); suffix != "" {
			return strings.ToLower(suffix)
		}
	}
	return ""
}

// firstSearchToken returns the first whitespace-separated token of a
// resolv.conf search/domain directive.
func firstSearchToken(search string) string {
	tokens := strings.Fields(search)
	if len(tokens) == 0 {
		return ""
	}
	return tokens[0]
}

// hostnameSuffix returns the portion of a hostname after the first dot, i.e.
// the DNS domain it belongs to. Returns "" for a short hostname.
func hostnameSuffix(hostname string) string {
	idx := strings.Index(hostname, ".")
	if idx < 0 || idx == len(hostname)-1 {
		return ""
	}
	return hostname[idx+1:]
}

// domainsOverlap reports whether a is a suffix of b, b a suffix of a, or they
// are equal (case-insensitive). This permits both exact matches and relationships
// where one is a subdomain of the other.
func domainsOverlap(a, b string) bool {
	la, lb := strings.ToLower(a), strings.ToLower(b)
	return la == lb || strings.HasSuffix(la, "."+lb) || strings.HasSuffix(lb, "."+la)
}

// runCorrelation performs the cross-source reconciliation and appends findings.
// It must be called after config, DNS, Kerberos, and hostname analysis have
// populated the report (i.e. after Phase 5 in analyzeData).
func runCorrelation(dirPath string, report *ReportData) {
	if !report.ADProviderMode {
		return
	}

	domain := canonicalADDomain(report)
	if domain == "" {
		return // not enough information to correlate safely (avoid false positives)
	}
	realm := strings.ToUpper(domain)

	// 1. AD provider configured but no krb5 default_realm -> likely not joined.
	krbRealm := strings.TrimSpace(report.KerberosRealm)
	if krbRealm == "" || strings.EqualFold(krbRealm, "Not configured") {
		msg := "[CRITICAL] An AD provider is configured but /etc/krb5.conf has no 'default_realm' for the domain '" + realm + "'. The host may not be joined to Active Directory (realm join / adcli join missing)."
		addConfigFinding(report, SevCritical, "join", msg, "krb5.conf", "[libdefaults].default_realm", 0, "")
		return
	}

	// 2. krb5 default_realm vs canonical AD realm mismatch.
	if !strings.EqualFold(realm, krbRealm) && !domainsOverlap(domain, krbRealm) {
		msg := "[CRITICAL] Kerberos 'default_realm = " + krbRealm + "' does not match the Active Directory domain '" + realm + "'. Kerberos auth against AD will fail. Fix default_realm (recommended: " + realm + ")."
		addConfigFinding(report, SevCritical, "krb5_realm", msg, "krb5.conf", "[libdefaults].default_realm", 0, "default_realm = "+krbRealm)
	}

	// 3. DNS search domain must be related to the AD domain.
	if report.SearchDomain != "" {
		searchToken := firstSearchToken(report.SearchDomain)
		if searchToken != "" && !domainsOverlap(searchToken, domain) {
			msg := "DNS search domain '" + searchToken + "' (resolv.conf) does not match the AD domain '" + domain + "'. DNS service-location (SRV) lookups for AD will fail or resolve to the wrong domain."
			addConfigFinding(report, SevError, "dns", msg, "resolv.conf", "search", 0, "search "+report.SearchDomain)
		}
	}

	// 4. Hostname FQDN suffix must be related to the AD domain.
	if report.Hostname != "" {
		short := !strings.Contains(report.Hostname, ".")
		if !short {
			if suffix := hostnameSuffix(report.Hostname); suffix != "" && !domainsOverlap(suffix, domain) {
				msg := "Hostname '" + report.Hostname + "' is not in the AD domain '" + domain + "'. AD joins/accounts are tied to the DNS domain; ensure the hostname is on a subdomain of the AD domain."
				addConfigFinding(report, SevError, "hostname", msg, "basic-environment.txt", "Hostname", 0, "Hostname: "+report.Hostname)
			}
		}
		// Short hostname is already flagged by analyzeHostnameAndFQDN.
	}
}