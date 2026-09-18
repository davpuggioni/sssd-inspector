// analyzer_auth_keytab_test.go
//
// Tests for analyzeKerberosAndKeytab (analyzer_auth.go, 0% coverage): the
// krb5.conf checks plus the four keytab evidence branches.
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

func TestKerberosAndKeytab_RC4LegacyWarning(t *testing.T) {
	dir := krb5Fixture(t, "[libdefaults]\n default_realm = CORP.EXAMPLE.COM\n default_tkt_enctypes = rc4-hmac\n", "KVNO Principal: host/host@CORP.EXAMPLE.COM\n")
	var report ReportData
	analyzeKerberosAndKeytab(dir, &report)
	if !containsString(report.Warnings, "[SECURITY] Legacy 'rc4-hmac'") {
		t.Errorf("expected the rc4-hmac legacy warning, got %v", report.Warnings)
	}
}

func TestKerberosAndKeytab_DefaultRealmParsed(t *testing.T) {
	dir := krb5Fixture(t, "[libdefaults]\n default_realm = CORP.EXAMPLE.COM\n", "KVNO Principal: host/host@CORP.EXAMPLE.COM\n")
	var report ReportData
	analyzeKerberosAndKeytab(dir, &report)
	if report.KerberosRealm != "CORP.EXAMPLE.COM" {
		t.Errorf("expected realm CORP.EXAMPLE.COM, got %q", report.KerberosRealm)
	}
}

func TestKerberosAndKeytab_DomainRealmMissingDot(t *testing.T) {
	dir := krb5Fixture(t, "[libdefaults]\n default_realm = CORP.EXAMPLE.COM\n[domain_realm]\n corp.example.com = CORP.EXAMPLE.COM\n", "KVNO Principal: host\n")
	var report ReportData
	analyzeKerberosAndKeytab(dir, &report)
	if !containsString(report.Warnings, "lacks a leading dot") {
		t.Errorf("expected the [domain_realm] leading-dot warning, got %v", report.Warnings)
	}
}

func TestKerberosAndKeytab_KeytabFound(t *testing.T) {
	dir := krb5Fixture(t, "", "Dec 10 12:00:00 server sssd: KVNO Principal: host/srv@CORP.EXAMPLE.COM\n")
	var report ReportData
	analyzeKerberosAndKeytab(dir, &report)
	if !report.KeytabFound {
		t.Errorf("expected KeytabFound=true on KVNO Principal evidence")
	}
	if containsString(report.Problems, "No Kerberos Keytab") {
		t.Errorf("must not report a missing keytab when the principal is present: %v", report.Problems)
	}
}

func TestKerberosAndKeytab_KeytabMissing(t *testing.T) {
	dir := krb5Fixture(t, "", "Dec 10 12:00:00 server sssd: ordinary line\n")
	var report ReportData
	analyzeKerberosAndKeytab(dir, &report)
	if report.KeytabFound {
		t.Errorf("KeytabFound must be false without principal evidence")
	}
	if !containsString(report.Problems, "No Kerberos Keytab (Machine Account) principal found") {
		t.Errorf("expected the missing-keytab problem, got %v", report.Problems)
	}
}

func TestKerberosAndKeytab_KeytabFileMissing(t *testing.T) {
	dir := krb5Fixture(t, "", "Dec 10 12:00:00 server sssd: Key table file '/etc/krb5.keytab' not found\n")
	var report ReportData
	analyzeKerberosAndKeytab(dir, &report)
	if !containsString(report.Problems, "/etc/krb5.keytab file is missing") {
		t.Errorf("expected the missing-keytab-file problem, got %v", report.Problems)
	}
}
