// logdir.go
//
// Raw-log analysis mode (-logdir): scans a directory (or a single file) of raw
// SSSD logs — e.g. /var/log/sssd/*.log — with the full analysis engine, but
// WITHOUT supportconfig metadata. Every field that would come from the
// supportconfig (system, packages, sssd.conf, service state) is marked
// constants.RawLogModeNA instead of being guessed, and the identity values
// needed for anonymization are harvested from the log lines themselves.
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"sssd-inspector/constants"
)

// maxHarvestLines bounds the per-file read used to harvest redaction tokens.
// Identity values repeat constantly in SSSD logs, so a bounded scan keeps raw
// mode cheap on huge log directories while still finding every domain,
// hostname and realm worth redacting.
const maxHarvestLines = 100000

// maxHarvestLineLen caps a single scanned line (1 MiB, matching the loader's
// max line length) so a pathological line cannot exhaust memory.
const maxHarvestLineLen = 1024 * 1024

// isRawSSDLLogFile reports whether name is a raw SSSD log file. Rotated files
// (sssd_example.com.log.1, sssd_example.com.log-20260918) are included;
// compressed archives (*.log.gz) are NOT, because the single-pass scanner
// reads plain text only.
func isRawSSDLLogFile(name string) bool {
	base := strings.ToLower(name)
	idx := strings.LastIndex(base, constants.RawSSDLLogSuffix)
	if idx < 0 {
		return false
	}
	rest := base[idx+len(constants.RawSSDLLogSuffix):]
	if rest == "" {
		return true // plain "*.log"
	}
	// Only the rotation suffixes ".log.<digits>" and ".log-<digits>" count:
	// ".log.txt" or ".log.gz" are not scanned.
	if rest[0] != '.' && rest[0] != '-' {
		return false
	}
	suffix := rest[1:]
	if suffix == "" {
		return false
	}
	for _, r := range suffix {
		if (r < '0' || r > '9') && r != '-' {
			return false
		}
	}
	return true
}

// collectLogFiles resolves the -logdir argument to the directory that holds
// the logs plus the (deterministic) list of raw log file names to scan —
// relative to that directory, which is what the single-pass scanner expects.
//
// It accepts a directory (the usual /var/log/sssd case) or a single log file.
// Errors are explicit: a missing path, a non-log file, or a directory without
// any SSSD log must never fall through to an empty (and therefore meaningless)
// analysis.
func collectLogFiles(path string) (dir string, files []string, err error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", nil, fmt.Errorf("error accessing log path: %w", err)
	}

	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return "", nil, fmt.Errorf("error reading log directory: %w", err)
		}
		dir = path
		for _, entry := range entries {
			if !entry.IsDir() && isRawSSDLLogFile(entry.Name()) {
				files = append(files, entry.Name())
			}
		}
	} else {
		if !isRawSSDLLogFile(filepath.Base(path)) {
			return "", nil, fmt.Errorf("path is not a raw SSSD log file (expected *%s): %s",
				constants.RawSSDLLogSuffix, path)
		}
		dir = filepath.Dir(path)
		files = append(files, filepath.Base(path))
	}

	if len(files) == 0 {
		return "", nil, fmt.Errorf("no raw SSSD log files (*%s, rotated included) found in: %s",
			constants.RawSSDLLogSuffix, path)
	}
	return dir, files, nil
}

// logPIITokens holds the identity values harvested from raw SSSD log lines.
// Order is first-seen (files are scanned in sorted order), so the primary
// candidates and the redaction list are deterministic across runs.
type logPIITokens struct {
	domains []string // [be[<domain>]], [domain/<domain>]
	realms  []string // realm=<REALM>, <user>@<REALM>
	hosts   []string // syslog prefix hosts, host/<fqdn>@<REALM>
}

// RegExps identifying identity values in raw SSSD log lines. SSSD prints the
// domain in bracketed messages ("Domain [x] is online", "[domain/x]"), the
// realm in krb5 principal/assignment context, and syslog prefixes the machine
// name — the only host/domain/realm sources available when no supportconfig
// is present.
var (
	domainBracketRe = regexp.MustCompile(`(?i)\bdomain \[([^\[\]\s]{4,})\]`)
	domainSectionRe = regexp.MustCompile(`\[domain/([^\[\]\s]+)\]`)
	realmAssignRe   = regexp.MustCompile(`(?i)\brealm\s*[=:]([A-Za-z][A-Za-z0-9.-]{2,})`)
	principalHostRe = regexp.MustCompile(`\bhost/([A-Za-z0-9][A-Za-z0-9.-]*)@[A-Za-z0-9]`)
	principalAtRe   = regexp.MustCompile(`\b[A-Za-z0-9._/-]+@([A-Z][A-Z0-9.-]{2,})\b`)
	ldapURLRe       = regexp.MustCompile(`(?i)ldaps?://([A-Za-z0-9][A-Za-z0-9.-]+)`)
	// The syslog prefix must end with "prog[pid]:" so a tag or message word
	// ("root", "kernel", "the") sitting in the host slot is never mistaken
	// for the machine name.
	syslogPrefixRe = regexp.MustCompile(`^[A-Z][a-z]{2}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}\s+([A-Za-z0-9][A-Za-z0-9_.-]*)\s+[A-Za-z0-9_.-]+\[\d+\]:`)
)

// nonIdentityValues are words that regularly occupy a value slot (host, tag,
// realm assignment) in log lines but are never identity values. Redacting one
// of them would corrupt the report (e.g. "realm supports ..." or "running as
// root").
var nonIdentityValues = map[string]bool{
	"domain": true, "domain_realm": true, "sssd": true, "krb5": true,
	"realm": true, "example.com": true, "localhost": true, "none": true,
	"unknown": true, "root": true, "kernel": true, "mail": true, "daemon": true,
	"systemd": true, "dbus": true, "cron": true, "auth": true, "failed": true,
	"error": true, "cannot": true, "supports": true, "offline": true, "online": true,
}

// looksLikeRealm rejects the common false positive of an assignment-style
// regex ("realm supports ..."): a realm is either dotted (corp.example.com)
// or written in uppercase (CORP.EXAMPLE.COM).
func looksLikeRealm(v string) bool {
	return strings.Contains(v, ".") || v == strings.ToUpper(v)
}

// collect appends value once if it passes the sanity checks: length, charset,
// blacklist, and — unless the caller explicitly saw a domain marker — at least
// one dot. The dot rule is what keeps "[be[ldap_id]]"-style service names out
// of the domain list, where they would be redacted as if they were a domain.
func collect(dst *[]string, seen map[string]bool, value string, allowNoDot bool) {
	if len(value) < 4 || seen[value] {
		return
	}
	if !strings.Contains(value, ".") && !allowNoDot {
		return
	}
	for _, r := range value {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' ||
			r == '.' || r == '-' || r == '_' || r == '/') {
			return
		}
	}
	if nonIdentityValues[strings.ToLower(value)] {
		return
	}
	seen[value] = true
	*dst = append(*dst, value)
}

// harvestPIITokens scans the selected raw log files (bounded) and extracts the
// domain/realm/hostname candidates that anonymizeReport must redact. It is
// best-effort: unreadable files or non-matching lines simply yield no tokens.
func harvestPIITokens(dir string, logFiles []string) logPIITokens {
	var tokens logPIITokens
	seen := map[string]map[string]bool{
		"domain": {}, "realm": {}, "host": {},
	}

	for _, name := range logFiles {
		f, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 0, 64*1024), maxHarvestLineLen)
		for lineNo := 0; scanner.Scan() && lineNo < maxHarvestLines; lineNo++ {
			line := scanner.Text()

			if m := domainBracketRe.FindStringSubmatch(line); m != nil {
				// Explicit "Domain [x]" marker: a single-label name is a
				// genuine domain here, so the dot requirement is lifted.
				collect(&tokens.domains, seen["domain"], m[1], true)
			}
			if m := domainSectionRe.FindStringSubmatch(line); m != nil {
				collect(&tokens.domains, seen["domain"], m[1], true)
			}
			if m := realmAssignRe.FindStringSubmatch(line); m != nil && looksLikeRealm(m[1]) {
				collect(&tokens.realms, seen["realm"], m[1], true)
			}
			if m := principalAtRe.FindStringSubmatch(line); m != nil && looksLikeRealm(m[1]) {
				collect(&tokens.realms, seen["realm"], m[1], true)
			}
			if m := principalHostRe.FindStringSubmatch(line); m != nil {
				collect(&tokens.hosts, seen["host"], m[1], true)
			}
			if m := ldapURLRe.FindStringSubmatch(line); m != nil {
				collect(&tokens.hosts, seen["host"], m[1], true)
			}
			if m := syslogPrefixRe.FindStringSubmatch(line); m != nil {
				collect(&tokens.hosts, seen["host"], m[1], true)
			}
		}
		f.Close()
	}
	return tokens
}

// Primary accessors: the values seeded into the report fields so the regular
// field-level redaction (SearchDomain/KerberosRealm/Hostname) works exactly as
// it does for a supportconfig, and so the report displays real values instead
// of blanks when they are known.
func (t logPIITokens) primaryDomain() string {
	if len(t.domains) > 0 {
		return t.domains[0]
	}
	return ""
}

func (t logPIITokens) primaryRealm() string {
	if len(t.realms) > 0 {
		return t.realms[0]
	}
	return ""
}

func (t logPIITokens) primaryHost() string {
	if len(t.hosts) > 0 {
		return t.hosts[0]
	}
	return ""
}

// redactTokens returns EVERY harvested value with its placeholder, so secondary
// domains and hosts (a second [be[...]] domain, a different host in messages)
// are redacted too — not just the primary ones seeded into the report fields.
//
// Hosts come FIRST: a FQDN host (dc01.example.test) must collapse to its
// placeholder before the bare domain (example.test) is replaced, otherwise only
// the domain part would be redacted and the host label ("dc01") would survive.
func (t logPIITokens) redactTokens() []redactToken {
	toks := make([]redactToken, 0, len(t.domains)+len(t.realms)+len(t.hosts))
	for _, h := range t.hosts {
		toks = append(toks, redactToken{Value: h, Replacement: "redacted-host"})
	}
	for _, d := range t.domains {
		toks = append(toks, redactToken{Value: d, Replacement: constants.DomainReplacement})
	}
	for _, r := range t.realms {
		toks = append(toks, redactToken{Value: r, Replacement: strings.ToUpper(constants.DomainReplacement)})
	}
	return toks
}

// analyzeLogsOnly performs the analysis on raw SSSD log files alone (used by
// -logdir). It runs the same engines as the supportconfig path — single-pass
// scan, temporal clusters, KB suggestions, sequence correlation, executive
// summary, correlation graph, anonymization — but skips every phase that
// requires supportconfig files, so it can never emit supportconfig-only
// findings such as "sssd.conf not found" or "sssd.service is not running".
func analyzeLogsOnly(dirPath string, logFiles []string, anonymize bool, progressFunc func(string, int)) ReportData {
	globalFileCache.Clear()

	var report ReportData
	report.Timestamp = time.Now().Format(constants.TimestampFormat)
	report.AppVersion = constants.AppVersion

	// No supportconfig metadata: mark every system-level field explicitly so
	// the report never presents an unknown value as a real one.
	report.SssdService = constants.RawLogModeNA
	report.WinbindService = constants.RawLogModeNA
	report.NscdStatus = constants.RawLogModeNA
	report.TimeService = constants.RawLogModeNA
	report.KernelVersion = constants.RawLogModeNA
	report.SLESRlease = constants.RawLogModeNA
	report.SCCStatus = constants.RawLogModeNA
	report.HardwareManufacturer = constants.RawLogModeNA
	report.HardwareModel = constants.RawLogModeNA
	report.Hypervisor = constants.RawLogModeNA
	report.VirtualIdentity = constants.RawLogModeNA
	report.SearchDomain = constants.RawLogModeNA
	report.KerberosRealm = constants.RawLogModeNA
	report.Hostname = constants.RawLogModeNA
	report.MACType = "Unknown"
	// Raw SSSD logs can only exist when the daemon has run; the
	// supportconfig-only "not installed / not running" checks are skipped.
	report.SssdInstalled = true

	// Harvest identity values BEFORE any analysis: without an sssd.conf the
	// log lines are the only source for the domain/realm/hostname that
	// anonymizeReport has to redact. Seeding the report fields makes the
	// regular field-level redaction work exactly as in supportconfig mode;
	// every harvested value also travels as an extra redactToken so secondary
	// domains/hosts are covered too.
	tokens := harvestPIITokens(dirPath, logFiles)
	if d := tokens.primaryDomain(); d != "" {
		report.SearchDomain = d
	}
	if rm := tokens.primaryRealm(); rm != "" {
		report.KerberosRealm = rm
	}
	if h := tokens.primaryHost(); h != "" {
		report.Hostname = h
	}

	if progressFunc != nil {
		progressFunc("Single-pass scanning raw SSSD logs...", 20)
	}

	// The KB corpus ships next to the binary; it does not depend on the
	// supportconfig being present (unlike the historical implementation, which
	// passed a nil list and silently disabled every KB feature).
	kbArticles := loadKBArticles(dirPath)

	singlePassResult := performSinglePassScanOnFiles(dirPath, logFiles, report.MACType, kbArticles)
	report.SSSDLogErrors = singlePassResult.SSSDLogErrors
	report.Timeline = singlePassResult.Timeline
	report.Problems = append(report.Problems, singlePassResult.Problems...)
	report.Warnings = append(report.Warnings, singlePassResult.Warnings...)
	if singlePassResult.KeytabFound {
		report.KeytabFound = true
	}

	// Temporal bursts and fuzzy KB suggestions work on the timeline alone.
	if progressFunc != nil {
		progressFunc("Running advanced correlation analysis...", 60)
	}
	report.TemporalClusters = analyzeTemporalClusters(report.Timeline)
	report.KBSuggestions = analyzeKBSuggestions(report.Timeline, kbArticles, report.MatchedTIDs)

	// Root-cause sequence correlation is timeline-driven as well (DNS/SRV
	// failover, Kerberos realm+clock skew, watchdog, backend offline) and must
	// run before the executive summary so its [CORRELATED ROOT CAUSE] findings
	// drive the headline — same order as analyzeData.
	if progressFunc != nil {
		progressFunc("Correlating root-cause sequences...", 75)
	}
	correlateSequences(report.Timeline, &report)

	report.Problems = deduplicateProblems(report.Problems)
	report.Warnings = deduplicateProblems(report.Warnings)

	computeExecutiveSummary(&report)

	finalizeReport(&report, dirPath, anonymize, progressFunc, tokens.redactTokens()...)

	return report
}
