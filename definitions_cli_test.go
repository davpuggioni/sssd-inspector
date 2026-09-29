// definitions_cli_test.go
//
// M1 tests for the CLI face of the Definitions Studio (definitions_cli.go),
// which both binaries route through. They pin the exit-code contract a CI gate
// depends on: 0 = definitions are fine, 1 = something would be skipped or the
// command failed, 2 = usage error.
package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"sssd-inspector/constants"
)

func boolPtr(v bool) *bool       { return &v }
func stringPtr(v string) *string { return &v }

// optsWith builds a cliOptions with only the requested flag "set"; every other
// definitions flag stays at its zero value, like a real parse would leave it.
func optsWith(definitionsInfo, validateRules bool, rulesTest string) *cliOptions {
	return &cliOptions{
		DefinitionsInfo: boolPtr(definitionsInfo),
		ValidateRules:   boolPtr(validateRules),
		RulesTest:       stringPtr(rulesTest),
	}
}

// TestProvidedFlags: the scan must see flags written after a positional path,
// with "=" values and with a double dash — flag.Parse stops at the first
// non-flag argument, so it cannot answer this question on its own.
func TestProvidedFlags(t *testing.T) {
	args := []string{
		"/tmp/supportconfig", // positional path: parsing stops here
		"-" + constants.FlagRulesTest + "=/tmp/sc",
		"--" + constants.FlagValidateRules,
		"-not-a-real-flag",
	}
	provided := providedFlags(args)

	if !provided[constants.FlagRulesTest] {
		t.Errorf("provided = %v, want %s", provided, constants.FlagRulesTest)
	}
	if !provided[constants.FlagValidateRules] {
		t.Errorf("provided = %v, want %s", provided, constants.FlagValidateRules)
	}
	if provided["not-a-real-flag"] {
		t.Errorf("provided = %v, must not contain unknown flags", provided)
	}
	if len(providedFlags(nil)) != 0 {
		t.Error("providedFlags(nil) must be empty")
	}
}

// TestRunDefinitionsCommandUnhandled: with none of the flags set the dispatcher
// must hand control back (otherwise, in the hybrid binary, the GUI would never
// start).
func TestRunDefinitionsCommandUnhandled(t *testing.T) {
	var stdout, stderr bytes.Buffer

	handled, code := runDefinitionsCommand(optsWith(false, false, ""), nil, &stdout, &stderr)
	if handled || code != 0 {
		t.Errorf("handled = %v, code = %d, want false/0", handled, code)
	}

	handled, code = runDefinitionsCommand(nil, nil, &stdout, &stderr)
	if handled || code != 0 {
		t.Errorf("with nil options: handled = %v, code = %d, want false/0", handled, code)
	}
}

// TestRunDefinitionsCommandInfo: -definitions-info prints the search paths.
func TestRunDefinitionsCommandInfo(t *testing.T) {
	wd := withTempWorkingDir(t)
	writeFile(t, wd, "rules.yaml", validRulesYAML)

	var stdout, stderr bytes.Buffer
	handled, code := runDefinitionsCommand(optsWith(true, false, ""), providedFlags([]string{"-definitions-info"}), &stdout, &stderr)
	if !handled || code != 0 {
		t.Fatalf("handled = %v, code = %d, want true/0 (stderr: %s)", handled, code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{"Definition search paths", "user scope:", "Rules loaded:", filepath.Join(wd, "rules.yaml")} {
		if !strings.Contains(out, want) {
			t.Errorf("output is missing %q:\n%s", want, out)
		}
	}
}

// TestRunDefinitionsCommandValidateOK: a machine with valid definitions exits 0.
func TestRunDefinitionsCommandValidateOK(t *testing.T) {
	wd := withTempWorkingDir(t)
	writeFile(t, wd, "rules.yaml", validRulesYAML)

	var stdout, stderr bytes.Buffer
	handled, code := runDefinitionsCommand(optsWith(false, true, ""), nil, &stdout, &stderr)
	if !handled || code != 0 {
		t.Fatalf("handled = %v, code = %d, want true/0 (stderr: %s)", handled, code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "[ok]") {
		t.Errorf("output = %q, want an [ok] verdict", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want no error output for valid definitions", stderr.String())
	}
}

// TestRunDefinitionsCommandValidateProblems: a rule the analysis would skip must
// fail the gate — that is the whole point of using it in CI.
func TestRunDefinitionsCommandValidateProblems(t *testing.T) {
	wd := withTempWorkingDir(t)
	writeFile(t, wd, "rules.yaml", `rules:
  - name: "broken"
    severity: "nonsense"
    message: "m"
    patterns: ["p"]
`)

	var stdout, stderr bytes.Buffer
	handled, code := runDefinitionsCommand(optsWith(false, true, ""), nil, &stdout, &stderr)
	if !handled || code != 1 {
		t.Fatalf("handled = %v, code = %d, want true/1 (stdout: %s)", handled, code, stdout.String())
	}
	if !strings.Contains(stdout.String(), "PROBLEMS") {
		t.Errorf("stdout = %q, want a PROBLEMS verdict", stdout.String())
	}
	if !strings.Contains(stderr.String(), "would skip") {
		t.Errorf("stderr = %q, want an explanation of the non-zero exit", stderr.String())
	}
}

// TestRunDefinitionsCommandRulesTest: the dry-run reports matches and evidence.
func TestRunDefinitionsCommandRulesTest(t *testing.T) {
	_, supportconfig := rulesTestFixture(t)

	var stdout, stderr bytes.Buffer
	handled, code := runDefinitionsCommand(optsWith(false, false, supportconfig), nil, &stdout, &stderr)
	if !handled || code != 0 {
		t.Fatalf("handled = %v, code = %d, want true/0 (stderr: %s)", handled, code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{"Rule dry-run against", "1 would fire", "[YES] fires", "evidence:"} {
		if !strings.Contains(out, want) {
			t.Errorf("output is missing %q:\n%s", want, out)
		}
	}
}

// TestRunDefinitionsCommandRulesTestErrors: an empty value is a usage error
// (exit 2) and an unreadable target is a failure (exit 1) — never a silent
// success, which would let a broken definition pass a pipeline check.
func TestRunDefinitionsCommandRulesTestErrors(t *testing.T) {
	wd := withTempWorkingDir(t)

	var stdout, stderr bytes.Buffer
	provided := map[string]bool{constants.FlagRulesTest: true}
	handled, code := runDefinitionsCommand(optsWith(false, false, ""), provided, &stdout, &stderr)
	if !handled || code != 2 {
		t.Errorf("empty -rules-test value: handled = %v, code = %d, want true/2", handled, code)
	}
	if !strings.Contains(stderr.String(), "requires the path") {
		t.Errorf("stderr = %q, want the usage error", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	missing := filepath.Join(wd, "does-not-exist")
	handled, code = runDefinitionsCommand(optsWith(false, false, missing), nil, &stdout, &stderr)
	if !handled || code != 1 {
		t.Errorf("missing target: handled = %v, code = %d, want true/1", handled, code)
	}
	if !strings.Contains(stderr.String(), "dry-run failed") {
		t.Errorf("stderr = %q, want the dry-run failure", stderr.String())
	}
}

// TestRenderRuleValidation pins the gate's rendering, including the KB
// diagnostics that must fail it just like a broken rule does.
func TestRenderRuleValidation(t *testing.T) {
	text, ok := RenderRuleValidation([]RuleValidationResult{{Label: "/tmp/good.yaml", Valid: true, RuleCount: 2}}, nil)
	if !ok || !strings.Contains(text, "[ok] /tmp/good.yaml: 2 rule(s) loaded") {
		t.Errorf("ok = %v, text = %q", ok, text)
	}

	bad := []RuleValidationResult{{
		Label:     "/tmp/bad.yaml",
		RuleCount: 1,
		Diagnostics: []Diagnostic{{
			File: "/tmp/bad.yaml", Line: 7, Message: "rule \"x\": invalid severity \"shout\", skipped",
		}},
	}}
	text, ok = RenderRuleValidation(bad, nil)
	if ok {
		t.Error("ok = true although a file has problems")
	}
	if !strings.Contains(text, "/tmp/bad.yaml:7") {
		t.Errorf("text = %q, want the file:line location", text)
	}

	text, ok = RenderRuleValidation(nil, nil)
	if !ok || !strings.Contains(text, "No rules files found") {
		t.Errorf("ok = %v, text = %q, want the empty-installation message", ok, text)
	}

	text, ok = RenderRuleValidation([]RuleValidationResult{{Label: "/tmp/good.yaml", Valid: true, RuleCount: 1}},
		[]Diagnostic{{File: "/tmp/kb_articles/broken.json", Line: 3, Message: "invalid KB article JSON"}})
	if ok {
		t.Error("ok = true although a KB article is malformed")
	}
	if !strings.Contains(text, "/tmp/kb_articles/broken.json:3") {
		t.Errorf("text = %q, want the KB diagnostic", text)
	}
}

// TestRenderRuleTest pins the dry-run rendering: a matched rule shows its
// evidence, an unmatched one is visibly marked as such.
func TestRenderRuleTest(t *testing.T) {
	res := RuleTestResult{
		TargetPath: "/sc",
		Total:      2,
		Matched:    1,
		Outcomes: []RuleTestOutcome{
			{Rule: AnalysisRule{Name: "fires", Severity: "critical"}, File: "/r/rules.yaml", Line: 2, Scope: ScopeUser,
				Matched: true, Evidence: "rc4-hmac", Message: "[RULE: fires] boom"},
			{Rule: AnalysisRule{Name: "quiet", Severity: "warning"}, File: "/r/rules.yaml", Line: 6, Scope: ScopeUser},
		},
	}
	out := RenderRuleTest(res)
	for _, want := range []string{
		"Rule dry-run against /sc",
		"2 rule(s) loaded, 1 would fire",
		"[YES] fires",
		"evidence: rc4-hmac",
		"[no ] quiet",
		"/r/rules.yaml:2 (user)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output is missing %q:\n%s", want, out)
		}
	}
}

// TestDefinitionRenderHelpers covers the small formatting pieces the CLI output
// depends on.
func TestDefinitionRenderHelpers(t *testing.T) {
	if got := definitionStatus(DefinitionFileInfo{Kind: KindRules, Exists: true, RuleCount: 3}); got != "3 rule(s)" {
		t.Errorf("definitionStatus(rules) = %q", got)
	}
	if got := definitionStatus(DefinitionFileInfo{Kind: KindKBArticles, Exists: true, ArticleCount: 5}); got != "5 article(s)" {
		t.Errorf("definitionStatus(kb) = %q", got)
	}
	if got := definitionStatus(DefinitionFileInfo{}); got != "not present" {
		t.Errorf("definitionStatus(missing) = %q", got)
	}
	if writableLabel(true) != "writable" || writableLabel(false) != "read-only" {
		t.Error("writableLabel mapping is wrong")
	}
	if orNone("") != "(unavailable)" || orNone("/x") != "/x" {
		t.Error("orNone mapping is wrong")
	}
	if indentLines("", "  ") != "" {
		t.Error("indentLines(\"\") must stay empty")
	}
	if got := indentLines("a\nb\n", "\t"); got != "\ta\n\tb\n" {
		t.Errorf("indentLines = %q", got)
	}
	if formatModTime(time.Time{}) != "" {
		t.Error("formatModTime(zero) must be empty")
	}
	if formatModTime(time.Unix(0, 0)) == "" {
		t.Error("formatModTime must render a real timestamp")
	}
	if _, _, err := scopeRoot("nonsense"); err == nil {
		t.Error("scopeRoot accepted an unknown scope")
	}
	if scopeRootLabel := ScopeSystem.Label(); scopeRootLabel != "system-wide" {
		t.Errorf("ScopeSystem.Label() = %q", scopeRootLabel)
	}
	if got := DefinitionScope("custom").Label(); got != "custom" {
		t.Errorf("unknown scope label = %q, want the raw value", got)
	}
}
