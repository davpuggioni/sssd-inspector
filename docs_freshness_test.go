// docs_freshness_test.go — the docs must not describe a codebase that no
// longer exists.
//
// The migration off the imperative layer finished, but the prose kept claiming
// there was a survivor ("CorrelationGraph.js mounted through a ref", DOMHelper,
// the browser test-runner, an `npm run lint` that was never configured). A
// wrong sentence in a Markdown file fails no test, so it survives for months
// and quietly misleads the next reader. These checks compare the docs against
// the tree, so a stale claim now breaks a build.
//
// It lives in Go rather than in the vitest suite on purpose: the frontend
// tsconfig is browser-only, with no @types/node, so a Node-API test there
// would mean weakening the type gate for the whole app.
package main

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// repoRoot is the directory holding frontend/ and docs/. The test runs from
// the package directory, which is the repository root.
const repoRoot = "."

func markdownFiles(t *testing.T) map[string]string {
	t.Helper()
	out := map[string]string{}
	roots := []string{
		filepath.Join(repoRoot, "docs"),
		filepath.Join(repoRoot, "frontend", "README.md"),
		filepath.Join(repoRoot, "README.md"),
	}
	for _, root := range roots {
		info, err := os.Stat(root)
		if os.IsNotExist(err) {
			continue // an optional entry, not a failure
		}
		if err != nil {
			t.Fatalf("stat %s: %v", root, err)
		}
		if !info.IsDir() {
			b, err := os.ReadFile(root)
			if err != nil {
				t.Fatalf("read %s: %v", root, err)
			}
			out[filepath.Base(root)] = string(b)
			continue
		}
		err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".md") {
				return nil
			}
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel(repoRoot, path)
			out[rel] = string(b)
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}
	if len(out) == 0 {
		t.Fatal("no markdown found; the test is not looking at anything")
	}
	return out
}

func readDoc(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot, "docs", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(b)
}

// TestNoJavaScriptLeftUnderSrc is the fact the docs now assert, checked
// directly: the last file was ported, so a .js reappearing means the prose
// ("everything is TypeScript", "no JavaScript remains") is wrong again.
func TestNoJavaScriptLeftUnderSrc(t *testing.T) {
	var found []string
	src := filepath.Join(repoRoot, "frontend", "src")
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		switch filepath.Ext(path) {
		case ".js", ".jsx":
			rel, _ := filepath.Rel(repoRoot, path)
			found = append(found, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", src, err)
	}
	if len(found) > 0 {
		t.Errorf("frontend/src/ contains JavaScript %v; the docs state that none remains", found)
	}
}

// TestDocsDoNotDescribeDeletedFilesAsExisting checks the specific claims that
// were wrong. Mentioning a deleted file as history is fine and expected, so
// only a claim that it still exists counts.
func TestDocsDoNotDescribeDeletedFilesAsExisting(t *testing.T) {
	dev := readDoc(t, "frontend-development.md")
	arch := readDoc(t, "frontend-architecture.md")

	forbidden := map[string][]string{
		"frontend-development.md": {
			"still mounted",              // claimed a ref-mounted survivor
			"UIComponents.js` are still", // claimed the deleted widget
			"DOMHelper.addEventListener", // deleted helper, in a code sample
			"TestUtils.assert",           // deleted test framework
			"test:coverage",              // npm script that does not exist
			"CorrelationGraph.js",
		},
		"frontend-architecture.md": {
			"CorrelationGraph.js` is the one imperative",
			"one imperative renderer, mounted via ref",
			"mounts CorrelationGraph.js through a ref",
		},
	}
	docs := map[string]string{
		"frontend-development.md":  dev,
		"frontend-architecture.md": arch,
	}
	for name, patterns := range forbidden {
		for _, pattern := range patterns {
			if strings.Contains(docs[name], pattern) {
				t.Errorf("docs/%s still claims %q, which no longer exists", name, pattern)
			}
		}
	}
}

// TestDocsMentionTheRealGraphRenderer guards the replacement, so a later edit
// cannot delete the accurate sentence and leave nothing behind.
func TestDocsMentionTheRealGraphRenderer(t *testing.T) {
	for _, name := range []string{"frontend-development.md", "frontend-architecture.md"} {
		if !strings.Contains(readDoc(t, name), "CorrelationGraph.tsx") {
			t.Errorf("docs/%s should describe CorrelationGraph.tsx, the current renderer", name)
		}
	}
}

// TestDocumentedNpmScriptsExist stops the docs from instructing a command that
// is not there: `npm run lint` was documented for months without a script.
func TestDocumentedNpmScriptsExist(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(repoRoot, "frontend", "package.json"))
	if err != nil {
		t.Fatalf("read package.json: %v", err)
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(b, &pkg); err != nil {
		t.Fatalf("parse package.json: %v", err)
	}
	if len(pkg.Scripts) == 0 {
		t.Fatal("no scripts parsed; the check is not looking at anything")
	}

	re := regexp.MustCompile(`npm run ([\w:-]+)`)
	for name, text := range markdownFiles(t) {
		for _, m := range re.FindAllStringSubmatch(text, -1) {
			if _, ok := pkg.Scripts[m[1]]; !ok {
				t.Errorf("%s documents `npm run %s`, which package.json does not define", name, m[1])
			}
		}
	}
}
