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
