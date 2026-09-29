// catalog_resolve.go
//
// Override resolution for the offline SSSD option catalog.
//
// The catalog is generated from upstream man pages and API definitions and
// embedded in the binary, which pins every build to one SSSD release. That is
// the right default — validation works on an air-gapped host with no files at
// all — but it is frozen at build time, and a support engineer working on a
// newer distro needs to validate against the release the customer runs.
//
// So a generated catalog.json dropped into a definitions root overrides the
// embedded copy, with the same precedence the rules loader gives its overrides
// (per-user before per-system), and the embedded catalog remains the fallback.
// A malformed override is reported as a diagnostic and skipped: the analysis
// falls back to the embedded catalog and the report's provenance line still
// names the release its claims are based on, so a skipped override can never
// silently change what "valid" means.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// CatalogOverrideFileName is the file a user drops into a definitions root to
// validate against a catalog newer than the one compiled into the binary. It is
// produced by the -gen-catalog flag from the upstream drop-zone.
const CatalogOverrideFileName = "catalog.json"

// catalogResolution records which layer the effective catalog came from, so the
// report and the Studio can state it instead of implying "the shipped one".
type catalogResolution struct {
	// Embedded is true when the catalog compiled into the binary is in use.
	Embedded bool
	// Path is the override file in use, empty when Embedded.
	Path string
	// Scope is the scope that provided the override.
	Scope DefinitionScope
}

// effective renders the resolution for humans (and for the report provenance).
func (r catalogResolution) effective() string {
	if r.Embedded || r.Path == "" {
		return "embedded"
	}
	return r.Path
}

// catalogOverrideCandidates returns the override locations in resolution order:
// the per-user root wins over the per-system one, and both win over the
// embedded catalog. Empty roots (no HOME, unconfigured) are skipped.
func catalogOverrideCandidates() []string {
	var out []string
	for _, root := range []string{userDefinitionsRoot(), systemDefinitionsRoot()} {
		if root == "" {
			continue
		}
		out = append(out, filepath.Join(root, CatalogOverrideFileName))
	}
	return out
}

// decodeCatalog parses and sanity-checks one catalog document. label names the
// origin ("embedded catalog" or a file path) so the error always says which
// copy is broken — the same validation for both layers, so an override can
// never be accepted for being "close enough".
func decodeCatalog(data []byte, label string) (*SssdCatalog, error) {
	var c SssdCatalog
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("%s is malformed: %w", label, err)
	}
	if len(c.Options) == 0 {
		return nil, fmt.Errorf("%s contains no options", label)
	}
	return &c, nil
}

// catalogOverrideCache memoises parsed overrides by content hash. An analysis
// must not re-parse a 250KB catalog on every configuration, but an edited
// override has to take effect on the next run — so the key is the content
// digest, not the path, and a fixed file invalidates itself.
var catalogOverrideCache sync.Map // "path:sha256" -> catalogOverrideCacheEntry

type catalogOverrideCacheEntry struct {
	cat *SssdCatalog
	err error
}

// loadOverrideCatalog reads, parses and caches one override file.
func loadOverrideCatalog(path string) (*SssdCatalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(data)
	key := path + ":" + hex.EncodeToString(sum[:])
	if hit, ok := catalogOverrideCache.Load(key); ok {
		entry := hit.(catalogOverrideCacheEntry)
		return entry.cat, entry.err
	}
	cat, perr := decodeCatalog(data, path)
	catalogOverrideCache.Store(key, catalogOverrideCacheEntry{cat: cat, err: perr})
	return cat, perr
}

// describeOverrideCatalog reports what one candidate holds, for the inventory:
// the option count when it parses, and the same diagnostic the analysis would
// produce when it does not.
func describeOverrideCatalog(path string) DefinitionFileInfo {
	fi := describeDefinitionPath(path, KindCatalog)
	if !fi.Exists || fi.IsDir {
		return fi
	}
	cat, err := loadOverrideCatalog(path)
	if err != nil {
		fi.Diagnostics = append(fi.Diagnostics, catalogOverrideDiagnostic(path, err))
		return fi
	}
	fi.OptionCount = len(cat.Options)
	return fi
}

// catalogOverrideDiagnostic is the report of a broken override. The message
// states the fallback explicitly, because "my new catalog is not being used" is
// exactly the question a user cannot answer from "permission denied".
func catalogOverrideDiagnostic(path string, err error) Diagnostic {
	return Diagnostic{
		File:     path,
		Message:  fmt.Sprintf("option catalog override skipped: %v — falling back to the catalog embedded in the binary", err),
		Severity: SevWarning,
	}
}

// loadOptionCatalog returns the catalog the analysis must validate against,
// the layer it came from, and a diagnostic for every override that was skipped.
func loadOptionCatalog() (*SssdCatalog, catalogResolution, []Diagnostic) {
	var diags []Diagnostic
	for _, candidate := range catalogOverrideCandidates() {
		if _, err := os.Stat(candidate); err != nil {
			continue // not present: the normal case, nothing to report
		}
		cat, err := loadOverrideCatalog(candidate)
		if err != nil {
			diags = append(diags, catalogOverrideDiagnostic(candidate, err))
			continue // try the next layer instead of validating against nothing
		}
		return cat, catalogResolution{Path: candidate, Scope: definitionScopeOf(candidate)}, diags
	}

	cat, err := loadEmbeddedCatalog()
	if err != nil {
		return nil, catalogResolution{Embedded: true}, diags
	}
	return cat, catalogResolution{Embedded: true}, diags
}
