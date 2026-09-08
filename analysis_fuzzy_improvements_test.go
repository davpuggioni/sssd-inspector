// analysis_fuzzy_improvements_test.go
package main

import (
	"testing"
)

// TestNormalizeKBTerm verifies the domain synonym dictionary canonicalises
// equivalent SSSD/AD vocabulary into a single term for lexical ranking.
func TestNormalizeKBTerm(t *testing.T) {
	cases := map[string]string{
		"krb5":      "kerberos",
		"krb":       "kerberos",
		"kdc":       "kerberos",
		"keytab":    "kerberos",
		"ntlm":      "kerberos",
		"tgt":       "ticket",
		"nsupdate":  "dns",
		"ddns":      "dns",
		"skew":      "clock",
		"reconnect": "connect",
		"starttls":  "tls",
		"ssl":       "tls",
	}
	for in, want := range cases {
		if got := normalizeKBTerm(in); got != want {
			t.Errorf("normalizeKBTerm(%q) = %q, want %q", in, got, want)
		}
	}
	// Non-synonym tokens must pass through unchanged.
	if got := normalizeKBTerm("realm"); got != "realm" {
		t.Errorf("normalizeKBTerm should pass 'realm' through, got %q", got)
	}
}

// TestTokenizeKB_SynonymsApplied confirms the tokenizer runs each kept token
// through the synonym dictionary (so ranking sees canonical vocabulary).
func TestTokenizeKB_SynonymsApplied(t *testing.T) {
	toks := tokenizeKB("krb5_child timeout KDC unreachable")
	foundKerberos := false
	for _, tok := range toks {
		if tok == "kerberos" {
			foundKerberos = true
		}
	}
	if !foundKerberos {
		t.Errorf("expected tokenized output to contain canonical 'kerberos', got %v", toks)
	}
}

// TestKbBM25_RelevantOverVerbose verifies BM25 ranks the short relevant article
// above a long article that only shares blank terms, thanks to document-length
// normalisation.
func TestKbBM25_RelevantOverVerbose(t *testing.T) {
	df := map[string]int{"kerberos": 2, "timeout": 2, "kdc": 1, "ad": 2}
	corpus := 3
	avg := 3
	query := map[string]float64{"kerberos": 1, "timeout": 1}

	// Short relevant doc: strong, concentrated lexical overlap.
	shortDoc := map[string]float64{"kerberos": 1, "timeout": 1}
	// Verbose doc: same terms but diluted by ~many unrelated tokens.
	verboseDoc := map[string]float64{"kerberos": 1, "timeout": 1, "ad": 3, "gpo": 2, "ldap": 2, "dns": 2}

	shortScore := kbBM25(query, shortDoc, 2, avg, corpus, df)
	verboseScore := kbBM25(query, verboseDoc, 8, avg, corpus, df)
	if shortScore <= 0 {
		t.Errorf("expected a positive BM25 score for the relevant doc, got %v", shortScore)
	}
	if verboseScore >= shortScore {
		t.Errorf("BM25 scoring must favour the short relevant doc (got short=%v verbose=%v)", shortScore, verboseScore)
	}
}
