// analyzer_auth_keytab_extra_test.go
//
// Remaining analyzeKerberosAndKeytab branch: stale keytab principal.
package main

import (
	"testing"
)

func TestKerberosAndKeytab_NoSuitablePrincipal(t *testing.T) {
	dir := krb5Fixture(t, "", "Dec 10 12:00:00 server sssd: No suitable principal found in keytab to authenticate as host\n")
	var report ReportData
	analyzeKerberosAndKeytab(dir, &report)
	if !containsString(report.Problems, "No suitable principal found in keytab") {
		t.Errorf("expected the stale-keytab problem, got %v", report.Problems)
	}
	if !containsString(report.Warnings, "Run 'kvno HOSTNAME$'") {
		t.Errorf("expected the kvno diagnostic hint, got %v", report.Warnings)
	}
}
