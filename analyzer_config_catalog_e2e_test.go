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

// TestCatalogOverrideChangesValidation is the decisive end-to-end property: a
// drop-in catalog must change what the analysis ACCEPTS, not merely what the
// Studio displays. The fixture uses ldap_uri, which the embedded catalog knows;
// an override that does not list it must turn that into an "unknown parameter"
// finding, and the report must name the override it validated against.
func TestCatalogOverrideChangesValidation(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"sssd.conf": "[sssd]\nservices = nss, pam\nldap_uri = ldap://dc01.example.com\n",
		"rpm.txt":   "sssd-2.9.4-150500.x86_64\n",
	})
	defer os.RemoveAll(dir)

	unknown := func(r ReportData, option string) bool {
		for _, f := range r.ConfigFindings {
			if strings.Contains(f.Message, "unknown parameter '"+option+"'") {
				return true
			}
		}
		return false
	}

	// Baseline: no override anywhere, so the embedded catalog decides.
	setUserDefinitionsRoot(t, t.TempDir())
	setSystemDefinitionsRoot(t, t.TempDir())
	baseline := analyzeData(dir, false, nil)
	if unknown(baseline, "ldap_uri") {
		t.Fatalf("the embedded catalog does not know ldap_uri; the fixture is wrong:\n%+v", baseline.ConfigFindings)
	}
	if baseline.CatalogProvenance == "" {
		t.Error("CatalogProvenance is empty: the report does not state what it was checked against")
	}

	// A minimal override that knows only "services": ldap_uri becomes unknown.
	userRoot := t.TempDir()
	setUserDefinitionsRoot(t, userRoot)
	writeCatalogOverride(t, userRoot, testCatalogJSON(t, "0.0.0-test", "services"))

	overridden := analyzeData(dir, false, nil)
	if !unknown(overridden, "ldap_uri") {
		t.Errorf("the override did not change validation: ldap_uri is unknown to it but was accepted.\nfindings: %+v", overridden.ConfigFindings)
	}
	if !strings.Contains(overridden.CatalogProvenance, "override in effect") {
		t.Errorf("CatalogProvenance = %q, want it to name the override in effect", overridden.CatalogProvenance)
	}
}

// TestBrokenCatalogOverrideReachesTheReport: a catalog that cannot be parsed
// must show up in the report diagnostics — the analysis keeps working against
// the embedded catalog, but the user is told their override was skipped.
func TestBrokenCatalogOverrideReachesTheReport(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"sssd.conf": "[sssd]\nservices = nss, pam\n",
		"rpm.txt":   "sssd-2.9.4-150500.x86_64\n",
	})
	defer os.RemoveAll(dir)

	userRoot := t.TempDir()
	setUserDefinitionsRoot(t, userRoot)
	setSystemDefinitionsRoot(t, t.TempDir())
	path := writeCatalogOverride(t, userRoot, "{ not json at all")

	report := analyzeData(dir, false, nil)
	for _, d := range report.Diagnostics {
		if d.File == path && strings.Contains(d.Message, "option catalog override skipped") {
			return
		}
	}
	t.Errorf("no override diagnostic reached the report; diagnostics: %+v", report.Diagnostics)
}
