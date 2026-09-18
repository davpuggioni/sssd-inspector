// kb_misc_test.go
//
// Small helpers around KB handling: truncateText (kb.go, 0%) and the
// Aho-Corasick cache Clear (ahocorasick.go, 0%).
package main

import (
	"testing"
)

func TestTruncateText(t *testing.T) {
	if got := truncateText("short", 10); got != "short" {
		t.Errorf("short string must pass through: %q", got)
	}
	if got := truncateText("exactly-ten!", 10); got != "exactly-te…" {
		t.Errorf("over-long string must truncate: %q", got)
	}
	if got := truncateText("this is too long", 7); got != "this is…" {
		t.Errorf("unexpected truncation: %q", got)
	}
	// Rune-safe: multi-byte characters must not be split.
	if got := truncateText("héllo wörld", 5); got != "héllo…" {
		t.Errorf("rune-unsafe truncation: %q", got)
	}
}

func TestACCache_ClearForcesRebuild(t *testing.T) {
	c := NewACCache()
	before := c.Get([]string{"alpha", "beta"})
	c.Clear()
	after := c.Get([]string{"alpha", "beta"})
	if before == after {
		t.Errorf("Clear must drop cached automatons")
	}
	var hits []int
	after.Match("xx beta yy", func(idx int) { hits = append(hits, idx) })
	if len(hits) != 1 || hits[0] != 1 {
		t.Errorf("rebuilt automaton must match 'beta' at index 1, got %v", hits)
	}
}

func TestCatOf_SubstringFallbacks(t *testing.T) {
	// catOf adds extra substring fallbacks beyond categoryFor: pin them.
	tests := []struct{ in, want string }{
		{"ssl connection reset", "tls"},
		{"kerberos ticket expired", "krb5"},
		{"provider went offline", "offline"},
		{"totally unrelated line", "general"},
	}
	for _, tc := range tests {
		if got := catOf(tc.in); got != tc.want {
			t.Errorf("catOf(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
