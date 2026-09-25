// version_drift_test.go
//
// Regression guard for the application version.
//
// The tool used to carry the version in three places: constants.AppVersion
// (printed by -v and re-copied onto the report by the CLI), a hardcoded
// "0.2.0" inside analyzeData() — which is what the GUI actually exported — and
// config.DefaultAppVersion. The CLI masked the drift, so GUI-generated
// reports shipped an app_version that disagreed with `-v`.
//
// Rule: constants.AppVersion is the ONLY place a version may appear as a
// string literal. Everything else references it, and the documents that state
// the current version must quote the same value.
package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"sssd-inspector/config"
	"sssd-inspector/constants"
)

// appVersionLiteral matches an AppVersion assigned a string literal, e.g.
// `report.AppVersion = "0.2.0"` or `AppVersion: "0.2.0"`. An assignment of
// the form `x = constants.AppVersion` does not match: no quote follows the
// operator, so the sanctioned references stay invisible to the guard.
var appVersionLiteral = regexp.MustCompile(`AppVersion\s*[:=]\s*"`)

// yamlVersionLiteral matches `version: "X"` as used by the YAML documents.
var yamlVersionLiteral = regexp.MustCompile(`version:\s*"([^"]+)"`)

// versionWalkSkip holds directories that never carry a version source worth
// checking: vendor/tooling trees, generated output, and constants/ itself —
// the sanctioned home of the one and only literal.
var versionWalkSkip = map[string]bool{
	".git":         true,
	"build":        true,
	"constants":    true, // the single sanctioned AppVersion literal lives here
	"dist":         true,
	"frontend":     true,
	"kb_articles":  true,
	"node_modules": true,
}

// TestAppVersionHasSingleLiteralSource fails when any non-test Go source
// assigns a string literal to AppVersion: it must come from constants.
func TestAppVersionHasSingleLiteralSource(t *testing.T) {
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if versionWalkSkip[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, line := range strings.Split(string(data), "\n") {
			if appVersionLiteral.MatchString(line) {
				t.Errorf("%s hardcodes an AppVersion literal — reference "+
					"constants.AppVersion instead:\n\t%s", path, strings.TrimSpace(line))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}
}

// TestConfigDefaultMatchesConstants: the YAML config default must advertise the
// same version as `-v` and as the report footer.
func TestConfigDefaultMatchesConstants(t *testing.T) {
	if config.DefaultAppVersion != constants.AppVersion {
		t.Errorf("config.DefaultAppVersion = %q, want constants.AppVersion %q",
			config.DefaultAppVersion, constants.AppVersion)
	}
	if v := config.DefaultConfig().App.Version; v != constants.AppVersion {
		t.Errorf("DefaultConfig().App.Version = %q, want %q", v, constants.AppVersion)
	}
}

// currentVersionDocs lists the documents that state the CURRENT version, as
// opposed to historical entries in docs/CHANGES.md or release-process
// examples, which deliberately quote whatever version they document.
var currentVersionDocs = []string{"config.yaml", "README.md", "docs/configuration.md"}

// TestCurrentDocsQuoteAppVersion keeps the shipped docs in step with the
// binary: every `version: "X"` they quote must be constants.AppVersion.
func TestCurrentDocsQuoteAppVersion(t *testing.T) {
	for _, doc := range currentVersionDocs {
		data, err := os.ReadFile(doc)
		if err != nil {
			t.Errorf("cannot read %s: %v", doc, err)
			continue
		}
		matches := yamlVersionLiteral.FindAllStringSubmatch(string(data), -1)
		if len(matches) == 0 {
			t.Errorf("%s no longer quotes a version — the guard or the doc is stale", doc)
			continue
		}
		for _, m := range matches {
			if m[1] != constants.AppVersion {
				t.Errorf("%s quotes version %q, want %q", doc, m[1], constants.AppVersion)
			}
		}
	}
}
