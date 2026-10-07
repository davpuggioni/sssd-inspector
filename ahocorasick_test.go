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

// TestACCanonicalIndexMapsDuplicates pins the contract callers rely on:
// Match reports CANONICAL indices, so CanonicalIndex is the only safe way to
// map a registered literal back to its metadata. It must resolve duplicates
// and case variants to the SAME index, and report unknown literals as absent.
func TestACCanonicalIndexMapsDuplicates(t *testing.T) {
	m := newACMatcher([]string{"alpha error", "beta error", "alpha error", "gamma error"})
	if m.Len() != 3 {
		t.Fatalf("expected 3 unique patterns, got %d", m.Len())
	}

	first, ok := m.CanonicalIndex("alpha error")
	if !ok {
		t.Fatal("CanonicalIndex must resolve a registered literal")
	}
	dup, ok := m.CanonicalIndex("alpha error")
	if !ok || dup != first {
		t.Errorf("duplicate literal must map to the same canonical index: got %d, want %d", dup, first)
	}
	upper, ok := m.CanonicalIndex("ALPHA ERROR")
	if !ok || upper != first {
		t.Errorf("case variant must map to the same canonical index: got %d, want %d", upper, first)
	}
	if _, ok := m.CanonicalIndex("never registered"); ok {
		t.Error("CanonicalIndex must report an unregistered literal as absent")
	}

	// The regression this protects: indexing a parallel metadata table with
	// the canonical index mis-attributes everything after the first duplicate.
	var got []string
	m.Match("gamma error happened", func(idx int) { got = append(got, m.patterns[idx]) })
	if len(got) != 1 || got[0] != "gamma error" {
		t.Errorf("Match must report the matched literal, got %v", got)
	}
}

// ---------------------------------------------------------------------------
// Cache-key regressions (forensic correctness)
// ---------------------------------------------------------------------------

// The old FNV-1a fingerprint was sensitive to slice ORDER: the same pattern
// set passed in two different orders produced two different keys, compiling
// duplicate automatons. Worse, a 64-bit collision would hand back an automaton
// built from a DIFFERENT pattern set, silently attributing matches to the
// wrong rule metadata. The key must be order- and case-independent, and
// distinct sets must never alias.
func TestACFingerprintOrderAndCaseIndependent(t *testing.T) {
	base := []string{"b error", "a error", "c error"}
	sameOrder := acFingerprint(base)
	reversed := acFingerprint([]string{"c error", "a error", "b error"})
	if sameOrder != reversed {
		t.Errorf("fingerprint must not depend on slice order: %s != %s", sameOrder, reversed)
	}
	cased := acFingerprint([]string{"B ERROR", "A error", "C Error"})
	if cased != sameOrder {
		t.Errorf("fingerprint must fold case: %s != %s", cased, sameOrder)
	}
}

func TestACFingerprintDistinguishesDifferentSets(t *testing.T) {
	a := acFingerprint([]string{"ldap error"})
	b := acFingerprint([]string{"ldap errors"})
	if a == b {
		t.Errorf("different pattern sets must never share a cache key (would return the wrong automaton)")
	}
	// A short pattern that is a prefix of a longer one must not collide either.
	c := acFingerprint([]string{"krb5"})
	d := acFingerprint([]string{"krb5", "krb5_keytab"})
	if c == d {
		t.Errorf("prefix pattern sets must not collide")
	}
}

func TestACCacheOrderIndependentSameInstance(t *testing.T) {
	c := NewACCache()
	a := c.Get([]string{"one", "two", "three"})
	b := c.Get([]string{"three", "one", "two"})
	if a != b {
		t.Error("the same pattern set in a different order must reuse one compiled automaton")
	}
	d := c.Get([]string{"ONE", "TWO", "THREE"})
	if d != a {
		t.Error("case variants of the same pattern set must reuse one compiled automaton")
	}
	other := c.Get([]string{"one", "two", "four"})
	if other == a {
		t.Error("a different pattern set must compile its own automaton, never alias a cached one")
	}
}

// ---------------------------------------------------------------------------
// Word-boundary verifier (stage-2 false-positive filter)
// ---------------------------------------------------------------------------

// collectWordBound is collectPatterns through MatchWordBound.
func collectWordBound(t *testing.T, patterns []string, text string) []string {
	t.Helper()
	m := newACMatcher(patterns)
	seen := make(map[int]bool)
	var hits []string
	m.MatchWordBound(text, func(idx int) {
		if !seen[idx] {
			seen[idx] = true
			hits = append(hits, m.patterns[idx])
		}
	})
	sort.Strings(hits)
	return hits
}

func TestACMatchWordBoundFiltersInsideWordHits(t *testing.T) {
	patterns := []string{"kdc", "ad ", "tgt failed"}
	// All three are substring-present but NOT at a word boundary.
	f := collectWordBound(t, patterns, "login failed_retry kdcluster badadjust tgt failed_verification")
	if len(f) != 0 {
		t.Errorf("word-boundary mode must reject inside-word matches, got %v", f)
	}
	// Same text must still be reported by plain Match (substring semantics
	// are unchanged for existing callers).
	plain := collectPatterns(t, patterns, "login failed_retry kdcluster badadjust tgt failed_verification")
	if len(plain) == 0 {
		t.Errorf("plain Match must keep substring semantics, got no hits")
	}
}

func TestACMatchWordBoundAcceptsRealBoundaries(t *testing.T) {
	patterns := []string{"kdc", "tgt failed"}
	got := collectWordBound(t, patterns, "KDC unreachable: tgt failed verification using key")
	want := []string{"kdc", "tgt failed"}
	if !equalStrings(got, want) {
		t.Errorf("word-boundary mode must accept true word-boundary matches: got %v, want %v", got, want)
	}
}

func TestACMatchWordBoundBoundaryIsPunctuationAware(t *testing.T) {
	// '.', '[', ':' and '-' separate words; '_' and alphanumerics do not.
	patterns := []string{"kdc"}
	for _, text := range []string{"x.kdc", "[kdc]", "kdc:", "kdc-2"} {
		if got := collectWordBound(t, patterns, text); len(got) != 1 {
			t.Errorf("punctuation around %q must count as a boundary, got %v", text, got)
		}
	}
	for _, text := range []string{"xkdc", "kdc2", "kdc_helper", "kdcx"} {
		if got := collectWordBound(t, patterns, text); len(got) != 0 {
			t.Errorf("%q must NOT count as a boundary, got %v", text, got)
		}
	}
}

func TestACMatchWordBoundAtTextEdges(t *testing.T) {
	patterns := []string{"kdc"}
	if got := collectWordBound(t, patterns, "kdc at start"); len(got) != 1 {
		t.Errorf("start-of-text boundary must match, got %v", got)
	}
	if got := collectWordBound(t, patterns, "at end kdc"); len(got) != 1 {
		t.Errorf("end-of-text boundary must match, got %v", got)
	}
}
