// logdir_test.go
//
// Regression tests for raw-log mode (-logdir): file selection, PII harvesting
// from log lines, the analyzeLogsOnly pipeline (including its anonymization
// guarantees) and the runLogDirAnalyze report generation.
//
// This mode exists because the -logdir feature was silently lost once before
// (dropped during a repository restructure while the flag surface was
// duplicated across two main() functions). Everything here is pinned so the
// feature — and specifically its anonymization, which has no sssd.conf to
// snapshot values from — cannot disappear again without a failing test.
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sssd-inspector/constants"
)

// writeRawLogFixture creates a directory shaped like /var/log/sssd: raw SSSD
// logs (including a rotated one), syslog-style "messages", and a non-log file
// that must be ignored.
//
// The content deliberately carries every identity value the anonymization
// tests must never see again — domain example.test, realm EXAMPLE.TEST, host
// testhost01 (plus its FQDN and the DC dc01.example.test) — and one trap for
// each harvesting false positive: the service name "[be[ldap_id]]" (must NOT
// be harvested as a domain) and a "root kernel:" syslog line (must NOT be
// harvested as a hostname).
func writeRawLogFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		// Raw SSSD log: clock-skew errors repeated inside a tight window
		// (enough for a temporal cluster) plus the identity-value carriers.
		"sssd_example.test.log": strings.Join([]string{
			"(2026-09-18 10:00:01): [sssd[be[ldap_id]]] [sdap_id_op_connect_done] (0x0040): Connection failed: Clock skew too great",
			"(2026-09-18 10:00:02): [krb5_child[1234]] [main] (0x0040): host/testhost01.example.test@EXAMPLE.TEST authentication failed",
			"(2026-09-18 10:00:03): [sssd[be[ldap_id]]] [sdap_get_servers] (0x0040): Domain [example.test] is offline",
			"(2026-09-18 10:00:04): [sssd[be[ldap_id]]] [sdap_connect] (0x0040): URI ldap://dc01.example.test failed: Connection refused",
			"(2026-09-18 10:00:05): [sssd[be[ldap_id]]] [sdap_id_op_connect_done] (0x0040): Connection failed: Clock skew too great",
			"(2026-09-18 10:00:06): [sssd[be[ldap_id]]] [sdap_id_op_connect_done] (0x0040): Connection failed: Clock skew too great",
			"(2026-09-18 10:00:07): [sssd[be[ldap_id]]] [sdap_id_op_connect_done] (0x0040): Connection failed: Clock skew too great",
			"(2026-09-18 10:00:08): [sssd[be[ldap_id]]] [sdap_id_op_connect_done] (0x0040): Connection failed: Clock skew too great",
			"(2026-09-18 10:00:09): [sssd[be[ldap_id]]] [sdap_id_op_connect_done] (0x0040): Connection failed: Clock skew too great",
		}, "\n"),
		// Rotated variant: must be collected (and scanned) like the main log.
		"sssd_example.test.log.1": "(2026-09-17 09:00:01): [sssd[be[ldap_id]]] [sdap_id_op_connect_done] (0x0040): Connection failed: Clock skew too great\n",
		// Syslog-formatted log: the host prefix is the only hostname source
		// (raw mode has no basic-environment.txt). It must end in .log to be
		// part of the -logdir input set, exactly like the SSSD logs.
		"messages.log": strings.Join([]string{
			"Aug 18 10:00:01 testhost01 sssd[33887]: [krb5_child] Clock skew too great with KDC",
			"Aug 18 10:00:02 testhost01 sssd[33888]: pam_sss authentication failed for user bob",
			"Aug 18 10:00:03 root kernel: [0.000000] boot noise that must not become a hostname",
		}, "\n"),
		// Non-log file: must never be picked up by collectLogFiles.
		"notes.txt": "raw notes, not a log\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// TestIsRawSSDLLogFile pins what counts as a raw SSSD log file: plain *.log
// and both rotation forms, and explicitly NOT compressed archives or the
// supportconfig file names (whose accidental inclusion would make -logdir
// silently analyse a supportconfig instead of raw logs).
func TestIsRawSSDLLogFile(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{"sssd_example.test.log", true},
		{"ldap_child.log", true},
		{"krb5_child.log", true},
		{"SSSD_EXAMPLE.TEST.LOG", true}, // extension match is case-insensitive
		{"sssd_example.test.log.1", true},
		{"sssd_example.test.log.12", true},
		{"sssd_example.test.log-20260918", true},
		{"sssd_example.test.log.gz", false}, // compressed: scanner reads plain text
		{"sssd_example.test.log.txt", false},
		{"sssd_example.test.logx", false},
		{"messages", false},
		{"messages.txt", false},
		{"sssd.txt", false},
		{"notes.txt", false},
	}
	for _, tc := range cases {
		if got := isRawSSDLLogFile(tc.name); got != tc.want {
			t.Errorf("isRawSSDLLogFile(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestCollectLogFiles_Directory: only the raw logs and their rotations, in a
// deterministic order, with names relative to the directory (the form the
// single-pass scanner expects).
func TestCollectLogFiles_Directory(t *testing.T) {
	dir := writeRawLogFixture(t)
	gotDir, files, err := collectLogFiles(dir)
	if err != nil {
		t.Fatalf("collectLogFiles(%q) failed: %v", dir, err)
	}
	if gotDir != dir {
		t.Errorf("dir = %q, want %q", gotDir, dir)
	}
	want := []string{"messages.log", "sssd_example.test.log", "sssd_example.test.log.1"}
	if len(files) != len(want) {
		t.Fatalf("files = %v, want %v", files, want)
	}
	for i := range want {
		if files[i] != want[i] {
			t.Errorf("files[%d] = %q, want %q (full list: %v)", i, files[i], want[i], files)
		}
	}
	for _, f := range files {
		if strings.HasSuffix(f, ".txt") {
			t.Errorf("non-log file %q must not be collected", f)
		}
	}
}

// TestCollectLogFiles_SingleFile: a single log file is accepted and resolved
// to its own directory, so -logdir also works on one extracted log.
func TestCollectLogFiles_SingleFile(t *testing.T) {
	dir := writeRawLogFixture(t)
	single := filepath.Join(dir, "sssd_example.test.log")
	gotDir, files, err := collectLogFiles(single)
	if err != nil {
		t.Fatalf("collectLogFiles(%q) failed: %v", single, err)
	}
	if gotDir != dir {
		t.Errorf("dir = %q, want %q", gotDir, dir)
	}
	if len(files) != 1 || files[0] != "sssd_example.test.log" {
		t.Errorf("files = %v, want [sssd_example.test.log]", files)
	}
}

// TestCollectLogFiles_MissingPath: a typo'd path must fail loudly instead of
// producing an empty (and therefore meaningless) report.
func TestCollectLogFiles_MissingPath(t *testing.T) {
	if _, _, err := collectLogFiles(filepath.Join(t.TempDir(), "does-not-exist")); err == nil {
		t.Error("expected an error for a missing path")
	}
}

// TestCollectLogFiles_NoLogFiles: a directory without any raw SSSD log must be
// rejected — the guard against running -logdir on the wrong input (a
// supportconfig, whose files are sssd.txt/messages/messages.txt).
func TestCollectLogFiles_NoLogFiles(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"sssd.txt":  "some supportconfig content\n",
		"messages":  "some syslog content\n",
		"notes.txt": "notes\n",
	})
	_, _, err := collectLogFiles(dir)
	if err == nil {
		t.Fatal("expected an error when no *.log files exist")
	}
	if !strings.Contains(err.Error(), "no raw SSSD log files") {
		t.Errorf("error = %v, want it to mention 'no raw SSSD log files'", err)
	}
}

// TestCollectLogFiles_NonLogFileRejects: a supportconfig file passed as the
// -logdir argument must not be treated as a raw log.
func TestCollectLogFiles_NonLogFileRejects(t *testing.T) {
	dir := setupMockDir(t, map[string]string{"sssd.txt": "content\n"})
	if _, _, err := collectLogFiles(filepath.Join(dir, "sssd.txt")); err == nil {
		t.Error("expected an error when the path is not a *.log file")
	}
}

// TestHarvestPIITokens_ExtractsIdentityValues: the values anonymization
// depends on must be found in raw logs — there is no sssd.conf to read them
// from in this mode.
func TestHarvestPIITokens_ExtractsIdentityValues(t *testing.T) {
	dir := writeRawLogFixture(t)
	_, files, err := collectLogFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	tokens := harvestPIITokens(dir, files)

	assertHas := func(kind string, list []string, want string) {
		t.Helper()
		for _, v := range list {
			if v == want {
				return
			}
		}
		t.Errorf("harvested %s = %v, want it to contain %q", kind, list, want)
	}
	assertHas("domain", tokens.domains, "example.test")
	assertHas("realm", tokens.realms, "EXAMPLE.TEST")
	assertHas("host", tokens.hosts, "testhost01")
	assertHas("host", tokens.hosts, "testhost01.example.test")
	assertHas("host", tokens.hosts, "dc01.example.test")

	if tokens.primaryDomain() == "" || tokens.primaryRealm() == "" || tokens.primaryHost() == "" {
		t.Errorf("primary accessors must be non-empty for this fixture: domain=%q realm=%q host=%q",
			tokens.primaryDomain(), tokens.primaryRealm(), tokens.primaryHost())
	}
}

// TestHarvestPIITokens_RejectsFalsePositives pins the anti-corruption rules:
// harvesting a service name, an English word after "realm=" or a syslog tag as
// an identity value would make the anonymizer rewrite harmless report text
// with placeholders.
func TestHarvestPIITokens_RejectsFalsePositives(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"a.log": strings.Join([]string{
			"(2026-09-18 10:00:01): [sssd[be[ldap_id]]] (0x0040): section loaded",
			"realm=supports fast negotiation requested",
			"Aug 18 10:00:01 root kernel: [0.000000] boot noise",
			"Domain [acme] is online", // single-label: explicit marker, must be kept
		}, "\n"),
	})
	_, files, err := collectLogFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	tokens := harvestPIITokens(dir, files)

	for _, d := range tokens.domains {
		if strings.EqualFold(d, "ldap_id") {
			t.Errorf("service name %q must never be harvested as a domain", d)
		}
	}
	for _, r := range tokens.realms {
		if r == "supports" {
			t.Errorf("English word %q must never be harvested as a realm", r)
		}
	}
	for _, h := range tokens.hosts {
		if h == "root" || h == "kernel" {
			t.Errorf("syslog tag %q must never be harvested as a hostname", h)
		}
	}
	if len(tokens.domains) != 1 || tokens.domains[0] != "acme" {
		t.Errorf("domains = %v, want exactly [acme] (explicit marker allows a single label)", tokens.domains)
	}
}

// TestAnalyzeLogsOnly_DetectsLogErrors: the whole point of raw-log mode is
// that the SSSD pattern engine still fires on *.log input.
func TestAnalyzeLogsOnly_DetectsLogErrors(t *testing.T) {
	dir := writeRawLogFixture(t)
	_, files, err := collectLogFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	report := analyzeLogsOnly(dir, files, false, nil)

	if len(report.SSSDLogErrors) == 0 {
		t.Fatal("raw-log analysis found no SSSD log errors; the pattern engine is not running on *.log files")
	}
	if len(report.Timeline) == 0 {
		t.Error("timeline must be populated from raw logs")
	}
	foundClockSkew := false
	for _, e := range report.SSSDLogErrors {
		if strings.Contains(e.Description, "Clock skew") {
			foundClockSkew = true
		}
	}
	if !foundClockSkew {
		t.Errorf("expected a 'Clock skew' log error, got %+v", report.SSSDLogErrors)
	}
}

// TestAnalyzeLogsOnly_NoSupportconfigFalsePositives: those findings can only be
// derived from a supportconfig; emitting them in raw-log mode would tell the
// user things that were never checked (and would bury real findings).
func TestAnalyzeLogsOnly_NoSupportconfigFalsePositives(t *testing.T) {
	dir := writeRawLogFixture(t)
	_, files, err := collectLogFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	report := analyzeLogsOnly(dir, files, false, nil)

	all := strings.Join(append(append([]string{}, report.Problems...), report.Warnings...), "\n")
	for _, banned := range []string{
		"sssd.conf or SSSD configuration block not found in the supportconfig",
		"sssd.service is not actively running",
		"No Kerberos Keytab (Machine Account) principal found",
	} {
		if strings.Contains(all, banned) {
			t.Errorf("supportconfig-only finding leaked into raw-log mode: %q", banned)
		}
	}
}

// TestAnalyzeLogsOnly_FieldsAndMarkers: unknown system fields carry the raw
// mode marker, while identity values are seeded from the logs themselves (so
// both the report and the anonymizer have real values to work with).
func TestAnalyzeLogsOnly_FieldsAndMarkers(t *testing.T) {
	dir := writeRawLogFixture(t)
	_, files, err := collectLogFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	report := analyzeLogsOnly(dir, files, false, nil)

	for field, value := range map[string]string{
		"SssdService":          report.SssdService,
		"WinbindService":       report.WinbindService,
		"NscdStatus":           report.NscdStatus,
		"TimeService":          report.TimeService,
		"KernelVersion":        report.KernelVersion,
		"HardwareManufacturer": report.HardwareManufacturer,
		"Hypervisor":           report.Hypervisor,
	} {
		if value != constants.RawLogModeNA {
			t.Errorf("%s = %q, want %q", field, value, constants.RawLogModeNA)
		}
	}
	if !report.SssdInstalled {
		t.Error("SssdInstalled must be true: raw SSSD logs can only exist if the daemon ran")
	}
	if report.SssdConfigFound {
		t.Error("SssdConfigFound must be false: no sssd.conf exists in raw-log mode")
	}
	if report.SearchDomain != "example.test" {
		t.Errorf("SearchDomain = %q, want the harvested domain", report.SearchDomain)
	}
	if report.KerberosRealm != "EXAMPLE.TEST" {
		t.Errorf("KerberosRealm = %q, want the harvested realm", report.KerberosRealm)
	}
	if report.Hostname != "testhost01" {
		t.Errorf("Hostname = %q, want the harvested syslog hostname", report.Hostname)
	}
}

// TestAnalyzeLogsOnly_SummaryClustersAndGraph: raw-log mode must produce the
// same derived artefacts as supportconfig mode (executive summary, temporal
// bursts, correlation graph) so the reports stay comparable.
func TestAnalyzeLogsOnly_SummaryClustersAndGraph(t *testing.T) {
	dir := writeRawLogFixture(t)
	_, files, err := collectLogFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	report := analyzeLogsOnly(dir, files, false, nil)

	s := report.Summary
	if s.LogErrorCount == 0 {
		t.Error("summary must count the log errors that were found")
	}
	if s.HealthScore < 0 || s.HealthScore > 100 {
		t.Errorf("HealthScore = %d, want 0..100", s.HealthScore)
	}
	if s.HealthScore == 100 {
		t.Error("HealthScore = 100 despite detected errors: the scoring did not run")
	}
	if len(report.TemporalClusters) == 0 {
		t.Error("expected at least one temporal cluster (clock-skew burst in the fixture)")
	}
	if len(report.Graph.Entities) == 0 {
		t.Error("correlation graph must contain the seeded identity entities")
	}
}

// assertNoRawPII fails when any of the fixture's identity values survives in
// the given serialized report. "example.com"/"redacted-host" are the expected
// placeholders; checking a couple of them as well keeps the assertion from
// passing vacuously.
func assertNoRawPII(t *testing.T, label, text string) {
	t.Helper()
	for _, token := range []string{
		"example.test", "EXAMPLE.TEST", "testhost01", "dc01",
	} {
		if strings.Contains(text, token) {
			t.Errorf("%s still contains raw PII token %q:\n%s", label, token, truncateForLog(text))
		}
	}
	if !strings.Contains(text, "example.com") && !strings.Contains(text, "redacted-host") {
		t.Errorf("%s contains neither expected placeholder (example.com / redacted-host) — "+
			"the redaction may not have run at all", label)
	}
}

// truncateForLog keeps a failing dump readable.
func truncateForLog(s string) string {
	if len(s) > 1200 {
		return s[:1200] + "...(truncated)"
	}
	return s
}

// TestAnalyzeLogsOnly_AnonymizeRedactsPII is THE guarantee of raw-log mode:
// with -anonymize the serialized report must not carry the domain, the realm,
// the short hostname or the DC name anywhere — not in the fields, not in the
// log examples, not in the timeline, clusters, KB suggestions or graph.
func TestAnalyzeLogsOnly_AnonymizeRedactsPII(t *testing.T) {
	dir := writeRawLogFixture(t)
	_, files, err := collectLogFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	report := analyzeLogsOnly(dir, files, true, nil)

	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	assertNoRawPII(t, "anonymized JSON report", string(data))
}

// TestAnalyzeLogsOnly_AnonymizeKeepsNAMarkers: the raw-mode placeholder must
// never be used as a redaction token — it would rewrite every field that
// carries it (this is the bug isRedactableToken exists for).
func TestAnalyzeLogsOnly_AnonymizeKeepsNAMarkers(t *testing.T) {
	// Fixture with logs that carry NO identity value at all, so every
	// system-level field keeps the N/A marker under anonymization.
	dir := setupMockDir(t, map[string]string{
		"plain.log": "(2026-09-18 10:00:01): [sssd[be[ldap_id]]] (0x0040): something went wrong\n",
	})
	_, files, err := collectLogFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	report := analyzeLogsOnly(dir, files, true, nil)

	if report.SssdService != constants.RawLogModeNA {
		t.Errorf("SssdService = %q, want the N/A marker to survive anonymization", report.SssdService)
	}
	if report.KernelVersion != constants.RawLogModeNA {
		t.Errorf("KernelVersion = %q, want the N/A marker to survive anonymization", report.KernelVersion)
	}
	if report.Hostname != constants.RawLogModeNA {
		t.Errorf("Hostname = %q, want the N/A marker (nothing to redact) — NOT a placeholder", report.Hostname)
	}
	if report.SearchDomain != constants.RawLogModeNA {
		t.Errorf("SearchDomain = %q, want the N/A marker", report.SearchDomain)
	}
	for _, e := range report.SSSDLogErrors {
		for _, ex := range e.Examples {
			if strings.Contains(ex, "redacted-host") || strings.Contains(ex, "example.com") {
				t.Errorf("placeholder injected into an example despite nothing to redact: %q", ex)
			}
		}
	}
}

// TestAnalyzeLogsOnly_WithoutAnonymizeKeepsValues: anti-over-redaction. The
// raw values must remain visible when the flag is off — a report the user did
// not ask to sanitize must stay fully diagnosable.
func TestAnalyzeLogsOnly_WithoutAnonymizeKeepsValues(t *testing.T) {
	dir := writeRawLogFixture(t)
	_, files, err := collectLogFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	report := analyzeLogsOnly(dir, files, false, nil)

	if report.Hostname != "testhost01" || report.SearchDomain != "example.test" || report.KerberosRealm != "EXAMPLE.TEST" {
		t.Errorf("identity fields were altered without -anonymize: host=%q domain=%q realm=%q",
			report.Hostname, report.SearchDomain, report.KerberosRealm)
	}
	data, _ := json.Marshal(report)
	if !strings.Contains(string(data), "testhost01") {
		t.Error("the raw hostname must appear in a non-anonymized report")
	}
}

// TestRunLogDirAnalyze_WritesRequestedFormats: the CLI contract — requested
// formats are written as <basename>_report.* in the working directory.
func TestRunLogDirAnalyze_WritesRequestedFormats(t *testing.T) {
	dir := writeRawLogFixture(t)
	work := t.TempDir()
	chdirInto(t, work)

	captureStdout(t, func() {
		if err := runLogDirAnalyze(dir, true, true, false, true); err != nil {
			t.Fatalf("runLogDirAnalyze failed: %v", err)
		}
	})

	base := filepath.Base(dir) // "sssd_example.test.log" fixture dir is a temp name
	for _, suffix := range []string{"_report.txt", "_report.html", "_report.json"} {
		matches, err := filepath.Glob(filepath.Join(work, "*"+suffix))
		if err != nil {
			t.Fatal(err)
		}
		if len(matches) != 1 {
			t.Errorf("expected exactly one %q report in %s, base name %q (got %v)",
				suffix, work, base, matches)
		}
	}
}

// TestRunLogDirAnalyze_DefaultsToStdoutOnly: -logdir mirrors -analyze — with
// no format flag nothing is written to disk (the report goes to stdout).
func TestRunLogDirAnalyze_DefaultsToStdoutOnly(t *testing.T) {
	dir := writeRawLogFixture(t)
	work := t.TempDir()
	chdirInto(t, work)

	captureStdout(t, func() {
		if err := runLogDirAnalyze(dir, false, false, false, false); err != nil {
			t.Fatalf("runLogDirAnalyze failed: %v", err)
		}
	})

	entries, err := os.ReadDir(work)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), "_report.txt") ||
			strings.HasSuffix(e.Name(), "_report.html") ||
			strings.HasSuffix(e.Name(), "_report.json") {
			t.Errorf("no report file must be written without format flags, found %q", e.Name())
		}
	}
}

// TestRunLogDirAnalyze_ErrorPaths: a missing path and a supportconfig-shaped
// directory must both fail with a clear error instead of an empty report.
func TestRunLogDirAnalyze_ErrorPaths(t *testing.T) {
	if err := runLogDirAnalyze(filepath.Join(t.TempDir(), "missing"), false, false, false, false); err == nil {
		t.Error("expected an error for a missing path")
	}

	supportconfig := setupMockDir(t, map[string]string{
		"sssd.txt": "config block\n",
		"messages": "syslog line\n",
	})
	err := runLogDirAnalyze(supportconfig, false, false, false, false)
	if err == nil {
		t.Fatal("expected an error when the directory has no raw *.log files")
	}
	if !strings.Contains(err.Error(), "no raw SSSD log files") {
		t.Errorf("error = %v, want it to mention the missing raw logs", err)
	}
}

// TestRunLogDirAnalyze_AnonymizedReportsHaveNoPII: the sanitization must hold
// for every format the user can share with a vendor, not just the JSON.
func TestRunLogDirAnalyze_AnonymizedReportsHaveNoPII(t *testing.T) {
	dir := writeRawLogFixture(t)
	work := t.TempDir()
	chdirInto(t, work)

	captureStdout(t, func() {
		if err := runLogDirAnalyze(dir, true, true, true, true); err != nil {
			t.Fatalf("runLogDirAnalyze failed: %v", err)
		}
	})

	for _, suffix := range []string{"_report.txt", "_report.html", "_report.json"} {
		matches, err := filepath.Glob(filepath.Join(work, "*"+suffix))
		if err != nil {
			t.Fatal(err)
		}
		if len(matches) != 1 {
			t.Fatalf("expected one %q report, got %v", suffix, matches)
		}
		data, err := os.ReadFile(matches[0])
		if err != nil {
			t.Fatal(err)
		}
		assertNoRawPII(t, "anonymized "+suffix, string(data))
	}
}

// TestAnalyzeData_IgnoresRawLogFiles is the isolation guard in the other
// direction: the supportconfig path must keep scanning only its own three
// files, so a stray *.log next to a supportconfig cannot change its findings.
func TestAnalyzeData_IgnoresRawLogFiles(t *testing.T) {
	dir := setupMockDir(t, map[string]string{
		"sssd_example.test.log": "(2026-09-18 10:00:01): [krb5_child] Clock skew too great\n",
		"ldap_child.log":        "(2026-09-18 10:00:02): [ldap_child] Invalid credentials\n",
	})
	report := analyzeData(dir, false, nil)

	if len(report.SSSDLogErrors) != 0 {
		t.Errorf("supportconfig analysis scanned raw *.log files: %+v", report.SSSDLogErrors)
	}
	if len(report.Timeline) != 0 {
		t.Errorf("supportconfig timeline must stay empty when only *.log files exist, got %v", report.Timeline)
	}
}
