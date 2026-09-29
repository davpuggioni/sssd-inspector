//go:build !cli

// app_definitions_test.go
//
// M1 tests for the Wails bindings of the Definitions Studio
// (app_definitions.go). The bindings are thin wrappers by design, so these
// tests assert the two things the frontend depends on: the payloads are the
// service payloads (whatever the GUI shows is what the analysis will do), and
// the one method that needs the Wails runtime fails cleanly instead of
// panicking when no frontend context is attached.
package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppDefinitionsBindings(t *testing.T) {
	wd := withTempWorkingDir(t)
	writeFile(t, wd, "rules.yaml", validRulesYAML)

	app := NewApp()

	if inv := app.ListDefinitions(); inv.RuleCount != 1 {
		t.Errorf("ListDefinitions: RuleCount = %d, want 1", inv.RuleCount)
	}

	if res := app.ValidateRuleYAML(validRulesYAML); !res.Valid || res.RuleCount != 1 {
		t.Errorf("ValidateRuleYAML(valid) = %+v, want a valid single-rule verdict", res)
	}
	if res := app.ValidateRuleYAML("rules:\n  - name: \"x\"\n    severity: [oops\n"); res.Valid {
		t.Errorf("ValidateRuleYAML(broken) = %+v, want an invalid verdict", res)
	}
	if res := app.ValidateRuleYAML("checks:\n  - name: x\n"); res.Valid {
		t.Errorf("ValidateRuleYAML(wrong top-level key) = %+v, want an invalid verdict", res)
	}

	userRoot := t.TempDir()
	setUserDefinitionsRoot(t, userRoot)
	// A distinct rule name and pattern: the fixture rule (working directory) and
	// this one both load, and only the fixture fires on the rc4 line below.
	userRules := strings.ReplaceAll(validRulesYAML, "legacy-rc4", "user-only-rule")
	userRules = strings.ReplaceAll(userRules, "rc4-hmac", "never-matches-here")
	save, err := app.SaveRuleYAML(userRules, "user")
	if err != nil {
		t.Fatalf("SaveRuleYAML: %v", err)
	}
	if !save.Saved || save.Path != filepath.Join(userRoot, "rules.yaml") {
		t.Fatalf("SaveRuleYAML = %+v, want a saved rules.yaml in %s", save, userRoot)
	}
	refused, err := app.SaveRuleYAML("rules: 42\n", "user")
	if err != nil {
		t.Fatalf("SaveRuleYAML(invalid) returned an error instead of a verdict: %v", err)
	}
	if refused.Saved || refused.Validation.Valid {
		t.Errorf("SaveRuleYAML(invalid) = %+v, want Saved=false with an invalid verdict", refused)
	}

	supportconfig := t.TempDir()
	writeFile(t, supportconfig, "sssd.conf", "[domain/x]\nkrb5_enctype = rc4-hmac\n")
	testRes, err := app.TestRulesAgainst(supportconfig)
	if err != nil {
		t.Fatalf("TestRulesAgainst: %v", err)
	}
	if testRes.Total != 2 || testRes.Matched != 1 {
		t.Errorf("TestRulesAgainst = %+v, want the legacy-rc4 rule to fire", testRes)
	}

	if info := app.GetCatalogInfo(); !info.Available || info.OptionCount == 0 {
		t.Errorf("GetCatalogInfo = %+v, want an available catalog", info)
	}
}

// TestAppOpenDefinitionsRoot: without a frontend context the method must return
// an error (the Wails runtime would log.Fatalf otherwise); with one, it creates
// the folder and hands it to the desktop handler — verified through the seam.
func TestAppOpenDefinitionsRoot(t *testing.T) {
	home := t.TempDir()
	root := filepath.Join(home, ".sssd-inspector")
	setUserDefinitionsRoot(t, root)

	app := NewApp()
	if err := app.OpenDefinitionsRoot("user"); err == nil {
		t.Error("OpenDefinitionsRoot succeeded without a Wails context")
	} else if !strings.Contains(err.Error(), "GUI") {
		t.Errorf("error = %v, want it to mention the missing GUI", err)
	}

	var opened string
	prev := browserOpenURLFn
	browserOpenURLFn = func(_ context.Context, url string) { opened = url }
	t.Cleanup(func() { browserOpenURLFn = prev })

	app.ctx = context.Background()
	if err := app.OpenDefinitionsRoot("user"); err != nil {
		t.Fatalf("OpenDefinitionsRoot: %v", err)
	}
	if want := "file://" + root; opened != want {
		t.Errorf("opened %q, want %q", opened, want)
	}
	if info, err := os.Stat(root); err != nil || !info.IsDir() {
		t.Errorf("the user definitions root was not created: %v", err)
	}

	if err := app.OpenDefinitionsRoot("bogus"); err == nil {
		t.Error("OpenDefinitionsRoot accepted an unknown scope")
	}
}
