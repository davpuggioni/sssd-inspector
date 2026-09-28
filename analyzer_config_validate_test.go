// analyzer_config_validate_test.go
//
// Direct tests for the AD validator branches (config_parser.go
// validateADSections, ~55% coverage): IP-valued ad_domain/ad_server,
// krb5_realm mismatch, ldap_sasl_mech, kerberos_method and
// use_fully_qualified_names.
package main

import (
	"testing"
)

// validateOneSection parses a single [domain/x] section and runs the AD
// validator against it, returning the populated report.
func validateOneSection(t *testing.T, body string) *ReportData {
	t.Helper()
	cfg := parseSssdConfig("[domain/ad]\n" + body)
	sec := cfg.Sections["domain/ad"]
	if sec == nil {
		t.Fatalf("failed to parse [domain/ad] section")
	}
	report := &ReportData{}
	validateADSections(sec, report)
	return report
}

func TestValidateADSections_EnumerateDeprecated(t *testing.T) {
	r := validateOneSection(t, "id_provider = ad\nenumerate = true\n")
	if !r.EnumerateIssue {
		t.Errorf("enumerate=true must set EnumerateIssue")
	}
	if !containsStringCategory(r, "enumerate") {
		t.Errorf("expected an 'enumerate' finding, got %v", findingCategories(r))
	}
	// Second section must not duplicate the message.
	cfg := parseSssdConfig("[domain/a]\nid_provider = ad\nenumerate = true\n\n[domain/b]\nid_provider = ad\nenumerate = true\n")
	r2 := &ReportData{}
	for _, name := range []string{"domain/a", "domain/b"} {
		validateADSections(cfg.Sections[name], r2)
	}
	n := 0
	for _, f := range r2.ConfigFindings {
		if f.Category == "enumerate" {
			n++
		}
	}
	if n != 1 {
		t.Errorf("enumerate deprecation must be reported once, got %d", n)
	}
}

func TestValidateADSections_AdDomainIPRejected(t *testing.T) {
	r := validateOneSection(t, "id_provider = ad\nad_domain = 10.0.0.5\n")
	if !containsStringCategory(r, "ad_domain") {
		t.Errorf("expected an 'ad_domain' IP finding, got %v", findingCategories(r))
	}
}

func TestValidateADSections_AdServerIPRisk(t *testing.T) {
	r := validateOneSection(t, "id_provider = ad\nad_domain = corp.example.com\nad_server = 10.0.0.5\n")
	if !containsStringCategory(r, "ad_server") {
		t.Errorf("expected an 'ad_server' SPN-risk finding, got %v", findingCategories(r))
	}
	r2 := validateOneSection(t, "id_provider = ad\nad_domain = corp.example.com\nad_server = dc01.corp.example.com\n")
	if containsStringCategory(r2, "ad_server") {
		t.Errorf("hostname ad_server must not raise a finding, got %v", findingCategories(r2))
	}
}

func TestValidateADSections_Krb5RealmMismatchCritical(t *testing.T) {
	r := validateOneSection(t, "id_provider = ad\nad_domain = corp.example.com\nkrb5_realm = OTHER.EXAMPLE.COM\n")
	found := false
	for _, f := range r.ConfigFindings {
		if f.Category == "krb5_realm" && f.Severity == SevCritical {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a critical krb5_realm finding, got %v", findingCategories(r))
	}
}

func TestValidateADSections_SaslMechAndMethod(t *testing.T) {
	r := validateOneSection(t, "id_provider = ad\nldap_sasl_mech = EXTERNAL\nkerberos_method = password\n")
	if !containsStringCategory(r, "ldap_sasl_mech") {
		t.Errorf("expected a 'ldap_sasl_mech' finding, got %v", findingCategories(r))
	}
	if !containsStringCategory(r, "kerberos_method") {
		t.Errorf("expected a 'kerberos_method' finding, got %v", findingCategories(r))
	}
}

func TestValidateADSections_UseFQDNFlag(t *testing.T) {
	r := validateOneSection(t, "id_provider = ad\nuse_fully_qualified_names = True\n")
	if !r.UseFQDNSet {
		t.Errorf("use_fully_qualified_names=True must set UseFQDNSet")
	}
}

// TestValidateADSections_GpoAccessControlDisabled is the regression test for
// the false positive on 'ad_gpo_access_control = disabled': the value is
// documented in sssd-ad(5) as the "do not evaluate GPOs at all" mode and must
// not be reported as a CONFIGURATION ERROR.
func TestValidateADSections_GpoAccessControlDisabled(t *testing.T) {
	for _, val := range []string{"disabled", "Disabled", "permissive", "enforcing"} {
		r := validateOneSection(t, "id_provider = ad\nad_gpo_access_control = "+val+"\n")
		if containsStringCategory(r, "gpo") {
			t.Errorf("ad_gpo_access_control = %s must be accepted, got %v", val, findingCategories(r))
		}
	}
}

// TestValidateADSections_GpoAccessControlInvalid keeps the guard for genuinely
// unsupported values.
func TestValidateADSections_GpoAccessControlInvalid(t *testing.T) {
	r := validateOneSection(t, "id_provider = ad\nad_gpo_access_control = yes\n")
	if !containsStringCategory(r, "gpo") {
		t.Errorf("ad_gpo_access_control = yes must be rejected, got %v", findingCategories(r))
	}
}

// containsStringCategory checks a finding category list (shares the
// substring semantics of containsString).
func containsStringCategory(r *ReportData, expected string) bool {
	return containsString(findingCategories(r), expected)
}
