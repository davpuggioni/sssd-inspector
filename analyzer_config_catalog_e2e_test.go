// analyzer_config_catalog_e2e_test.go
//
// End-to-end guard for the catalog-backed sssd.conf validation: a realistic
// configuration containing a typo, an out-of-enum value, a wrong type and a
// valid (previously mis-flagged) GPO mode must surface the right findings, at
// the right severity, in the final report.
package main

import (
	"os"
	"strings"
	"testing"
)

// catalogE2EFixture writes a supportconfig-like directory whose sssd.conf
// mixes four deliberate problems with correct settings.
func catalogE2EFixture(t *testing.T) string {
	t.Helper()
	conf := "[sssd]\n" +
		"services = nss, pam\n" +
		"domains = example.com\n" +
		"\n[domain/example.com]\n" +
		"id_provider = ad\n" +
		"ldap_uri = ldap://dc01.example.com\n" +
		"ldap_url = ldap://dc02.example.com\n" + // typo of ldap_uri
		"ldap_schema = rfc2309\n" + // not a documented schema
		"enumerate = maybe\n" + // not a boolean
		"ad_gpo_access_control = disabled\n" // valid since SSSD 1.14
	return setupMockDir(t, map[string]string{
		"sssd.conf": conf,
		"sssd.txt":  "sssd[be[example.com]]: Starting up\n",
		"rpm.txt":   "sssd-2.9.4-150500.x86_64\n",
	})
}

func TestAnalyzeData_SurfacesCatalogFindings(t *testing.T) {
	dir := catalogE2EFixture(t)
	defer os.RemoveAll(dir)

	report := analyzeData(dir, false, nil)
	if !report.SssdConfigFound {
		t.Fatal("sssd.conf was not picked up by the analyzer")
	}

	warnings := strings.Join(report.Warnings, "\n")
	for _, want := range []string{"ldap_url", "ldap_uri", "ldap_schema", "boolean"} {
		if !strings.Contains(warnings, want) {
			t.Errorf("report warnings are missing %q; got:\n%s", want, warnings)
		}
	}

	// The whole point of the catalog is to turn "silently ignored" settings
	// into warnings, never into hard errors.
	for _, p := range report.Problems {
		for _, bad := range []string{"ldap_url", "ldap_schema", "enumerate", "ad_gpo_access_control"} {
			if strings.Contains(p, bad) {
				t.Errorf("%q must be a warning, not a problem: %s", bad, p)
			}
		}
	}
}

// TestAnalyzeData_AcceptsGpoAccessControlDisabled is the end-to-end twin of
// TestValidateADSections_GpoAccessControlDisabled: the value must produce no
// finding at all in the final report.
func TestAnalyzeData_AcceptsGpoAccessControlDisabled(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"sssd.conf": "[sssd]\nservices = nss, pam\n\n[domain/example.com]\nid_provider = ad\nad_gpo_access_control = disabled\n",
		"rpm.txt":   "sssd-2.9.4-150500.x86_64\n",
	})
	defer os.RemoveAll(dir)

	report := analyzeData(dir, false, nil)
	for _, f := range report.ConfigFindings {
		if strings.Contains(f.Message, "ad_gpo_access_control") {
			t.Errorf("ad_gpo_access_control = disabled produced a finding: %s", f.Message)
		}
	}
}
