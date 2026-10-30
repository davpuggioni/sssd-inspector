// golden_scenarios_test.go
//
// The golden-scenario corpus (Percorso A).
//
// Unlike the unit tests, which assert that a function does what its author
// believed, these scenarios assert that the ANALYSIS reports what a human says
// is actually wrong with a host. That makes the false-positive rate measurable
// — the number that decides whether a support engineer trusts the tool — and it
// catches the class of bug unit tests structurally cannot: a detector that
// fires when it should not, or stops firing on a real-shaped configuration.
//
// Authoring workflow:
//
//  1. create testdata/scenarios/<id>/input/ with the supportconfig-shaped files
//     of a host you actually diagnosed, injecting ONE known defect;
//  2. go test -run TestGoldenScenarios -update-golden
//  3. REVIEW THE DIFF. This step is the real work: you are deciding whether
//     each reported finding is genuinely justified. Blindly accepting the
//     generated manifest would encode today's bugs as "expected".
//
// Run with -v to print the corpus-wide precision/recall table.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

var updateGolden = flag.Bool("update-golden", false,
	"regenerate the findings section of every scenario's expect.yaml from the current analysis output")

// scenariosRoot is where the corpus lives.
const scenariosRoot = "testdata/scenarios"

func discoverScenarios(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(scenariosRoot)
	if err != nil {
		t.Fatalf("no scenario corpus at %s: %v", scenariosRoot, err)
	}
	var dirs []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(scenariosRoot, e.Name(), "expect.yaml")); err != nil {
			continue
		}
		dirs = append(dirs, e.Name())
	}
	sort.Strings(dirs)
	if len(dirs) == 0 {
		t.Fatalf("no scenarios found under %s", scenariosRoot)
	}
	return dirs
}

// runScenario analyses one scenario input and returns the report.
func runScenario(t *testing.T, id string) ReportData {
	t.Helper()
	input := filepath.Join(scenariosRoot, id, "input")
	if _, err := os.Stat(input); err != nil {
		t.Fatalf("scenario %s has no input/ directory: %v", id, err)
	}
	globalFileCache.Clear()
	return analyzeData(context.Background(), input, false, nil)
}

// renderGolden rewrites the findings block from the actual report, keeping the
// hand-written parts (title, notes, must_not_fire, known_gaps) and any
// severity the author already pinned deliberately.
func renderGolden(exp *ScenarioExpect, rep *ReportData) *ScenarioExpect {
	prior := map[string]string{}
	for _, f := range exp.Findings {
		prior[f.RuleID] = f.Severity
	}
	counts := map[string]int{}
	for i := range rep.ConfigFindings {
		counts[rep.ConfigFindings[i].RuleID]++
	}
	var order []string
	for id := range counts {
		order = append(order, id)
	}
	sort.Strings(order)

	out := &ScenarioExpect{
		ID:          exp.ID,
		Title:       exp.Title,
		Notes:       exp.Notes,
		Problems:    exp.Problems,
		Warnings:    exp.Warnings,
		MustNotFire: exp.MustNotFire,
		KnownGaps:   exp.KnownGaps,
	}
	for _, id := range order {
		sev := ""
		for i := range rep.ConfigFindings {
			if rep.ConfigFindings[i].RuleID == id {
				sev = severityName(rep.ConfigFindings[i].Severity)
				break
			}
		}
		if p, ok := prior[id]; ok && p != "" {
			sev = p
		}
		out.Findings = append(out.Findings, ExpectedFinding{
			RuleID: id, Severity: sev, MinCount: counts[id],
		})
	}
	return out
}

// scenarioScore accumulates the corpus-wide confusion counts.
//
// Counting is per EXPECTATION, not per finding instance: a per-domain rule that
// legitimately fires three times is one expectation, not two false positives.
// A false positive is a whole RULE the scenario never asked for, which is the
// failure that actually erodes trust.
type scenarioScore struct {
	tp, fp, fn int
	byKind     map[string]int
	gaps       []string
	debt       []string
}

func newScenarioScore() *scenarioScore {
	return &scenarioScore{byKind: map[string]int{}}
}

func (s *scenarioScore) add(m scenarioMismatch) {
	s.byKind[m.Kind]++
	switch m.Kind {
	case "false_positive", "forbidden_fired", "contract_violation":
		s.fp++
	case "false_negative":
		s.fn++
	}
}

func (s *scenarioScore) precision() float64 {
	if s.tp+s.fp == 0 {
		return 1
	}
	return float64(s.tp) / float64(s.tp+s.fp)
}

func (s *scenarioScore) recall() float64 {
	if s.tp+s.fn == 0 {
		return 1
	}
	return float64(s.tp) / float64(s.tp+s.fn)
}

// countSatisfied returns how many expectation entries the report meets.
func countSatisfied(exp *ScenarioExpect, rep *ReportData) int {
	n := 0
	for _, f := range exp.Findings {
		want := f.MinCount
		if want == 0 {
			want = 1
		}
		if f.Severity == "" {
			if countRule(rep, f.RuleID) >= want {
				n++
			}
			continue
		}
		if countRuleAt(rep, f.RuleID, scenarioSeverityNames[strings.ToLower(f.Severity)]) >= want {
			n++
		}
	}
	return n
}

func TestGoldenScenarios(t *testing.T) {
	ids := discoverScenarios(t)
	score := newScenarioScore()
	failing := 0

	for _, id := range ids {
		id := id
		t.Run(id, func(t *testing.T) {
			expectPath := filepath.Join(scenariosRoot, id, "expect.yaml")
			exp, err := loadScenarioExpect(expectPath)
			if err != nil {
				t.Fatalf("%v", err)
			}
			rep := runScenario(t, id)

			if *updateGolden {
				data, err := yaml.Marshal(renderGolden(exp, &rep))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(expectPath, data, 0644); err != nil {
					t.Fatal(err)
				}
				t.Logf("rewrote %s — REVIEW THE DIFF before committing", expectPath)
				return
			}

			mismatches := compareScenario(exp, &rep)
			score.tp += countSatisfied(exp, &rep)
			for _, g := range exp.KnownGaps {
				score.gaps = append(score.gaps, fmt.Sprintf("%s/%s: %s", id, g.RuleID, g.Reason))
			}

			// A "redundant" mismatch is a rule that fires but only repeats
			// another one. The finding is not wrong, so it must not fail the
			// run: it is surfaced as debt, not as an error.
			var hard, soft []scenarioMismatch
			for _, m := range mismatches {
				if m.Kind == "redundant" {
					soft = append(soft, m)
					continue
				}
				hard = append(hard, m)
			}
			if len(soft) > 0 {
				for _, m := range soft {
					score.debt = append(score.debt, fmt.Sprintf("%s/%s", id, m.Finding))
				}
			}
			if len(hard) == 0 {
				return
			}

			failing++
			var b strings.Builder
			fmt.Fprintf(&b, "scenario %q — %s\n", exp.ID, exp.Title)
			if exp.Notes != "" {
				fmt.Fprintf(&b, "  injected cause: %s\n", exp.Notes)
			}
			for _, m := range hard {
				if m.Finding != "" {
					fmt.Fprintf(&b, "  [%s] %s: %s\n", m.Kind, m.Finding, m.Detail)
				} else {
					fmt.Fprintf(&b, "  [%s] %s\n", m.Kind, m.Detail)
				}
			}
			t.Error(b.String())
			for _, m := range hard {
				score.add(m)
			}
		})
	}

	if *updateGolden {
		return
	}

	t.Logf("scenarios=%d satisfied=%d false_positives=%d false_negatives=%d "+
		"precision=%.3f recall=%.3f",
		len(ids), score.tp, score.fp, score.fn, score.precision(), score.recall())
	for kind, n := range score.byKind {
		t.Logf("  %s: %d", kind, n)
	}
	for _, g := range score.gaps {
		t.Logf("  KNOWN GAP %s", g)
	}
	for _, d := range score.debt {
		t.Logf("  REDUNDANT %s", d)
	}
	if failing > 0 {
		t.Errorf("%d scenario(s) diverged from their ground truth", failing)
	}
	if len(score.gaps) > 0 || len(score.debt) > 0 {
		t.Logf("%d known gap(s) and %d redundancy(ies) tolerated: they are tracked debt, not correctness",
			len(score.gaps), len(score.debt))
	}
}
