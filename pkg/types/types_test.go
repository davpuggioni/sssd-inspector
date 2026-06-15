// Package types provides shared data types for SSSD Inspector
package types

import (
	"encoding/json"
	"testing"
)

// TestReportData_JSONRoundTrip verifies that ReportData serializes/deserializes correctly
func TestReportData_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		data ReportData
	}{
		{
			name: "empty report",
			data: ReportData{},
		},
		{
			name: "full report",
			data: ReportData{
				AppVersion:    "1.0.0",
				Timestamp:     "01:01:2026 12:00:00",
				SupportCaseID: "SR#12345",
				SssdInstalled: true,
			},
		},
		{
			name: "report with errors",
			data: ReportData{
				SSSDLogErrors: []SSSDLogError{
					{Description: "Error 1", Examples: []string{"line1", "line2"}},
					{Description: "Error 2", Examples: nil},
				},
				Problems: []string{"Problem A", "Problem B"},
			},
		},
		{
			name: "report with KB articles",
			data: ReportData{
				MatchedTIDs: []TIDArticle{
					{TIDID: "TID-001", Title: "Test Article", Evidence: []string{"ev1"}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(tt.data)
			if err != nil {
				t.Fatalf("json.Marshal failed: %v", err)
			}

			var decoded ReportData
			if err := json.Unmarshal(encoded, &decoded); err != nil {
				t.Fatalf("json.Unmarshal failed: %v", err)
			}

			// Verify evidence is stripped (has json:"-" tag)
			if len(decoded.MatchedTIDs) > 0 && len(decoded.MatchedTIDs[0].Evidence) > 0 {
				t.Error("Evidence should be empty after JSON round-trip (json:\"-\" tag)")
			}
		})
	}
}

// TestSSSDLogError_ZeroValue verifies default zero values
func TestSSSDLogError_ZeroValue(t *testing.T) {
	var e SSSDLogError
	if e.Description != "" {
		t.Errorf("expected empty Description, got %q", e.Description)
	}
	if e.Examples != nil {
		t.Errorf("expected nil Examples, got %v", e.Examples)
	}
}

// TestTimelineEvent verifies TimelineEvent structure
func TestTimelineEvent(t *testing.T) {
	ev := TimelineEvent{
		Timestamp: "2026-04-01 12:00:00",
		Message:   "Test event",
		RawLog:    "(2026-04-01 12:00:00) test log line",
	}

	if ev.Timestamp != "2026-04-01 12:00:00" {
		t.Errorf("unexpected Timestamp: %q", ev.Timestamp)
	}
	if ev.Message != "Test event" {
		t.Errorf("unexpected Message: %q", ev.Message)
	}
	if ev.RawLog != "(2026-04-01 12:00:00) test log line" {
		t.Errorf("unexpected RawLog: %q", ev.RawLog)
	}
}

// TestTIDArticle_JSON verifies serialization of KB articles
func TestTIDArticle_JSON(t *testing.T) {
	article := TIDArticle{
		TIDID:          "TID-0001",
		Title:          "Test Article",
		URL:            "https://example.com/0001",
		Description:    "Description",
		LogPatterns:    []string{"error.*timeout"},
		ConfigPatterns: []string{"debug_level"},
		Evidence:       []string{"line1"}, // should be excluded by json:"-"
	}

	data, err := json.Marshal(article)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded TIDArticle
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.TIDID != "TID-0001" {
		t.Errorf("expected TID-0001, got %q", decoded.TIDID)
	}
	if decoded.Evidence != nil {
		t.Error("Evidence field should be nil after JSON round-trip")
	}
}
