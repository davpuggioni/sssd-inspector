// analyzer_correlate_test.go
//
// Regression tests for the pure helpers of the cross-source correlation
// engine (analyzer_correlate.go, 18% coverage) plus the false-positive
// guards of runCorrelation.
package main

import (
	"testing"
)

func TestCanonicalADDomain_Precedence(t *testing.T) {
	tests := []struct {
		name   string
		report ReportData
		want   string
	}{
		{"explicit ad_domain wins", ReportData{AdDomain: "EXAMPLE.com", SearchDomain: "other.com", Hostname: "h.other.com"}, "example.com"},
		{"leading dot stripped", ReportData{AdDomain: ".corp.example.com"}, "corp.example.com"},
		{"search domain fallback, first token", ReportData{SearchDomain: "corp.example.com other.example.com"}, "corp.example.com"},
		{"hostname suffix fallback", ReportData{Hostname: "host.corp.example.com"}, "corp.example.com"},
		{"search beats hostname", ReportData{SearchDomain: "a.example.com", Hostname: "host.b.example.com"}, "a.example.com"},
		{"short hostname yields nothing", ReportData{Hostname: "shorthost"}, ""},
		{"trailing dot hostname yields nothing", ReportData{Hostname: "host."}, ""},
		{"empty report yields nothing", ReportData{}, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := canonicalADDomain(&tc.report); got != tc.want {
				t.Errorf("canonicalADDomain() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFirstSearchToken(t *testing.T) {
	tests := []struct{ in, want string }{
		{"corp.example.com", "corp.example.com"},
		{"corp.example.com other.example.com", "corp.example.com"},
		{"  corp.example.com\tother", "corp.example.com"},
		{"", ""},
		{"   ", ""},
	}
	for _, tc := range tests {
		if got := firstSearchToken(tc.in); got != tc.want {
			t.Errorf("firstSearchToken(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestHostnameSuffix(t *testing.T) {
	tests := []struct{ in, want string }{
		{"host.corp.example.com", "corp.example.com"},
		{"host.example.com", "example.com"},
		{"shorthost", ""},
		{"host.", ""},
		{"", ""},
	}
	for _, tc := range tests {
		if got := hostnameSuffix(tc.in); got != tc.want {
			t.Errorf("hostnameSuffix(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestDomainsOverlap(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"corp.example.com", "CORP.EXAMPLE.COM", true},
		{"corp.example.com", "corp.example.com", true},
		{"sub.corp.example.com", "corp.example.com", true},
		{"corp.example.com", "sub.corp.example.com", true},
		{"corp.example.com", "other.example.com", false},
		{"corp.example.com", "notcorp.example.com", false},
		{"", "", true},
	}
	for _, tc := range tests {
		if got := domainsOverlap(tc.a, tc.b); got != tc.want {
			t.Errorf("domainsOverlap(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

// findingCategories returns the categories of all structured findings.
func findingCategories(r *ReportData) []string {
	var out []string
	for _, f := range r.ConfigFindings {
		out = append(out, f.Category)
	}
	return out
}

// TestRunCorrelation_NotADProvider_NoFindings: without an AD provider the
// engine must stay silent (no false positives on IPA/LDAP/file setups).
func TestRunCorrelation_NotADProvider_NoFindings(t *testing.T) {
	r := &ReportData{ADProviderMode: false, AdDomain: "corp.example.com"}
	runCorrelation(t.TempDir(), r)
	if len(r.ConfigFindings) != 0 {
		t.Errorf("expected no findings for a non-AD host, got %v", findingCategories(r))
	}
}

// TestRunCorrelation_WithoutDomainInfo_NoFindings: without any domain signal
// the engine must not guess (false-positive guard).
func TestRunCorrelation_WithoutDomainInfo_NoFindings(t *testing.T) {
	r := &ReportData{ADProviderMode: true}
	runCorrelation(t.TempDir(), r)
	if len(r.ConfigFindings) != 0 {
		t.Errorf("expected no findings without domain information, got %v", findingCategories(r))
	}
}
