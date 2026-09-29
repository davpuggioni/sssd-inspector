// config_semantics.go
//
// Provider-aware semantic rules for sssd.conf: cross-option checks the
// syntactic catalog validator cannot express (does this setting make sense
// given THIS domain's provider?). Small, dependency-free, deterministic.
package main

import (
	"fmt"
	"strings"
)

// SemanticRule is one provider-aware cross-option check.
type SemanticRule struct {
	ID       string
	Category string
	Check    func(cfg *ParsedConfig, report *ReportData)
}

// semanticRules is the registry. New rules are added here with a test each.
var semanticRules = []SemanticRule{
	{ID: "semantic:ldap-uri-without-ldap-provider", Category: "provider", Check: ruleLDAPURIRequiresLDAPProvider},
	{ID: "semantic:dead-bind-credentials", Category: "provider", Check: ruleDeadBindCredentials},
	{ID: "semantic:dead-ldap-options", Category: "provider", Check: ruleDeadLDAPOptions},
}

// validateConfigSemantics runs every registered semantic rule.
func validateConfigSemantics(cfg *ParsedConfig, report *ReportData) {
	for _, rule := range semanticRules {
		rule.Check(cfg, report)
	}
}

// domainProvider returns the normalized id_provider of a domain section.
func domainProvider(sec *SssdSection) string {
	if v, _, ok := firstOption(sec, "id_provider"); ok {
		return strings.ToLower(strings.TrimSpace(v))
	}
	return ""
}

// ldapCapableProviders speak LDAP for ldap_uri/search-base style options.
var ldapCapableProviders = map[string]bool{
	"ldap": true, "ad": true, "ipa": true,
}

func ruleLDAPURIRequiresLDAPProvider(cfg *ParsedConfig, report *ReportData) {
	for _, name := range cfg.Order {
		sec := cfg.Sections[name]
		if sec.Kind != "domain" {
			continue
		}
		prov := domainProvider(sec)
		if prov == "" || ldapCapableProviders[prov] {
			continue
		}
		for _, key := range []string{"ldap_uri", "ldap_search_base", "ldap_backup_uri"} {
			if v, line, ok := firstOption(sec, key); ok && strings.TrimSpace(v) != "" {
				msg := fmt.Sprintf("CONFIGURATION WARNING: '%s' is set in domain '%s' but id_provider = %s, which never contacts an LDAP server. SSSD ignores the setting; move it to an ldap/ad/ipa domain or fix the provider.",
					key, strings.TrimPrefix(name, "domain/"), prov)
				addConfigFindingEx(report, SevWarning, "provider", msg, "sssd.conf",
					name+"."+key, line, key+" = "+v,
					"semantic:ldap-uri-without-ldap-provider", ConfidenceHeuristic, "sssd-ldap(5)")
			}
		}
	}
}

// bindCredentialOptions are only consumed on LDAP-backed authentication.
var bindCredentialOptions = []string{
	"ldap_default_bind_dn", "ldap_default_authtok", "ldap_default_authtok_type",
}

func ruleDeadBindCredentials(cfg *ParsedConfig, report *ReportData) {
	for _, name := range cfg.Order {
		sec := cfg.Sections[name]
		if sec.Kind != "domain" {
			continue
		}
		auth := ""
		if v, _, ok := firstOption(sec, "auth_provider"); ok {
			auth = strings.ToLower(strings.TrimSpace(v))
		} else {
			auth = domainProvider(sec)
		}
		if auth == "" || ldapCapableProviders[auth] || auth == "krb5" || auth == "proxy" {
			continue
		}
		for _, key := range bindCredentialOptions {
			if v, line, ok := firstOption(sec, key); ok && strings.TrimSpace(v) != "" {
				msg := fmt.Sprintf("CONFIGURATION WARNING: '%s' is set in domain '%s' but auth_provider = %s, which never performs an LDAP bind. The credential is dead configuration; remove it or fix the provider.",
					key, strings.TrimPrefix(name, "domain/"), auth)
				addConfigFindingEx(report, SevWarning, "provider", msg, "sssd.conf",
					name+"."+key, line, key+" = ***",
					"semantic:dead-bind-credentials", ConfidenceHeuristic, "sssd-ldap(5)")
			}
		}
	}
}

func ruleDeadLDAPOptions(cfg *ParsedConfig, report *ReportData) {
	for _, name := range cfg.Order {
		sec := cfg.Sections[name]
		if sec.Kind != "domain" {
			continue
		}
		prov := domainProvider(sec)
		if prov == "" || ldapCapableProviders[prov] {
			continue
		}
		for _, key := range []string{"ldap_schema", "ldap_tls_cacert", "ldap_tls_reqcert", "ldap_search_timeout", "ldap_network_timeout", "ldap_opt_timeout"} {
			if v, line, ok := firstOption(sec, key); ok && strings.TrimSpace(v) != "" {
				msg := fmt.Sprintf("CONFIGURATION WARNING: '%s' is set in domain '%s' but id_provider = %s, which never performs LDAP searches. SSSD ignores the setting.",
					key, strings.TrimPrefix(name, "domain/"), prov)
				addConfigFindingEx(report, SevWarning, "provider", msg, "sssd.conf",
					name+"."+key, line, key+" = "+v,
					"semantic:dead-ldap-options", ConfidenceHeuristic, "sssd-ldap(5)")
			}
		}
	}
}
