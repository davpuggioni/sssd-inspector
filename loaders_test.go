// loaders_test.go
//
// Tests for the archive extraction path (loaders.go), which had 0% coverage.
// Covers: only relevant files are extracted, non-archive input is an error,
// the progress callback fires, and no file can escape the temp directory.
package main

import (
	"archive/tar"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ulikunitz/xz"
)

// tarEntry describes one file to pack into the test archive.
type tarEntry struct {
	name    string
	content string
}

// writeTXZ builds a .tar.xz archive at path containing the given entries.
func writeTXZ(t *testing.T, path string, entries []tarEntry) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	xw, err := xz.NewWriter(f)
	if err != nil {
		f.Close()
		t.Fatal(err)
	}
	tw := tar.NewWriter(xw)
	for _, e := range entries {
		if err := tw.WriteHeader(&tar.Header{
			Name:     e.name,
			Mode:     0644,
			Size:     int64(len(e.content)),
			Typeflag: tar.TypeReg,
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(e.content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := xw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

// supportconfigEntries returns a representative fake supportconfig payload.
func supportconfigEntries() []tarEntry {
	return []tarEntry{
		{name: "srv/supportconfig-abc123/sssd.txt", content: "Failed to initialize credentials using keytab [MEMORY:/etc/krb5.keytab]: Preauthentication failed\n"},
		{name: "srv/supportconfig-abc123/basic-environment.txt", content: "Hostname: host.example.com\n"},
		{name: "srv/supportconfig-abc123/rpm.txt", content: "sssd-2.9.4-150500.x86_64\n"},
	}
}

// TestExtractArchiveToTemp_ExtractsRelevantFiles pins the extraction contract:
// relevant files land flat in the temp dir; irrelevant files are skipped.
func TestExtractArchiveToTemp_ExtractsRelevantFiles(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "supportconfig.tar.xz")
	entries := append([]tarEntry{}, supportconfigEntries()...)
	entries = append(entries, tarEntry{name: "srv/supportconfig-abc123/tarball-journal.bin", content: "\x00\x01binary-noise\n"})
	writeTXZ(t, archive, entries)

	var progress []string
	dir, err := extractArchiveToTemp(archive, func(msg string, pct int) {
		progress = append(progress, msg)
	})
	if err != nil {
		t.Fatalf("extractArchiveToTemp failed: %v", err)
	}
	defer os.RemoveAll(dir)

	for _, name := range []string{"sssd.txt", "basic-environment.txt", "rpm.txt"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected relevant file %s to be extracted: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "tarball-journal.bin")); !os.IsNotExist(err) {
		t.Errorf("irrelevant file must be skipped during extraction")
	}
	if len(progress) == 0 {
		t.Errorf("expected the progress callback to fire at least once")
	}
}

// TestExtractArchiveToTemp_InvalidArchive: garbage input must be an error,
// never a silent empty result.
func TestExtractArchiveToTemp_InvalidArchive(t *testing.T) {
	garbage := filepath.Join(t.TempDir(), "not-an-archive.tar.xz")
	if err := os.WriteFile(garbage, []byte("this is not xz data"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := extractArchiveToTemp(garbage, nil); err == nil {
		t.Errorf("expected an error for a non-xz archive, got nil")
	}
}

// TestExtractArchiveToTemp_PathTraversal: nested directory entries must be
// flattened (Base) and cannot escape the extraction directory.
func TestExtractArchiveToTemp_PathTraversal(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "traversal.tar.xz")
	writeTXZ(t, archive, []tarEntry{
		{name: "../../sssd.txt", content: "escaped content\n"},
		{name: "a/b/c/basic-environment.txt", content: "Hostname: sneaky.example.com\n"},
	})

	dir, err := extractArchiveToTemp(archive, nil)
	if err != nil {
		t.Fatalf("extractArchiveToTemp failed: %v", err)
	}
	defer os.RemoveAll(dir)

	parent := filepath.Dir(dir)
	leftovers, _ := filepath.Glob(filepath.Join(parent, "sssd.txt"))
	if len(leftovers) > 0 {
		t.Errorf("extraction escaped the temp dir: %v", leftovers)
	}
	// Even the traversal entry (relevant basename) must resolve INSIDE dir.
	got, err := os.ReadFile(filepath.Join(dir, "sssd.txt"))
	if err != nil || !strings.Contains(string(got), "escaped content") {
		t.Errorf("relevant basename escaping attempt should be flattened into the temp dir: %v, %q", err, got)
	}
}
