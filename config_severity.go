// config_severity.go
//
// Central severity policy for configuration findings.
package main

import "strings"

// Confidence layers for findings. "api"/"man"/"curated" mirror the catalog
// OptionMeta.Confidence values; "heuristic" marks hand-written structural
// checks and "log" marks findings derived from log evidence.
const (
	ConfidenceHeuristic = "heuristic"
	ConfidenceLog       = "log"
)

// errorCapableRules allowlists the rule IDs permitted to raise SevError or
// SevCritical. Every other rule is capped at SevWarning because SSSD ignores
// unknown or misplaced options and keeps running. Only conditions that
// provably break the daemon may escalate.
//
// Pre-existing validators keep their historical severity: their rule IDs are
// grandfathered here so the central policy changes nothing for them. Only
// newly introduced rules start capped until proven breaking.
var errorCapableRules = map[string]bool{
	"structure:no-domains":               true,
	"structure:services":                 true,
	"domain:missing-id-provider":         true,
	"domain:invalid-id-provider":         true,
	"domain:inherit-from":                true,
	"idmap:invalid-range":                true,
	"idmap:overlap":                      true,
	"access:simple-allow-without-filter": true,
	"syntax:duplicate-key":               true,
	"semantic:missing-id-provider":       true,
	"semantic:invalid-id-provider":       true,
	"correlate:krb5-no-realm":            true,
	"correlate:krb5-realm-mismatch":      true,
	"correlate:krb5-clock-skew":          true,
	"correlate:dns-search-mismatch":      true,
	"correlate:hostname-not-fqdn":        true,
	"correlate:dns-srv-failure":          true,
	"correlate:offline-after-online":     true,
	"log:ad-domain-unresolvable":         true,
	"log:ad-server-unresolvable":         true,
	"log:unsupported-sasl-mech":          true,
	"log:kerberos-method-secret":         true,
	"log:krb5-validate-disabled":         true,
	"log:starttls-on-ad":                 true,
	"log:invalid-gpo-value":              true,
}

// catalogSeverity enforces the central severity policy: a finding that wants
// more than SevWarning keeps it only when its rule is allowlisted above.
// User-authored YAML rules ("rule:...") keep their declared severity: the
// author explicitly chose it, so the policy does not second-guess them.
func catalogSeverity(want Severity, confidence, ruleID string) Severity {
	if want <= SevWarning {
		return want
	}
	if strings.HasPrefix(ruleID, "rule:") {
		return want
	}
	if errorCapableRules[ruleID] {
		return want
	}
	return SevWarning
}
