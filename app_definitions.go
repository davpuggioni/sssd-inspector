//go:build !cli

// app_definitions.go
//
// Wails bindings for the Definitions Studio. Every method is a thin wrapper
// around definitions_service.go, which the CLI flags (-definitions-info,
// -validate-rules, -rules-test) and the analysis itself also call — so the
// GUI, the CLI and the analysis can never disagree about what a definition
// file does.
//
// Only OpenDefinitionsRoot needs the Wails runtime (to launch the desktop file
// manager); it returns an error instead of panicking when no frontend context
// is attached, which is the case in unit tests and in CLI invocations of the
// hybrid binary.
package main

import (
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"sssd-inspector/constants"
)

// browserOpenURLFn opens a URL (here: a file:// directory) with the desktop's
// default handler. It is a seam so the "show me the folder" affordance can be
// unit-tested without launching a file manager.
var browserOpenURLFn = runtime.BrowserOpenURL

// ListDefinitions returns the definition search inventory: every location the
// analysis inspects for rules and KB articles, what was found there, whether
// the user may write there, and every input that would be skipped.
func (a *App) ListDefinitions() DefinitionsInventory {
	return ListDefinitions()
}

// ValidateRuleYAML validates rules YAML edited in the Studio without saving it.
// The verdict is carried in the result: a malformed document is not an error,
// it is a validation result the editor renders inline.
func (a *App) ValidateRuleYAML(content string) RuleValidationResult {
	return ValidateRuleYAML(content)
}

// SaveRuleYAML validates and saves rules YAML into the user or system scope.
// Invalid documents are refused (Saved=false) with the diagnostics in

// ReadRuleYAML returns the current content of the rules document in the given
// scope, so the Studio opens what is in effect rather than an empty editor.
// A scope without a rules.yaml yet is not an error: Exists is false.
func (a *App) ReadRuleYAML(scope string) (RuleDocument, error) {
	return ReadRuleYAML(scope)
}

// OpenCatalogFile opens a file chooser for a generated catalog.json, so
// InstallCatalog gets a path from the OS dialog instead of a typed one.
func (a *App) OpenCatalogFile() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("cannot open the file browser: the GUI is not running")
	}
	return openFileDialogFn(a.ctx, runtime.OpenDialogOptions{
		Title: "Select a generated catalog.json",
		Filters: []runtime.FileFilter{
			{DisplayName: "Catalog (catalog.json)", Pattern: "catalog.json"},
			{DisplayName: constants.FilterAllFiles, Pattern: constants.PatternAllFiles},
		},
	})
}

// InstallCatalog installs a generated catalog.json as the override for a scope.
// The document is validated with the loader's own decoder, the write is atomic
// and the previous catalog is kept as catalog.json.bak.
func (a *App) InstallCatalog(path string, scope string) (DefinitionSaveResult, error) {
	return InstallCatalog(path, scope)
}

// Validation, and the previous content is preserved as rules.yaml.bak.
func (a *App) SaveRuleYAML(content string, scope string) (DefinitionSaveResult, error) {
	return SaveRuleYAML(content, scope)
}

// TestRulesAgainst dry-runs the loaded rules against a supportconfig directory
// or archive and reports which rules would fire, with the evidence line that
// made each match. Nothing is written and no report is produced.
func (a *App) TestRulesAgainst(targetPath string) (RuleTestResult, error) {
	return TestRulesAgainst(targetPath)
}

// GetCatalogInfo reports the embedded SSSD option catalog: upstream release,
// generation date and size. Read-only; see GetCatalogInfo.
func (a *App) GetCatalogInfo() CatalogInfo {
	return GetCatalogInfo()
}

// OpenDefinitionsRoot creates (if needed) and opens the definitions directory
// of the given scope in the desktop file manager, so a user can drop definition
// files in without having to know the path by heart.
func (a *App) OpenDefinitionsRoot(scope string) error {
	root, _, err := scopeRoot(scope)
	if err != nil {
		return err
	}
	if a.ctx == nil {
		return fmt.Errorf("cannot open the definitions folder: the GUI is not running")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("cannot create %s: %w%s", root, err, writeHint(root))
	}
	browserOpenURLFn(a.ctx, "file://"+root)
	return nil
}
