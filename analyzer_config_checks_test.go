// analyzer_config_checks_test.go
package main

import (
	"strings"
	"testing"
)

func TestValidateConfigStructure_NoDomains(t *testing.T) {
	cfg := parseSssdConfig("[sssd]\nservices = nss, pam\n")
	var report ReportData
	validateConfigStructure(cfg, &report)

	found := false
	for _, f := range report.ConfigFindings {
		if f.Category == "structure" && strings.Contains(f.Message, "no [domain/NAME] sections") {
			found = true
		}
	}
	if !found {
		t.Errorf("missing-domain configuration was not detected")
	}
}

func TestValidateConfigStructure_ServicesMissingPam(t *testing.T) {
	cfg := parseSssdConfig("[sssd]\nservices = nss\n\n[domain/example.com]\nid_provider = ad\n")
	var report ReportData
	validateConfigStructure(cfg, &report)

	found := false
	for _, f := range report.ConfigFindings {
		if strings.Contains(f.Message, "does not include 'pam'") {
			found = true
		}
	}
	if !found {
		t.Errorf("missing pam responder was not detected")
	}
}

func TestValidateConfigStructure_HealthyConfig(t *testing.T) {
	cfg := parseSssdConfig("[sssd]\nservices = nss, pam, ssh\n\n[domain/example.com]\nid_provider = ad\n")
	var report ReportData
	validateConfigStructure(cfg, &report)

	if len(report.ConfigFindings) != 0 {
		t.Errorf("healthy config produced findings: %+v", report.ConfigFindings)
	}
}

func TestValidateIDMapRanges_Overlap(t *testing.T) {
	cfg := parseSssdConfig("[domain/a.example.com]\nid_provider = ad\nldap_idmap_range_min = 10000\nldap_idmap_range_max = 20000\n\n" +
		"[domain/b.example.com]\nid_provider = ad\nldap_idmap_range_min = 15000\nldap_idmap_range_max = 30000\n")
	var report ReportData
	validateIDMapRanges(cfg, &report)

	found := false
	for _, f := range report.ConfigFindings {
		if f.Category == "idmap" && strings.Contains(f.Message, "overlap") {
			found = true
		}
	}
	if !found {
		t.Errorf("overlapping idmap ranges were not detected")
	}
}

// TestValidateIDMapRanges_RealOptionNames is the regression test for the
// silent no-op bug: the checker used to look up 'ldap_idmap_min_id' /
// 'ldap_idmap_max_id', which are not SSSD options, so a real sssd.conf using
// the documented names was never validated.
func TestValidateIDMapRanges_RealOptionNames(t *testing.T) {
	cfg := parseSssdConfig("[domain/a.example.com]\nid_provider = ad\nldap_idmap_range_min = 10000\nldap_idmap_range_max = 20000\n\n" +
		"[domain/b.example.com]\nid_provider = ad\nldap_idmap_range_min = 15000\nldap_idmap_range_max = 30000\n")
	var report ReportData
	validateIDMapRanges(cfg, &report)

	if len(report.ConfigFindings) == 0 {
		t.Fatalf("documented ldap_idmap_range_min/max options must be validated; got no findings")
	}
	for _, f := range report.ConfigFindings {
		if f.Category != "idmap" {
			t.Errorf("unexpected finding category %q: %s", f.Category, f.Message)
		}
	}
}

// TestValidateIDMapRanges_AdjacentRangesAreNotOverlapping documents that
// ldap_idmap_range_max is exclusive, so [10000,20000) and [20000,30000) are
// disjoint and must not be reported.
func TestValidateIDMapRanges_AdjacentRangesAreNotOverlapping(t *testing.T) {
	cfg := parseSssdConfig("[domain/a.example.com]\nid_provider = ad\nldap_idmap_range_min = 10000\nldap_idmap_range_max = 20000\n\n" +
		"[domain/b.example.com]\nid_provider = ad\nldap_idmap_range_min = 20000\nldap_idmap_range_max = 30000\n")
	var report ReportData
	validateIDMapRanges(cfg, &report)

	if len(report.ConfigFindings) != 0 {
		t.Errorf("adjacent (non-overlapping) ranges must be accepted: %+v", report.ConfigFindings)
	}
}

// TestValidateIDMapRanges_LegacyAliasStillWorks keeps backwards compatibility
// with configs using the older idmap_range_min/max spelling.
func TestValidateIDMapRanges_LegacyAliasStillWorks(t *testing.T) {
	cfg := parseSssdConfig("[domain/a.example.com]\nid_provider = ad\nidmap_range_min = 10000\nidmap_range_max = 20000\n\n" +
		"[domain/b.example.com]\nid_provider = ad\nidmap_range_min = 15000\nidmap_range_max = 30000\n")
	var report ReportData
	validateIDMapRanges(cfg, &report)

	found := false
	for _, f := range report.ConfigFindings {
		if f.Category == "idmap" && strings.Contains(f.Message, "overlap") {
			found = true
		}
	}
	if !found {
		t.Errorf("legacy idmap_range_min/max aliases must still be validated")
	}
}

func TestValidateIDMapRanges_InvertedRange(t *testing.T) {
	cfg := parseSssdConfig("[domain/example.com]\nid_provider = ad\nldap_idmap_range_min = 30000\nldap_idmap_range_max = 20000\n")
	var report ReportData
	validateIDMapRanges(cfg, &report)

	found := false
	for _, f := range report.ConfigFindings {
		if f.Category == "idmap" && strings.Contains(f.Message, "invalid ID mapping range") {
			found = true
		}
	}
	if !found {
		t.Errorf("inverted idmap range was not detected")
	}
}

func TestValidateIDMapRanges_DisabledMappingIgnored(t *testing.T) {
	// With ldap_id_mapping = False the ranges are irrelevant: no findings.
	cfg := parseSssdConfig("[domain/a.example.com]\nid_provider = ad\nldap_id_mapping = False\nldap_idmap_range_min = 10000\nldap_idmap_range_max = 20000\n\n" +
		"[domain/b.example.com]\nid_provider = ad\nldap_id_mapping = False\nldap_idmap_range_min = 15000\nldap_idmap_range_max = 30000\n")
	var report ReportData
	validateIDMapRanges(cfg, &report)

	if len(report.ConfigFindings) != 0 {
		t.Errorf("disabled id_mapping should not produce findings: %+v", report.ConfigFindings)
	}
}

// TestValidateDomainStructure_MissingIDProvider mirrors SSSD's
// check_domain_id_provider: a domain without the mandatory id_provider is
// flagged as a configuration error (validated from src/util/sss_ini.c).
func TestValidateDomainStructure_MissingIDProvider(t *testing.T) {
	cfg := parseSssdConfig("[sssd]\nservices = nss, pam\n\n[domain/example.com]\nkrb5_realm = EXAMPLE.COM\n")
	var report ReportData
	validateDomainStructure(cfg, &report)

	found := false
	for _, f := range report.ConfigFindings {
		if strings.Contains(f.Message, "missing the mandatory 'id_provider'") {
			found = true
		}
	}
	if !found {
		t.Errorf("domain missing id_provider was not flagged")
	}
}

// TestValidateDomainStructure_InvalidProvider flags an unsupported provider.
func TestValidateDomainStructure_InvalidProvider(t *testing.T) {
	cfg := parseSssdConfig("[sssd]\n\n[domain/x]\nid_provider = bogus\n")
	var report ReportData
	validateDomainStructure(cfg, &report)

	found := false
	for _, f := range report.ConfigFindings {
		if strings.Contains(f.Message, "has invalid id_provider") {
			found = true
		}
	}
	if !found {
		t.Errorf("invalid id_provider value was not flagged")
	}
}

// TestValidateDomainStructure_InheritFrom mirrors SSSD's
// check_domain_inherit_from: inherit_from inside a [domain/*] section is not
// permitted.
func TestValidateDomainStructure_InheritFrom(t *testing.T) {
	cfg := parseSssdConfig("[domain/example.com]\nid_provider = ad\ninherit_from = default\n")
	var report ReportData
	validateDomainStructure(cfg, &report)

	found := false
	for _, f := range report.ConfigFindings {
		if strings.Contains(f.Message, "inherit_from") && strings.Contains(f.Message, "not permitted") {
			found = true
		}
	}
	if !found {
		t.Errorf("inherit_from inside a domain was not flagged")
	}
}
