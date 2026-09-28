// catalog_gen_test.go
//
// Tests for the offline catalog generator (`-gen-catalog`): API *.conf parsing,
// DocBook XML parsing, roff fallback (plain and gzipped), the curated enum
// enrichment and the markdown renderer.
package main

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withCatalogOutput redirects catalogOutputDir into a temporary directory so a
// generation run never overwrites the committed sssd_catalog/ artifacts.
func withCatalogOutput(t *testing.T) string {
	t.Helper()
	out := t.TempDir()
	prev := catalogOutputDir
	catalogOutputDir = out
	t.Cleanup(func() { catalogOutputDir = prev })
	return out
}

// writeFiles materialises a map[relative path]content tree under a temp dir.
func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, content := range files {
		full := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// loadGeneratedCatalog runs the generator against srcDir and parses the JSON it
// wrote into outDir.
func loadGeneratedCatalog(t *testing.T, srcDir, outDir string) *SssdCatalog {
	t.Helper()
	if err := RunGenerateCatalog(srcDir); err != nil {
		t.Fatalf("RunGenerateCatalog(%q) error = %v", srcDir, err)
	}
	raw, err := os.ReadFile(filepath.Join(outDir, "catalog.json"))
	if err != nil {
		t.Fatalf("catalog.json was not written: %v", err)
	}
	var cat SssdCatalog
	if err := json.Unmarshal(raw, &cat); err != nil {
		t.Fatalf("catalog.json is not valid JSON: %v", err)
	}
	return &cat
}

func TestRunGenerateCatalog_MissingSourceDir(t *testing.T) {
	withCatalogOutput(t)
	err := RunGenerateCatalog(filepath.Join(t.TempDir(), "does-not-exist"))
	if err == nil {
		t.Fatal("a missing source directory must be an error")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("unexpected error message: %v", err)
	}
	// The caller is told where to look for instructions.
	if !strings.Contains(err.Error(), "README.md") {
		t.Errorf("error should point at the drop-zone README: %v", err)
	}
}

func TestRunGenerateCatalog_NoOptionsFound(t *testing.T) {
	withCatalogOutput(t)
	src := writeFiles(t, map[string]string{"notes.txt": "nothing to parse here\n"})
	err := RunGenerateCatalog(src)
	if err == nil {
		t.Fatal("a source directory without options must be an error")
	}
	if !strings.Contains(err.Error(), "no SSSD configuration options found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunGenerateCatalog_DefaultsToUpstreamDir(t *testing.T) {
	withCatalogOutput(t)

	// Run in an empty working directory so "./upstream" is guaranteed absent:
	// this proves that an empty source argument resolves to the drop-zone path.
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(prevWD); err != nil {
			t.Fatalf("cannot restore the working directory: %v", err)
		}
	}()

	err = RunGenerateCatalog("")
	if err == nil {
		t.Fatal("no ./upstream drop-zone exists here, so RunGenerateCatalog(\"\") must fail")
	}
	if !strings.Contains(err.Error(), "upstream") {
		t.Errorf("the default source directory should be 'upstream', got: %v", err)
	}
}

// TestRunGenerateCatalog_FromAPIConfAndXML covers the primary pipeline: the API
// *.conf files provide names/types/defaults, the DocBook XML adds docs and
// defaults, and the curated table fills in the enumerations.
func TestRunGenerateCatalog_FromAPIConfAndXML(t *testing.T) {
	out := withCatalogOutput(t)
	src := writeFiles(t, map[string]string{
		"version.m4": "m4_define([VERSION_NUMBER], [9.9.9-test])\n",
		"sssd.api.conf": "# comment\n" +
			"[service]\n" +
			"timeout = int, None, false\n" +
			"debug_level = int, None, false\n" +
			"\n[domain]\n" +
			"id_provider = str, None, true\n" +
			"ldap_schema = str, None, false, rfc2307\n" +
			"\n[provider/ad]\n" +
			"ad_gpo_access_control = str, None, false, permissive\n",
		"sssd.api.d/sssd-ad.conf": "[provider/ad]\n" +
			"ad_domain = str, None, false\n" +
			"ldap_idmap_range_min = int, None, false, 200000\n",
		"src/man/sssd-ad.5.xml": "<refentry><varlistentry>" +
			"<term>ad_gpo_access_control (string)</term>" +
			"<listitem><para>Operation mode for GPO-based access control.</para>" +
			"<para>Default: permissive</para></listitem></varlistentry>" +
			"</refentry>",
	})

	cat := loadGeneratedCatalog(t, src, out)

	if cat.Version != "9.9.9-test" {
		t.Errorf("Version = %q, want 9.9.9-test (from version.m4)", cat.Version)
	}
	// The API conf is authoritative for name and type.
	if got := cat.Options["timeout"].Type; got != "int" {
		t.Errorf("timeout type = %q, want int", got)
	}
	if got := cat.Options["ad_domain"].Type; got != "string" {
		t.Errorf("ad_domain type = %q, want string (str maps to string)", got)
	}
	// The 4th API field is the default.
	if got := cat.Options["ad_gpo_access_control"].Default; got != "permissive" {
		t.Errorf("ad_gpo_access_control default = %q, want permissive", got)
	}
	if got := cat.Options["ldap_idmap_range_min"].Default; got != "200000" {
		t.Errorf("ldap_idmap_range_min default = %q, want 200000", got)
	}
	// "None" must not leak in as a literal default.
	if got := cat.Options["timeout"].Default; got == "None" {
		t.Errorf("timeout default must not be the literal 'None'")
	}
	// Curated enumerations are applied on top.
	if got := cat.Options["ldap_schema"].Values; len(got) == 0 {
		t.Errorf("ldap_schema should have curated values")
	}
	if got := cat.Options["ad_gpo_access_control"].Values; len(got) != 3 {
		t.Errorf("ad_gpo_access_control values = %v, want 3 entries", got)
	}
	// The XML pass adds the documentation.
	if got := cat.Options["ad_gpo_access_control"].Doc; !strings.Contains(got, "GPO-based access control") {
		t.Errorf("ad_gpo_access_control doc = %q, want the XML prose", got)
	}
	// Sections and sources are tracked and sorted. The API files name the
	// provider sections "[provider/ad]", but in sssd.conf those settings live
	// in "[domain/ad]", so the generator normalises the name.
	if _, ok := cat.Sections["domain/ad"]; !ok {
		t.Errorf("domain/ad section is missing: %v", cat.Sections)
	}
	for s := range cat.Sections {
		if strings.HasPrefix(s, "provider") {
			t.Errorf("section %q was not normalised to its sssd.conf name", s)
		}
	}
	for i := 1; i < len(cat.Sources); i++ {
		if cat.Sources[i-1] > cat.Sources[i] {
			t.Errorf("Sources are not sorted: %v", cat.Sources)
		}
	}
	if len(cat.OptionNames) != len(cat.Options) {
		t.Errorf("OptionNames (%d) != Options (%d)", len(cat.OptionNames), len(cat.Options))
	}
	for i := 1; i < len(cat.OptionNames); i++ {
		if cat.OptionNames[i-1] > cat.OptionNames[i] {
			t.Errorf("OptionNames are not sorted: %v", cat.OptionNames)
		}
	}
	// The markdown companion is written too and mentions the options.
	md, err := os.ReadFile(filepath.Join(out, "catalog.md"))
	if err != nil {
		t.Fatalf("catalog.md was not written: %v", err)
	}
	for _, want := range []string{"ad_gpo_access_control", "## Sections", "All Options (Alphabetical)"} {
		if !strings.Contains(string(md), want) {
			t.Errorf("catalog.md is missing %q", want)
		}
	}
}

// TestRunGenerateCatalog_RoffFallback covers the compiled-man-page path, both
// plain (.5) and gzipped (.5.gz).
func TestRunGenerateCatalog_RoffFallback(t *testing.T) {
	out := withCatalogOutput(t)

	plain := ".TH sssd-test 5\n" +
		".SH NAME\nsssd-test\n" +
		".TP\n\\fBldap_uri\\fR (string)\nComma separated list of URIs.\n" +
		".TP 4\n\\fBdebug_level\\fR (integer)\nDebug level.\n"
	src := writeFiles(t, map[string]string{"sssd-test.5": plain})

	// A gzipped page with a third option, to exercise the gzip reader.
	gzPath := filepath.Join(src, "sssd-gz.5.gz")
	f, err := os.Create(gzPath)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	if _, err := gz.Write([]byte(".TP\n\\fBkrb5_realm\\fR (string)\nKerberos realm.\n")); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	cat := loadGeneratedCatalog(t, src, out)

	for _, want := range []string{"ldap_uri", "debug_level", "krb5_realm"} {
		if _, ok := cat.Options[want]; !ok {
			t.Errorf("roff parsing missed the option %q: %v", want, cat.OptionNames)
		}
	}
	if got := cat.Options["ldap_uri"].Type; got != "string" {
		t.Errorf("ldap_uri type = %q, want string (from the roff parenthesised type)", got)
	}
	if got := cat.Options["debug_level"].Type; got != "int" {
		t.Errorf("debug_level type = %q, want int (integer maps to int)", got)
	}
	if len(cat.Sources) != 2 {
		t.Errorf("Sources = %v, want both the .5 and the .5.gz page", cat.Sources)
	}
}

// TestRunGenerateCatalog_CorruptGzipIsSkipped ensures a broken archive does not
// abort the whole generation.
func TestRunGenerateCatalog_CorruptGzipIsSkipped(t *testing.T) {
	out := withCatalogOutput(t)
	src := writeFiles(t, map[string]string{
		"sssd-broken.5.gz": "this is not gzip data at all",
		"sssd-ok.5":        ".TP\n\\fBldap_uri\\fR (string)\nURI list.\n",
	})
	cat := loadGeneratedCatalog(t, src, out)
	if _, ok := cat.Options["ldap_uri"]; !ok {
		t.Errorf("the valid page must still be parsed: %v", cat.OptionNames)
	}
}

func TestMapAPIType(t *testing.T) {
	cases := map[string]string{
		"bool": "bool", "BOOLEAN": "bool",
		"int": "int", "Integer": "int",
		"list": "list",
		"str":  "string", "string": "string", "": "string", "weird": "string",
	}
	for in, want := range cases {
		if got := mapAPIType(in); got != want {
			t.Errorf("mapAPIType(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSliceContains(t *testing.T) {
	list := []string{"a", "b"}
	if !sliceContains(list, "a") || sliceContains(list, "c") || sliceContains(nil, "a") {
		t.Errorf("sliceContains behaves unexpectedly")
	}
}

func TestEnrichKnownEnums(t *testing.T) {
	cat := &SssdCatalog{Options: map[string]OptionMeta{
		"ldap_schema": {Name: "ldap_schema", Type: "string"},
	}}
	enrichKnownEnums(cat)
	if len(cat.Options["ldap_schema"].Values) == 0 {
		t.Errorf("ldap_schema did not receive curated values")
	}
	// Options absent from the catalog must not be invented.
	if _, ok := cat.Options["ad_gpo_access_control"]; ok {
		t.Errorf("enrichKnownEnums must not create options that were not parsed")
	}
}

func TestGenerateCatalogMarkdown_EmptyCatalog(t *testing.T) {
	md := generateCatalogMarkdown(&SssdCatalog{
		Version:   "0.0.0",
		Generated: "2026-01-01",
		Sections:  map[string][]string{},
		Options:   map[string]OptionMeta{},
	})
	for _, want := range []string{
		"# SSSD Configuration Options Reference Catalog",
		"Total Options: **0**",
		"## All Options (Alphabetical)",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("markdown output is missing %q", want)
		}
	}
}
