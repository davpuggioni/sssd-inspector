// analysis_graph_test.go — tests for the correlation graph engine (P9) and the
// diff/comparison engine (P5). No external dependencies.
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ─── helpers ────────────────────────────────────────────────────────────────

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("writeFile %s: %v", name, err)
	}
}

// ─── buildCorrelationGraph ─────────────────────────────────────────────────

func TestBuildCorrelationGraph_EmptyReport(t *testing.T) {
	r := &ReportData{ConfigFindings: []ConfigFinding{}}
	g := buildCorrelationGraph(r)

	if len(g.Entities) != 0 || len(g.Findings) != 0 || len(g.Sources) != 0 || len(g.Edges) != 0 {
		t.Errorf("empty report should yield empty graph, got entities=%d findings=%d sources=%d edges=%d",
			len(g.Entities), len(g.Findings), len(g.Sources), len(g.Edges))
	}
}

func TestBuildCorrelationGraph_BasicStructure(t *testing.T) {
	r := &ReportData{
		AdDomain:      "example.com",
		KerberosRealm: "OTHER.NET",
		Hostname:      "host01",
		SearchDomain:  "other.net",
		ConfigFindings: []ConfigFinding{
			{
				Severity: SevCritical, Category: "krb5_realm",
				Message:    "Kerberos default_realm mismatch",
				SourcePath: "krb5.conf", SourceKey: "[libdefaults].default_realm",
				SourceLine: 5, Evidence: "default_realm = OTHER.NET",
			},
			{
				Severity: SevError, Category: "dns",
				Message:    "DNS search domain mismatch",
				SourcePath: "sssd.conf", SourceKey: "ldap_uri",
				SourceLine: 12, Evidence: "ldap_uri = ld://dc.other.net",
			},
		},
	}
	g := buildCorrelationGraph(r)

	if len(g.Entities) < 4 {
		t.Errorf("expected at least 4 top-level entities, got %d", len(g.Entities))
	}
	if len(g.Findings) != 2 {
		t.Errorf("expected 2 findings, got %d", len(g.Findings))
	}
	if len(g.Sources) != 2 {
		t.Errorf("expected 2 sources, got %d", len(g.Sources))
	}
	if len(g.Edges) < 2 {
		t.Errorf("expected at least 2 edges, got %d", len(g.Edges))
	}
}

func TestBuildCorrelationGraph_NoOrphanEdges(t *testing.T) {
	r := &ReportData{
		AdDomain:      "example.com",
		KerberosRealm: "EXAMPLE.COM",
		ConfigFindings: []ConfigFinding{
			{
				Severity: SevWarning, Category: "ad_server",
				Message:    "ad_server is IP",
				SourcePath: "sssd.conf", SourceLine: 8,
				Evidence: "ad_server = 192.168.1.10",
			},
		},
	}
	g := buildCorrelationGraph(r)

	valid := make(map[string]bool)
	for _, e := range g.Entities {
		valid[e.ID] = true
	}
	for _, f := range g.Findings {
		valid[f.ID] = true
	}
	for _, s := range g.Sources {
		valid[s.ID] = true
	}
	for i, edge := range g.Edges {
		if !valid[edge.From] {
			t.Errorf("edge %d references unknown From node %q", i, edge.From)
		}
		if !valid[edge.To] {
			t.Errorf("edge %d references unknown To node %q", i, edge.To)
		}
	}
}

func TestBuildCorrelationGraph_DedupEntities(t *testing.T) {
	r := &ReportData{
		ConfigFindings: []ConfigFinding{
			{
				Severity: SevCritical, Category: "krb5_realm",
				Message:    "realm mismatch #1",
				SourcePath: "krb5.conf", SourceLine: 3,
				Evidence: "default_realm = EXAMPLE.COM",
			},
			{
				Severity: SevCritical, Category: "krb5_realm",
				Message:    "realm mismatch #2",
				SourcePath: "krb5.conf", SourceLine: 7,
				Evidence: "default_realm = EXAMPLE.COM",
			},
		},
	}
	g := buildCorrelationGraph(r)

	realmEntities := 0
	for _, e := range g.Entities {
		if e.Kind == "realm" {
			realmEntities++
		}
	}
	if realmEntities > 1 {
		t.Errorf("expected 1 deduped realm entity, got %d", realmEntities)
	}
}

func TestBuildCorrelationGraph_TemporalClusters(t *testing.T) {
	r := &ReportData{
		AdDomain: "example.com",
		TemporalClusters: []TemporalCluster{
			{
				Description:  "SSSD Offline: Provider forced offline",
				EventCount:   4,
				SampleRawLog: "(2026-01-01 10:00:00) Backend 'ad.example.com' is offline",
			},
		},
	}
	g := buildCorrelationGraph(r)

	found := false
	for _, f := range g.Findings {
		if f.Category == "temporal" && f.EventCount == 4 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected temporal cluster finding with EventCount=4")
	}
}

func TestBuildCorrelationGraph_KBSuggestions(t *testing.T) {
	r := &ReportData{
		KBSuggestions: []KBSuggestion{
			{
				TIDID:      "000012345",
				Title:      "SSSD fails to start after upgrade",
				Score:      0.72,
				SampleLine: "Unknown message in log file",
			},
		},
	}
	g := buildCorrelationGraph(r)

	found := false
	for _, f := range g.Findings {
		if f.Category == "kb_suggestion" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected kb_suggestion finding in graph")
	}
}

func TestBuildCorrelationGraph_PIIRedaction(t *testing.T) {
	r := &ReportData{
		AdDomain:     "corp.example.org",
		Hostname:     "prod-server-01",
		SearchDomain: "corp.example.org",
		ConfigFindings: []ConfigFinding{
			{
				Severity: SevError, Category: "ad_server",
				Message: "ad_server is IP", SourcePath: "sssd.conf",
				SourceLine: 5, Evidence: "ad_server = 10.0.0.5",
			},
		},
	}
	// Build graph from raw data (PII still present), then anonymize in-place.
	r.Graph = buildCorrelationGraph(r)
	anonymizeReport(r, "")

	for _, e := range r.Graph.Entities {
		if strings.Contains(e.Value, "corp.example.org") || strings.Contains(e.Value, "prod-server-01") {
			t.Errorf("PII leaked into entity value %q", e.Value)
		}
	}
	for _, f := range r.Graph.Findings {
		if strings.Contains(f.Evidence, "10.0.0.5") || strings.Contains(f.Evidence, "corp.example.org") {
			t.Errorf("PII leaked into finding evidence %q", f.Evidence)
		}
	}
	for _, s := range r.Graph.Sources {
		if strings.Contains(s.LineText, "10.0.0.5") || strings.Contains(s.LineText, "corp.example.org") {
			t.Errorf("PII leaked into source text %q", s.LineText)
		}
	}
}

// TestBuildCorrelationGraph_NoNullSlices guards against a frontend regression
// where an empty Sources slice serialized to JSON "sources": null, crashing
// the SVG renderer ("null is not an object evaluating 'n.sources.length'").
// Findings without source lines are exactly the case that used to leave
// Sources nil.
func TestBuildCorrelationGraph_NoNullSlices(t *testing.T) {
	r := &ReportData{
		ConfigFindings: []ConfigFinding{
			{
				Severity: SevWarning, Category: "dns",
				Message: "no source line", SourcePath: "",
				SourceLine: 0, Evidence: "",
			},
		},
	}
	g := buildCorrelationGraph(r)

	if len(g.Findings) == 0 {
		t.Fatalf("expected at least one finding even without a source line")
	}
	if g.Sources == nil {
		t.Error("Sources must be a non-nil slice (serializes to JSON []) even when empty")
	}
	if g.Entities == nil || g.Edges == nil || g.Findings == nil {
		t.Error("Entities/Findings/Edges must be non-nil slices")
	}
}

func TestCompareConfigs_NoChanges(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "sssd.conf", "[sssd]\ndomains = example.com\n\n[domain/example.com]\nid_provider = ad\n")

	progress := func(msg string, pct int) {}
	report := CompareConfigs(dir, dir, progress)

	if report.ScoreDelta != 0 {
		t.Errorf("comparing same dir should give 0 delta, got %d", report.ScoreDelta)
	}
	if len(report.OnlyInA) != 0 || len(report.OnlyInB) != 0 {
		t.Errorf("comparing same dir should yield no diff, got onlyA=%d onlyB=%d",
			len(report.OnlyInA), len(report.OnlyInB))
	}
}

func TestCompareConfigs_WithChanges(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()
	// A has an ad_server issue; B has it fixed
	writeFile(t, dirA, "sssd.conf", "[sssd]\ndomains = example.com\n\n[domain/example.com]\nid_provider = ad\nad_server = 192.168.1.10\n")
	writeFile(t, dirB, "sssd.conf", "[sssd]\ndomains = example.com\n\n[domain/example.com]\nid_provider = ad\nad_server = dc.example.com\n")

	progress := func(msg string, pct int) {}
	report := CompareConfigs(dirA, dirB, progress)

	if len(report.OnlyInA) == 0 {
		t.Errorf("expected findings only in A (ad_server=IP), got none")
	}
	if report.ScoreDelta <= 0 {
		t.Errorf("expected B to be healthier than A (positive delta), got %d", report.ScoreDelta)
	}
}
