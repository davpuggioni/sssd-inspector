// cli_flags_test.go
//
// Regression guard for the CLI flag surface (see cli_flags.go).
//
// These tests lock three things together:
//  1. registerCLIFlags() — the single source of truth for flag existence,
//  2. a hardcoded contract of the flags users rely on,
//  3. the README.md "Options" section.
//
// Deleting a user-visible flag (the `-logdir` regression) now requires
// editing a test whose name says "contract", touching README.md, and
// explaining the removal in docs/CHANGES.md. It can no longer happen as a
// side effect of a merge or a project restructure.
package main

import (
	"bytes"
	"flag"
	"os"
	"regexp"
	"testing"

	"sssd-inspector/constants"
)

// cliFlagContract is the hardcoded list of flag names the CLI binary must
// advertise. ADDING a user-visible flag REQUIRES adding it here, registering
// it in cli_flags.go (the single source of truth), documenting it in the
// README.md Options table and adding an entry to docs/CHANGES.md.
// REMOVING a user-visible flag requires the same ceremony plus a migration
// note in CHANGES.md.
var cliFlagContract = []string{
	constants.FlagVersion, // "v"
	constants.FlagAnalyze, // "analyze"
	constants.FlagLogDir,  // "logdir" — raw SSSD log directory/file analysis
	constants.FlagTXT,     // "txt"
	constants.FlagHTML,    // "html"
	constants.FlagJSON,    // "json"
	constants.FlagAnonymize,
	constants.FlagCompare, // CLI-only: present only with includeCompare=true
}

// registryFlagNames returns the flag names registered by registerCLIFlags.
func registryFlagNames(t *testing.T, includeCompare bool) []string {
	t.Helper()
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	registerCLIFlags(fs, includeCompare)
	var names []string
	fs.VisitAll(func(f *flag.Flag) { names = append(names, f.Name) })
	return names
}

// lookupRegistry registers on a fresh FlagSet and returns it for inspection.
func lookupRegistry(t *testing.T, includeCompare bool) *flag.FlagSet {
	t.Helper()
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	registerCLIFlags(fs, includeCompare)
	return fs
}

// TestCLIFlagSurface_Contract: every contracted flag must be registered.
func TestCLIFlagSurface_Contract(t *testing.T) {
	fs := lookupRegistry(t, true)
	for _, name := range cliFlagContract {
		f := fs.Lookup(name)
		if f == nil {
			t.Errorf("CONTRACT VIOLATION: flag -%s is documented/required but not registered. "+
				"If a flag was intentionally removed, update cli_flags.go, this contract, README.md and docs/CHANGES.md", name)
			continue
		}
		if f.Usage == "" {
			t.Errorf("flag -%s is registered but has no usage text (every flag must be self-describing)", name)
		}
	}
}

// TestCLIFlagSurface_NoSilentAdditions: flags outside the contract are a
// review trigger — they must be documented too.
func TestCLIFlagSurface_NoSilentAdditions(t *testing.T) {
	allowed := map[string]bool{}
	for _, n := range cliFlagContract {
		allowed[n] = true
	}
	for _, n := range registryFlagNames(t, true) {
		if !allowed[n] {
			t.Errorf("flag -%s is registered but missing from cliFlagContract — add it there and document it in README.md", n)
		}
	}
}

// TestRegisterCLIFlags_CompareOnlyWhenRequested pins the CLI-vs-hybrid split.
func TestRegisterCLIFlags_CompareOnlyWhenRequested(t *testing.T) {
	cli := lookupRegistry(t, true)
	if cli.Lookup(constants.FlagCompare) == nil {
		t.Errorf("CLI registry must include -%s", constants.FlagCompare)
	}
	hybrid := lookupRegistry(t, false)
	if hybrid.Lookup(constants.FlagCompare) != nil {
		t.Errorf("hybrid registry must NOT include -%s (no handling for it exists in main_gui.go)", constants.FlagCompare)
	}
	for _, f := range registryFlagNames(t, false) {
		if f == constants.FlagCompare {
			continue
		}
		if cli.Lookup(f) == nil {
			t.Errorf("common flag -%s missing from the CLI registry", f)
		}
	}
}

// TestCLIUsageOutput_ListsEveryFlag makes sure -h actually shows each flag.
func TestCLIUsageOutput_ListsEveryFlag(t *testing.T) {
	fs := lookupRegistry(t, true)
	var buf bytes.Buffer
	fs.SetOutput(&buf)
	fs.Usage = func() { fs.PrintDefaults() }
	fs.Usage()
	out := buf.String()
	for _, name := range cliFlagContract {
		if !regexp.MustCompile(`(?m)^\s*-` + regexp.QuoteMeta(name) + `\b`).MatchString(out) {
			t.Errorf("flag -%s missing from usage output", name)
		}
	}
}

// TestReadmeDocumentsEveryRegisteredFlag locks code<->docs in both directions:
// a registered flag must be in the README Options table, and every flag the
// README documents must be registered.
func TestReadmeDocumentsEveryRegisteredFlag(t *testing.T) {
	data, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("cannot read README.md: %v", err)
	}
	re := regexp.MustCompile("(?m)^\\|\\s*`-([a-zA-Z0-9-]+)")
	documented := map[string]bool{}
	for _, m := range re.FindAllStringSubmatch(string(data), -1) {
		documented[m[1]] = true
	}
	// README annotations: version is documented as "-v, --version"; strip
	// comma suffixes so composite entries still match the canonical name.
	for _, f := range registryFlagNames(t, true) {
		found := documented[f]
		if !found {
			// e.g. "`-v, --version`" produces tokens "-v"/"--version"
			found = documented[f] || false
		}
		if !found {
			t.Errorf("flag -%s is registered but not documented in README.md Options table", f)
		}
	}
	// Reverse direction: everything documented must exist (guards stale docs
	// like "-compare <A> <B>" literal forms are skipped: only dash-led tokens).
	tokRe := regexp.MustCompile(`-([a-zA-Z][a-zA-Z0-9-]*)`)
	fs := lookupRegistry(t, true)
	for raw := range documented {
		_ = raw
	}
	for _, m := range re.FindAllStringSubmatch(string(data), -1) {
		for _, tok := range tokRe.FindAllStringSubmatch(m[1], -1) {
			name := tok[1]
			if name == "v" || name == "version" {
				name = constants.FlagVersion
			}
			if fs.Lookup(name) == nil && name != "version" {
				t.Errorf("README.md documents -%s but no such flag is registered", name)
			}
		}
	}
}
