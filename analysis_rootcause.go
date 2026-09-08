// analysis_rootcause.go
// Root-cause breakdown for the report (P4a): groups every diagnostic signal
// (structured config findings and log-error categories) under the coarse
// root-cause domain it points to, so the report reads as "one root cause, many
// symptoms" instead of a flat list. This is the same correlation mindset as
// correlateSequences, applied to the whole report rather than a single burst.
package main

import (
	"fmt"
	"sort"
)

// rootCauseGroup is one row of the root-cause breakdown: a coarse SSSD failure
// domain (dns, krb5, tls, offline, ...) with the config and log signals that
// point to it and a weighted strength.
type rootCauseGroup struct {
	Category      string   // coarse root-cause category, e.g. "dns"
	Severity      int      // weighted strength (higher = more likely root)
	ConfigSignals []string // human-readable messages from config findings
	LogSignals    []string // SSDLLogError descriptions pointing at this domain
}

// rootCauseSeverityWeight assigns an ordinal weight to each severity so the
// breakdown can rank rows.
var rootCauseSeverityWeight = map[Severity]int{
	SevWarning:  1,
	SevError:    2,
	SevCritical: 3,
}

// severityTag returns a compact severity label for the breakdown table.
func severityTag(sev Severity) string {
	switch sev {
	case SevCritical:
		return "CRITICAL"
	case SevError:
		return "ERROR"
	default:
		return "WARN"
	}
}

// rootCauseBreakdown groups every signal in the report under its coarse
// root-cause category. Rows are ordered by weighted strength descending so the
// most-likely root cause appears first in the report.
func rootCauseBreakdown(r ReportData) []rootCauseGroup {
	groups := make(map[string]*rootCauseGroup)
	push := func(cat, configMsg, logMsg string, weight int) {
		g, ok := groups[cat]
		if !ok {
			nw := &rootCauseGroup{Category: cat}
			groups[cat] = nw
			g = nw
		}
		g.Severity += weight
		if configMsg != "" {
			g.ConfigSignals = append(g.ConfigSignals, configMsg)
		}
		if logMsg != "" {
			g.LogSignals = append(g.LogSignals, logMsg)
		}
	}

	for _, f := range r.ConfigFindings {
		w := rootCauseSeverityWeight[f.Severity]
		if w == 0 {
			w = 1
		}
		push(f.Category, fmt.Sprintf("[%s] %s", severityTag(f.Severity), f.Message), "", w)
	}
	for _, e := range r.SSSDLogErrors {
		cat := categoryFor(e.Description)
		if cat == "general" || cat == "" {
			continue
		}
		if w, ok := logCategorySeverity[cat]; ok {
			push(cat, "", e.Description, w)
		} else {
			push(cat, "", e.Description, 1)
		}
	}

	out := make([]rootCauseGroup, 0, len(groups))
	for _, g := range groups {
		out = append(out, *g)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Severity != out[j].Severity {
			return out[i].Severity > out[j].Severity
		}
		return out[i].Category < out[j].Category
	})
	return out
}

// buildRootCauseBreakdownLines renders the breakdown as plain-text report rows.
func buildRootCauseBreakdownLines(r ReportData) []string {
	rows := rootCauseBreakdown(r)
	if len(rows) == 0 {
		return []string{" No root-cause signals beyond healthy findings."}
	}
	var lines []string
	for _, row := range rows {
		hdr := fmt.Sprintf("[%s] %s (weight %d)", severityTagRow(row.Severity), row.Category, row.Severity)
		lines = append(lines, " "+hdr)
		for _, c := range row.ConfigSignals {
			lines = append(lines, "     config: "+c)
		}
		for _, l := range row.LogSignals {
			lines = append(lines, "     log:    "+l)
		}
	}
	return lines
}

// severityTagRow maps a numeric weight ceiling to a coarse severity label for
// the text breakdown header.
func severityTagRow(weight int) string {
	switch {
	case weight >= 4:
		return "HIGH"
	case weight >= 2:
		return "MED"
	default:
		return "LOW"
	}
}
