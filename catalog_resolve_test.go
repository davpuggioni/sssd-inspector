// catalog_resolve_test.go
//
// Tests for the option-catalog override. The contract under test is the one
// that matters to a user with a newer SSSD release than the binary knows about:
// a valid override is really used by the analysis, and a broken one is
// reported and falls back instead of silently changing what "valid" means.
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// testCatalogJSON builds a minimal but valid catalog document: decodeCatalog
// requires at least one option, and the analysis consults option_names for
// typo suggestions, so both are filled in.
func testCatalogJSON(t *testing.T, version string, optionNames ...string) string {
	t.Helper()
	if len(optionNames) == 0 {
		optionNames = []string{"ldap_uri"}
	}
	doc := SssdCatalog{
		Source:      "SSSD upstream",
		Version:     version,
		Generated:   "2026-09-29",
		Sources:     []string{"ad_modified_defaults.xml"},
		Sections:    map[string][]string{"sssd": optionNames},
		Options:     map[string]OptionMeta{},
		OptionNames: optionNames,
	}
	for _, name := range optionNames {
		doc.Options[name] = OptionMeta{Type: "string"}
	}
	sort.Strings(doc.Sections["sssd"])
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal test catalog: %v", err)
	}
	return string(raw)
}

// writeCatalogOverride drops a catalog.json into a definitions root.
func writeCatalogOverride(t *testing.T, root, content string) string {
	t.Helper()
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, CatalogOverrideFileName)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestOptionCatalogPrefersTheUserOverride: the per-user copy wins over the
// per-system one, and both win over the embedded catalog.
func TestOptionCatalogPrefersTheUserOverride(t *testing.T) {
	userRoot := t.TempDir()
	systemRoot := t.TempDir()
	setUserDefinitionsRoot(t, userRoot)
	setSystemDefinitionsRoot(t, systemRoot)

	writeCatalogOverride(t, systemRoot, testCatalogJSON(t, "2.13.0"))
	userPath := writeCatalogOverride(t, userRoot, testCatalogJSON(t, "2.15.0"))

	cat, res, diags := loadOptionCatalog()
	if len(diags) != 0 {
		t.Errorf("diagnostics = %+v, want none for two valid overrides", diags)
	}
	if cat == nil || cat.Version != "2.15.0" {
		t.Fatalf("effective catalog = %+v, want the per-user override (2.15.0)", cat)
	}
	if res.Embedded || res.Path != userPath {
		t.Errorf("resolution = %+v, want the per-user override %s", res, userPath)
	}
	if res.Scope != ScopeUser {
		t.Errorf("scope = %q, want %q", res.Scope, ScopeUser)
	}
}

// TestOptionCatalogFallsBackToSystemThenEmbedded: with only a system override
// present it wins; with none, the embedded catalog is used and the resolution
// says so explicitly.
func TestOptionCatalogFallsBackToSystemThenEmbedded(t *testing.T) {
	userRoot := t.TempDir()
	systemRoot := t.TempDir()
	setUserDefinitionsRoot(t, userRoot)
	setSystemDefinitionsRoot(t, systemRoot)

	cat, res, _ := loadOptionCatalog()
	if cat == nil || !res.Embedded || res.Path != "" {
		t.Fatalf("with no override: catalog=%v resolution=%+v, want the embedded one", cat != nil, res)
	}

	sysPath := writeCatalogOverride(t, systemRoot, testCatalogJSON(t, "2.13.0"))
	cat, res, diags := loadOptionCatalog()
	if len(diags) != 0 {
		t.Errorf("diagnostics = %+v, want none for a valid override", diags)
	}
	if cat == nil || cat.Version != "2.13.0" || res.Path != sysPath || res.Scope != ScopeSystem {
		t.Fatalf("system override ignored: catalog=%+v resolution=%+v", cat, res)
	}
}

// TestBrokenCatalogOverrideIsReportedAndFallsBack is the safety property: a
// broken override must never be half-applied. The user gets a diagnostic naming
// the file and the fallback, and validation continues against the embedded
// catalog — instead of silently validating against nothing.
func TestBrokenCatalogOverrideIsReportedAndFallsBack(t *testing.T) {
	root := t.TempDir()
	setUserDefinitionsRoot(t, root)
	setSystemDefinitionsRoot(t, t.TempDir())
	path := writeCatalogOverride(t, root, `{"options": [ this is not a map`)

	cat, res, diags := loadOptionCatalog()
	if !res.Embedded {
		t.Errorf("resolution = %+v, want the embedded fallback for a malformed override", res)
	}
	if cat == nil {
		t.Fatal("no catalog after a malformed override: validation would silently stop happening")
	}
	if len(diags) != 1 {
		t.Fatalf("diagnostics = %+v, want exactly one for the malformed override", diags)
	}
	d := diags[0]
	if d.File != path {
		t.Errorf("diagnostic file = %q, want %q", d.File, path)
	}
	for _, want := range []string{"skipped", "falling back", "malformed"} {
		if !strings.Contains(d.Message, want) {
			t.Errorf("diagnostic message %q does not mention %q", d.Message, want)
		}
	}
	if d.Severity != SevWarning {
		t.Errorf("severity = %d, want SevWarning", d.Severity)
	}
}

// TestCatalogOverrideWithNoOptionsIsRejected: a syntactically valid document
// with an empty options map is exactly the "silently no validation" trap.
func TestCatalogOverrideWithNoOptionsIsRejected(t *testing.T) {
	root := t.TempDir()
	setUserDefinitionsRoot(t, root)
	setSystemDefinitionsRoot(t, t.TempDir())
	path := writeCatalogOverride(t, root, `{"version":"9.9","sources":[],"options":{},"sections":{}}`)

	cat, res, diags := loadOptionCatalog()
	if cat == nil || !res.Embedded {
		t.Fatalf("an option-less override was accepted: res=%+v", res)
	}
	if len(diags) != 1 || !strings.Contains(diags[0].Message, "no options") {
		t.Fatalf("diagnostics = %+v, want one reporting the empty catalog (%s)", diags, path)
	}
}

// TestCatalogOverrideCacheFollowsContent: the cache is keyed by content, so
// editing the override in place takes effect without restarting the process
// (the GUI stays open across analyses).
func TestCatalogOverrideCacheFollowsContent(t *testing.T) {
	root := t.TempDir()
	setUserDefinitionsRoot(t, root)
	setSystemDefinitionsRoot(t, t.TempDir())
	path := writeCatalogOverride(t, root, testCatalogJSON(t, "2.14.0"))

	cat, _, _ := loadOptionCatalog()
	if cat == nil || cat.Version != "2.14.0" {
		t.Fatalf("first load = %+v, want 2.14.0", cat)
	}

	writeCatalogOverride(t, root, testCatalogJSON(t, "2.16.0"))
	cat, _, _ = loadOptionCatalog()
	if cat == nil || cat.Version != "2.16.0" {
		t.Fatalf("after editing %s: catalog = %+v, want the new 2.16.0", path, cat)
	}
}

// TestGetCatalogInfoReportsTheOverride: the Studio and -definitions-info must
// state which catalog is in effect, not imply the embedded one.
func TestGetCatalogInfoReportsTheOverride(t *testing.T) {
	root := t.TempDir()
	setUserDefinitionsRoot(t, root)
	setSystemDefinitionsRoot(t, t.TempDir())

	embedded := GetCatalogInfo()
	if !embedded.Available {
		t.Fatalf("embedded catalog unavailable: %+v", embedded)
	}
	if embedded.UsingOverride || embedded.Effective != "embedded" {
		t.Errorf("with no override: Effective=%q UsingOverride=%v, want \"embedded\"/false",
			embedded.Effective, embedded.UsingOverride)
	}
	if len(embedded.OverridePaths) == 0 {
		t.Error("OverridePaths is empty: the Studio could not tell a user where to drop a catalog")
	}

	writeCatalogOverride(t, root, testCatalogJSON(t, "2.16.0"))
	withOverride := GetCatalogInfo()
	if !withOverride.UsingOverride || withOverride.Version != "2.16.0" {
		t.Errorf("CatalogInfo = %+v, want the 2.16.0 override reported as in effect", withOverride)
	}
	if withOverride.OptionCount != 1 {
		t.Errorf("OptionCount = %d, want 1 from the override", withOverride.OptionCount)
	}
}

// TestListDefinitionsIncludesCatalogCandidates: "where do I put a catalog?" is
// the same question the other rows of the inventory answer.
func TestListDefinitionsIncludesCatalogCandidates(t *testing.T) {
	userRoot := t.TempDir()
	systemRoot := t.TempDir()
	setUserDefinitionsRoot(t, userRoot)
	setSystemDefinitionsRoot(t, systemRoot)
	path := writeCatalogOverride(t, userRoot, testCatalogJSON(t, "2.16.0"))

	inv := ListDefinitions()
	found := false
	for _, f := range inv.Files {
		if f.Kind != KindCatalog {
			continue
		}
		found = true
		if f.Path == path && (f.OptionCount != 1 || len(f.Diagnostics) != 0) {
			t.Errorf("catalog row = %+v, want the override with 1 option and no diagnostics", f)
		}
	}
	if !found {
		t.Errorf("no catalog candidate in the inventory; files: %+v", inv.Files)
	}
}

// TestListDefinitionsReportsABrokenOverride: the per-location diagnostic is
// merged into the inventory, so the Studio and the stderr banner both see it.
func TestListDefinitionsReportsABrokenOverride(t *testing.T) {
	userRoot := t.TempDir()
	setUserDefinitionsRoot(t, userRoot)
	setSystemDefinitionsRoot(t, t.TempDir())
	path := writeCatalogOverride(t, userRoot, "{ broken")

	inv := ListDefinitions()
	for _, d := range inv.Diagnostics {
		if d.File == path && strings.Contains(d.Message, "option catalog override skipped") {
			return
		}
	}
	t.Errorf("no override diagnostic in the inventory for %s; got %+v", path, inv.Diagnostics)
}

// TestInstallCatalogMakesTheOverrideEffective is the point of the operation:
// installing a catalog must change what the analysis validates against, not
// merely copy a file somewhere.
func TestInstallCatalogMakesTheOverrideEffective(t *testing.T) {
	userRoot := t.TempDir()
	setUserDefinitionsRoot(t, userRoot)
	setSystemDefinitionsRoot(t, t.TempDir())

	// A source catalog outside the definitions roots, as the file chooser
	// would hand it over.
	source := filepath.Join(t.TempDir(), "catalog.json")
	if err := os.WriteFile(source, []byte(testCatalogJSON(t, "9.9.9")), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := InstallCatalog(source, "user")
	if err != nil {
		t.Fatalf("InstallCatalog: %v", err)
	}
	if !res.Saved || res.Scope != ScopeUser {
		t.Fatalf("result = %+v, want a saved user-scope catalog", res)
	}
	if res.Path != filepath.Join(userRoot, CatalogOverrideFileName) {
		t.Errorf("Path = %q, want %q", res.Path, filepath.Join(userRoot, CatalogOverrideFileName))
	}

	cat, resolution, diags := loadOptionCatalog()
	if cat == nil || cat.Version != "9.9.9" {
		t.Fatalf("effective catalog after install = %+v, want the installed 9.9.9", cat)
	}
	if resolution.Embedded || resolution.Scope != ScopeUser {
		t.Errorf("resolution = %+v, want the user override", resolution)
	}
	if len(diags) != 0 {
		t.Errorf("diagnostics = %+v, want none after a clean install", diags)
	}
}

// TestInstallCatalogRefusesAnUnusableDocument: a file the loader would skip must
// never be installed, otherwise the user believes in a catalog that does
// nothing. Nothing is written in that case.
func TestInstallCatalogRefusesAnUnusableDocument(t *testing.T) {
	userRoot := t.TempDir()
	setUserDefinitionsRoot(t, userRoot)
	setSystemDefinitionsRoot(t, t.TempDir())

	for _, tc := range []struct{ name, body string }{
		{"malformed", "{ not json"},
		{"no options", `{"version":"9.9","options":{},"sections":{}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := filepath.Join(t.TempDir(), "catalog.json")
			if err := os.WriteFile(source, []byte(tc.body), 0o644); err != nil {
				t.Fatal(err)
			}
			res, err := InstallCatalog(source, "user")
			if err == nil {
				t.Fatalf("InstallCatalog accepted an unusable catalog: %+v", res)
			}
			if !strings.Contains(err.Error(), "cannot be used as a catalog override") {
				t.Errorf("error = %v, want it to say the file was refused", err)
			}
			if _, statErr := os.Stat(filepath.Join(userRoot, CatalogOverrideFileName)); !os.IsNotExist(statErr) {
				t.Error("a refused catalog was written to the definitions root")
			}
		})
	}
}

// TestInstallCatalogKeepsThePreviousCatalog: the override is the one file a
// user cannot regenerate on an air-gapped host, so the previous one is backed
// up like a rules file.
func TestInstallCatalogKeepsThePreviousCatalog(t *testing.T) {
	userRoot := t.TempDir()
	setUserDefinitionsRoot(t, userRoot)
	setSystemDefinitionsRoot(t, t.TempDir())

	first := filepath.Join(t.TempDir(), "first.json")
	if err := os.WriteFile(first, []byte(testCatalogJSON(t, "2.13.0")), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := InstallCatalog(first, "user"); err != nil {
		t.Fatalf("first install: %v", err)
	}

	second := filepath.Join(t.TempDir(), "second.json")
	if err := os.WriteFile(second, []byte(testCatalogJSON(t, "2.15.0")), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := InstallCatalog(second, "user")
	if err != nil {
		t.Fatalf("second install: %v", err)
	}
	if res.Backup != res.Path+".bak" {
		t.Errorf("Backup = %q, want %q", res.Backup, res.Path+".bak")
	}
	if got := readFileString(t, res.Backup); got != testCatalogJSON(t, "2.13.0") {
		t.Errorf("backup content = %.40q, want the previously installed catalog", got)
	}
	if got := readFileString(t, res.Path); got != testCatalogJSON(t, "2.15.0") {
		t.Errorf("installed content = %.40q, want the new catalog", got)
	}
}

// TestInstallCatalogScopeErrors: the same scope contract as SaveRuleYAML, and
// an empty selection is a refusal rather than a silent success.
func TestInstallCatalogScopeErrors(t *testing.T) {
	setUserDefinitionsRoot(t, t.TempDir())
	setSystemDefinitionsRoot(t, t.TempDir())
	source := filepath.Join(t.TempDir(), "catalog.json")
	if err := os.WriteFile(source, []byte(testCatalogJSON(t, "2.15.0")), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := InstallCatalog(source, ""); err == nil {
		t.Error("InstallCatalog accepted an empty scope")
	} else if !strings.Contains(err.Error(), "user") || !strings.Contains(err.Error(), "system") {
		t.Errorf("error = %v, want it to name the accepted scopes", err)
	}
	if _, err := InstallCatalog("", "user"); err == nil {
		t.Error("InstallCatalog accepted an empty file selection")
	} else if !strings.Contains(err.Error(), "no catalog file selected") {
		t.Errorf("error = %v, want it to say no file was chosen", err)
	}
	if _, err := InstallCatalog(filepath.Join(t.TempDir(), "missing.json"), "user"); err == nil {
		t.Error("InstallCatalog accepted a path that does not exist")
	}
}

// readFileString is a small test helper for asserting exact file content.
func readFileString(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}
