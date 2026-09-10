// export_roundtrip_test.go
//
// Regression test for the GUI Export TXT / Export JSON / Export PDF path.
//
// The frontend sends the whole `currentReport` JS object back to Go, so
// SaveTXT/SaveJSON must tolerate whatever Analyze serializes — including
// the P7 aggregated timeline rows (Occurrences > 1, Samples). This test
// simulates the exact JSON round trip and runs the three builders.
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// roundTripReport simulates the Wails JSON round trip: Go serializes the
// analysis result (Analyze -> JS), the browser sends the same object back
// (JS -> SaveTXT/SaveJSON), and Go deserializes into ReportData.
func roundTripReport(t *testing.T, r ReportData) ReportData {
	t.Helper()
	wire, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal analysis result: %v", err)
	}
	var back ReportData
	dec := json.NewDecoder(strings.NewReader(string(wire)))
	if err := dec.Decode(&back); err != nil {
		t.Fatalf("deserialize report after Wails round trip: %v", err)
	}
	return back
}

// TestExportRoundTrip_AggregatedTimeline verifies the export builders
// survive a full report containing P7 aggregated timeline rows.
func TestExportRoundTrip_AggregatedTimeline(t *testing.T) {
	r := ReportData{
		AppVersion:      "test",
		Timestamp:       "2024-01-01 00:00:00",
		SupportCaseID:   "SR-1",
		KerberosRealm:   "EXAMPLE.COM",
		Nameservers:     []string{"10.0.0.1"},
		KeytabFound:     true,
		SssdConfigFound: true,
		SSSDLogErrors: []SSSDLogError{
			{Description: "Backend Offline: marked OFFLINE.", Examples: []string{"sssd: Going offline!"}},
		},
		Timeline: []TimelineEvent{
			{
				Timestamp:   "2024-01-01 00:00:01",
				Message:     "Backend Offline: marked OFFLINE.",
				RawLog:      "sssd: Going offline!",
				Occurrences: 200,
				Samples:     []string{"sssd: Going offline! pid=1", "sssd: Going offline! pid=2"},
			},
		},
		TemporalClusters: []TemporalCluster{
			{Description: "Backend Offline: marked OFFLINE.", EventCount: 200, WindowStart: "2024-01-01 00:00:01", WindowEnd: "2024-01-01 00:05:00", SampleRawLog: "sssd: Going offline!"},
		},
		ConfigFindings: []ConfigFinding{
			{Severity: SevError, Category: "offline", Message: "[CORRELATED ROOT CAUSE] Backend went OFFLINE", SourcePath: "timeline", SourceKey: "sequence:offline"},
		},
		Problems: []string{"sssd.service is not actively running."},
		Warnings: []string{"[TUNING] hint"},
		MatchedTIDs: []TIDArticle{
			{TIDID: "TID-1", Title: "T", Description: "D", URL: "https://example.invalid"},
		},
		Summary: ExecutiveSummary{HealthScore: 40, ErrorCount: 1, ProblemCount: 1, WarningCount: 1, LogErrorCount: 1, TopCategory: "offline", TopCategoryHits: 3, Headline: "[ROOT CAUSE] offline"},
	}

	back := roundTripReport(t, r)

	if len(back.Timeline) != 1 {
		t.Fatalf("timeline rows after round trip = %d, want 1", len(back.Timeline))
	}
	if back.Timeline[0].Occurrences != 200 {
		t.Errorf("occurrences after round trip = %d, want 200", back.Timeline[0].Occurrences)
	}
	if len(back.Timeline[0].Samples) != 2 {
		t.Errorf("samples after round trip = %d, want 2", len(back.Timeline[0].Samples))
	}

	// SaveJSON path
	data, err := buildJSONReport(back)
	if err != nil {
		t.Fatalf("buildJSONReport after round trip: %v", err)
	}
	if !strings.Contains(string(data), "occurrences") {
		t.Error("JSON export lost the aggregated occurrence count")
	}

	// SaveTXT path
	txt := buildTextReport(back)
	if !strings.Contains(txt, "Backend Offline") {
		t.Error("TXT export missing timeline-derived content after round trip")
	}

	// HTML path (writeHTMLReportFile must not fail on aggregated rows)
	dir := t.TempDir()
	fname := filepath.Join(dir, "report.html")
	writeHTMLReportFile(back, fname)
	html, err := os.ReadFile(fname)
	if err != nil {
		t.Fatalf("read html: %v", err)
	}
	for _, frag := range []string{"Event Timeline", "200 total occurrence(s)", "Root-Cause Incidents", "Temporal Clusters"} {
		if !strings.Contains(string(html), frag) {
			t.Errorf("HTML export missing %q after round trip", frag)
		}
	}
}
