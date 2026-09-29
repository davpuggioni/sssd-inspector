// config_semantics_test.go
//
// Regression tests for the provider-aware semantic rules and the central
// severity policy.
package main

import (
	"strings"
	"testing"
)

func semanticFindings(t *testing.T, body string) *ReportData {
	t.Helper()
	cfg := parseSssdConfig(body)
	report := &ReportData{}
	validateConfigSemantics(cfg, report)
	return report
}

func hasRuleID(r *ReportData, ruleID string) bool {
	for _, f := range r.ConfigFindings {
		if f.RuleID == ruleID {
			return true
		}
	}
	return false
}

// The user's real LDAP supportconfig must stay clean: ldap_uri and the bind
// credentials on an id_provider = ldap domain are live settings.
func TestSemantic_LDAPDomainIsClean(t *testing.T) {
	body := "[domain/LDAP]\n" +
		"id_provider = ldap\n" +
		"auth_provider = ldap\n" +
		"ldap_uri = ldaps://ldaps.intranet.eon.com:636/\n" +
		"ldap_search_base = ou=Users,ou=EEA,o=EON,c=DE\n" +
		"ldap_default_bind_dn = cn=D29424,ou=Users,ou=EEA,o=EON,c=DE\n" +
		"ldap_default_authtok = secret\n" +
		"ldap_schema = rfc2307bis\n" +
		"ldap_tls_cacert = /etc/ca.crt\n"
	r := semanticFindings(t, body)
	if len(r.ConfigFindings) != 0 {
		t.Errorf("healthy LDAP domain produced semantic findings: %+v", r.ConfigFindings)
	}
}

// ldap_uri on a files domain is ignored by SSSD and must be flagged.
func TestSemantic_LDAPURIOnFilesDomain(t *testing.T) {
	r := semanticFindings(t, "[domain/files]\nid_provider = files\nldap_uri = ldap://x\n")
	if !hasRuleID(r, "semantic:ldap-uri-without-ldap-provider") {
		t.Fatalf("ldap_uri on files domain not flagged: %+v", r.ConfigFindings)
	}
	for _, f := range r.ConfigFindings {
		if f.Severity != SevWarning {
			t.Errorf("semantic finding severity = %v, want SevWarning", f.Severity)
		}
		if f.RuleID == "" || f.Confidence == "" || f.DocRef == "" {
			t.Errorf("finding misses evidence contract fields: %+v", f)
		}
	}
}

// Bind credentials with auth_provider = none are dead configuration.
func TestSemantic_DeadBindCredentials(t *testing.T) {
	r := semanticFindings(t, "[domain/x]\nid_provider = ldap\nauth_provider = none\nldap_default_bind_dn = cn=y\n")
	if !hasRuleID(r, "semantic:dead-bind-credentials") {
		t.Fatalf("dead bind DN not flagged: %+v", r.ConfigFindings)
	}
	// The secret itself must never be echoed into evidence.
	for _, f := range r.ConfigFindings {
		if strings.Contains(f.Evidence, "cn=y") {
			t.Errorf("evidence leaks the bind DN value: %q", f.Evidence)
		}
	}
}

// Bind credentials on an ldap domain are live and must stay silent.
func TestSemantic_LiveBindCredentialsSilent(t *testing.T) {
	r := semanticFindings(t, "[domain/x]\nid_provider = ldap\nldap_default_bind_dn = cn=y\n")
	if hasRuleID(r, "semantic:dead-bind-credentials") {
		t.Errorf("live bind DN wrongly flagged: %+v", r.ConfigFindings)
	}
}

// Search tuning on a proxy domain is ignored by SSSD.
func TestSemantic_DeadLDAPOptionsOnProxy(t *testing.T) {
	r := semanticFindings(t, "[domain/x]\nid_provider = proxy\nproxy_lib_name = files\nldap_schema = rfc2307bis\n")
	if !hasRuleID(r, "semantic:dead-ldap-options") {
		t.Fatalf("dead ldap_schema on proxy not flagged: %+v", r.ConfigFindings)
	}
}

// Severity policy: a non-allowlisted rule can never escalate.
func TestSeverityPolicy_CapsUnknownRules(t *testing.T) {
	if got := catalogSeverity(SevError, ConfidenceMan, "catalog:unknown-option"); got != SevWarning {
		t.Errorf("catalogSeverity(Error, man, unknown-option) = %v, want SevWarning", got)
	}
	if got := catalogSeverity(SevCritical, ConfidenceHeuristic, "something:new"); got != SevWarning {
		t.Errorf("catalogSeverity(Critical, heuristic, new) = %v, want SevWarning", got)
	}
	if got := catalogSeverity(SevError, ConfidenceHeuristic, "domain:invalid-id-provider"); got != SevError {
		t.Errorf("allowlisted rule was capped: got %v, want SevError", got)
	}
	if got := catalogSeverity(SevWarning, ConfidenceMan, "catalog:unknown-option"); got != SevWarning {
		t.Errorf("warnings must pass through unchanged, got %v", got)
	}
}

// Evidence contract: every catalog validator finding carries rule, confidence
// and doc attribution so a false positive is traceable.
func TestEvidenceContract_CatalogFindings(t *testing.T) {
	r := catalogFindings(t, "[domain/x]\nid_provider = ad\nldap_url = ldap://dc02\nldap_schema = rfc2309\n")
	if len(r.ConfigFindings) == 0 {
		t.Fatal("expected catalog findings, got none")
	}
	for _, f := range r.ConfigFindings {
		if f.RuleID == "" {
			t.Errorf("finding has no rule_id: %+v", f)
		}
		if f.Confidence == "" {
			t.Errorf("finding has no confidence: %+v", f)
		}
	}
}

// Catalog confidence: API-sourced options outrank man-only ones.
func TestCatalog_ConfidenceTags(t *testing.T) {
	cat, err := loadEmbeddedCatalog()
	if err != nil {
		t.Fatalf("loadEmbeddedCatalog() error = %v", err)
	}
	if got := cat.Options["ldap_uri"].Confidence; got != ConfidenceAPI {
		t.Errorf("ldap_uri confidence = %q, want %q", got, ConfidenceAPI)
	}
	if got := cat.Options["config_file_version"].Confidence; got != ConfidenceCurated {
		t.Errorf("config_file_version confidence = %q, want %q", got, ConfidenceCurated)
	}
	for name, opt := range cat.Options {
		if opt.Confidence == "" {
			t.Errorf("option %q has no confidence tag", name)
			break
		}
	}
}
