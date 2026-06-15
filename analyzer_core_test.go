// analyzer_core_test.go
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sssd-inspector/pkg/analysis"
	"sssd-inspector/pkg/types"
)

// Shared Helper: Checks if a specific string exists in a slice
func containsString(slice []string, expected string) bool {
	for _, item := range slice {
		if strings.Contains(item, expected) {
			return true
		}
	}
	return false
}

// Shared Helper: Sets up a mock temporary directory to test the streaming engine
func setupMockDir(t *testing.T, fileMap map[string]string) string {
	dir, err := os.MkdirTemp("", "test-sssd-*")
	if err != nil {
		t.Fatal(err)
	}
	for name, content := range fileMap {
		err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// Test: Verify that SSSD version extraction works correctly
func TestGetSSSDVersion(t *testing.T) {
	packages := []string{"sssd-2.10.2-150700.9.17.1.x86_64", "sssd-ad-2.10.2"}
	major, minor := analysis.GetSSSDVersion(packages)

	if major != 2 || minor != 10 {
		t.Errorf("Expected version 2.10, but got %d.%d", major, minor)
	}
}

// Test: Verify Deep PII Redaction (IPv6, MAC, Emails)
func TestAnonymizeReport_DeepPII(t *testing.T) {
	report := types.ReportData{
		SearchDomain:  "suse.com",
		KerberosRealm: "SUSE.COM",
		SSSDLogErrors: []types.SSSDLogError{
			{
				Description: "Connection to AD failed for IPv6 2001:0db8:85a3:0000:0000:8a2e:0370:7334",
				Examples: []string{
					"Processing group user.name@suse.com",
					"Hardware fault at MAC address 00:1B:44:11:3A:B7",
				},
			},
		},
	}

	analysis.AnonymizeReport(&report)

	if strings.Contains(report.SSSDLogErrors[0].Description, "2001:0db8:85a3") {
		t.Errorf("Failed to redact IPv6 address")
	}
	if strings.Contains(report.SSSDLogErrors[0].Examples[0], "user.name@suse.com") {
		t.Errorf("Failed to redact UPN/Email address")
	}
	if strings.Contains(report.SSSDLogErrors[0].Examples[1], "00:1B:44:11:3A:B7") {
		t.Errorf("Failed to redact MAC address")
	}
	if !strings.Contains(report.SSSDLogErrors[0].Examples[0], "example.com") {
		t.Errorf("Failed to mask internal domain to example.com")
	}
}

// Test: Verify sssd.conf config snippet gets domain redacted as well,
// including domains from "domains =", "ad_domain =", and "ldap_search_base ="
// that may differ from SearchDomain and KerberosRealm.
func TestAnonymizeReport_ConfigSnippetDomain(t *testing.T) {
	report := types.ReportData{
		SearchDomain:  "company.com",
		KerberosRealm: "COMPANY.COM",
		SSSDConfigSnippet: `[sssd]
domains = sub.corp.company.com
services = nss, pam

[domain/sub.corp.company.com]
ad_domain = sub.corp.company.com
ldap_search_base = dc=sub,dc=corp,dc=company,dc=com
krb5_realm = CORP.COMPANY.COM
`,
		Problems: []string{
			"Using ldap_search_base dc=sub,dc=corp,dc=company,dc=com",
			"Domain sub.corp.company.com",
		},
	}

	analysis.AnonymizeReport(&report)

	// The domains "sub.corp.company.com" and "CORP.COMPANY.COM" from sssd.conf
	// should be redacted in the config snippet. The raw "dc=..." characters
	// inside ldap_search_base are not a domain match (they lack dots).
	if strings.Contains(report.SSSDConfigSnippet, "sub.corp.company.com") {
		t.Errorf("SSSDConfigSnippet still contains unredacted 'sub.corp.company.com': %q", report.SSSDConfigSnippet)
	}
	if strings.Contains(report.SSSDConfigSnippet, "CORP.COMPANY") {
		t.Errorf("SSSDConfigSnippet still contains unredacted 'CORP.COMPANY': %q", report.SSSDConfigSnippet)
	}
	// Verify all domains were redacted — check no original domain values remain
	if strings.Contains(report.SSSDConfigSnippet, "company.com") || strings.Contains(report.SSSDConfigSnippet, "COMPANY") {
		t.Errorf("SSSDConfigSnippet still contains unredacted domain: %q", report.SSSDConfigSnippet)
	}
	// Verify the values were replaced with example.com/EXAMPLE.COM
	if !strings.Contains(report.SSSDConfigSnippet, "example.com") {
		t.Errorf("SSSDConfigSnippet should contain redacted domain 'example.com', got: %q", report.SSSDConfigSnippet)
	}
	if !strings.Contains(report.SSSDConfigSnippet, "EXAMPLE.COM") {
		t.Errorf("SSSDConfigSnippet should have 'CORP.COMPANY.COM' redacted to 'EXAMPLE.COM', got: %q", report.SSSDConfigSnippet)
	}
	for _, p := range report.Problems {
		if strings.Contains(p, "sub.corp.company.com") {
			t.Errorf("Problem still contains unredacted domain: %q", p)
		}
	}
}

// Test: Verify mixed-case domain redaction (e.g., "Corp.Example.Com")
// This tests the case-insensitive regex replacement for domains
func TestAnonymizeReport_MixedCaseDomain(t *testing.T) {
	// Test with domain containing mixed case like "Corp.Suse.Com"
	report := types.ReportData{
		SearchDomain:  "corp.suse.com",
		KerberosRealm: "SUSE.COM",
		Problems: []string{
			"In Corp.Suse.Com domain, Kerberos ticket expired",
			"In CORP.SUSE.COM domain, LDAP connection failed",
			"In corp.suse.com domain, everything is fine",
			"Kerberos realm SUSE.COM is unreachable",
			"Kerberos realm suse.com is unreachable",
			"Kerberos realm SuSe.CoM is unreachable",
		},
	}

	analysis.AnonymizeReport(&report)

	// All variants should now be redacted to "example.com" / "EXAMPLE.COM"
	for _, p := range report.Problems {
		if strings.Contains(p, "suse") || strings.Contains(p, "SUSE") || strings.Contains(p, "SuSe") || strings.Contains(p, "CoM") {
			t.Errorf("Mixed-case domain not fully redacted: got %q", p)
		}
	}
	if !strings.Contains(report.Problems[0], "example.com") {
		t.Errorf("Mixed-case 'Corp.Suse.Com' should be redacted to example.com, got: %q", report.Problems[0])
	}
	if !strings.Contains(report.Problems[3], "EXAMPLE.COM") {
		t.Errorf("Kerberos realm should be redacted to EXAMPLE.COM, got: %q", report.Problems[3])
	}
}
