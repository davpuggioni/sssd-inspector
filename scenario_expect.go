// scenario_expect.go
//
// Ground-truth manifest for the golden-scenario corpus.
//
// A scenario is a directory under testdata/scenarios/ holding an "input"
// subtree laid out exactly like an extracted supportconfig, plus an
// expect.yaml describing what the analysis MUST report. The author knows which
// defect they injected, so the expectation is ground truth by construction —
// which is what finally makes a false-positive RATE measurable, something the
// unit tests cannot do on their own.
//
// The matching key is the RuleID, never the message: messages get reworded,
// RuleIDs are the stable contract (types.go requires one on every finding).
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// scenarioSeverityNames maps the manifest spelling to the typed Severity.
var scenarioSeverityNames = map[string]Severity{
	"warning":  SevWarning,
	"error":    SevError,
	"critical": SevCritical,
}

// ExpectedFinding asserts one finding the scenario expects to be reported.
type ExpectedFinding struct {
	RuleID string `yaml:"rule_id"`
	// Severity is optional; when empty the rule must fire at any severity.
	Severity string `yaml:"severity,omitempty"`
	// MinCount defaults to 1 (a per-domain rule may legitimately fire more).
	MinCount int `yaml:"min_count,omitempty"`
}

// KnownGap records a finding that SHOULD be reported but is not yet: visible
// debt. It does not fail the suite, but it is printed and counted, so it
// cannot be forgotten.
type KnownGap struct {
	RuleID string `yaml:"rule_id"`
	Reason string `yaml:"reason"`
}

// ScenarioExpect is the parsed expect.yaml of one scenario.
type ScenarioExpect struct {
	ID    string `yaml:"id"`
	Title string `yaml:"title"`
	Notes string `yaml:"notes,omitempty"`

	Findings    []ExpectedFinding `yaml:"findings,omitempty"`
	Problems    []string          `yaml:"problems,omitempty"`
	Warnings    []string          `yaml:"warnings,omitempty"`
	MustNotFire []string          `yaml:"must_not_fire,omitempty"`
	// KnownRedundancies lists rules that DO fire here but duplicate another
	// rule in the same scenario — the same defect reported twice, usually at
	// two different severities. Unlike must_not_fire these are tolerated: the
	// suite does not fail, because the finding is not wrong, only redundant.
	// They are tracked debt, because double-reporting one misconfiguration is
	// noise the reader pays for on every run.
	KnownRedundancies []KnownGap `yaml:"known_redundancies,omitempty"`
	KnownGaps         []KnownGap `yaml:"known_gaps,omitempty"`
}

// loadScenarioExpect reads and validates one expect.yaml.
func loadScenarioExpect(path string) (*ScenarioExpect, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var e ScenarioExpect
	if err := yaml.Unmarshal(data, &e); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if e.ID == "" {
		return nil, fmt.Errorf("%s: 'id' is required", path)
	}
	if e.Title == "" {
		return nil, fmt.Errorf("%s: 'title' is required (a scenario nobody can review is not ground truth)", path)
	}
	for i := range e.Findings {
		f := &e.Findings[i]
		if f.RuleID == "" {
			return nil, fmt.Errorf("%s: findings[%d] has no rule_id", path, i)
		}
		if f.Severity != "" {
			if _, ok := scenarioSeverityNames[strings.ToLower(f.Severity)]; !ok {
				return nil, fmt.Errorf("%s: findings[%d] severity %q is not one of warning|error|critical", path, i, f.Severity)
			}
		}
		if f.MinCount < 0 {
			return nil, fmt.Errorf("%s: findings[%d] min_count must be >= 0", path, i)
		}
	}
	for _, id := range expMustNotFire(&e) {
		if id == "" {
			return nil, fmt.Errorf("%s: must_not_fire contains an empty rule_id", path)
		}
	}
	for i, g := range e.KnownGaps {
		if g.RuleID == "" || g.Reason == "" {
			return nil, fmt.Errorf("%s: known_gaps[%d] needs both rule_id and reason", path, i)
		}
	}
	for i, g := range e.KnownRedundancies {
		if g.RuleID == "" || g.Reason == "" {
			return nil, fmt.Errorf("%s: known_redundancies[%d] needs both rule_id and reason", path, i)
		}
	}
	return &e, nil
}

func expMustNotFire(e *ScenarioExpect) []string { return e.MustNotFire }

// scenarioMismatch describes one difference between expectation and reality.
type scenarioMismatch struct {
	Kind    string // false_positive | false_negative | severity_mismatch | forbidden_fired | contract_violation
	Finding string
	Detail  string
}

// compareScenario computes the mismatches between the expectation and a real
// report. Comparison is order-insensitive on purpose: a golden scenario must
// not become flaky because findings are emitted in a different order, and
// ordering regressions are already covered by the determinism tests.
//
// Rules:
//   - a false positive is a reported finding no expectation covers;
//   - a false negative is an unmet expectation;
//   - must_not_fire catches documented near-misses that must stay silent.
func compareScenario(exp *ScenarioExpect, rep *ReportData) []scenarioMismatch {
	var out []scenarioMismatch

	// --- expectations -------------------------------------------------------
	satisfied := make([]bool, len(exp.Findings))
	for i, f := range exp.Findings {
		want := f.MinCount
		if want == 0 {
			want = 1
		}
		if f.Severity == "" {
			if n := countRule(rep, f.RuleID); n < want {
				out = append(out, scenarioMismatch{
					Kind:    "false_negative",
					Finding: f.RuleID,
					Detail:  fmt.Sprintf("expected >=%d occurrence(s) at any severity, got %d", want, n),
				})
			} else {
				satisfied[i] = true
			}
			continue
		}
		wantSev := scenarioSeverityNames[strings.ToLower(f.Severity)]
		if n := countRuleAt(rep, f.RuleID, wantSev); n < want {
			// Distinguish "not reported at all" from "reported at the wrong
			// severity": the second is a severity regression and worth seeing
			// distinctly in the failure output.
			if other := countRule(rep, f.RuleID); other > 0 {
				out = append(out, scenarioMismatch{
					Kind:    "severity_mismatch",
					Finding: f.RuleID,
					Detail: fmt.Sprintf("expected %s x%d, got: %s",
						f.Severity, want, describeSeverities(rep, f.RuleID)),
				})
			} else {
				out = append(out, scenarioMismatch{
					Kind:    "false_negative",
					Finding: f.RuleID,
					Detail:  fmt.Sprintf("expected %s x%d, got 0", f.Severity, want),
				})
			}
		} else {
			satisfied[i] = true
		}
	}

	// --- false positives ----------------------------------------------------
	covered := make(map[string]bool, len(exp.Findings))
	for _, f := range exp.Findings {
		covered[f.RuleID] = true
	}
	seen := map[string]bool{}
	for i := range rep.ConfigFindings {
		f := &rep.ConfigFindings[i]
		if f.RuleID == "" {
			out = append(out, scenarioMismatch{
				Kind:   "contract_violation",
				Detail: "a finding was reported with an empty RuleID (types.go requires one)",
			})
			continue
		}
		if covered[f.RuleID] || seen[f.RuleID] {
			continue
		}
		seen[f.RuleID] = true
		out = append(out, scenarioMismatch{
			Kind:    "false_positive",
			Finding: f.RuleID,
			Detail: fmt.Sprintf("reported as %s but not expected: %s",
				severityName(f.Severity), truncate(f.Message, 100)),
		})
	}

	// --- must_not_fire ------------------------------------------------------
	for _, banned := range exp.MustNotFire {
		if countRule(rep, banned) > 0 {
			out = append(out, scenarioMismatch{
				Kind:    "forbidden_fired",
				Finding: banned,
				Detail:  "listed in must_not_fire but fired anyway",
			})
		}
	}

	// --- free-text expectations --------------------------------------------
	problems := strings.Join(rep.Problems, "\n")
	warnings := strings.Join(rep.Warnings, "\n")
	for _, want := range exp.Problems {
		if !strings.Contains(problems, want) {
			out = append(out, scenarioMismatch{Kind: "false_negative", Detail: "missing from Problems: " + want})
		}
	}
	for _, want := range exp.Warnings {
		if !strings.Contains(warnings, want) {
			out = append(out, scenarioMismatch{Kind: "false_negative", Detail: "missing from Warnings: " + want})
		}
	}

	// --- tracked debt -------------------------------------------------------
	// Applied last, once every mismatch is known: a known_gap silences a false
	// negative (the rule SHOULD have fired but does not), a known_redundancy
	// downgrades a false positive to "redundant" (the rule fires but only
	// repeats what another rule already said). Both require a reason in the
	// manifest, and both only apply to the exact rule they name — an
	// over-broad entry cannot mask an unrelated regression.
	gaps := map[string]bool{}
	for _, g := range exp.KnownGaps {
		gaps[g.RuleID] = true
	}
	redundant := map[string]bool{}
	for _, g := range exp.KnownRedundancies {
		redundant[g.RuleID] = true
	}
	filtered := out[:0]
	for _, m := range out {
		switch {
		case m.Kind == "false_negative" && gaps[m.Finding]:
			continue
		case m.Kind == "false_positive" && redundant[m.Finding]:
			m.Kind = "redundant"
		}
		filtered = append(filtered, m)
	}
	out = filtered

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return out[i].Finding < out[j].Finding
	})
	return out
}

// countRule counts findings with the given RuleID.
func countRule(rep *ReportData, ruleID string) int {
	n := 0
	for i := range rep.ConfigFindings {
		if rep.ConfigFindings[i].RuleID == ruleID {
			n++
		}
	}
	return n
}

// countRuleAt counts findings with the given RuleID and Severity.
func countRuleAt(rep *ReportData, ruleID string, sev Severity) int {
	n := 0
	for i := range rep.ConfigFindings {
		if rep.ConfigFindings[i].RuleID == ruleID && rep.ConfigFindings[i].Severity == sev {
			n++
		}
	}
	return n
}

// describeSeverities renders "warning x2, error x1" for a rule, so a severity
// regression says what it actually became.
func describeSeverities(rep *ReportData, ruleID string) string {
	counts := map[Severity]int{}
	for i := range rep.ConfigFindings {
		if rep.ConfigFindings[i].RuleID == ruleID {
			counts[rep.ConfigFindings[i].Severity]++
		}
	}
	order := []Severity{SevCritical, SevError, SevWarning}
	var parts []string
	for _, s := range order {
		if counts[s] > 0 {
			parts = append(parts, fmt.Sprintf("%s x%d", severityName(s), counts[s]))
		}
	}
	if len(parts) == 0 {
		return "nothing"
	}
	return strings.Join(parts, ", ")
}

// severityName renders a Severity for reports and manifests.
func severityName(s Severity) string {
	switch s {
	case SevCritical:
		return "critical"
	case SevError:
		return "error"
	default:
		return "warning"
	}
}

// truncate shortens s for a one-line report.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
