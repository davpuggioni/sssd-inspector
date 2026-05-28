// analyzer_core_test.go
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	major, minor := getSSSDVersion(packages)

	if major != 2 || minor != 10 {
		t.Errorf("Expected version 2.10, but got %d.%d", major, minor)
	}
}

// Test: Verify Deep PII Redaction (IPv6, MAC, Emails)
func TestAnonymizeReport_DeepPII(t *testing.T) {
	report := ReportData{
		SearchDomain:  "suse.com",
		KerberosRealm: "SUSE.COM",
		SSSDLogErrors: []SSSDLogError{
			{
				Description: "Connection to AD failed for IPv6 2001:0db8:85a3:0000:0000:8a2e:0370:7334",
				Examples: []string{
					"Processing group user.name@suse.com",
					"Hardware fault at MAC address 00:1B:44:11:3A:B7",
				},
			},
		},
	}

	anonymizeReport(&report)

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
