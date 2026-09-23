// anonymize_syslog_host_test.go
//
// Regression tests for the SSSD Error Logs leak (reported with screenshot):
// Examples/RawLog lines carry the bare syslog hostname
// ("Aug 18 ... testhost01 ldap_child[1]: ..."), which matched neither the FQDN
// report.Hostname nor the "<short>.example.com" form, so the server name
// stayed visible in anonymized reports.
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// syslogShortHostFixture builds a supportconfig with an uppercase-FQDN
// hostname and syslog-style log lines carrying the bare short hostname.
func syslogShortHostFixture(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "syslog-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	files := map[string]string{
		"basic-environment.txt": "Hostname: TESTHOST01.example.test\nKernel: Linux testhost01 5.14.21-150500.55.44-default #1 SMP x86_64\n",
		"rpm.txt":               "sssd-2.9.4-150500.x86_64\n",
		"sssd.conf":             "[sssd]\nservices = nss, pam\n\n[domain/example.test]\nad_domain = example.test\nid_provider = ad\n",
		"sssd.txt": "Aug 18 10:00:01 testhost01 ldap_child[33887]: Failed to initialize credentials using keytab [MEMORY:/etc/krb5.keytab]: Preauthentication failed.\n" +
			"Aug 18 10:00:02 testhost01 ldap_child[33888]: (0x0020): SRV lookup for _ldap._tcp.example.test failed, KDC 192.0.2.10 unreachable\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// TestSyslogShortHostnameIsRedactedInLogExamples reproduces the reported
// screenshot: the bare syslog hostname must read "redacted-host".
func TestSyslogShortHostnameIsRedactedInLogExamples(t *testing.T) {
	report := analyzeData(syslogShortHostFixture(t), true, nil)
	if len(report.SSSDLogErrors) == 0 {
		t.Skip("fixture produced no log errors; adjust fixture")
	}
	for i, e := range report.SSSDLogErrors {
		for j, ex := range e.Examples {
			if strings.Contains(strings.ToLower(ex), "testhost01") {
				t.Errorf("sssd_log_errors[%d].examples[%d] leaks the syslog hostname: %.160q", i, j, ex)
			}
		}
	}
}

// TestTimelineAndClustersHaveNoRawPII: derived raw-log excerpts (timeline
// RawLog/Samples, cluster SampleRawLog, KB SampleLine) must be scrubbed like
// the log examples they come from.
func TestTimelineAndClustersHaveNoRawPII(t *testing.T) {
	report := analyzeData(syslogShortHostFixture(t), true, nil)
	check := func(where, s string) {
		t.Helper()
		lower := strings.ToLower(s)
		for _, raw := range []string{"testhost01", "example.test", "192.0.2.10"} {
			if strings.Contains(lower, strings.ToLower(raw)) {
				t.Errorf("%s leaks %q: %.160q", where, raw, s)
			}
		}
	}
	for i, ev := range report.Timeline {
		check("timeline["+itoaSmall(i)+"].raw_log", ev.RawLog)
		for j, s := range ev.Samples {
			check("timeline["+itoaSmall(i)+"].samples["+itoaSmall(j)+"]", s)
		}
	}
	for i, c := range report.TemporalClusters {
		check("temporal_clusters["+itoaSmall(i)+"].sample_raw_log", c.SampleRawLog)
	}
	for i, k := range report.KBSuggestions {
		check("kb_suggestions["+itoaSmall(i)+"].sample_line", k.SampleLine)
	}
}

// TestNonAnonymizedReportKeepsRawLines: without -anonymize nothing is
// redacted (guard against over-redaction leaking into the default path).
func TestNonAnonymizedReportKeepsRawLines(t *testing.T) {
	report := analyzeData(syslogShortHostFixture(t), false, nil)
	found := false
	for _, e := range report.SSSDLogErrors {
		for _, ex := range e.Examples {
			if strings.Contains(ex, "testhost01") {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("non-anonymized report lost the raw syslog hostname; fixture broken?")
	}
}

// TestShortHostnameFromKernelLineWithoutHostnameField is the exact reported
// scenario: basic-environment.txt has NO "Hostname:" line (only
// "Kernel: <node> ..."), and the log lines carry that bare nodename as the
// syslog prefix. The nodename must still be redacted.
func TestShortHostnameFromKernelLineWithoutHostnameField(t *testing.T) {
	dir, err := os.MkdirTemp("", "kernelfb-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	files := map[string]string{
		"basic-environment.txt": "Kernel: testhost01 5.14.21-150500.55.44-default #1 SMP x86_64\n",
		"rpm.txt":               "sssd-2.9.4-150500.x86_64\n",
		"sssd.conf":             "[sssd]\nservices = nss, pam\n\n[domain/corp.example]\nad_domain = corp.example\nid_provider = ad\n",
		"sssd.txt":              "2026-08-18T09:19:05.898657+02:00 testhost01 ldap_child[33887]: Failed to initialize credentials using keytab [MEMORY:/etc/krb5.keytab]: Preauthentication failed.\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	report := analyzeData(dir, true, nil)
	if len(report.SSSDLogErrors) == 0 {
		t.Skip("fixture produced no log errors; adjust fixture")
	}
	for _, e := range report.SSSDLogErrors {
		for _, ex := range e.Examples {
			if strings.Contains(strings.ToLower(ex), "testhost01") {
				t.Errorf("kernel-nodename fallback failed, leak: %.160q", ex)
			}
		}
	}
}
