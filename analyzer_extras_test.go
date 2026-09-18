// analyzer_extras_test.go
//
// Pin-down tests for small core helpers that had 0% or partial coverage:
// findingKeys (dedup key for config findings), categoryFor (log-error ->
// root-cause category) and chainable (sequence adjacency predicate).
package main

import (
	"testing"
)

func TestFindingKeys_Deduplicates(t *testing.T) {
	r := &ReportData{ConfigFindings: []ConfigFinding{
		{Severity: SevCritical, Category: "krb5_realm", Message: "realm mismatch"},
		{Severity: SevCritical, Category: "krb5_realm", Message: "realm mismatch"},
		{Severity: SevError, Category: "dns", Message: "search mismatch"},
	}}
	keys := findingKeys(r)
	if len(keys) != 2 {
		t.Fatalf("expected 2 deduped keys, got %v", keys)
	}
	if keys[0] != "krb5_realm|realm mismatch" || keys[1] != "dns|search mismatch" {
		t.Errorf("unexpected keys: %v", keys)
	}
	if got := findingKeys(&ReportData{}); len(got) != 0 {
		t.Errorf("empty report must yield no keys, got %v", got)
	}
}

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

func TestChainable(t *testing.T) {
	link := func(idx int, epoch int64, hasEpoch bool) seqLink {
		return seqLink{idx: idx, epoch: epoch, hasEpoch: hasEpoch}
	}
	if !chainable(link(0, 100, true), link(1, 103, true)) {
		t.Errorf("adjacent links inside the time gap must chain")
	}
	if chainable(link(0, 100, true), link(1, 100+sequenceClusterGapSeconds+1, true)) {
		t.Errorf("links beyond the time gap must not chain")
	}
	if chainable(link(1, 0, false), link(0, 0, false)) {
		t.Errorf("non-increasing index must not chain")
	}
	if !chainable(link(4, 0, false), link(5, 0, false)) {
		t.Errorf("consecutive indexes without epochs must chain")
	}
	if chainable(link(4, 0, false), link(6, 0, false)) {
		t.Errorf("non-consecutive indexes without epochs must not chain")
	}
	// Only one side has an epoch: falls back to index adjacency.
	if !chainable(link(7, 100, true), link(8, 0, false)) {
		t.Errorf("mixed epoch/no-epoch consecutive indexes must chain")
	}
}
