// analyzer_correlate_checks_test.go
//
// runCorrelation contradiction checks: the four findings (join, krb5_realm,
// dns, hostname) and the no-false-positive anchor for a consistent setup.
package main

import (
	"testing"
)

// TestRunCorrelation_MissingRealm_CriticalJoin pins the "not joined"
// critical: AD provider configured but krb5.conf has no default_realm.
func TestRunCorrelation_MissingRealm_CriticalJoin(t *testing.T) {
	r := &ReportData{ADProviderMode: true, AdDomain: "corp.example.com", KerberosRealm: "Not configured"}
	runCorrelation(t.TempDir(), r)
	cats := findingCategories(r)
	if !containsString(cats, "join") {
		t.Errorf("expected a critical 'join' finding, got %v", cats)
	}
	for _, f := range r.ConfigFindings {
		if f.Category == "join" && f.Severity != SevCritical {
			t.Errorf("join finding must be SevCritical, got %v", f.Severity)
		}
	}
}

// TestRunCorrelation_RealmMismatch_Critical pins the realm contradiction.
func TestRunCorrelation_RealmMismatch_Critical(t *testing.T) {
	r := &ReportData{ADProviderMode: true, AdDomain: "corp.example.com", KerberosRealm: "OTHER.EXAMPLE.COM", SearchDomain: "corp.example.com", Hostname: "host.corp.example.com"}
	runCorrelation(t.TempDir(), r)
	if !containsString(findingCategories(r), "krb5_realm") {
		t.Errorf("expected a 'krb5_realm' finding, got %v", findingCategories(r))
	}
}

// TestRunCorrelation_RealmSubdomain_OK documents the intended tolerance:
// realm equal to a parent domain of the AD domain is not a contradiction.
func TestRunCorrelation_RealmSubdomain_OK(t *testing.T) {
	r := &ReportData{ADProviderMode: true, AdDomain: "sub.corp.example.com", KerberosRealm: "CORP.EXAMPLE.COM", SearchDomain: "sub.corp.example.com", Hostname: "host.sub.corp.example.com"}
	runCorrelation(t.TempDir(), r)
	if containsString(findingCategories(r), "krb5_realm") {
		t.Errorf("subdomain/parent-domain realm relation must not raise krb5_realm, got %v", findingCategories(r))
	}
}

// TestRunCorrelation_SearchDomainMismatch_Error pins the DNS guard.
func TestRunCorrelation_SearchDomainMismatch_Error(t *testing.T) {
	r := &ReportData{ADProviderMode: true, AdDomain: "corp.example.com", KerberosRealm: "CORP.EXAMPLE.COM", SearchDomain: "totally-unrelated.net", Hostname: "host.corp.example.com"}
	runCorrelation(t.TempDir(), r)
	if !containsString(findingCategories(r), "dns") {
		t.Errorf("expected a 'dns' finding, got %v", findingCategories(r))
	}
}

// TestRunCorrelation_HostnameOutsideDomain_Error pins the hostname guard.
func TestRunCorrelation_HostnameOutsideDomain_Error(t *testing.T) {
	r := &ReportData{ADProviderMode: true, AdDomain: "corp.example.com", KerberosRealm: "CORP.EXAMPLE.COM", SearchDomain: "corp.example.com", Hostname: "host.wrong-domain.net"}
	runCorrelation(t.TempDir(), r)
	if !containsString(findingCategories(r), "hostname") {
		t.Errorf("expected a 'hostname' finding, got %v", findingCategories(r))
	}
}

// TestRunCorrelation_ConsistentSetup_NoFindings is the no-false-positive
// anchor: an aligned setup must produce zero findings.
func TestRunCorrelation_ConsistentSetup_NoFindings(t *testing.T) {
	r := &ReportData{ADProviderMode: true, AdDomain: "corp.example.com", KerberosRealm: "CORP.EXAMPLE.COM", SearchDomain: "corp.example.com", Hostname: "host.corp.example.com"}
	runCorrelation(t.TempDir(), r)
	if len(r.ConfigFindings) != 0 {
		t.Errorf("consistent setup must produce no findings, got %v", findingCategories(r))
	}
}
