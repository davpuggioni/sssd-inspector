// config_parser.go
// Phase A: A real INI parser for sssd.conf plus an Active Directory oriented
// typed-option validator. Replaces the fragile substring-based detection that
// used strings.Contains on the whole line (e.g. it could flag any line
// containing "ad", such as "username = admin").
package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// SssdKeyVal is a single key/value occurrence with the physical source line.
type SssdKeyVal struct {
	Value string
	Line  int // 1-based line in the parsed content
}

// SssdSection models one [section] of sssd.conf. Keys are stored lowercased;
// multiple occurrences are preserved so duplicate-key detection works.
type SssdSection struct {
	Name    string
	Kind    string // "sssd", "domain", "pam", "nss", "sudo", "ssh", "autofs", ...
	Options map[string][]SssdKeyVal
}

// ParsedConfig is the in-memory representation of an sssd.conf file.
type ParsedConfig struct {
	Raw        []string // original lines (trimmed) for evidence rendering
	Sections   map[string]*SssdSection
	Order      []string       // section header order (stable iteration)
	HasAD      bool           // true if at least one domain uses id_provider = ad
	AdSections []*SssdSection // the AD provider sections
}

// parseSssdConfig parses the INI-format content of sssd.conf.
func parseSssdConfig(content string) *ParsedConfig {
	cfg := &ParsedConfig{
		Sections: make(map[string]*SssdSection),
	}
	if content == "" {
		return cfg
	}

	var cur *SssdSection
	for i, rawLine := range strings.Split(content, "\n") {
		lineNum := i + 1
		line := strings.TrimSpace(rawLine)
		cfg.Raw = append(cfg.Raw, line)

		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Section header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			name := strings.TrimSpace(strings.Trim(line, "[]"))
			if sec, ok := cfg.Sections[name]; ok {
				cur = sec
				continue
			}
			cur = &SssdSection{
				Name:    name,
				Kind:    sectionKind(name),
				Options: make(map[string][]SssdKeyVal),
			}
			cfg.Sections[name] = cur
			cfg.Order = append(cfg.Order, name)
			continue
		}

		// key = value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // not a recognized option line
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}
		if cur == nil {
			cur = &SssdSection{
				Name:    "[global]",
				Kind:    "global",
				Options: make(map[string][]SssdKeyVal),
			}
			cfg.Sections[cur.Name] = cur
			cfg.Order = append(cfg.Order, cur.Name)
		}
		cur.Options[key] = append(cur.Options[key], SssdKeyVal{Value: value, Line: lineNum})
	}

	// Resolve AD provider sections.
	for _, name := range cfg.Order {
		sec := cfg.Sections[name]
		if sec.Kind != "domain" {
			continue
		}
		if vals := sec.Options["id_provider"]; len(vals) > 0 && strings.EqualFold(strings.TrimSpace(vals[0].Value), "ad") {
			cfg.HasAD = true
			cfg.AdSections = append(cfg.AdSections, sec)
		}
	}
	return cfg
}

// sectionKind derives the logical section kind from a header name.
func sectionKind(name string) string {
	switch {
	case strings.HasPrefix(name, "domain/"):
		return "domain"
	default:
		if idx := strings.Index(name, "/"); idx > 0 {
			return name[:idx]
		}
		return name
	}
}

// validateDomainStructure applies the domain-level structural checks that
// mirror SSSD's own sss_ini.c custom validators (verified in the C source:
// check_domain_id_provider, check_domain_inherit_from):
//   - id_provider is mandatory and must be one of ad|ipa|ldap|proxy|simple.
//   - inherit_from is NOT permitted inside a [domain/*] section (SSSD rejects
//     it there; it is only valid under the [sssd] section to reuse defaults).
//
// These checks run for every provider type, not just AD, so a misconfigured
// LDAP/IPA/proxy domain is flagged before log analysis.
// fallbackIdProviders is the id_provider list used when the embedded catalog is
// unavailable or carries no enumeration. "files" belongs here: it is documented
// and a real [domain/files] section uses it.
var fallbackIdProviders = []string{"ad", "ipa", "ldap", "proxy", "simple", "files"}

// idProviderList returns the accepted id_provider values and a human readable
// list for the error message.
//
// The values come from the embedded option catalog rather than from a literal
// baked into the code: the literal drifted from the documentation and rejected
// "files", so a correct [domain/files] section was reported as a fatal
// CONFIGURATION ERROR and sank the health score. Deriving the list from the
// single source of truth removes the possibility of drift.
func idProviderList() (map[string]bool, string) {
	if cat, err := loadEmbeddedCatalog(); err == nil && cat != nil {
		if opt, ok := cat.Options["id_provider"]; ok && len(opt.Values) > 0 {
			allowed := make(map[string]bool, len(opt.Values))
			for _, v := range opt.Values {
				allowed[strings.ToLower(v)] = true
			}
			return allowed, strings.Join(opt.Values, ", ")
		}
	}
	allowed := make(map[string]bool, len(fallbackIdProviders))
	for _, v := range fallbackIdProviders {
		allowed[v] = true
	}
	return allowed, strings.Join(fallbackIdProviders, ", ")
}

func validateDomainStructure(cfg *ParsedConfig, report *ReportData) {
	allowedProviders, provList := idProviderList()
	for _, name := range cfg.Order {
		sec := cfg.Sections[name]
		if sec.Kind != "domain" {
			continue
		}
		// id_provider must be present and valid.
		if vals := sec.Options["id_provider"]; len(vals) == 0 {
			msg := fmt.Sprintf("CONFIGURATION ERROR: domain '%s' is missing the mandatory 'id_provider' option. SSSD will not start this domain until a provider (%s) is set.", strings.TrimPrefix(name, "domain/"), provList)
			addConfigFindingEx(report, SevError, "config", msg, "sssd.conf", name+".id_provider", 0, "", "domain:missing-id-provider", ConfidenceHeuristic, "sssd.conf(5)")
		} else {
			prov := strings.ToLower(strings.TrimSpace(strings.Split(vals[0].Value, ",")[0]))
			if !allowedProviders[prov] {
				msg := fmt.Sprintf("CONFIGURATION ERROR: domain '%s' has invalid id_provider '%s'. Valid values are: %s.", strings.TrimPrefix(name, "domain/"), vals[0].Value, provList)
				addConfigFindingEx(report, SevError, "config", msg, "sssd.conf", name+".id_provider", vals[0].Line, "id_provider = "+vals[0].Value, "domain:invalid-id-provider", ConfidenceHeuristic, "sssd.conf(5)")
			}
		}
		// inherit_from is not allowed in per-domain sections.
		if v, line, ok := firstOption(sec, "inherit_from"); ok && v != "" {
			msg := fmt.Sprintf("CONFIGURATION ERROR: 'inherit_from' is not permitted inside a [domain/*] section (domain '%s'). It is only valid under [sssd] to reuse a domain template. Move the setting or remove it.", strings.TrimPrefix(name, "domain/"))
			addConfigFindingEx(report, SevError, "config", msg, "sssd.conf", name+".inherit_from", line, "inherit_from = "+v, "domain:inherit-from", ConfidenceHeuristic, "sssd.conf(5)")
		}
	}
}

// validateADConfig runs the typed AD option validator against a parsed config.
// It centralizes the AD checks that were previously scattered (and buggy) in
// analyzeSSSDConfig.
func validateADConfig(cfg *ParsedConfig, report *ReportData) {
	if !cfg.HasAD {
		return
	}
	report.ADProviderMode = true

	// Whole-config tuning hints (preserve existing behavior).
	hasTokengroupsFalse := false
	hasTimeout := false
	for _, name := range cfg.Order {
		sec := cfg.Sections[name]
		if v, _, ok := firstOption(sec, "ldap_use_tokengroups"); ok && strings.EqualFold(v, "false") {
			hasTokengroupsFalse = true
		}
		if _, _, ok := firstOption(sec, "timeout"); ok {
			hasTimeout = true
		}
	}
	if !hasTokengroupsFalse {
		report.Warnings = append(report.Warnings, "[TUNING] If AD users authenticate but fail authorization (missing groups), consider setting 'ldap_use_tokengroups = False'.")
	}
	if !hasTimeout {
		report.Warnings = append(report.Warnings, "[TUNING] No LDAP timeout specified. Adding 'timeout = 30' can help stabilize slow Active Directory connections.")
	}

	// simple_allow_* / access_provider reconciliation (whole-config, as before).
	hasSimpleAllow, hasSimpleAllowGroups, accessProviderSimple := false, false, false
	for _, name := range cfg.Order {
		sec := cfg.Sections[name]
		if _, _, ok := firstOption(sec, "simple_allow_groups"); ok {
			hasSimpleAllowGroups, hasSimpleAllow = true, true
		} else if _, _, ok := firstOption(sec, "simple_allow_users"); ok {
			hasSimpleAllow = true
		}
		if v, _, ok := firstOption(sec, "access_provider"); ok && strings.EqualFold(v, "simple") {
			accessProviderSimple = true
		}
	}
	if hasSimpleAllow && !accessProviderSimple {
		msg := "CONFIGURATION ERROR: 'simple_allow_users' or 'simple_allow_groups' is used in sssd.conf, but 'access_provider = simple' is not set (e.g., using 'ad'). These parameters will be ignored. Use ad_access_filter instead."
		addConfigFindingEx(report, SevError, "access", msg, "sssd.conf", "simple_allow", 0, "", "access:simple-allow-without-filter", ConfidenceHeuristic, "sssd-simple(5)")
	}
	if hasSimpleAllowGroups {
		report.Warnings = append(report.Warnings, "[DIAGNOSTIC HINT] 'simple_allow_groups' is active. If users authenticate but fail authorization, test by commenting it out and using 'simple_allow_users = <username>' to isolate group resolution issues.")
	}

	// Per-AD-section checks.
	for _, sec := range cfg.AdSections {
		validateADSections(sec, report)
	}
}

// validateDuplicateKeys reports duplicate parameter occurrences across ALL sections
// of sssd.conf, regardless of the provider kind. Duplicate settings make SSSD's
// behavior order-dependent and unpredictable, so this check applies to the full config
// — not just AD domains. (In the legacy scanner was this "duplicate parameter" check;
// Phase A moved it to its own ungated pass so it still runs for every sssd.conf.)
func validateDuplicateKeys(cfg *ParsedConfig, report *ReportData) {
	for _, name := range cfg.Order {
		sec := cfg.Sections[name]
		for key, vals := range sec.Options {
			if key == "debug_level" || len(vals) < 2 {
				continue
			}
			msg := fmt.Sprintf("CONFIGURATION ERROR: Duplicate parameter '%s' found in section %s of sssd.conf. SSSD may behave unpredictably.", key, sec.Name)
			addConfigFindingEx(report, SevError, "duplicate", msg, "sssd.conf", sec.Name+"."+key, vals[1].Line, fmt.Sprintf("%s = %s", key, vals[1].Value), "syntax:duplicate-key", ConfidenceHeuristic, "sssd.conf(5)")
		}
	}
}

// firstOption returns the value and source line of the first occurrence of the
// lowercased key in the section. ok is false if the key is absent.
func firstOption(sec *SssdSection, key string) (string, int, bool) {
	vals, ok := sec.Options[strings.ToLower(key)]
	if !ok || len(vals) == 0 {
		return "", 0, false
	}
	return strings.TrimSpace(vals[0].Value), vals[0].Line, true
}

// addConfigFinding records a structured finding and mirrors it into the legacy
// Problems/Warnings string slices so existing text/HTML/CLI paths pick it up.
// It is the compatibility shim over addConfigFindingEx: findings built this
// way carry the heuristic confidence layer and no rule/doc attribution.
func addConfigFinding(report *ReportData, sev Severity, category, message, sourcePath, sourceKey string, sourceLine int, evidence string) {
	addConfigFindingEx(report, sev, category, message, sourcePath, sourceKey, sourceLine, evidence, "", ConfidenceHeuristic, "")
}

// addConfigFindingEx is the attributed finding constructor every validator
// should use: ruleID names the check, confidence names the ground-truth layer
// that decided, and docRef points at the backing documentation. Severity is
// capped by the central policy in catalogSeverity so a weak knowledge layer
// can never escalate a finding above SevWarning on its own.
func addConfigFindingEx(report *ReportData, sev Severity, category, message, sourcePath, sourceKey string, sourceLine int, evidence string, ruleID, confidence, docRef string) {
	sev = catalogSeverity(sev, confidence, ruleID)
	report.ConfigFindings = append(report.ConfigFindings, ConfigFinding{
		Severity:   sev,
		Category:   category,
		Message:    message,
		SourcePath: sourcePath,
		SourceKey:  sourceKey,
		SourceLine: sourceLine,
		Evidence:   evidence,
		RuleID:     ruleID,
		Confidence: confidence,
		DocRef:     docRef,
	})
	switch sev {
	case SevError, SevCritical:
		report.Problems = append(report.Problems, message)
	default:
		report.Warnings = append(report.Warnings, message)
	}
}

// isIPAddress reports whether s is a syntactically valid IPv4/IPv6 address.
func isIPAddress(s string) bool {
	return net.ParseIP(strings.Trim(s, "[]")) != nil
}

// validateADSections performs the AD provider specific option checks for one
// [domain/*] section whose id_provider is ad.
func validateADSections(sec *SssdSection, report *ReportData) {
	prefix := sec.Name + "."

	// enumerate = true (deprecated for AD).
	if v, line, ok := firstOption(sec, "enumerate"); ok && strings.EqualFold(v, "true") {
		if !report.EnumerateIssue {
			report.EnumerateIssue = true
			msg := "[DEPRECATION] 'enumerate = true' is set in sssd.conf. This causes severe performance issues, is deprecated for AD/IPA, and is unsupported in SSSD 2.10+."
			addConfigFindingEx(report, SevWarning, "enumerate", msg, "sssd.conf", prefix+"enumerate", line, "enumerate = "+v, "ad:enumerate-deprecated", ConfidenceHeuristic, "sssd-ad(5)")
		}
	}

	// use_fully_qualified_names (informational flag).
	if v, _, ok := firstOption(sec, "use_fully_qualified_names"); ok && strings.EqualFold(v, "true") {
		report.UseFQDNSet = true
	}

	// ad_domain: capture canonical value; flag IP or malformed.
	if v, line, ok := firstOption(sec, "ad_domain"); ok && v != "" {
		canonical := strings.ToLower(strings.TrimPrefix(v, "."))
		if report.AdDomain == "" {
			report.AdDomain = canonical
		}
		if isIPAddress(strings.TrimPrefix(v, ".")) {
			msg := fmt.Sprintf("CONFIGURATION ERROR: 'ad_domain' is set to the IP address '%s'. It must be the AD DNS domain name, not an IP.", v)
			addConfigFindingEx(report, SevError, "ad_domain", msg, "sssd.conf", prefix+"ad_domain", line, "ad_domain = "+v, "log:ad-domain-unresolvable", ConfidenceLog, "sssd-ad(5)")
		}
	}

	// ad_server / ad_hostname must not be an IP (GSSAPI SPN risk).
	checkServerIP := func(key string) {
		if v, line, ok := firstOption(sec, key); ok && v != "" && isIPAddress(v) {
			msg := fmt.Sprintf("Kerberos SPN Risk: '%s = %s' is configured as an IP address instead of a hostname. Kerberos (GSSAPI) requires hostnames to request tickets.", key, v)
			addConfigFindingEx(report, SevError, "ad_server", msg, "sssd.conf", prefix+key, line, key+" = "+v, "log:ad-server-unresolvable", ConfidenceLog, "sssd-ad(5)")
		}
	}
	checkServerIP("ad_server")
	checkServerIP("ad_hostname")

	// krb5_realm vs ad_domain must match.
	if realm, rLine, rOK := firstOption(sec, "krb5_realm"); rOK && realm != "" {
		if report.AdDomain != "" && !strings.EqualFold(realm, report.AdDomain) {
			msg := fmt.Sprintf("[CRITICAL] 'krb5_realm = %s' does not match the AD domain '%s'. Kerberos authentication against AD will fail. Fix krb5_realm (recommended: %s).", realm, report.AdDomain, strings.ToUpper(report.AdDomain))
			addConfigFindingEx(report, SevCritical, "krb5_realm", msg, "sssd.conf", prefix+"krb5_realm", rLine, "krb5_realm = "+realm, "correlate:krb5-realm-mismatch", ConfidenceLog, "sssd-ad(5)")
		}
	}

	// ldap_sasl_mech must be GSSAPI for AD.
	if v, line, ok := firstOption(sec, "ldap_sasl_mech"); ok && v != "" && !strings.EqualFold(v, "gssapi") {
		msg := fmt.Sprintf("CONFIGURATION ERROR: 'ldap_sasl_mech = %s' is incompatible with the ad provider, which requires GSSAPI.", v)
		addConfigFindingEx(report, SevError, "ldap_sasl_mech", msg, "sssd.conf", prefix+"ldap_sasl_mech", line, "ldap_sasl_mech = "+v, "log:unsupported-sasl-mech", ConfidenceLog, "sssd-ldap(5)")
	}

	// kerberos_method must be keytab or secrets.
	if v, line, ok := firstOption(sec, "kerberos_method"); ok && v != "" {
		lv := strings.ToLower(v)
		if lv != "keytab" && lv != "secrets" {
			msg := fmt.Sprintf("CONFIGURATION ERROR: 'kerberos_method = %s' is invalid. Allowed values are 'keytab' or 'secrets'.", v)
			addConfigFindingEx(report, SevError, "kerberos_method", msg, "sssd.conf", prefix+"kerberos_method", line, "kerberos_method = "+v, "log:kerberos-method-secret", ConfidenceLog, "sssd-ad(5)")
		} else if lv == "secrets" {
			report.Warnings = append(report.Warnings, "[DIAGNOSTIC HINT] 'kerberos_method = secrets' is set. SSSD will obtain the machine credentials from the secrets store; ensure the keytab principal is still valid and updateable by adcli.")
		}
	}

	// ldap_id_mapping = false requires RFC2307 attrs.
	if v, line, ok := firstOption(sec, "ldap_id_mapping"); ok && strings.EqualFold(v, "false") {
		msg := "[WARNING] 'ldap_id_mapping = False' is set. AD logins will fail silently unless UNIX attributes (uidNumber, gidNumber) are manually populated in Active Directory (RFC2307)."
		addConfigFindingEx(report, SevWarning, "ldap_id_mapping", msg, "sssd.conf", prefix+"ldap_id_mapping", line, "ldap_id_mapping = "+v, "log:manual-id-mapping", ConfidenceLog, "sssd-ldap(5)")
	}

	// case_sensitive = True on AD is a classic "user not found" source.
	if v, line, ok := firstOption(sec, "case_sensitive"); ok && strings.EqualFold(v, "true") {
		msg := "[WARNING] 'case_sensitive = True' is set on an AD provider. Active Directory object names are case-insensitive; this commonly causes spurious 'user not found' / authorization failures. Consider removing it or setting 'False'."
		addConfigFindingEx(report, SevWarning, "case_sensitive", msg, "sssd.conf", prefix+"case_sensitive", line, "case_sensitive = "+v, "log:case-sensitive-on-ad", ConfidenceLog, "sssd-ad(5)")
	}

	// krb5_validate = false is a security risk.
	if v, line, ok := firstOption(sec, "krb5_validate"); ok && strings.EqualFold(v, "false") {
		msg := "[SECURITY RISK] 'krb5_validate = false' is set. This disables KDC spoofing protection. If used to bypass the AD RC4 bug, remove this and fix the AD operatingSystemVersion attribute or update local crypto policies instead."
		addConfigFindingEx(report, SevError, "krb5_validate", msg, "sssd.conf", prefix+"krb5_validate", line, "krb5_validate = "+v, "log:krb5-validate-disabled", ConfidenceLog, "sssd-krb5(5)")
	}

	validateADAdvancedOptions(sec, report, prefix)
}

// validateADAdvancedOptions implements the Phase 2 AD-specific checks that
// require the section context (GPO, site discovery, machine-account password
// renewal, StartTLS incompatibility, ad_hostname consistency).
func validateADAdvancedOptions(sec *SssdSection, report *ReportData, prefix string) {
	// ldap_id_use_start_tls is incompatible with the ad provider, which
	// always uses SASL/GSSAPI over its own connection.
	if v, line, ok := firstOption(sec, "ldap_id_use_start_tls"); ok && strings.EqualFold(v, "true") {
		msg := "CONFIGURATION ERROR: 'ldap_id_use_start_tls = true' is set on an 'ad' provider. The AD provider uses SASL/GSSAPI and does not honour StartTLS; this option is ignored at best and rejected at worst."
		addConfigFindingEx(report, SevError, "tls", msg, "sssd.conf", prefix+"ldap_id_use_start_tls", line, "ldap_id_use_start_tls = "+v, "log:starttls-on-ad", ConfidenceLog, "sssd-ldap(5)")
	}

	// ad_gpo_access_control accepts 'disabled', 'permissive' or 'enforcing'.
	// 'disabled' was added in SSSD 1.14 and is valid (verified against
	// sssd-ad(5): "It can be set to 'disabled', 'permissive' or
	// 'enforcing'"); the previous check rejected it and produced a false
	// CONFIGURATION ERROR.
	if v, line, ok := firstOption(sec, "ad_gpo_access_control"); ok && v != "" {
		lv := strings.ToLower(v)
		if lv != "disabled" && lv != "permissive" && lv != "enforcing" {
			msg := fmt.Sprintf("CONFIGURATION ERROR: 'ad_gpo_access_control = %s' is invalid. Allowed values are 'disabled', 'permissive' or 'enforcing'. SSSD may refuse to start or fall back to permissive mode.", v)
			addConfigFindingEx(report, SevError, "gpo", msg, "sssd.conf", prefix+"ad_gpo_access_control", line, "ad_gpo_access_control = "+v, "log:invalid-gpo-value", ConfidenceLog, "sssd-ad(5)")
		}
	}

	// ad_site only takes effect when DNS site discovery is enabled.
	if _, _, siteOK := firstOption(sec, "ad_site"); siteOK {
		if v, line, ok := firstOption(sec, "ad_enable_dns_sites"); ok && strings.EqualFold(v, "false") {
			msg := "CONFIGURATION WARNING: 'ad_site' is configured together with 'ad_enable_dns_sites = false'. The static site is still used, but automatic DC failover across sites is disabled; on site outage the client cannot locate another DC."
			addConfigFindingEx(report, SevWarning, "ad_site", msg, "sssd.conf", prefix+"ad_enable_dns_sites", line, "ad_enable_dns_sites = "+v, "log:static-site-no-failover", ConfidenceLog, "sssd-ad(5)")
		}
	}

	// ad_machine_account_password_renewal_opts must be 'N:M' (days, hours).
	if v, line, ok := firstOption(sec, "ad_machine_account_password_renewal_opts"); ok && v != "" {
		parts := strings.Split(v, ":")
		valid := len(parts) == 2
		if valid {
			for _, p := range parts {
				if _, err := strconv.Atoi(strings.TrimSpace(p)); err != nil {
					valid = false
					break
				}
			}
		}
		if !valid {
			msg := fmt.Sprintf("CONFIGURATION WARNING: 'ad_machine_account_password_renewal_opts = %s' is malformed. The expected format is '<renewal days>:<renewal hours>' (e.g. '30:4'); SSSD falls back to the defaults.", v)
			addConfigFindingEx(report, SevWarning, "machine_account", msg, "sssd.conf", prefix+"ad_machine_account_password_renewal_opts", line, "ad_machine_account_password_renewal_opts = "+v, "log:malformed-renewal-opts", ConfidenceLog, "sssd-ad(5)")
		}
	}

	// ad_hostname should match the host FQDN recorded in the supportconfig.
	if v, line, ok := firstOption(sec, "ad_hostname"); ok && v != "" && report.Hostname != "" {
		if !strings.EqualFold(v, report.Hostname) {
			msg := fmt.Sprintf("CONFIGURATION WARNING: 'ad_hostname = %s' does not match the system hostname '%s'. SPNs and the machine account keytab are tied to the real hostname; a mismatch causes Kerberos 'Server not found in Kerberos database' errors.", v, report.Hostname)
			addConfigFindingEx(report, SevWarning, "ad_hostname", msg, "sssd.conf", prefix+"ad_hostname", line, "ad_hostname = "+v, "log:ad-hostname-mismatch", ConfidenceLog, "sssd-ad(5)")
		}
	}
}
