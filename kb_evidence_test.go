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
// NOTE on the line chosen here: "dc01.example.com" survives anonymization
// because the search-domain mask does not match inside a subdomain. That is a
// pre-existing, report-wide limitation (SSSDLogError.Examples and
// TimelineEvent.RawLog behave identically), not something shipping the
// evidence introduces — so this test pins parity rather than inventing a
// stronger guarantee. Narrowing that mask is a separate, report-wide change.
func TestKBEvidenceIsScrubbedByAnonymization(t *testing.T) {
	line := "Aug 18 10:00:00 host01.example.com sssd: bind to ldap://dc01.example.com:389 from 10.0.0.5 failed"
	report := ReportData{
		Hostname:      "host01.example.com",
		SearchDomain:  "example.com",
		MatchedTIDs:   []TIDArticle{{TIDID: "TID-TEST-3", Evidence: []string{line}}},
		SSSDLogErrors: []SSSDLogError{{Description: "d", Examples: []string{line}}},
	}
	anonymizeReport(&report, "")

	got := report.MatchedTIDs[0].Evidence[0]
	if strings.Contains(got, "host01.example.com") {
		t.Errorf("anonymization left the hostname in the KB evidence: %s", got)
	}
	if strings.Contains(got, "10.0.0.5") {
		t.Errorf("anonymization left the IP in the KB evidence: %s", got)
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
