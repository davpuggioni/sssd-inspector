// analysis_chains_test.go
//
// Tests for the sequence-chain emitters (analysis_sequences.go): the
// overload/restart-storm correlation and the timelineEpoch helper. These
// pin the root-cause consolidation behavior against regressions.
package main

import (
	"testing"
)

func TestEmitOverloadChain_Fires(t *testing.T) {
	links := []seqLink{
		{idx: 0, message: "Child terminated by own WATCHDOG", rawLog: "sssd_be terminated by own WATCHDOG"},
		{idx: 1, message: "LDAP packet overlarge", rawLog: "ldap packet overlarge, max 100"},
	}
	var report ReportData
	emitted := map[string]bool{}
	if !emitOverloadChain(links, &report, emitted) {
		t.Fatalf("expected the overload chain to fire")
	}
	if !containsStringCategory(&report, "ad_server") {
		t.Errorf("expected an 'ad_server' root-cause finding, got %v", findingCategories(&report))
	}
	// Second call with the same key must be a no-op (dedup guard).
	if emitOverloadChain(links, &report, emitted) {
		t.Errorf("already-emitted chain must not fire twice")
	}
}

func TestEmitOverloadChain_RequiresBothSignals(t *testing.T) {
	watchdogOnly := []seqLink{{idx: 0, message: "terminated by own WATCHDOG", rawLog: "x watchdog y"}}
	var r1 ReportData
	if emitOverloadChain(watchdogOnly, &r1, map[string]bool{}) {
		t.Errorf("watchdog alone must not fire the overload chain")
	}
	overloadOnly := []seqLink{{idx: 0, message: "ldap overlarge packet", rawLog: "overlarge"}}
	var r2 ReportData
	if emitOverloadChain(overloadOnly, &r2, map[string]bool{}) {
		t.Errorf("overload signal alone must not fire the chain")
	}
	if len(r1.ConfigFindings)+len(r2.ConfigFindings) != 0 {
		t.Errorf("no findings expected, got %v + %v", r1.ConfigFindings, r2.ConfigFindings)
	}
}

func TestTimelineEpoch_ParsesAndRejects(t *testing.T) {
	epoch, ok := timelineEpoch(TimelineEvent{Timestamp: "2024-05-01 12:00:00"})
	if !ok || epoch <= 0 {
		t.Errorf("expected a valid epoch, got %d, %v", epoch, ok)
	}
	if _, ok := timelineEpoch(TimelineEvent{Timestamp: "not-a-timestamp"}); ok {
		t.Errorf("garbage timestamp must be rejected")
	}
	if _, ok := timelineEpoch(TimelineEvent{}); ok {
		t.Errorf("empty timestamp must be rejected")
	}
}
