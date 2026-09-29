// definitions_service_test.go
//
// M1 regression tests for the Definitions Studio service layer
// (definitions_service.go). The contract they pin is the one a support
// engineer relies on:
//
//   - ListDefinitions tells the truth about where the inspector looks;
//   - ValidateRuleYAML agrees with the loader (same bytes → same verdict),
//     including the silently-dead cases ("no rules: header", duplicate names,
//     nameless rules) that M0 could only report after the fact;
//   - SaveRuleYAML never leaves the user with a rules file that silently does
//     nothing, and keeps the previous one as .bak;
//   - TestRulesAgainst uses the real matcher, so "it fires here" means "it
//     fires in the report".
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setSystemDefinitionsRoot mirrors setUserDefinitionsRoot (see
// definitions_diagnostics_test.go) for the system-wide root.
func setSystemDefinitionsRoot(t *testing.T, dir string) {
	t.Helper()
	prev := systemDefinitionsRoot
	systemDefinitionsRoot = func() string { return dir }
	t.Cleanup(func() { systemDefinitionsRoot = prev })
}

// chdirForTest chdirs into dir and restores the previous working directory.
func chdirForTest(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(old); err != nil {
			t.Errorf("restore wd: %v", err)
		}
	})
}

// validRulesYAML is a minimal valid rules document used across these tests.
const validRulesYAML = `rules:
  - name: "legacy-rc4"
    severity: "warning"
    category: "crypto"
    files: ["sssd.conf"]
    message: "Legacy RC4 enctype found"
    patterns: ["rc4-hmac"]
`

// TestListDefinitionsReportsEverySearchPath pins the answer to the question the
// whole feature exists for: "where does the inspector look for my rules?".
func TestListDefinitionsReportsEverySearchPath(t *testing.T) {
	wd := withTempWorkingDir(t)
	writeFile(t, wd, "rules.yaml", validRulesYAML)

	userRoot := t.TempDir()
	writeFile(t, userRoot, "rules.yaml", `rules:
  - name: "user-rule"
    severity: "error"
    message: "user scope message"
    patterns: ["needle-in-user-scope"]
`)
	setUserDefinitionsRoot(t, userRoot)

	inv := ListDefinitions()

	if inv.UserRoot != userRoot {
		t.Errorf("UserRoot = %q, want %q", inv.UserRoot, userRoot)
	}
	if inv.RuleCount != 2 {
		t.Fatalf("RuleCount = %d, want 2 (working-dir rule + user rule)", inv.RuleCount)
	}
	if inv.ArticleCount == 0 {
		t.Error("ArticleCount = 0: the embedded KB corpus must be counted")
	}

	// Provenance: every loaded rule knows the file, line and scope it came from.
	names := map[DefinitionScope]string{}
	for _, ri := range inv.Rules {
		names[ri.Scope] = ri.Rule.Name
		if ri.Line <= 0 {
			t.Errorf("rule %q: Line = %d, want the source line", ri.Rule.Name, ri.Line)
		}
		if !filepath.IsAbs(ri.File) {
			t.Errorf("rule %q: File = %q, want an absolute path the GUI can open", ri.Rule.Name, ri.File)
		}
	}
	if names[ScopeWorkDir] != "legacy-rc4" {
		t.Errorf("working-dir rule = %q, want legacy-rc4 (provenance: %v)", names[ScopeWorkDir], names)
	}
	if names[ScopeUser] != "user-rule" {
		t.Errorf("user rule = %q, want user-rule (provenance: %v)", names[ScopeUser], names)
	}

	var foundWorkDir, foundEmbedded bool
	for _, f := range inv.Files {
		if f.Kind == KindRules && f.Scope == ScopeWorkDir && f.Exists && f.RuleCount == 1 {
			foundWorkDir = true
			if !f.Writable {
				t.Error("working-dir rules.yaml reported read-only, but the test just wrote it")
			}
			if f.ModTime == "" {
				t.Error("working-dir rules.yaml has no ModTime")
			}
		}
		if f.Scope == ScopeEmbedded {
			foundEmbedded = true
			if f.Writable {
				t.Error("embedded definitions must never be reported as writable")
			}
			if !strings.HasPrefix(f.Path, "embedded://") {
				t.Errorf("embedded path = %q, want an embedded:// scheme", f.Path)
			}
		}
	}
	if !foundWorkDir {
		t.Error("inventory does not describe the working-directory rules.yaml")
	}
	if !foundEmbedded {
		t.Error("inventory does not describe the embedded KB corpus")
	}
	if len(inv.Diagnostics) != 0 {
		t.Errorf("Diagnostics = %+v, want none for valid definitions", inv.Diagnostics)
	}
}

// TestListDefinitionsDeduplicatesDiagnostics: the same bad file is seen twice
// (per-location scan + aggregate loaders) and must be reported once.
func TestListDefinitionsDeduplicatesDiagnostics(t *testing.T) {
	wd := withTempWorkingDir(t)
	writeFile(t, wd, "rules.yaml", `rules:
  - name: "bad-severity"
    severity: "shout"
    message: "nope"
    patterns: ["x"]
`)

	inv := ListDefinitions()
	if len(inv.Diagnostics) != 1 {
		t.Fatalf("Diagnostics = %d entries (%+v), want exactly 1", len(inv.Diagnostics), inv.Diagnostics)
	}
	if inv.RuleCount != 0 {
		t.Errorf("RuleCount = %d, want 0: the only rule is invalid", inv.RuleCount)
	}
	if !strings.Contains(inv.Diagnostics[0].Message, "invalid severity") {
		t.Errorf("diagnostic = %q, want the invalid-severity message", inv.Diagnostics[0].Message)
	}
}

// TestListDefinitionsFlagsCrossFileDuplicateNames: the same rule name in two
// files (a copy of ./rules.yaml dropped into the user scope) loads twice with
// indistinguishable findings — that must be reported, not silently doubled.
func TestListDefinitionsFlagsCrossFileDuplicateNames(t *testing.T) {
	wd := withTempWorkingDir(t)
	writeFile(t, wd, "rules.yaml", validRulesYAML)

	userRoot := t.TempDir()
	setUserDefinitionsRoot(t, userRoot)
	writeFile(t, userRoot, "rules.yaml", validRulesYAML) // same rule name, other scope

	inv := ListDefinitions()
	if inv.RuleCount != 2 {
		t.Fatalf("RuleCount = %d, want 2 (both copies load: rules are additive)", inv.RuleCount)
	}
	if len(inv.Diagnostics) != 1 {
		t.Fatalf("Diagnostics = %+v, want exactly one duplicate-name warning", inv.Diagnostics)
	}
	d := inv.Diagnostics[0]
	if !strings.Contains(d.Message, "legacy-rc4") || !strings.Contains(d.Message, "also defined in") {
		t.Errorf("diagnostic = %q, want it to name the rule and the other file", d.Message)
	}
	if d.Line != 2 {
		t.Errorf("diagnostic line = %d, want 2 (the second definition)", d.Line)
	}
}

// TestDefinitionScopeOf prioritises the dedicated definition roots over the
// generic "under the working directory" match: with HOME == cwd (a real setup
// when the inspector is launched from a home directory) a user rule must still
// report as the user scope.
func TestDefinitionScopeOf(t *testing.T) {
	home := t.TempDir()
	setUserDefinitionsRoot(t, filepath.Join(home, ".sssd-inspector"))
	chdirForTest(t, home)

	userFile := filepath.Join(home, ".sssd-inspector", "rules.yaml")
	if got := definitionScopeOf(userFile); got != ScopeUser {
		t.Errorf("definitionScopeOf(%q) = %q, want %q", userFile, got, ScopeUser)
	}

	wdFile := filepath.Join(home, "rules.yaml")
	if got := definitionScopeOf(wdFile); got != ScopeWorkDir {
		t.Errorf("definitionScopeOf(%q) = %q, want %q", wdFile, got, ScopeWorkDir)
	}

	if got := definitionScopeOf("/etc/sssd-inspector/rules.yaml"); got == ScopeSystem {
		t.Error("system scope matched although TestMain disables the system root")
	}
	if !sameOrUnder(filepath.Join(home, "sub", "file"), home) {
		t.Error("sameOrUnder must accept a nested path")
	}
	if sameOrUnder(home+"x/file", home) {
		t.Error("sameOrUnder must not treat a sibling prefix as nested")
	}
	if sameOrUnder(home, "") {
		t.Error("sameOrUnder must never match an empty root")
	}
}

// TestValidateRuleYAML pins the verdicts of the Studio's "Validate" button.
// Every expectation is shared with the loader (parseRulesDocument), so an
// editor that says "valid" cannot be lying about what the analysis will load.
func TestValidateRuleYAML(t *testing.T) {
	cases := []struct {
		name      string
		content   string
		wantRules int
		wantInMsg string
		wantLine  int
	}{
		{
			name:      "valid document",
			content:   validRulesYAML,
			wantRules: 1,
		},
		{
			name:    "comments-only template is a valid zero-rule document",
			content: "# template\n#  - name: \"example\"\n",
		},
		{
			name:    "empty editor content",
			content: "",
		},
		{
			name:    "rules list present but empty",
			content: "rules: []\n",
		},
		{
			name:      "rules: header forgotten (top-level sequence)",
			content:   "- name: \"x\"\n  severity: \"warning\"\n  message: \"m\"\n  patterns: [\"p\"]\n",
			wantInMsg: "top-level",
		},
		{
			name:      "wrong top-level key",
			content:   "checks:\n  - name: \"x\"\n",
			wantInMsg: "no top-level \"rules\" list found (top-level keys: checks)",
		},
		{
			name:      "rules value is not a list",
			content:   "rules: 42\n",
			wantInMsg: "the top-level \"rules\" value must be a list, found a scalar value",
		},
		{
			name:      "rules entry is not a mapping",
			content:   "rules:\n  - just-a-string\n",
			wantInMsg: "rule #1 (line 2) must be a mapping",
		},
		{
			name: "invalid severity",
			content: `rules:
  - name: "ok"
    severity: "warning"
    message: "m"
    patterns: ["p"]
  - name: "bad"
    severity: "shout"
    message: "m"
    patterns: ["p"]
`,
			wantRules: 1, // "ok" still loads; only the broken rule is skipped
			wantLine:  6,
			wantInMsg: "invalid severity",
		},
		{
			name: "nameless rule",
			content: `rules:
  - severity: "warning"
    message: "m"
    patterns: ["p"]
`,
			wantLine:  2,
			wantInMsg: "missing name",
		},
		{
			name: "missing patterns",
			content: `rules:
  - name: "no-patterns"
    severity: "warning"
    message: "m"
`,
			wantLine:  2,
			wantInMsg: "missing patterns or message",
		},
		{
			name: "duplicate rule names",
			content: `rules:
  - name: "dup"
    severity: "warning"
    message: "m"
    patterns: ["p"]
  - name: "dup"
    severity: "warning"
    message: "m"
    patterns: ["p"]
`,
			// A duplicate name is reported but NOT skipped: findings keyed by
			// the rule name would be ambiguous, which the user must know.
			wantRules: 2,
			wantLine:  6,
			wantInMsg: "duplicate name",
		},
		{
			name:      "syntactically broken YAML",
			content:   "rules:\n  - name: \"broken\"\n   severity: bad-indent\n",
			wantInMsg: "invalid rules file",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := ValidateRuleYAML(tc.content)
			wantValid := tc.wantInMsg == ""
			if res.Valid != wantValid {
				t.Errorf("Valid = %v, want %v (diagnostics: %+v)", res.Valid, wantValid, res.Diagnostics)
			}
			if res.RuleCount != tc.wantRules {
				t.Errorf("RuleCount = %d, want %d", res.RuleCount, tc.wantRules)
			}
			if res.Label != "<editor>" {
				t.Errorf("Label = %q, want <editor>", res.Label)
			}
			if tc.wantInMsg == "" {
				return
			}
			var found bool
			for _, d := range res.Diagnostics {
				if !strings.Contains(d.Message, tc.wantInMsg) {
					continue
				}
				found = true
				if tc.wantLine > 0 && d.Line != tc.wantLine {
					t.Errorf("diagnostic line = %d, want %d (%s)", d.Line, tc.wantLine, d.Message)
				}
			}
			if !found {
				t.Errorf("no diagnostic contains %q (got %+v)", tc.wantInMsg, res.Diagnostics)
			}
		})
	}
}

// assertFileContent fails the test unless path holds exactly want.
func assertFileContent(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(got) != want {
		t.Errorf("%s = %q, want %q", path, got, want)
	}
}

// TestSaveRuleYAMLUserScope covers the happy path, the .bak guarantee and the
// atomic replacement (no leftover temp file).
func TestSaveRuleYAMLUserScope(t *testing.T) {
	root := t.TempDir()
	setUserDefinitionsRoot(t, root)
	wantPath := filepath.Join(root, "rules.yaml")

	res, err := SaveRuleYAML(validRulesYAML, "user")
	if err != nil {
		t.Fatalf("SaveRuleYAML: %v", err)
	}
	if !res.Saved || res.Path != wantPath || res.Scope != ScopeUser {
		t.Fatalf("result = %+v, want a successful save into %s", res, wantPath)
	}
	if res.Bytes != len(validRulesYAML) {
		t.Errorf("Bytes = %d, want %d", res.Bytes, len(validRulesYAML))
	}
	if !res.Validation.Valid {
		t.Errorf("Validation = %+v, want the saved document to be valid", res.Validation)
	}
	assertFileContent(t, wantPath, validRulesYAML)

	// Second save: the previous content must survive as .bak.
	second := validRulesYAML + "# second revision\n"
	res2, err := SaveRuleYAML(second, "user")
	if err != nil {
		t.Fatalf("second SaveRuleYAML: %v", err)
	}
	if res2.Backup != wantPath+".bak" {
		t.Errorf("Backup = %q, want %q", res2.Backup, wantPath+".bak")
	}
	assertFileContent(t, wantPath, second)
	assertFileContent(t, wantPath+".bak", validRulesYAML)

	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("temp file %q left behind by the atomic write", e.Name())
		}
	}
}

// TestSaveRuleYAMLSystemScope: the system scope writes into its own root (in
// production /etc/sssd-inspector, writable only as root).
func TestSaveRuleYAMLSystemScope(t *testing.T) {
	root := t.TempDir()
	setSystemDefinitionsRoot(t, root)

	res, err := SaveRuleYAML(validRulesYAML, "system")
	if err != nil {
		t.Fatalf("SaveRuleYAML(system): %v", err)
	}
	if !res.Saved || res.Scope != ScopeSystem || res.Path != filepath.Join(root, "rules.yaml") {
		t.Fatalf("result = %+v, want a saved system-scope rules.yaml in %s", res, root)
	}
	assertFileContent(t, filepath.Join(root, "rules.yaml"), validRulesYAML)
}

// TestSaveRuleYAMLRefusesInvalidContent: a document the analysis would skip
// must not be written — leaving the user with definitions that silently do
// nothing is the failure mode the service exists to prevent. The existing file
// must survive untouched, with no backup and no partial write.
func TestSaveRuleYAMLRefusesInvalidContent(t *testing.T) {
	root := t.TempDir()
	setUserDefinitionsRoot(t, root)
	target := filepath.Join(root, "rules.yaml")
	writeFile(t, root, "rules.yaml", validRulesYAML)

	broken := "rules:\n  - name: \"x\"\n    severity: \"nonsense\"\n    message: \"m\"\n    patterns: [\"p\"]\n"
	res, err := SaveRuleYAML(broken, "user")
	if err != nil {
		t.Fatalf("SaveRuleYAML returned an error for an invalid document: %v", err)
	}
	if res.Saved {
		t.Fatal("Saved = true for a document the analysis would skip")
	}
	if res.Validation.Valid {
		t.Error("Validation.Valid = true for a document the analysis would skip")
	}
	if res.Path != target {
		t.Errorf("Path = %q, want the target the user tried to save to (%s)", res.Path, target)
	}
	assertFileContent(t, target, validRulesYAML)
	if _, err := os.Stat(target + ".bak"); !os.IsNotExist(err) {
		t.Error("a backup was created for a save that never happened")
	}
}

// TestSaveRuleYAMLScopeErrors: only the two scope roots are writable targets,
// and an unavailable root is an error, not a silent no-op.
func TestSaveRuleYAMLScopeErrors(t *testing.T) {
	if _, err := SaveRuleYAML(validRulesYAML, "compiled"); err == nil {
		t.Error("SaveRuleYAML accepted an unknown scope")
	} else if !strings.Contains(err.Error(), "user") || !strings.Contains(err.Error(), "system") {
		t.Errorf("error = %v, want it to name the accepted scopes", err)
	}

	setUserDefinitionsRoot(t, "")
	if _, err := SaveRuleYAML(validRulesYAML, "user"); err == nil {
		t.Error("SaveRuleYAML accepted the user scope with no HOME")
	} else if !strings.Contains(err.Error(), "per-user") {
		t.Errorf("error = %v, want it to explain that the per-user root is unknown", err)
	}

	setSystemDefinitionsRoot(t, "")
	if _, err := SaveRuleYAML(validRulesYAML, "system"); err == nil {
		t.Error("SaveRuleYAML accepted the system scope with no configured root")
	}
}

// TestSaveRuleYAMLReadOnlyDirectory: the failure must be an error (not a
// partial file), and the hint must be actionable for the system root.
func TestSaveRuleYAMLReadOnlyDirectory(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: directory permissions are not enforced")
	}
	root := t.TempDir()
	setUserDefinitionsRoot(t, root)
	if err := os.Chmod(root, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(root, 0o755) })

	if _, err := SaveRuleYAML(validRulesYAML, "user"); err == nil {
		t.Error("SaveRuleYAML succeeded in a read-only directory")
	}
	if _, err := os.Stat(filepath.Join(root, "rules.yaml")); !os.IsNotExist(err) {
		t.Error("a rules.yaml appeared in a read-only directory")
	}
}

// rulesTestFixture writes a rules.yaml into a fresh working directory (so

// TestReadRuleYAMLReturnsTheCurrentDocument: the Studio opens what is in
// effect, so reading back must return exactly the bytes that were written,
// with the resolved scope and path.
func TestReadRuleYAMLReturnsTheCurrentDocument(t *testing.T) {
	root := t.TempDir()
	setUserDefinitionsRoot(t, root)

	doc, err := ReadRuleYAML("user")
	if err != nil {
		t.Fatalf("ReadRuleYAML: %v", err)
	}
	if doc.Exists || doc.Content != "" || doc.Bytes != 0 {
		t.Errorf("ReadRuleYAML on a fresh root = %+v, want an empty, non-existing document", doc)
	}
	if doc.Path != filepath.Join(root, "rules.yaml") || doc.Scope != ScopeUser {
		t.Errorf("ReadRuleYAML path/scope = %q/%q, want %q/%q", doc.Path, doc.Scope, filepath.Join(root, "rules.yaml"), ScopeUser)
	}

	if _, err := SaveRuleYAML(validRulesYAML, "user"); err != nil {
		t.Fatalf("SaveRuleYAML: %v", err)
	}
	doc, err = ReadRuleYAML("user")
	if err != nil {
		t.Fatalf("ReadRuleYAML after save: %v", err)
	}
	if !doc.Exists || doc.Content != validRulesYAML || doc.Bytes != len(validRulesYAML) {
		t.Errorf("ReadRuleYAML after save = %+v, want the saved document verbatim", doc)
	}
}

// TestReadRuleYAMLScopeErrors: an unknown scope, an unset per-user root and an
// unconfigured system root are errors, never a silent empty document.
func TestReadRuleYAMLScopeErrors(t *testing.T) {
	setUserDefinitionsRoot(t, t.TempDir())
	setSystemDefinitionsRoot(t, t.TempDir())

	if _, err := ReadRuleYAML("compiled"); err == nil {
		t.Error("ReadRuleYAML accepted an unknown scope")
	} else if !strings.Contains(err.Error(), "user") || !strings.Contains(err.Error(), "system") {
		t.Errorf("error = %v, want it to name the accepted scopes", err)
	}

	setUserDefinitionsRoot(t, "")
	if _, err := ReadRuleYAML("user"); err == nil {
		t.Error("ReadRuleYAML accepted the user scope with no HOME")
	}

	setSystemDefinitionsRoot(t, "")
	if _, err := ReadRuleYAML("system"); err == nil {
		t.Error("ReadRuleYAML accepted the system scope with no configured root")
	}
}

// discovery finds it) plus a supportconfig holding one matching line.
func rulesTestFixture(t *testing.T) (rulesDir, supportconfigDir string) {
	t.Helper()
	rulesDir = withTempWorkingDir(t)
	writeFile(t, rulesDir, "rules.yaml", `rules:
  - name: "fires"
    severity: "critical"
    category: "crypto"
    files: ["sssd.conf"]
    message: "RC4 enctype present"
    patterns: ["rc4-hmac"]
  - name: "never-fires"
    severity: "warning"
    files: ["sssd.conf"]
    message: "not present in this supportconfig"
    patterns: ["this-pattern-does-not-exist"]
`)
	supportconfigDir = t.TempDir()
	writeFile(t, supportconfigDir, "sssd.conf", "[domain/example.com]\nkrb5_enctype = rc4-hmac\n")
	return rulesDir, supportconfigDir
}

// TestTestRulesAgainstReportsMatchesAndEvidence: the dry-run must agree with
// the real matcher and hand back the evidence line, so a user can see WHY a
// rule fired.
func TestTestRulesAgainstReportsMatchesAndEvidence(t *testing.T) {
	_, supportconfig := rulesTestFixture(t)

	res, err := TestRulesAgainst(supportconfig)
	if err != nil {
		t.Fatalf("TestRulesAgainst: %v", err)
	}
	if res.Total != 2 || res.Matched != 1 {
		t.Fatalf("Total = %d, Matched = %d, want 2 loaded and exactly 1 fired", res.Total, res.Matched)
	}
	if len(res.Outcomes) != 2 {
		t.Fatalf("Outcomes = %d, want one per loaded rule", len(res.Outcomes))
	}
	byName := map[string]RuleTestOutcome{}
	for _, o := range res.Outcomes {
		byName[o.Rule.Name] = o
	}

	fired := byName["fires"]
	if !fired.Matched {
		t.Fatalf("outcome for \"fires\" = %+v, want a match", fired)
	}
	if !strings.Contains(fired.Evidence, "rc4-hmac") {
		t.Errorf("Evidence = %q, want the offending config line", fired.Evidence)
	}
	if !strings.Contains(fired.Message, "[RULE: fires]") {
		t.Errorf("Message = %q, want the rendered rule finding", fired.Message)
	}
	if !filepath.IsAbs(fired.File) || fired.Line <= 0 {
		t.Errorf("provenance = %s:%d, want an absolute file and its line", fired.File, fired.Line)
	}
	if fired.Scope != ScopeWorkDir {
		t.Errorf("Scope = %q, want %q for a rule found in the working directory", fired.Scope, ScopeWorkDir)
	}

	if quiet := byName["never-fires"]; quiet.Matched {
		t.Errorf("outcome for \"never-fires\" = %+v, want no match", quiet)
	}
}

// TestTestRulesAgainstWritesNothing: the dry-run is a probe — the analyzed
// directory and the definition roots must be untouched.
func TestTestRulesAgainstWritesNothing(t *testing.T) {
	rulesDir, supportconfig := rulesTestFixture(t)

	before, err := os.ReadDir(supportconfig)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := TestRulesAgainst(supportconfig); err != nil {
		t.Fatalf("TestRulesAgainst: %v", err)
	}
	after, err := os.ReadDir(supportconfig)
	if err != nil {
		t.Fatal(err)
	}
	if len(before) != len(after) {
		t.Errorf("target directory changed: %d entries before, %d after", len(before), len(after))
	}
	if entries, err := os.ReadDir(rulesDir); err != nil {
		t.Fatal(err)
	} else if len(entries) != 1 {
		t.Errorf("definitions directory gained files: %v", entries)
	}
}

// TestTestRulesAgainstErrorsAndEmptyRuleSet covers the two edges the GUI hits:
// a path that cannot be read, and a machine with no rules installed at all.
func TestTestRulesAgainstErrorsAndEmptyRuleSet(t *testing.T) {
	wd := withTempWorkingDir(t)

	if _, err := TestRulesAgainst(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("TestRulesAgainst accepted a missing target path")
	}

	res, err := TestRulesAgainst(wd)
	if err != nil {
		t.Fatalf("TestRulesAgainst with no rules installed: %v", err)
	}
	if res.Total != 0 || res.Matched != 0 || len(res.Outcomes) != 0 {
		t.Errorf("result = %+v, want no outcomes when no rules are installed", res)
	}
}

// TestGetCatalogInfo: the Studio shows the catalog provenance next to a
// finding, so an unavailable catalog must be reported, not hidden.
func TestGetCatalogInfo(t *testing.T) {
	info := GetCatalogInfo()
	if !info.Available {
		t.Fatalf("catalog unavailable: %s", info.Error)
	}
	if info.OptionCount == 0 || info.SectionCount == 0 {
		t.Errorf("catalog counts = %d options / %d sections, want both > 0", info.OptionCount, info.SectionCount)
	}
	if info.Source == "" || info.Version == "" {
		t.Errorf("catalog provenance = %q / %q, want both filled", info.Source, info.Version)
	}
}

// A nil Go slice marshals to JSON null, and the generated front-end models
// declare plain arrays (rules: RuleInfo[]). A null field crashed the React
// Definitions Studio on the most common installation there is: no custom rules.
// This test pins the wire contract at the boundary that owns it, so the GUI
// never receives a null array in the first place.
func TestServicePayloadsNeverMarshalArraysAsNull(t *testing.T) {
	wd := withTempWorkingDir(t) // an empty definitions root: every list is nil

	cases := []struct {
		name    string
		payload any
	}{
		{"ListDefinitions", ListDefinitions()},
		{"GetCatalogInfo", GetCatalogInfo()},
		{"ValidateRuleYAML", ValidateRuleYAML(validRulesYAML)},
		{"ValidateRuleYAML empty", ValidateRuleYAML("")},
		{"TestRulesAgainst", dryRunResult(t, wd)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var decoded map[string]any
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			for field, value := range decoded {
				if value == nil {
					t.Errorf("field %q marshals as null, want [] or omitted", field)
				}
			}
		})
	}
}

// dryRunResult runs TestRulesAgainst against an empty target directory. An
// unreadable target is still a payload worth pinning, so the error is logged
// and the result the service returned is checked as-is.
func dryRunResult(t *testing.T, target string) RuleTestResult {
	t.Helper()
	res, err := TestRulesAgainst(target)
	if err != nil {
		t.Logf("dry-run target unreadable, pinning the result it returned anyway: %v", err)
	}
	return res
}
