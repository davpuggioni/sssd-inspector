// config_catalog.go
//
// Offline validation of sssd.conf against the option catalog embedded in the
// binary (sssd_catalog/catalog.json, produced by "-gen-catalog").
//
// The catalog is generated from the upstream SSSD man pages and API
// definitions, so the inspector can flag:
//
//  1. Unknown parameters (almost always typos: 'ldap_url' instead of
//     'ldap_uri', 'enumerate_new_users' instead of 'enumerate', ...).
//  2. Wrong value type for a known option (a string where an integer is
//     expected, a boolean option set to a free-form string, ...).
//  3. Values outside the documented enumeration of an option
//     (e.g. 'ldap_schema = rfc2309', 'ad_gpo_access_control = yes').
//
// Severity policy (agreed with the report format): a typo or a bad value is a
// SevWarning because SSSD keeps running with the option ignored (or falls back
// to the default); only syntax errors and fatal structural problems are
// SevError. This keeps the "Problems" section reserved for real breakage.
package main

import (
	"embed"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// embeddedCatalogFS holds the generated option catalog compiled into the
// binary so validation works on an air-gapped host with no extra files.
//
//go:embed sssd_catalog/catalog.json
var embeddedCatalogFS embed.FS

var (
	catalogOnce sync.Once
	catalogData *SssdCatalog
	catalogErr  error
)

// loadEmbeddedCatalog lazily parses the catalog compiled into the binary
// exactly once. The embedded copy cannot change while the process runs, so it
// needs no cache invalidation; overrides do (see catalog_resolve.go).
func loadEmbeddedCatalog() (*SssdCatalog, error) {
	catalogOnce.Do(func() {
		raw, err := embeddedCatalogFS.ReadFile("sssd_catalog/catalog.json")
		if err != nil {
			catalogErr = fmt.Errorf("embedded catalog unavailable: %w", err)
			return
		}
		catalogData, catalogErr = decodeCatalog(raw, "embedded catalog")
	})
	return catalogData, catalogErr
}

// catalogOptionNames returns the sorted list of known option names; used for
// typo suggestions and for "did you mean" hints.
func catalogOptionNames(c *SssdCatalog) []string {
	if len(c.OptionNames) > 0 {
		return c.OptionNames
	}
	names := make([]string, 0, len(c.Options))
	for k := range c.Options {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// validateConfigAgainstCatalog checks every key of every section against the
// option catalog in effect (an override in a definitions root, else the
// embedded one — see catalog_resolve.go). It is deliberately conservative:
// findings are SevWarning (option ignored by SSSD) and never guilt-trip a
// working configuration.
func validateConfigAgainstCatalog(cfg *ParsedConfig, report *ReportData) {
	cat, res, diags := loadOptionCatalog()
	// A skipped override is a report-level problem, not a silent detail: the
	// analysis must be able to answer "was my catalog used?".
	report.AddDiagnostics(diags)
	if cat == nil {
		// No catalog at all (broken build) -> stay silent rather than emit noise.
		return
	}
	names := catalogOptionNames(cat)
	prov := catalogProvenance(cat)
	// Publish the audit trail on the report itself, so it is visible even when
	// no configuration finding is raised. The override, when in effect, is
	// named too: which file produced the verdict is part of the verdict.
	report.CatalogProvenance = prov + catalogSourceNote(res)

	for _, name := range cfg.Order {
		sec := cfg.Sections[name]
		for key, vals := range sec.Options {
			opt, known := cat.Options[key]
			if !known {
				reportUnknownOption(report, prov, names, sec, key, vals)
				continue
			}
			// A known option in a section that does not accept it is a
			// different (and far more common) mistake than a typo, so it gets
			// its own message: the key exists, the placement does not.
			if msg := wrongSectionReason(opt, sec); msg != "" {
				line, evidence := firstKeyVal(vals, key)
				addConfigFindingEx(report, SevWarning, "config_section", msg, "sssd.conf",
					sec.Name+"."+key, line, evidence, "catalog:option-in-wrong-section", opt.Confidence, docRefForSource(opt.Source))
				continue
			}
			for _, kv := range vals {
				if v := invalidValueReason(opt, kv.Value); v != "" {
					msg := fmt.Sprintf("CONFIGURATION WARNING: '%s = %s' in section [%s] is not valid for the '%s' option (%s). SSSD ignores the setting or falls back to the default '%s'. %s",
						key, kv.Value, sec.Name, opt.Type, v, opt.Default, prov)
					addConfigFindingEx(report, SevWarning, "config_value", msg, "sssd.conf",
						sec.Name+"."+key, kv.Line, key+" = "+kv.Value, "catalog:invalid-value", opt.Confidence, docRefForSource(opt.Source))
				}
			}
		}
	}
}

// knownSectionFamilies is the set of section kinds the option catalog
// documents, derived from the catalog's own section map rather than from a
// literal here, so the two cannot drift.
func knownSectionFamilies(c *SssdCatalog) map[string]bool {
	families := make(map[string]bool)
	for name := range c.Sections {
		f := sectionFamily(name)
		if f == "" || f == "*" {
			continue
		}
		families[f] = true
	}
	// "*" in the catalog means "any section", which is a statement about
	// option placement, not a section kind in its own right.
	return families
}

// validateSectionHeaders checks that every [section] in the file is a section
// SSSD documents.
//
// The catalog's keys are man-page section names ("domain/ldap/id"), not the
// literal headers a user writes, so a raw membership test would be wrong in
// both directions: it would reject the perfectly valid [domain/example.com],
// and it would accept nothing useful. The check is therefore on the section
// KIND — the part before the first slash — which is what SSSD itself dispatches
// on.
//
// An unknown kind is worth a warning: SSSD ignores an unrecognised section
// entirely, so the options inside it never take effect, and the operator
// usually typed the header wrong. It is a warning rather than an error because
// a section this SSSD build does not document is not proof of a mistake.
func validateSectionHeaders(cfg *ParsedConfig, report *ReportData) {
	cat, res, diags := loadOptionCatalog()
	report.AddDiagnostics(diags)
	if cat == nil {
		return // no catalog: stay silent rather than guess
	}
	families := knownSectionFamilies(cat)
	// The override in effect is part of the verdict: a section rejected by a
	// locally generated catalog is a weaker claim than one rejected upstream.
	prov := catalogProvenance(cat) + catalogSourceNote(res)
	names := catalogSectionNames(cat)

	for _, name := range cfg.Order {
		sec := cfg.Sections[name]
		if sec.Kind == "" {
			continue
		}
		if families[sec.Kind] {
			continue
		}
		suggestion := suggestSectionName(names, sec.Kind)
		msg := fmt.Sprintf("CONFIGURATION WARNING: section [%s] is not a section SSSD documents (%s). SSSD ignores the whole section, so every option inside it has no effect. %s",
			name, strings.Join(sortedKeys(families), ", "), prov)
		if suggestion != "" {
			msg = fmt.Sprintf("CONFIGURATION WARNING: section [%s] is not a section SSSD documents. Did you mean [%s]? SSSD ignores the whole section, so every option inside it has no effect. %s",
				name, suggestion, prov)
		}
		addConfigFindingEx(report, SevWarning, "config_section", msg, "sssd.conf",
			name, 0, "["+name+"]", "catalog:unknown-section", ConfidenceHeuristic, "sssd.conf(5)")
	}
}

// catalogSectionNames returns the documented section kinds, for typo
// suggestions. A wildcard entry is not a kind and is filtered out.
func catalogSectionNames(c *SssdCatalog) []string {
	families := knownSectionFamilies(c)
	names := make([]string, 0, len(families))
	for f := range families {
		names = append(names, f)
	}
	sort.Strings(names)
	return names
}

// suggestSectionName finds the closest documented section kind, if one is
// close enough to be a credible correction. The same edit-distance budget as
// option names: a distant match would be a misleading suggestion.
func suggestSectionName(names []string, target string) string {
	return suggestOptionName(names, target)
}

// sortedKeys returns the keys of m in a stable order, for message building.
func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// firstKeyVal returns the line number and rendered evidence of the first
// occurrence of a key, used by findings that are not value-specific.
func firstKeyVal(vals []SssdKeyVal, key string) (int, string) {
	if len(vals) == 0 {
		return 0, ""
	}
	return vals[0].Line, key + " = " + vals[0].Value
}

// sectionFamily returns the part of a section name before the first slash, so
// that "domain/ad" and "domain/example.com" both belong to the "domain"
// family. The provider part is a qualifier that this validator deliberately
// ignores: whether ldap_uri applies to a given [domain/...] section depends on
// its id_provider, and the catalog does not model provider negotiation.
func sectionFamily(name string) string {
	if i := strings.IndexByte(name, '/'); i > 0 {
		return name[:i]
	}
	return name
}

// wrongSectionReason returns a non-empty explanation when a documented option
// is used in an sssd.conf section of a different family. SSSD silently ignores
// the setting there, exactly as for an unknown parameter, but the remediation
// is completely different: the key must be moved, not renamed.
func wrongSectionReason(opt OptionMeta, sec *SssdSection) string {
	allowed := opt.Sections
	if len(allowed) == 0 {
		return "" // no section information: never guess
	}
	family := sectionFamily(sec.Kind)
	if family == "" {
		family = sectionFamily(sec.Name)
	}
	for _, s := range allowed {
		if s == "*" || s == sec.Name || s == family {
			return ""
		}
		// Same family (e.g. the option is documented for domain/ad and the
		// key is used in [domain/example.com]): acceptable, the provider
		// decides whether it applies.
		if sectionFamily(s) == family {
			return ""
		}
	}
	target := strings.Join(allowed, "], [")
	return fmt.Sprintf("CONFIGURATION WARNING: '%s' is not valid in section [%s]; it is only valid in section [%s]. SSSD ignores the setting where it is written.",
		opt.Name, sec.Name, target)
}

// catalogProvenance renders the audit trail of the embedded catalog, so every
// finding states which documentation release it was checked against. Without
// it the report asserts "not part of the SSSD option list for this release"
// without ever saying which release, or how many options were compared.
func catalogProvenance(cat *SssdCatalog) string {
	version := cat.Version
	if version == "" {
		version = "unknown"
	}
	return fmt.Sprintf("[checked against the SSSD %s option catalog: %d options from %d documentation sources]",
		version, len(cat.Options), len(cat.Sources))
}

// catalogSourceNote names the override in effect, if any. The release alone is
// the claim; the file is the audit trail behind it, and a reader comparing two
// reports needs both.
func catalogSourceNote(res catalogResolution) string {
	if res.Embedded || res.Path == "" {
		return ""
	}
	return fmt.Sprintf(" (override in effect: %s)", res.Path)
}

// reportUnknownOption emits the typo finding, including the closest known
// option name when one is close enough to be a credible correction.
func reportUnknownOption(report *ReportData, prov string, names []string, sec *SssdSection, key string, vals []SssdKeyVal) {
	line, evidence := firstKeyVal(vals, key)
	if suggestion := suggestOptionName(names, key); suggestion != "" {
		msg := fmt.Sprintf("CONFIGURATION WARNING: unknown parameter '%s' in section [%s]. Did you mean '%s'? SSSD silently ignores unknown parameters, so this setting has no effect. %s",
			key, sec.Name, suggestion, prov)
		addConfigFindingEx(report, SevWarning, "config_unknown", msg, "sssd.conf", sec.Name+"."+key, line, evidence, "catalog:unknown-option", ConfidenceAPI, "sssd.conf(5)")
		return
	}
	msg := fmt.Sprintf("CONFIGURATION WARNING: unknown parameter '%s' in section [%s]. It is not part of the SSSD option list for this release; SSSD silently ignores it (check for a typo). %s",
		key, sec.Name, prov)
	addConfigFindingEx(report, SevWarning, "config_unknown", msg, "sssd.conf", sec.Name+"."+key, line, evidence, "catalog:unknown-option", ConfidenceAPI, "sssd.conf(5)")
}

// docRefForSource converts a catalog Source file name into a man page
// reference for the evidence contract (e.g. "sssd-ldap.5.xml" ->
// "sssd-ldap(5)", "sssd-ldap.conf" stays as is).
func docRefForSource(source string) string {
	s := strings.TrimSpace(source)
	if s == "" || s == "curated" {
		return s
	}
	if i := strings.LastIndex(s, "/"); i >= 0 {
		s = s[i+1:]
	}
	if strings.HasSuffix(s, ".xml") {
		base := strings.TrimSuffix(s, ".xml")
		if i := strings.LastIndex(base, "."); i > 0 {
			return base[:i] + "(" + base[i+1:] + ")"
		}
	}
	return s
}

// invalidValueReason returns a non-empty human-readable explanation when the
// raw value does not satisfy the option's documented type or enumeration.
func invalidValueReason(opt OptionMeta, raw string) string {
	val := strings.TrimSpace(raw)
	if val == "" {
		return ""
	}
	// For list options SSSD accepts comma or space separated items; validate
	// the first token against the enumeration when one is documented.
	first := val
	if idx := strings.IndexAny(val, ", \t"); idx > 0 {
		first = val[:idx]
	}

	switch opt.Type {
	case "bool":
		if !isBoolToken(val) {
			return fmt.Sprintf("a boolean is expected (true/false)")
		}
	case "int":
		if !isIntToken(val) {
			return fmt.Sprintf("an integer is expected")
		}
	}

	if len(opt.Values) > 0 && !containsFold(opt.Values, first) {
		return fmt.Sprintf("allowed values are: %s", strings.Join(opt.Values, ", "))
	}
	return ""
}

// isBoolToken accepts every spelling SSSD itself accepts for booleans.
func isBoolToken(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "true", "false", "yes", "no", "1", "0", "on", "off":
		return true
	}
	return false
}

// isIntToken validates an optional sign followed by decimal digits.
func isIntToken(v string) bool {
	v = strings.TrimSpace(v)
	if v == "" {
		return false
	}
	if v[0] == '+' || v[0] == '-' {
		v = v[1:]
	}
	if v == "" {
		return false
	}
	for i := 0; i < len(v); i++ {
		if v[i] < '0' || v[i] > '9' {
			return false
		}
	}
	return true
}

// containsFold reports whether the slice contains s, ignoring case.
func containsFold(list []string, s string) bool {
	for _, item := range list {
		if strings.EqualFold(strings.TrimSpace(item), strings.TrimSpace(s)) {
			return true
		}
	}
	return false
}

// suggestOptionName returns the closest known option name within a small edit
// distance, or "" when nothing is close enough to be a credible typo fix.
func suggestOptionName(names []string, target string) string {
	const maxDistance = 2
	best := ""
	bestDist := maxDistance + 1
	for _, cand := range names {
		// Cheap rejection: very different lengths cannot be within the budget.
		if abs(len(cand)-len(target)) > maxDistance {
			continue
		}
		d := levenshtein(strings.ToLower(cand), strings.ToLower(target))
		if d < bestDist || (d == bestDist && best != "" && cand < best) {
			bestDist = d
			best = cand
		}
	}
	if bestDist > maxDistance {
		return ""
	}
	return best
}

// levenshtein computes the classic edit distance between two strings.
func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := 0; j <= len(b); j++ {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			m := prev[j] + 1
			if cur[j-1]+1 < m {
				m = cur[j-1] + 1
			}
			if prev[j-1]+cost < m {
				m = prev[j-1] + cost
			}
			cur[j] = m
		}
		prev, cur = cur, prev
	}
	return prev[len(b)]
}

// abs returns the absolute value of an int.
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
