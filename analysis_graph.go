// analysis_graph.go — builds the interactive correlation graph and computes
// the delta between two analyses (Diff Mode, P5 / P9).
//
// The graph is a three-layer structure:
//   Entities (nouns)  →  Findings (verbs)  →  Sources (physical evidence)
//
// It is built AFTER PII anonymization so the frontend receives ready-to-render
// data. Every entity/finding/source has a stable, deduplicated ID; edges never
// reference missing node IDs.

package main

import (
	"fmt"
	"hash/fnv"
	"regexp"
	"strings"
)

// entityID produces a stable, case-insensitive dedup key for an entity.
func entityID(kind, value string) string {
	return fmt.Sprintf("entity:%s:%s", kind, strings.ToLower(value))
}

// sourceID produces a stable ID for a physical evidence node.
func sourceID(path string, line int) string {
	return fmt.Sprintf("source:%s:%d", strings.ToLower(path), line)
}

// findingID derives a stable ID from a finding's category + message so that
// the same diagnostic appearing multiple times collapses into one node.
func findingID(cat, msg string) string {
	h := fnv.New32a()
	h.Write([]byte(cat + "|" + msg))
	return fmt.Sprintf("finding:%08x", h.Sum32())
}

// entityValueRegex extracts the right-hand side of "key = value" lines.
var entityValueRegex = regexp.MustCompile(`(?i)^\s*([a-z0-9_.\-]+)\s*=\s*(.+?)\s*$`)

// buildCorrelationGraph walks the report's findings, entities and sources and
// assembles a CorrelationGraph. Called AFTER anonymizeReport.
func buildCorrelationGraph(r *ReportData) CorrelationGraph {
	var g CorrelationGraph
	// Initialize every slice explicitly so an empty graph serializes to [] in
	// JSON instead of null. The frontend consumes graph.{entities,findings,
	// sources,edges}.length directly; a null slice would crash it.
	g.Entities = make([]GraphEntity, 0)
	g.Findings = make([]GraphFindingNode, 0)
	g.Sources = make([]GraphSourceNode, 0)
	g.Edges = make([]GraphEdge, 0)
	entityIndex := make(map[string]int)
	sourceIndex := make(map[string]int)
	findingIndex := make(map[string]int)

	addEntity := func(kind, label, value, srcPath string, sev Severity) string {
		id := entityID(kind, value)
		if idx, ok := entityIndex[id]; ok {
			if sev > g.Entities[idx].Severity {
				g.Entities[idx].Severity = sev
			}
			return id
		}
		idx := len(g.Entities)
		entityIndex[id] = idx
		g.Entities = append(g.Entities, GraphEntity{
			ID: id, Kind: kind, Label: label, Value: value,
			Severity: sev, SourcePath: srcPath,
		})
		return id
	}

	addSource := func(path string, line int, lineText string) string {
		id := sourceID(path, line)
		if _, ok := sourceIndex[id]; ok {
			return id
		}
		idx := len(g.Sources)
		sourceIndex[id] = idx
		g.Sources = append(g.Sources, GraphSourceNode{
			ID: id, SourcePath: path, SourceLine: line, LineText: lineText,
		})
		return id
	}

	addFinding := func(cat, msg, srcPath, srcKey string, srcLine int, ev string, evtCount int, sev Severity) string {
		id := findingID(cat, msg)
		if _, ok := findingIndex[id]; ok {
			return id
		}
		idx := len(g.Findings)
		findingIndex[id] = idx
		g.Findings = append(g.Findings, GraphFindingNode{
			ID: id, Category: cat, Message: msg, Severity: sev,
			SourcePath: srcPath, SourceKey: srcKey, SourceLine: srcLine,
			Evidence: ev, EventCount: evtCount,
		})
		return id
	}

	addEdge := func(from, to, kind string) {
		g.Edges = append(g.Edges, GraphEdge{From: from, To: to, Kind: kind})
	}

	// 1. Top-level entities from ReportData
	if r.AdDomain != "" && r.AdDomain != "Not configured" {
		addEntity("domain", "AD Domain", r.AdDomain, "", SevWarning)
	}
	if r.KerberosRealm != "" && r.KerberosRealm != "Not configured" {
		addEntity("realm", "Kerberos Realm", r.KerberosRealm, "", SevWarning)
	}
	if r.Hostname != "" && r.Hostname != "Unknown" {
		addEntity("hostname", "Hostname", r.Hostname, "", SevWarning)
	}
	if r.SearchDomain != "" && r.SearchDomain != "None" {
		addEntity("dns", "DNS Search Domain", r.SearchDomain, "resolv.conf", SevWarning)
	}

	// 2. ConfigFindings
	for i := range r.ConfigFindings {
		f := &r.ConfigFindings[i]
		fID := addFinding(f.Category, f.Message, f.SourcePath, f.SourceKey, f.SourceLine, f.Evidence, 0, f.Severity)

		if f.Evidence != "" {
			if m := entityValueRegex.FindStringSubmatch(f.Evidence); m != nil {
				key := strings.ToLower(m[1])
				val := m[2]
				switch {
				case strings.Contains(key, "domain"):
					eID := addEntity("domain", "AD Domain", val, f.SourcePath, f.Severity)
					addEdge(eID, fID, "entity_to_finding")
				case strings.Contains(key, "realm"):
					eID := addEntity("realm", "Kerberos Realm", val, f.SourcePath, f.Severity)
					addEdge(eID, fID, "entity_to_finding")
				case strings.Contains(key, "server") || strings.Contains(key, "hostname"):
					eID := addEntity("ad_server", "AD Server", val, f.SourcePath, f.Severity)
					addEdge(eID, fID, "entity_to_finding")
				case strings.Contains(key, "search"):
					eID := addEntity("dns", "DNS Search", val, f.SourcePath, f.Severity)
					addEdge(eID, fID, "entity_to_finding")
				}
			}
		}

		if f.SourcePath != "" && f.SourceLine > 0 {
			sID := addSource(f.SourcePath, f.SourceLine, f.Evidence)
			addEdge(fID, sID, "finding_to_source")
		}
	}

	// 3. Temporal clusters
	for i := range r.TemporalClusters {
		c := &r.TemporalClusters[i]
		msg := fmt.Sprintf("Temporal burst: %s", c.Description)
		fID := addFinding("temporal", msg, "", "", 0, c.SampleRawLog, c.EventCount, SevWarning)
		for ei := range g.Entities {
			ent := &g.Entities[ei]
			if ent.Value != "" && strings.Contains(strings.ToLower(c.Description), strings.ToLower(ent.Value)) {
				addEdge(ent.ID, fID, "entity_to_finding")
			}
		}
	}

	// 4. SSSD log errors
	for i := range r.SSSDLogErrors {
		e := &r.SSSDLogErrors[i]
		fID := addFinding("log_error", e.Description, "", "", 0, "", 0, SevWarning)
		for ei := range g.Entities {
			ent := &g.Entities[ei]
			if ent.Value != "" && strings.Contains(strings.ToLower(e.Description), strings.ToLower(ent.Value)) {
				addEdge(ent.ID, fID, "entity_to_finding")
			}
		}
	}

	// 5. KB suggestions
	for i := range r.KBSuggestions {
		kb := &r.KBSuggestions[i]
		msg := fmt.Sprintf("KB suggestion: %s (%.0f%%)", kb.Title, kb.Score*100)
		addFinding("kb_suggestion", msg, "", "", 0, kb.SampleLine, 0, SevWarning)
	}

	return g
}
