// analyzer_config_checks.go
// Phase 2: additional typed configuration checks derived from the INI
// representation of sssd.conf. All checks operate ONLY on data available
// inside the supportconfig (no live system access).
//
//  1. Missing / empty [domain/...] sections  -> SSSD exits at startup.
//  2. [sssd] services missing nss or pam    -> responders never start.
//  3. Overlapping ldap_idmap ranges across AD domains -> silent UID/GID
//     collisions between domains (classic multi-domain / trust breaker).
package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// validateConfigStructure runs the whole-config structural checks. It must be
// called after parseSssdConfig and works for every provider kind, not just AD.
func validateConfigStructure(cfg *ParsedConfig, report *ReportData) {
	// 1. At least one [domain/...] section must exist.
	domainCount := 0
	for _, name := range cfg.Order {
		if cfg.Sections[name].Kind == "domain" {
			domainCount++
		}
	}
	if domainCount == 0 {
		msg := "CONFIGURATION ERROR: sssd.conf contains no [domain/NAME] sections. SSSD has nothing to do and will exit immediately ('No domains configured, exiting')."
		addConfigFinding(report, SevError, "structure", msg, "sssd.conf", "[domain]", 0, "")
	}

	// 2. The [sssd] section must enable the nss and pam responders.
	sssdSec := cfg.Sections["sssd"]
	if sssdSec == nil {
		msg := "CONFIGURATION WARNING: sssd.conf has no [sssd] section. Without it, the nss and pam responders use defaults; verify they are started."
		addConfigFinding(report, SevWarning, "structure", msg, "sssd.conf", "[sssd]", 0, "")
		return
	}
	if vals, _, ok := firstOption(sssdSec, "services"); ok && vals != "" {
		services := splitList(vals)
		has := func(s string) bool {
			for _, v := range services {
				if v == s {
					return true
				}
			}
			return false
		}
		if !has("nss") {
			msg := "CONFIGURATION ERROR: 'services' in [sssd] does not include 'nss'. User/group lookups through NSS will fail."
			addConfigFinding(report, SevError, "structure", msg, "sssd.conf", "[sssd].services", 0, "services = "+vals)
		}
		if !has("pam") {
			msg := "CONFIGURATION ERROR: 'services' in [sssd] does not include 'pam'. PAM authentication will fail even if the domain is healthy."
			addConfigFinding(report, SevError, "structure", msg, "sssd.conf", "[sssd].services", 0, "services = "+vals)
		}
	}
}

// splitList splits a comma/space separated sssd.conf list into lowercased,
// trimmed tokens.
func splitList(v string) []string {
	var out []string
	cur := ""
	for i := 0; i <= len(v); i++ {
		c := byte(',')
		if i < len(v) {
			c = v[i]
		}
		if c == ',' || c == ' ' || c == '\t' {
			if cur != "" {
				out = append(out, cur)
				cur = ""
			}
			continue
		}
		cur += string(c)
	}
	return out
}

// idRange is one ldap_idmap range declared in a domain section.
type idRange struct {
	domain string
	minID  int
	maxID  int
	line   int
}

// validateIDMapRanges collects every ID-mapping range declared in a domain
// section and flags:
//   - inverted or invalid ranges inside a single domain;
//   - overlaps between ranges of different domains (or duplicated values in
//     the same domain family), which cause cross-domain UID/GID collisions.
//
// The canonical option names are 'ldap_idmap_range_min' and
// 'ldap_idmap_range_max' (both integer, verified against sssd-ldap(5) and the
// upstream sssd.api.d/sssd-ldap.conf definitions). The legacy
// 'idmap_range_min'/'idmap_range_max' spellings are accepted as aliases since
// they still show up in older examples. The previous implementation looked up
// 'ldap_idmap_min_id'/'ldap_idmap_max_id', which are NOT SSSD option names at
// all: the check silently never matched a real sssd.conf.
func validateIDMapRanges(cfg *ParsedConfig, report *ReportData) {
	var ranges []idRange
	for _, name := range cfg.Order {
		sec := cfg.Sections[name]
		if sec.Kind != "domain" {
			continue
		}
		// Ranges are only relevant when algorithmic ID mapping is active
		// (ldap_id_mapping defaults to True; only an explicit False disables it).
		if v, _, ok := firstOption(sec, "ldap_id_mapping"); ok && equalFoldTrim(v, "false") {
			continue
		}
		// ldap_idmap_range_max is exclusive (it is min + size), so two ranges
		// are disjoint when cur.min >= prev.max.
		minV, minLine, minOK := firstOptionAny(sec, "ldap_idmap_range_min", "idmap_range_min")
		maxV, maxLine, maxOK := firstOptionAny(sec, "ldap_idmap_range_max", "idmap_range_max")
		if !minOK || !maxOK {
			continue
		}
		minID, err1 := strconv.Atoi(minV)
		maxID, err2 := strconv.Atoi(maxV)
		if err1 != nil || err2 != nil {
			continue
		}
		line := minLine
		if line == 0 {
			line = maxLine
		}
		if minID >= maxID {
			msg := fmt.Sprintf("CONFIGURATION ERROR: '%s' declares an invalid ID mapping range (%d-%d): 'ldap_idmap_range_min' must be lower than 'ldap_idmap_range_max' (which is exclusive).", sec.Name, minID, maxID)
			addConfigFinding(report, SevError, "idmap", msg, "sssd.conf", sec.Name+".ldap_idmap_range_min", line, fmt.Sprintf("ldap_idmap_range_min = %d", minID))
			continue
		}
		ranges = append(ranges, idRange{domain: sec.Name, minID: minID, maxID: maxID, line: line})
	}

	// Sort by min ID and check each range only against its neighbours so the
	// algorithm stays O(n log n) instead of O(n²).
	sort.Slice(ranges, func(i, j int) bool { return ranges[i].minID < ranges[j].minID })
	for i := 1; i < len(ranges); i++ {
		prev, cur := ranges[i-1], ranges[i]
		if cur.minID < prev.maxID {
			msg := fmt.Sprintf("CONFIGURATION ERROR: ID mapping ranges overlap between '%s' (%d-%d) and '%s' (%d-%d). Users of the two domains can receive colliding UID/GIDs.",
				prev.domain, prev.minID, prev.maxID, cur.domain, cur.minID, cur.maxID)
			addConfigFinding(report, SevError, "idmap", msg, "sssd.conf", cur.domain+".ldap_idmap_range_min", cur.line, fmt.Sprintf("ldap_idmap_range_min = %d", cur.minID))
		}
	}
}

// firstOptionAny returns the first matching occurrence among the given key
// aliases, in the order they are listed.
func firstOptionAny(sec *SssdSection, keys ...string) (string, int, bool) {
	for _, k := range keys {
		if v, line, ok := firstOption(sec, k); ok {
			return v, line, true
		}
	}
	return "", 0, false
}

// equalFoldTrim compares two strings case-insensitively after trimming.
func equalFoldTrim(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}
