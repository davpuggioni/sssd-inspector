package analysis

import (
	"strings"
	"testing"

	"sssd-inspector/pkg/types"
)

// TestDeduplicateProblems tests the DeduplicateProblems function
func TestDeduplicateProblems(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected int
	}{
		{"empty slice", nil, 0},
		{"no duplicates", []string{"a", "b", "c"}, 3},
		{"with duplicates", []string{"a", "b", "a", "c", "b"}, 3},
		{"all duplicates", []string{"x", "x", "x"}, 1},
		{"single element", []string{"only"}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeduplicateProblems(tt.input)
			if len(got) != tt.expected {
				t.Errorf("DeduplicateProblems(%v) returned %d elements, want %d", tt.input, len(got), tt.expected)
			}
		})
	}
}

// TestDeduplicateProblems_Order verifies that order is preserved
func TestDeduplicateProblems_Order(t *testing.T) {
	input := []string{"first", "second", "first", "third", "second"}
	got := DeduplicateProblems(input)

	expected := []string{"first", "second", "third"}
	for i, v := range expected {
		if got[i] != v {
			t.Errorf("expected element %d to be %q, got %q", i, v, got[i])
		}
	}
}

// TestGetSSSDVersion tests version extraction from package strings
func TestGetSSSDVersion(t *testing.T) {
	tests := []struct {
		name       string
		packages   []string
		wantMajor  int
		wantMinor  int
		expectZero bool
	}{
		{"empty packages", nil, 0, 0, true},
		{"no sssd package", []string{"openssl-1.1.1"}, 0, 0, true},
		{"sssd-2.10.2", []string{"sssd-2.10.2-150700.9.17.1.x86_64"}, 2, 10, false},
		{"sssd 1.16.5", []string{"sssd 1.16.5-150000.1.3.1.x86_64"}, 1, 16, false},
		{"multiple packages", []string{"sssd-client-2.9.0", "sssd-2.9.0-1.x86_64"}, 2, 9, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			major, minor := GetSSSDVersion(tt.packages)
			if tt.expectZero {
				if major != 0 || minor != 0 {
					t.Errorf("GetSSSDVersion(%v) = (%d, %d), want (0, 0)", tt.packages, major, minor)
				}
				return
			}
			if major != tt.wantMajor || minor != tt.wantMinor {
				t.Errorf("GetSSSDVersion(%v) = (%d, %d), want (%d, %d)", tt.packages, major, minor, tt.wantMajor, tt.wantMinor)
			}
		})
	}
}

// TestAnonymizeReport_IPAddresses verifies IP address redaction
func TestAnonymizeReport_IPAddresses(t *testing.T) {
	report := &types.ReportData{
		Problems: []string{"Server 192.168.1.100 is unreachable"},
		Warnings: []string{"Check DNS on 10.0.0.1"},
	}

	AnonymizeReport(report)

	for _, p := range report.Problems {
		if strings.Contains(p, "192.168.1.100") {
			t.Errorf("IP address was not redacted in: %s", p)
		}
		if !strings.Contains(p, "XXX.XXX.XXX.XXX") {
			t.Errorf("Expected XXX.XXX.XXX.XXX in: %s", p)
		}
	}
	for _, w := range report.Warnings {
		if strings.Contains(w, "10.0.0.1") {
			t.Errorf("IP address was not redacted in: %s", w)
		}
	}
}

// TestAnonymizeReport_MacAddresses verifies MAC address redaction
func TestAnonymizeReport_MacAddresses(t *testing.T) {
	report := &types.ReportData{
		Problems: []string{"MAC: 00:11:22:33:44:55"},
	}

	AnonymizeReport(report)

	if strings.Contains(report.Problems[0], "00:11:22:33:44:55") {
		t.Errorf("MAC address was not redacted")
	}
	if !strings.Contains(report.Problems[0], "XX:XX:XX:XX:XX:XX") {
		t.Errorf("Expected XX:XX:XX:XX:XX:XX in: %s", report.Problems[0])
	}
}

// TestAnonymizeReport_SensitiveFields verifies redaction of sensitive metadata
func TestAnonymizeReport_SensitiveFields(t *testing.T) {
	report := &types.ReportData{
		HardwareManufacturer: "Dell Inc.",
		HardwareModel:        "PowerEdge R740",
		VirtualIdentity:      "VMware Virtual Platform",
		SCCStatus:            "Registered",
	}

	AnonymizeReport(report)

	if report.HardwareManufacturer != "[REDACTED]" {
		t.Errorf("HardwareManufacturer should be REDACTED, got %q", report.HardwareManufacturer)
	}
	if report.HardwareModel != "[REDACTED]" {
		t.Errorf("HardwareModel should be REDACTED, got %q", report.HardwareModel)
	}
	if report.VirtualIdentity != "[REDACTED]" {
		t.Errorf("VirtualIdentity should be REDACTED, got %q", report.VirtualIdentity)
	}
	if report.SCCStatus != "[REDACTED]" {
		t.Errorf("SCCStatus should be REDACTED, got %q", report.SCCStatus)
	}
}

// TestAnonymizeReport_Email verifies email redaction
func TestAnonymizeReport_Email(t *testing.T) {
	report := &types.ReportData{
		Problems: []string{"Contact admin@example.com for support"},
	}

	AnonymizeReport(report)

	if strings.Contains(report.Problems[0], "admin@example.com") {
		t.Errorf("email was not redacted")
	}
	if !strings.Contains(report.Problems[0], "[REDACTED_USER]@example.com") {
		t.Errorf("Expected REDACTED_USER in: %s", report.Problems[0])
	}
}

// TestAnonymizeReport_Timeline verifies timeline event sanitization
func TestAnonymizeReport_Timeline(t *testing.T) {
	report := &types.ReportData{
		Timeline: []types.TimelineEvent{
			{Message: "Error connecting to 192.168.1.1", RawLog: "from 10.0.0.1"},
		},
	}

	AnonymizeReport(report)

	if strings.Contains(report.Timeline[0].Message, "192.168.1.1") {
		t.Errorf("IP in Timeline message was not redacted")
	}
	if strings.Contains(report.Timeline[0].RawLog, "10.0.0.1") {
		t.Errorf("IP in Timeline raw log was not redacted")
	}
}

// TestAnonymizeReport_ConfigSnippet tests domain redaction in config
func TestAnonymizeReport_ConfigSnippet(t *testing.T) {
	report := &types.ReportData{
		SearchDomain: "corp.example.com",
		SSSDConfigSnippet: `[domain/corp.example.com]
ad_domain = corp.example.com
`,
	}

	AnonymizeReport(report)

	if strings.Contains(report.SSSDConfigSnippet, "corp.example.com") {
		t.Errorf("domain in config snippet was not redacted")
	}
}

// TestAnonymizeReport_MixedCaseDomain tests case-sensitive domain replacement
func TestAnonymizeReport_MixedCaseDomain(t *testing.T) {
	report := &types.ReportData{
		SearchDomain:  "CORP.COMPANY.COM",
		KerberosRealm: "CORP.COMPANY.COM",
	}

	AnonymizeReport(report)

	if report.SearchDomain != "EXAMPLE.COM" {
		t.Errorf("SearchDomain should be EXAMPLE.COM, got %q", report.SearchDomain)
	}
}
