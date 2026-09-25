// test_fixture_hygiene_test.go
//
// Repository rule: test fixtures, test comments, CI smoke data and
// documentation must only ever use reserved/generic names — RFC 2606/6761
// domains (example.test, example.com, ...), documentation IPs (192.0.2.0/24)
// and invented hostnames (testhost01) — NEVER a real company domain or a
// hostname taken from a machine that was analysed.
//
// Real values did slip in once (a company domain, an internal hostname, a
// private IP). This guard keeps them out for good: every banned value is
// listed below, and a new one must be added to the list the moment it is
// discovered.
package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// bannedRealTokens are values that were (or could be) copied from real
// environments. They are assembled from parts so this guard file does not
// contain the literals it bans — otherwise a plain grep would trip on the
// guard itself.
var bannedRealTokens = []string{
	strings.Join([]string{"intra", "swm", "de"}, "."), // real company domain
	"svweb" + "devi01", // internal hostname
	"web" + "dev01",    // hostname from log examples
	"srv" + "-prod-07", // prod-looking hostname
	"10" + ".44.12.10", // private IP from real logs
	"su" + "se.com",    // real company domain used as test data
}

// docURLHosts are real hosts that may legitimately appear as knowledge-base /
// product documentation URLs inside tests and docs. They are stripped before
// the ban check, so any OTHER occurrence of a real domain still fails.
var docURLHosts = []string{
	"support.scc.suse.com",
	"www.suse.com",
	"lists.suse.com",
}

// hygieneTarget reports whether the file is covered by the rule: Go sources
// (including tests), Markdown docs, and CI workflows. YAML elsewhere
// (config.yaml carries the author's real e-mail) is out of scope, as is the
// bundled KB data (real documentation URLs by design).
func hygieneTarget(path string) bool {
	if base := filepath.Base(path); base == "test_fixture_hygiene_test.go" {
		return false // this guard must not ban itself
	}
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".go"):
		return true
	case strings.HasSuffix(lower, ".md"):
		return true
	case strings.HasPrefix(lower, filepath.Join(".github", "workflows")+"/"):
		return strings.HasSuffix(lower, ".yml") || strings.HasSuffix(lower, ".yaml")
	}
	return false
}

// TestRepositoryUsesOnlyGenericTestNames scans the repository for real-world
// domain/hostname/IP values used as test data or in comments.
func TestRepositoryUsesOnlyGenericTestNames(t *testing.T) {
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules", "frontend", "kb_articles", "dist":
				return filepath.SkipDir // out of scope (JS sources / KB data / build output)
			}
			return nil
		}
		if !hygieneTarget(path) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := strings.ToLower(string(data))
		for _, host := range docURLHosts {
			text = strings.ReplaceAll(text, strings.ToLower(host), "")
		}
		for _, token := range bannedRealTokens {
			if strings.Contains(text, strings.ToLower(token)) {
				t.Errorf("%s contains the banned real-world value %q — use a reserved/generic test name instead "+
					"(example.test, testhost01, 192.0.2.10)", path, token)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}
}
