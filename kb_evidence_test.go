// kb_evidence_test.go
//
// TIDArticle.Evidence used to be tagged `json:"-"`, so the log lines that
// triggered a KB match never reached the report and the "Evidence Found" block
// in the GUI could not render. These tests pin both halves of shipping it: it
// must be in the JSON, and anonymization must scrub it like every other log
// excerpt.
package main

import (
	"encoding/json"
	"strings"
	"testing"

	"sssd-inspector/constants"
)

// TestKBEvidenceReachesTheJSONReport: the evidence has to be in the payload
// the frontend receives; without the key the GUI can only link the TID.
func TestKBEvidenceReachesTheJSONReport(t *testing.T) {
	report := ReportData{
		MatchedTIDs: []TIDArticle{{
			TIDID:    "TID-TEST-1",
			Title:    "Legacy RC4 enctype",
			Evidence: []string{"Aug 18 10:00:00 host01 sssd: rc4-hmac in use"},
		}},
	}
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(raw), `"evidence"`) {
		t.Errorf("report JSON has no evidence key: %s", raw)
	}
	if !strings.Contains(string(raw), "rc4-hmac in use") {
		t.Errorf("the matched log line is missing from the report: %s", raw)
	}
}

// TestKBEvidenceOmittedWhenEmpty: an article matched only through config
// patterns has no log line to show, and the key must not appear at all.
func TestKBEvidenceOmittedWhenEmpty(t *testing.T) {
	report := ReportData{MatchedTIDs: []TIDArticle{{TIDID: "TID-TEST-2"}}}
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(raw), `"evidence"`) {
		t.Errorf("an empty evidence list was serialized: %s", raw)
	}
}

// TestKBEvidenceIsScrubbedByAnonymization is the safety half: shipping log
// excerpts must not bypass anonymization, and must be treated exactly like the
// excerpts already in the report.
//
// The domain here is a realistic one, NOT the replacement token: masking
// replaces the domain with constants.DomainReplacement ("example.com"), so a
// fixture whose SearchDomain already is "example.com" would make the
// substitution a no-op and "prove" a leak that does not exist. That mistake
// briefly had this test asserting the wrong thing.
func TestKBEvidenceIsScrubbedByAnonymization(t *testing.T) {
	line := "Aug 18 10:00:00 host01.corp.example sssd: bind to ldap://dc01.corp.example:389 from 10.0.0.5 failed"
	report := ReportData{
		Hostname:      "host01.corp.example",
		SearchDomain:  "corp.example",
		MatchedTIDs:   []TIDArticle{{TIDID: "TID-TEST-3", Evidence: []string{line}}},
		SSSDLogErrors: []SSSDLogError{{Description: "d", Examples: []string{line}}},
	}
	anonymizeReport(&report, "")

	got := report.MatchedTIDs[0].Evidence[0]
	if strings.Contains(got, "corp.example") {
		t.Errorf("anonymization left the domain (or one of its subdomains) in the KB evidence: %s", got)
	}
	if strings.Contains(got, "10.0.0.5") {
		t.Errorf("anonymization left the IP in the KB evidence: %s", got)
	}
	// The host label is kept, the domain is not: that is what makes the line
	// still diagnostic after redaction.
	if !strings.Contains(got, "dc01."+constants.DomainReplacement) {
		t.Errorf("expected the subdomain to survive as dc01.%s, got: %s", constants.DomainReplacement, got)
	}
	if got != report.SSSDLogErrors[0].Examples[0] {
		t.Errorf("KB evidence is scrubbed differently from other log excerpts:\nevidence: %s\nexamples:  %s",
			got, report.SSSDLogErrors[0].Examples[0])
	}
	// The non-sensitive part must survive: an evidence block full of
	// placeholders would be useless.
	if !strings.Contains(got, "bind to") || !strings.Contains(got, "failed") {
		t.Errorf("anonymization destroyed the diagnostic content: %s", got)
	}
}

// TestAnonymizationMasksSubdomainsOfEveryRedactedDomain pins the property the
// previous test only implied: the domain, the realm and the AD domain are
// replaced as substrings, so a fully-qualified host built from them cannot
// survive. Anonymization is a privacy boundary — this test is what would catch
// a future "tidy-up" that switches those ReplaceAll calls for a word-bounded
// match and quietly starts leaking the customer domain through host names.
func TestAnonymizationMasksSubdomainsOfEveryRedactedDomain(t *testing.T) {
	const host = "dc01.corp.example:389"
	tests := []struct {
		name  string
		field func(*ReportData) *string
		value string
	}{
		// The search domain and the realm are masked as plain substrings, which
		// already covers every host under them.
		{"search domain", func(r *ReportData) *string { return &r.SearchDomain }, "corp.example"},
		{"kerberos realm", func(r *ReportData) *string { return &r.KerberosRealm }, "corp.example"},
		// The AD domain is a different story: "dc01.corp.example" does not
		// contain "ad.corp.example", so the parent domain has to be masked too
		// or the customer's domain leaks through the AD's own host names.
		{"ad domain", func(r *ReportData) *string { return &r.AdDomain }, "ad.corp.example"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			report := ReportData{}
			*tc.field(&report) = tc.value
			report.SSSDLogErrors = []SSSDLogError{{
				Description: "d",
				Examples:    []string{"bind to " + host + " from 10.0.0.5"},
			}}
			anonymizeReport(&report, "")

			got := report.SSSDLogErrors[0].Examples[0]
			if strings.Contains(strings.ToLower(got), "corp.example") {
				t.Errorf("%s: the domain survived inside a subdomain: %s", tc.name, got)
			}
			// A realm is upper case by convention, so it is replaced with the
			// upper-case marker; the search domain keeps its case. Either way the
			// replacement token must be what is left.
			if !strings.Contains(strings.ToUpper(got), strings.ToUpper(constants.DomainReplacement)) {
				t.Errorf("%s: expected the domain to be replaced by %q, got: %s",
					tc.name, constants.DomainReplacement, got)
			}
		})
	}
}

// TestParentDomainOnlyForMultiLabelDomains: the helper must not strip a domain
// down to a single label, or it would mask the bare word "corp" in unrelated
// text. Two labels already cover their subdomains as a substring.
func TestParentDomainOnlyForMultiLabelDomains(t *testing.T) {
	tests := []struct{ in, want string }{
		{"ad.corp.example", "corp.example"},
		{"a.b.c.example", "c.example"},
		{"corp.example", ""}, // two labels: the value is already the parent
		{"corp", ""},         // no dot at all
		{"", ""},
	}
	for _, tc := range tests {
		if got := parentDomain(tc.in); got != tc.want {
			t.Errorf("parentDomain(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestAnonymizationKeepsTwoLabelAdDomainReadable: a two-label AD domain must
// still be masked, and the DC host under it must become the replacement token —
// not the bare word "corp".
func TestAnonymizationKeepsTwoLabelAdDomainReadable(t *testing.T) {
	report := ReportData{AdDomain: "corp.example"}
	report.SSSDLogErrors = []SSSDLogError{{
		Description: "d",
		Examples:    []string{"site dc01.corp.example joined"},
	}}
	anonymizeReport(&report, "")

	got := report.SSSDLogErrors[0].Examples[0]
	if !strings.Contains(got, "dc01."+constants.DomainReplacement) {
		t.Errorf("expected dc01.%s, got: %s", constants.DomainReplacement, got)
	}
}
