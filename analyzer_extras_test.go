// analyzer_extras_test.go
//
// Pin-down tests for small core helpers: categoryFor (log-error -> root-cause
// category), which must stay deterministic.
package main

import (
	"testing"
)

func TestCategoryFor_ExtraFallbacks(t *testing.T) {
	// Cases beyond analysis_sequences_test.go: lowercase free-text
	// fallbacks, the remaining domains, and the "general" default.
	tests := []struct{ in, want string }{
		{"weird kerberos preauth failure", "krb5"},
		{"ldap size limit exceeded", "ldap"},
		{"cache database on disk full", "db"},
		{"access denied for user", "access"},
		{"rc4 crypto negotiation", "crypto"},
		{"something utterly unexpected", "general"},
	}
	for _, tc := range tests {
		if got := categoryFor(tc.in); got != tc.want {
			t.Errorf("categoryFor(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
