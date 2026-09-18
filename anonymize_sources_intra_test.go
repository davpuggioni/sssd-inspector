// anonymize_sources_intra_test.go
//
// Regression tests for the provenance PII leak: ConfigFindings (and their
// graph copies) carry SourceKey values like "domain/<domain>.<key>" whose
// section name is the raw domain. Scrubbing Message/Evidence alone left the
// domain visible in the HTML "Configuration Findings (with provenance)"
// section. These tests pin that SourceKey/SourcePath are redacted too.

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// provenanceFixture builds a supportconfig whose sssd.conf has a per-domain
// section (so SourceKey embeds the raw domain, e.g.
// "domain/intra.swm.de.ldap_id_mapping") plus an FQDN hostname report value.
func provenanceFixture(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "prov-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	files := map[string]string{
		"sssd.conf": "[sssd]\nservices = nss, pam\n\n" +
			"[domain/intra.swm.de]\nad_domain = intra.swm.de\nldap_id_mapping = False\nid_provider = ad\n",
		"rpm.txt":  "sssd-2.9.4-150500.x86_64\n",
		"sssd.txt": "(0x0020): some line mentioning intra.swm.de\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func failIfContainsRawDomain(t *testing.T, where, s string) {
	t.Helper()
	for _, raw := range []string{"intra.swm.de", "INTRA.SWM.DE"} {
		if strings.Contains(s, raw) {
			t.Errorf("%s leaks raw domain %q: %.160q", where, raw, s)
		}
	}
}

func itoaSmall(i int) string {
	if i == 0 {
		return "0"
	}
	var b [8]byte
	p := len(b)
	for n := i; n > 0; n /= 10 {
		p--
		b[p] = byte('0' + n%10)
	}
	return string(b[p:])
}

// TestAnonymizedSourceKeysHaveNoRawDomain: no ConfigFinding SourceKey may
// expose the internal domain when anonymization is on.
func TestAnonymizedSourceKeysHaveNoRawDomain(t *testing.T) {
	report := analyzeData(provenanceFixture(t), true, nil)
	if len(report.ConfigFindings) == 0 {
		t.Skip("fixture produced no config findings; adjust fixture")
	}
	for i, f := range report.ConfigFindings {
		failIfContainsRawDomain(t, "config_findings["+itoaSmall(i)+"].source_key", f.SourceKey)
		failIfContainsRawDomain(t, "config_findings["+itoaSmall(i)+"].source_path", f.SourcePath)
		failIfContainsRawDomain(t, "config_findings["+itoaSmall(i)+"].message", f.Message)
		failIfContainsRawDomain(t, "config_findings["+itoaSmall(i)+"].evidence", f.Evidence)
	}
}

// TestAnonymizedGraphProvenanceHasNoRawDomain: graph findings (copies of
// ConfigFindings) and graph entities must be redacted the same way.
func TestAnonymizedGraphProvenanceHasNoRawDomain(t *testing.T) {
	report := analyzeData(provenanceFixture(t), true, nil)
	for i, f := range report.Graph.Findings {
		if f.SourceKey == "" {
			continue
		}
		failIfContainsRawDomain(t, "graph.findings["+itoaSmall(i)+"].source_key", f.SourceKey)
		failIfContainsRawDomain(t, "graph.findings["+itoaSmall(i)+"].source_path", f.SourcePath)
	}
	for i, e := range report.Graph.Entities {
		failIfContainsRawDomain(t, "graph.entities["+itoaSmall(i)+"].value", e.Value)
		failIfContainsRawDomain(t, "graph.entities["+itoaSmall(i)+"].id", e.ID)
	}
}

// TestSourceKeyKeepsActionableSection: provenance must still tell the user
// WHERE to act — "domain/example.com.<key>" — so only the domain is
// replaced, not the whole key.
func TestSourceKeyKeepsActionableSection(t *testing.T) {
	report := analyzeData(provenanceFixture(t), true, nil)
	found := false
	for _, f := range report.ConfigFindings {
		if strings.HasPrefix(f.SourceKey, "domain/") {
			found = true
			if !strings.Contains(f.SourceKey, "example.com") {
				t.Errorf("source_key %q lost the actionable domain/section structure", f.SourceKey)
			}
		}
	}
	if !found {
		t.Skip("no domain-scoped source_key in fixture")
	}
}

// TestHostnameEntityHasNoShortHostname: the graph hostname entity must not
// leak the server short name ("srv-07.example.com").
func TestHostnameEntityHasNoShortHostname(t *testing.T) {
	dir := provenanceFixture(t)
	const hostLine = "Hostname: srv-prod-07.intra.swm.de\n"
	f, err := os.OpenFile(filepath.Join(dir, "basic-environment.txt"),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(hostLine); err != nil {
		t.Fatal(err)
	}
	f.Close()

	report := analyzeData(dir, true, nil)
	withHost := false
	for _, e := range report.Graph.Entities {
		if e.Kind != "hostname" {
			continue
		}
		withHost = true
		if strings.Contains(strings.ToLower(e.Value), "srv-prod-07") {
			t.Errorf("hostname entity value %q leaks the server short name", e.Value)
		}
		if strings.Contains(strings.ToLower(e.ID), "srv-prod-07") {
			t.Errorf("hostname entity id %q leaks the server short name", e.ID)
		}
	}
	if !withHost {
		t.Skip("no hostname entity produced; adjust fixture")
	}
}
