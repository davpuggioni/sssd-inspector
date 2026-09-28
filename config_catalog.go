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
	"encoding/json"
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

// loadEmbeddedCatalog lazily parses the embedded catalog exactly once.
func loadEmbeddedCatalog() (*SssdCatalog, error) {
	catalogOnce.Do(func() {
		raw, err := embeddedCatalogFS.ReadFile("sssd_catalog/catalog.json")
		if err != nil {
			catalogErr = fmt.Errorf("embedded catalog unavailable: %w", err)
			return
		}
		var c SssdCatalog
		if err := json.Unmarshal(raw, &c); err != nil {
			catalogErr = fmt.Errorf("embedded catalog is malformed: %w", err)
			return
		}
		if len(c.Options) == 0 {
			catalogErr = fmt.Errorf("embedded catalog contains no options")
			return
		}
		catalogData = &c
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
// embedded catalog. It is deliberately conservative: findings are SevWarning
// (option ignored by SSSD) and never guilt-trip a working configuration.
func validateConfigAgainstCatalog(cfg *ParsedConfig, report *ReportData) {
	cat, err := loadEmbeddedCatalog()
	if err != nil || cat == nil {
		// No catalog (broken build) -> stay silent rather than emit noise.
		return
	}
	names := catalogOptionNames(cat)

	for _, name := range cfg.Order {
		sec := cfg.Sections[name]
		for key, vals := range sec.Options {
			opt, known := cat.Options[key]
			if !known {
				reportUnknownOption(report, names, sec, key, vals)
				continue
			}
			for _, kv := range vals {
				if v := invalidValueReason(opt, kv.Value); v != "" {
					msg := fmt.Sprintf("CONFIGURATION WARNING: '%s = %s' in section [%s] is not valid for the '%s' option (%s). SSSD ignores the setting or falls back to the default '%s'.",
						key, kv.Value, sec.Name, opt.Type, v, opt.Default)
					addConfigFinding(report, SevWarning, "config_value", msg, "sssd.conf",
						sec.Name+"."+key, kv.Line, key+" = "+kv.Value)
				}
			}
		}
	}
}

// reportUnknownOption emits the typo finding, including the closest known
// option name when one is close enough to be a credible correction.
func reportUnknownOption(report *ReportData, names []string, sec *SssdSection, key string, vals []SssdKeyVal) {
	line := 0
	evidence := ""
	if len(vals) > 0 {
		line = vals[0].Line
		evidence = key + " = " + vals[0].Value
	}
	if suggestion := suggestOptionName(names, key); suggestion != "" {
		msg := fmt.Sprintf("CONFIGURATION WARNING: unknown parameter '%s' in section [%s]. Did you mean '%s'? SSSD silently ignores unknown parameters, so this setting has no effect.",
			key, sec.Name, suggestion)
		addConfigFinding(report, SevWarning, "config_unknown", msg, "sssd.conf", sec.Name+"."+key, line, evidence)
		return
	}
	msg := fmt.Sprintf("CONFIGURATION WARNING: unknown parameter '%s' in section [%s]. It is not part of the SSSD option list for this release; SSSD silently ignores it (check for a typo).", key, sec.Name)
	addConfigFinding(report, SevWarning, "config_unknown", msg, "sssd.conf", sec.Name+"."+key, line, evidence)
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
