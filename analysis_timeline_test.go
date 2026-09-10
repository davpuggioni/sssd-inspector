// analysis_timeline_test.go
package main

import (
	"strings"
	"testing"
)

// TestTimelineAggregator_CollapsesRepeats verifies the P7 bounded timeline:
// 10.000 identical matches collapse into ONE row with Occurrences = 10.000
// (memory stays flat instead of one event per line).
func TestTimelineAggregator_CollapsesRepeats(t *testing.T) {
	agg := newTimelineAggregator()
	for i := 0; i < 10000; i++ {
		agg.add("2024-01-01 00:00:01", "Backend Offline: marked OFFLINE.", "sssd: Going offline!")
	}
	if len(agg.events) != 1 {
		t.Fatalf("expected 1 aggregated row, got %d", len(agg.events))
	}
	if agg.events[0].Occurrences != 10000 {
		t.Errorf("occurrences = %d, want 10000", agg.events[0].Occurrences)
	}
}

// TestTimelineAggregator_DistinctRowsAndSamples verifies distinct
// (timestamp, message) pairs stay separate and per-row samples are capped.
func TestTimelineAggregator_DistinctRowsAndSamples(t *testing.T) {
	agg := newTimelineAggregator()
	agg.add("2024-01-01 00:00:01", "Event A", "raw-a1")
	agg.add("2024-01-01 00:00:01", "Event A", "raw-a2")
	agg.add("2024-01-01 00:00:01", "Event A", "raw-a3")
	agg.add("2024-01-01 00:00:01", "Event A", "raw-a4") // 4th distinct sample dropped
	agg.add("2024-01-01 00:00:02", "Event A", "raw-a1") // different ts -> separate row
	if len(agg.events) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(agg.events))
	}
	if agg.events[0].Occurrences != 4 {
		t.Errorf("row0 occurrences = %d, want 4", agg.events[0].Occurrences)
	}
	if len(agg.events[0].Samples) != maxTimelineSamples {
		t.Errorf("row0 samples = %d, want cap %d", len(agg.events[0].Samples), maxTimelineSamples)
	}
}

// TestTimelineAggregator_CapDropsDistinct verifies the global row cap:
// beyond maxTimelineEvents, new DISTINCT events are dropped but counted.
func TestTimelineAggregator_CapDropsDistinct(t *testing.T) {
	agg := newTimelineAggregator()
	for i := 0; i < maxTimelineEvents+10; i++ {
		agg.add("2024-01-01 00:00:01", strings.Repeat("E", 8)+string(rune('a'+i%26))+string(rune('0'+i/26%10))+string(rune('A'+i/260%26)), "raw")
	}
	if len(agg.events) > maxTimelineEvents {
		t.Errorf("rows = %d, want <= %d", len(agg.events), maxTimelineEvents)
	}
	if agg.dropped == 0 {
		t.Error("expected dropped > 0 once the cap is exceeded")
	}
}

// TestTimelineTotalOccurrences_BackwardCompat verifies pre-aggregation rows
// (Occurrences == 0, as built by hand in older tests) still count as 1.
func TestTimelineTotalOccurrences_BackwardCompat(t *testing.T) {
	events := []TimelineEvent{
		{Timestamp: "2024-01-01 00:00:01", Message: "A", RawLog: "a"},
		{Timestamp: "2024-01-01 00:00:02", Message: "B", RawLog: "b", Occurrences: 5},
	}
	if got := timelineTotalOccurrences(events); got != 6 {
		t.Errorf("total = %d, want 6", got)
	}
}

// TestAnalyzeTemporalClusters_AggregatedVolume verifies a burst collapsed
// into one aggregated row still forms a cluster via its Occurrences count.
func TestAnalyzeTemporalClusters_AggregatedVolume(t *testing.T) {
	timeline := []TimelineEvent{
		{Timestamp: "2026-01-01 10:00:00", Message: "Kerberos: Clock skew too great", RawLog: "raw1", Occurrences: 5},
	}
	clusters := analyzeTemporalClusters(timeline)
	if len(clusters) != 1 {
		t.Fatalf("expected 1 cluster, got %d: %+v", len(clusters), clusters)
	}
	if clusters[0].EventCount != 5 {
		t.Errorf("cluster count = %d, want 5", clusters[0].EventCount)
	}
}
