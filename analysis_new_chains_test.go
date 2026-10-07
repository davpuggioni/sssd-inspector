// analysis_new_chains_test.go
//
// Regression tests for the R5-R8 sequence-chain emitters added in Phase 6d.2
// (TLS->SASL, idmap->downstream, GPO->PAM, enumeration->watchdog). Each chain
// must: fire when BOTH links are present, stay silent when only one is (the
// conservative >=2-signal rule), file itself under its own category, and
// dedup on a second call.
package main

import (
	"strings"
	"testing"
)

// requireFinding asserts a finding with the given category exists.
func requireFinding(t *testing.T, r *ReportData, cat string) {
	t.Helper()
	if !containsStringCategory(r, cat) {
		t.Errorf("expected a %q root-cause finding, got %v", cat, findingCategories(r))
	}
}

// forbidFinding asserts no finding carries the given category.
func forbidFinding(t *testing.T, r *ReportData, cat string) {
	t.Helper()
	if containsStringCategory(r, cat) {
		t.Errorf("no %q finding expected, got %v", cat, findingCategories(r))
	}
}

// hasRule asserts a finding with the given RuleID exists.
func hasRule(r *ReportData, ruleID string) bool {
	for _, f := range r.ConfigFindings {
		if f.RuleID == ruleID {
			return true
		}
	}
	return false
}

// --- R5: TLS -> SASL -------------------------------------------------------

func TestEmitTLSChain_Fires(t *testing.T) {
	links := []seqLink{
		{idx: 0, message: "Could not start TLS encryption", rawLog: "ldap_start_tls failed"},
		{idx: 1, message: "LDAP SASL bind failed", rawLog: "ldap_sasl_interactive_bind_s failed"},
	}
	var report ReportData
	emitted := map[string]bool{}
	emitTLSChain(links, &report, emitted)
	requireFinding(t, &report, "tls")
	if !hasRule(&report, "correlate:tls-before-sasl") {
		t.Errorf("expected correlate:tls-before-sasl rule, got %v", findingCategories(&report))
	}
	// Dedup: second call must be a no-op.
	emitTLSChain(links, &report, emitted)
	if n := len(report.ConfigFindings); n != 1 {
		t.Errorf("chain must emit exactly once, got %d findings", n)
	}
}

func TestEmitTLSChain_RequiresBothSignals(t *testing.T) {
	tlsOnly := []seqLink{{idx: 0, message: "Could not start TLS encryption", rawLog: "x"}}
	var r1 ReportData
	emitTLSChain(tlsOnly, &r1, map[string]bool{})
	forbidFinding(t, &r1, "tls")

	saslOnly := []seqLink{{idx: 0, message: "LDAP SASL bind failed", rawLog: "y"}}
	var r2 ReportData
	emitTLSChain(saslOnly, &r2, map[string]bool{})
	if len(r1.ConfigFindings)+len(r2.ConfigFindings) != 0 {
		t.Errorf("single-link chains must stay silent, got %v + %v", r1.ConfigFindings, r2.ConfigFindings)
	}
}

func TestEmitTLSChain_MatchesUnderscoreLogVariants(t *testing.T) {
	// Real SSSD writes "ldap_sasl_interactive_bind_s failed" and
	// "ldap_start_tls failed" — underscores, not spaces.
	links := []seqLink{
		{idx: 0, message: "sdap_async_sys_connect", rawLog: "ldap_start_tls failed: connect"},
		{idx: 1, message: "sdap_connection_setup", rawLog: "ldap_sasl_interactive_bind_s failed"},
	}
	var report ReportData
	emitTLSChain(links, &report, map[string]bool{})
	requireFinding(t, &report, "tls")
}

// --- R6: idmap -> downstream ------------------------------------------------

func TestEmitIDMapChain_Fires(t *testing.T) {
	links := []seqLink{
		{idx: 0, message: "Cannot map SID to UID", rawLog: "sid lookup failed: out of range"},
		{idx: 1, message: "User lookup", rawLog: "no such user for uid 4294967295"},
	}
	var report ReportData
	emitIDMapChain(links, &report, map[string]bool{})
	requireFinding(t, &report, "idmap")
	if !hasRule(&report, "correlate:idmap-downstream") {
		t.Errorf("expected correlate:idmap-downstream rule")
	}
}

func TestEmitIDMapChain_RequiresDownstreamSignal(t *testing.T) {
	idmapOnly := []seqLink{{idx: 0, message: "sid mapping failed", rawLog: "invalid sid"}}
	var r1 ReportData
	emitIDMapChain(idmapOnly, &r1, map[string]bool{})
	forbidFinding(t, &r1, "idmap")

	downstreamOnly := []seqLink{{idx: 0, message: "lookup", rawLog: "no such user"}}
	var r2 ReportData
	emitIDMapChain(downstreamOnly, &r2, map[string]bool{})
	if len(r1.ConfigFindings)+len(r2.ConfigFindings) != 0 {
		t.Errorf("single-link idmap chains must stay silent")
	}
}

// --- R7: GPO -> PAM --------------------------------------------------------

func TestEmitGPOChain_Fires(t *testing.T) {
	links := []seqLink{
		{idx: 0, message: "GPO evaluation", rawLog: "failed to read GPO from SYSVOL"},
		{idx: 1, message: "pam_sss", rawLog: "authentication failure; user not permitted"},
	}
	var report ReportData
	emitGPOChain(links, &report, map[string]bool{})
	requireFinding(t, &report, "gpo")
	if !hasRule(&report, "correlate:gpo-before-pam") {
		t.Errorf("expected correlate:gpo-before-pam rule")
	}
}

func TestEmitGPOChain_RequiresBothSignals(t *testing.T) {
	gpoOnly := []seqLink{{idx: 0, message: "gpo", rawLog: "gpo error"}}
	var r1 ReportData
	emitGPOChain(gpoOnly, &r1, map[string]bool{})
	forbidFinding(t, &r1, "gpo")

	pamOnly := []seqLink{{idx: 0, message: "pam", rawLog: "authentication failure"}}
	var r2 ReportData
	emitGPOChain(pamOnly, &r2, map[string]bool{})
	if len(r1.ConfigFindings)+len(r2.ConfigFindings) != 0 {
		t.Errorf("single-link GPO chains must stay silent")
	}
}

// --- R8: enumeration -> watchdog -------------------------------------------

func TestEmitEnumWatchdogChain_Fires(t *testing.T) {
	links := []seqLink{
		{idx: 0, message: "sssd_nss", rawLog: "overlarge group with 40000 members"},
		{idx: 1, message: "watchdog", rawLog: "terminated by own WATCHDOG"},
	}
	var report ReportData
	emitEnumWatchdogChain(links, &report, map[string]bool{})
	requireFinding(t, &report, "overload")
	if !hasRule(&report, "correlate:enum-watchdog") {
		t.Errorf("expected correlate:enum-watchdog rule")
	}
}

func TestEmitEnumWatchdogChain_DoesNotDoubleEmitOverload(t *testing.T) {
	// The R2 overload chain already owns the "overload" key; R8 must respect
	// it so one report does not show two competing overload root causes.
	links := []seqLink{
		{idx: 0, message: "sssd_nss", rawLog: "huge group"},
		{idx: 1, message: "watchdog", rawLog: "terminated by own WATCHDOG"},
	}
	var report ReportData
	emitted := map[string]bool{"overload": true} // R2 fired first
	emitEnumWatchdogChain(links, &report, emitted)
	forbidFinding(t, &report, "overload")
}

func TestEmitEnumWatchdogChain_RequiresBothSignals(t *testing.T) {
	watchdogOnly := []seqLink{{idx: 0, message: "watchdog", rawLog: "terminated by own WATCHDOG"}}
	var r1 ReportData
	emitEnumWatchdogChain(watchdogOnly, &r1, map[string]bool{})
	enumOnly := []seqLink{{idx: 0, message: "nss", rawLog: "overlarge group"}}
	var r2 ReportData
	emitEnumWatchdogChain(enumOnly, &r2, map[string]bool{})
	if len(r1.ConfigFindings)+len(r2.ConfigFindings) != 0 {
		t.Errorf("single-link enum/watchdog chains must stay silent")
	}
}

// --- Integration: correlateSequences dispatches R5-R8 ----------------------

// TestCorrelateSequences_DispatchesNewChains builds a synthetic timeline long
// enough for minTimelineEvents and verifies the dispatcher wires the new
// emitters (regression: emitters defined but never called).
func TestCorrelateSequences_DispatchesNewChains(t *testing.T) {
	timeline := []TimelineEvent{
		{Timestamp: "2024-05-01 10:00:00", Message: "TLS failure", RawLog: "ldap_start_tls failed: certificate"},
		{Timestamp: "2024-05-01 10:00:01", Message: "SASL bind failure", RawLog: "ldap_sasl_interactive_bind_s failed"},
		{Timestamp: "2024-05-01 10:00:02", Message: "GPO failure", RawLog: "failed to read GPO"},
		{Timestamp: "2024-05-01 10:00:03", Message: "PAM denial", RawLog: "authentication failure"},
	}
	var report ReportData
	correlateSequences(timeline, &report)

	if !hasRule(&report, "correlate:tls-before-sasl") {
		t.Errorf("dispatcher must invoke emitTLSChain; findings: %v", findingCategories(&report))
	}
	if !hasRule(&report, "correlate:gpo-before-pam") {
		t.Errorf("dispatcher must invoke emitGPOChain; findings: %v", findingCategories(&report))
	}
	// Every correlated finding must carry evidence (provenance contract).
	for _, f := range report.ConfigFindings {
		if strings.HasPrefix(f.Message, "[CORRELATED ROOT CAUSE]") && f.Evidence == "" {
			t.Errorf("correlated finding %s must carry evidence", f.RuleID)
		}
	}
}

// TestCorrelateSequences_TooShortTimelineStaysSilent pins the abstention floor.
func TestCorrelateSequences_TooShortTimelineStaysSilent(t *testing.T) {
	timeline := []TimelineEvent{
		{Timestamp: "2024-05-01 10:00:00", Message: "TLS failure", RawLog: "ldap_start_tls failed"},
		{Timestamp: "2024-05-01 10:00:01", Message: "SASL bind failure", RawLog: "ldap_sasl_interactive_bind_s failed"},
	}
	var report ReportData
	correlateSequences(timeline, &report)
	if len(report.ConfigFindings) != 0 {
		t.Errorf("below minTimelineEvents the engine must abstain, got %v", findingCategories(&report))
	}
}
