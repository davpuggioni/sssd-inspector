// ahocorasick_test.go - Tests for the Aho-Corasick multi-pattern matcher.

package main

import (
	"sort"
	"testing"
)

// collectMatches runs the automaton over text and returns the matched
// (deduplicated) pattern strings.
func collectPatterns(t *testing.T, patterns []string, text string) []string {
	t.Helper()
	m := newACMatcher(patterns)
	seen := make(map[int]bool)
	var hits []string
	m.Match(text, func(idx int) {
		if !seen[idx] {
			seen[idx] = true
			hits = append(hits, m.patterns[idx])
		}
	})
	sort.Strings(hits)
	return hits
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestACSimpleMatches(t *testing.T) {
	patterns := []string{"sssd", "krb5", "keytab"}
	m := newACMatcher(patterns)
	if m.Len() != 3 {
		t.Fatalf("expected 3 patterns, got %d", m.Len())
	}

	got := collectPatterns(t, patterns, "sssd_be[krb5] failed to read /etc/krb5.keytab")
	want := []string{"keytab", "krb5", "sssd"}
	if !equalStrings(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	if got := collectPatterns(t, patterns, "nothing relevant here"); len(got) != 0 {
		t.Errorf("expected no matches, got %v", got)
	}
}

func TestACOverlappingPatterns(t *testing.T) {
	// "service key not available" overlaps with "not available" and "key not".
	patterns := []string{"service key not available", "not available", "key not"}
	got := collectPatterns(t, patterns, "GSSAPI Error: service key not available for host/x")
	want := []string{"key not", "not available", "service key not available"}
	if !equalStrings(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestACSuffixChain(t *testing.T) {
	// Classic Aho-Corasick failure-link case: "she", "he", "hers", "his".
	patterns := []string{"she", "he", "hers", "his"}
	got := collectPatterns(t, patterns, "ushers")
	want := []string{"he", "hers", "she"}
	if !equalStrings(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestACCaseInsensitive(t *testing.T) {
	patterns := []string{"TERMINATED BY OWN WATCHDOG"}
	got := collectPatterns(t, patterns, "Child process terminated by own watchdog, killing")
	if len(got) != 1 || got[0] != "terminated by own watchdog" {
		t.Errorf("case-insensitive match failed, got %v", got)
	}
}

func TestACMultipleOccurrences(t *testing.T) {
	m := newACMatcher([]string{"sssd"})
	count := 0
	m.Match("sssd and sssd and sssd", func(int) { count++ })
	if count != 3 {
		t.Errorf("expected 3 invocations, got %d", count)
	}
}

func TestACDuplicatePatternsDeduplicated(t *testing.T) {
	// The scanner can register the same literal twice (error + quick list);
	// it must map to one canonical index so handlers fire once per line.
	m := newACMatcher([]string{"tgt failed", "TGT failed", "other"})
	if m.Len() != 2 {
		t.Fatalf("expected dedup to 2 patterns, got %d", m.Len())
	}
	count := 0
	m.Match("TGT failed verification", func(int) { count++ })
	if count != 1 {
		t.Errorf("expected 1 invocation after dedup, got %d", count)
	}
}

func TestACEmptyAndNoPatterns(t *testing.T) {
	m := newACMatcher([]string{"", ""})
	if m.Len() != 0 {
		t.Errorf("expected 0 patterns, got %d", m.Len())
	}
	m.Match("any text", func(int) { t.Error("no patterns should match") })

	empty := newACMatcher(nil)
	empty.Match("whatever", func(int) { t.Error("no patterns should match") })
}

func TestACCacheReturnsSameInstance(t *testing.T) {
	pats := []string{"one", "two"}
	a := globalACCache.Get(pats)
	b := globalACCache.Get([]string{"one", "two"})
	if a != b {
		t.Error("ACCache.Get must return the same compiled automaton for identical pattern sets")
	}
}
