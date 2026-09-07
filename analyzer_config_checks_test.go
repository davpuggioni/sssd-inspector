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
	cfg := parseSssdConfig("[domain/a.example.com]\nid_provider = ad\nldap_idmap_min_id = 10000\nldap_idmap_max_id = 20000\n\n" +
		"[domain/b.example.com]\nid_provider = ad\nldap_idmap_min_id = 15000\nldap_idmap_max_id = 30000\n")
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

func TestValidateIDMapRanges_InvertedRange(t *testing.T) {
	cfg := parseSssdConfig("[domain/example.com]\nid_provider = ad\nldap_idmap_min_id = 30000\nldap_idmap_max_id = 20000\n")
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
	cfg := parseSssdConfig("[domain/a.example.com]\nid_provider = ad\nldap_id_mapping = False\nldap_idmap_min_id = 10000\nldap_idmap_max_id = 20000\n\n" +
		"[domain/b.example.com]\nid_provider = ad\nldap_id_mapping = False\nldap_idmap_min_id = 15000\nldap_idmap_max_id = 30000\n")
	var report ReportData
	validateIDMapRanges(cfg, &report)

	if len(report.ConfigFindings) != 0 {
		t.Errorf("disabled id_mapping should not produce findings: %+v", report.ConfigFindings)
	}
}
