// analyzer_kerberos_live_test.go
//
// Tests for the PRODUCTION Kerberos/keytab path.
//
// These used to call analyzeKerberosAndKeytab, which nothing reaches since the
// single-pass rewrite. The live split is:
//   - krb5.conf parsing  -> analyzeKerberosConfig   (analyzer_singlepass.go)
//   - keytab evidence    -> performSinglePassScan's Keytab* flags, folded into
//     report.Problems/Warnings by the scanner and by
//     analyzeData.
package main

import (
	"os"
	"testing"
)

// krb5Fixture returns a fixture dir whose etc.txt carries a krb5.conf
// section plus the given log lines in sssd.txt.
func krb5Fixture(t *testing.T, krb5Conf, sssdLog string) string {
	t.Helper()
	etcContent := "# /etc/krb5.conf\n" + krb5Conf + "#==[ command ]#==\n"
	dir := setupMockDir(t, map[string]string{
		"etc.txt":  etcContent,
		"sssd.txt": sssdLog,
	})
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

// scanKeytab runs the production scanner and mirrors how analyzeData folds the
// keytab flags into the report, so these tests assert on what a real analysis
// would emit rather than on the scanner struct alone.
func scanKeytab(t *testing.T, dir string) (*SinglePassResult, ReportData) {
	t.Helper()
	res := performSinglePassScan(dir, "AppArmor", nil)

	report := ReportData{KeytabFound: res.KeytabFound}
	report.Problems = append(report.Problems, res.Problems...)
	report.Warnings = append(report.Warnings, res.Warnings...)

	// analyzeData:124-133 — the missing-machine-account problem is raised only
	// when the logs carry no keytab evidence at all.
	if !res.KeytabFound && !res.KeytabNotFound && !res.NoKeytabPrincipal {
		report.Problems = append(report.Problems,
			"No Kerberos Keytab (Machine Account) principal found. AD join might be broken.")
	}
	return res, report
}

func TestKerberos_RC4LegacyWarning(t *testing.T) {
	dir := krb5Fixture(t, "[libdefaults]\n default_realm = CORP.EXAMPLE.COM\n default_tkt_enctypes = rc4-hmac\n", "KVNO Principal: host/host@CORP.EXAMPLE.COM\n")
	var report ReportData
	analyzeKerberosConfig(dir, &report)
	if !containsString(report.Warnings, "[SECURITY] Legacy 'rc4-hmac'") {
		t.Errorf("expected the rc4-hmac legacy warning, got %v", report.Warnings)
	}
}

func TestKerberos_DefaultRealmParsed(t *testing.T) {
	dir := krb5Fixture(t, "[libdefaults]\n default_realm = CORP.EXAMPLE.COM\n", "KVNO Principal: host/host@CORP.EXAMPLE.COM\n")
	var report ReportData
	analyzeKerberosConfig(dir, &report)
	if report.KerberosRealm != "CORP.EXAMPLE.COM" {
		t.Errorf("expected realm CORP.EXAMPLE.COM, got %q", report.KerberosRealm)
	}
}

func TestKerberos_DomainRealmMissingDot(t *testing.T) {
	dir := krb5Fixture(t, "[libdefaults]\n default_realm = CORP.EXAMPLE.COM\n[domain_realm]\n corp.example.com = CORP.EXAMPLE.COM\n", "KVNO Principal: host\n")
	var report ReportData
	analyzeKerberosConfig(dir, &report)
	if !containsString(report.Warnings, "lacks a leading dot") {
		t.Errorf("expected the [domain_realm] leading-dot warning, got %v", report.Warnings)
	}
}

func TestKerberos_KeytabFound(t *testing.T) {
	dir := krb5Fixture(t, "", "Dec 10 12:00:00 server sssd: KVNO Principal: host/srv@CORP.EXAMPLE.COM\n")
	_, report := scanKeytab(t, dir)
	if !report.KeytabFound {
		t.Errorf("expected KeytabFound=true on KVNO Principal evidence")
	}
	if containsString(report.Problems, "No Kerberos Keytab") {
		t.Errorf("must not report a missing keytab when the principal is present: %v", report.Problems)
	}
}

func TestKerberos_KeytabMissing(t *testing.T) {
	dir := krb5Fixture(t, "", "Dec 10 12:00:00 server sssd: ordinary line\n")
	_, report := scanKeytab(t, dir)
	if report.KeytabFound {
		t.Errorf("KeytabFound must be false without principal evidence")
	}
	if !containsString(report.Problems, "No Kerberos Keytab (Machine Account) principal found") {
		t.Errorf("expected the missing-keytab problem, got %v", report.Problems)
	}
}

func TestKerberos_KeytabFileMissing(t *testing.T) {
	dir := krb5Fixture(t, "", "Dec 10 12:00:00 server sssd: Key table file '/etc/krb5.keytab' not found\n")
	_, report := scanKeytab(t, dir)
	if !containsString(report.Problems, "/etc/krb5.keytab file is missing") {
		t.Errorf("expected the missing-keytab-file problem, got %v", report.Problems)
	}
}

// Remaining branch: stale keytab principal.
func TestKerberos_NoSuitablePrincipal(t *testing.T) {
	dir := krb5Fixture(t, "", "Dec 10 12:00:00 server sssd: No suitable principal found in keytab to authenticate as host\n")
	_, report := scanKeytab(t, dir)
	if !containsString(report.Problems, "No suitable principal found in keytab") {
		t.Errorf("expected the stale-keytab problem, got %v", report.Problems)
	}
	if !containsString(report.Warnings, "Run 'kvno HOSTNAME$'") {
		t.Errorf("expected the kvno diagnostic hint, got %v", report.Warnings)
	}
}
