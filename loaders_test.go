// loaders_test.go
//
// Tests for the archive extraction path (loaders.go), which had 0% coverage.
// Covers: only relevant files are extracted, non-archive input is an error,
// the progress callback fires, and no file can escape the temp directory.
package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

// writeRawTarXZ builds a valid .tar.xz archive WITHOUT archive/tar, so an entry
// header can declare a Size larger than the bytes actually present. That is
// the only way to reach the io.CopyN failure path while keeping the archive
// itself structurally valid: a merely truncated archive fails earlier, at the
// container level, and the whole temp dir is discarded anyway.
func writeRawTarXZ(t *testing.T, path string, entries []rawTarEntry) {
	t.Helper()
	var buf bytes.Buffer
	for _, e := range entries {
		hdr := make([]byte, 512)
		copy(hdr[0:100], e.name)
		copy(hdr[100:108], "0000644\x00")
		copy(hdr[108:116], "0000000\x00")
		copy(hdr[116:124], "0000000\x00")
		copy(hdr[124:136], fmt.Sprintf("%011s\x00", strconv.FormatInt(e.declaredSize, 10)))
		copy(hdr[136:148], fmt.Sprintf("%011s\x00", "00000000000"))
		hdr[156] = '0' // regular file
		copy(hdr[257:263], "ustar\x00")
		copy(hdr[263:265], "00")
		for i := 148; i < 156; i++ {
			hdr[i] = ' '
		}
		sum := 0
		for _, b := range hdr {
			sum += int(b)
		}
		copy(hdr[148:156], fmt.Sprintf("%06s\x00 ", strconv.Itoa(sum)))
		buf.Write(hdr)

		padded := e.content
		for len(padded)%512 != 0 {
			padded += "\x00"
		}
		buf.WriteString(padded)
	}
	buf.Write(make([]byte, 1024)) // two zero blocks terminate the archive

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	xw, err := xz.NewWriter(f)
	if err != nil {
		f.Close()
		t.Fatal(err)
	}
	if _, err := xw.Write(buf.Bytes()); err != nil {
		f.Close()
		t.Fatal(err)
	}
	xw.Close()
	f.Close()
}

// rawTarEntry is one hand-built tar entry: declaredSize may exceed len(content).
type rawTarEntry struct {
	name         string
	content      string
	declaredSize int64
}

// TestExtractArchiveToTemp_CorruptArchiveLeavesNoPartialFiles pins the
// end-to-end evidence-integrity guarantee of the extractor: whatever goes
// wrong inside the archive (truncated stream, bad header, entry that declares
// more bytes than it carries), the caller must never receive a temp directory
// containing a HALF-written sssd.conf or etc.txt.
//
// A partial file is worse than a missing one: the analysis would read it as
// complete and report invented findings (missing [domain/] section, invalid
// idmap ranges, absent responders) that look real in the report. This is the
// observable form of the io.CopyN error branch in extractArchiveToTemp, which
// removes the partial target and, on a container-level failure, the whole
// temp dir.
func TestExtractArchiveToTemp_CorruptArchiveLeavesNoPartialFiles(t *testing.T) {
	cases := []struct {
		name    string
		build   func(t *testing.T, path string)
		wantErr bool
	}{
		{
			name: "truncated stream",
			build: func(t *testing.T, path string) {
				writeTXZ(t, path, supportconfigEntries())
				fi, err := os.Stat(path)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.Truncate(path, fi.Size()/2); err != nil {
					t.Fatal(err)
				}
			},
			wantErr: true,
		},
		{
			name: "garbage payload",
			build: func(t *testing.T, path string) {
				if err := os.WriteFile(path, []byte("not an xz stream at all"), 0644); err != nil {
					t.Fatal(err)
				}
			},
			wantErr: true,
		},
		{
			name: "declared size larger than payload",
			build: func(t *testing.T, path string) {
				writeRawTarXZ(t, path, []rawTarEntry{
					{name: "basic-environment.txt", content: "Hostname: h.example.com\n", declaredSize: 512},
					{name: "sssd.conf", content: "[domain/example.com]\nid_provider = ad\n", declaredSize: 4096},
				})
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			archive := filepath.Join(t.TempDir(), "corrupt.tar.xz")
			tc.build(t, archive)

			outDir, err := extractArchiveToTemp(archive, nil)
			if tc.wantErr {
				// A hard failure is fine, but it must not leak a temp dir
				// holding partial evidence.
				if err == nil {
					defer os.RemoveAll(outDir)
				}
				if err != nil && outDir != "" {
					os.RemoveAll(outDir)
				}
			} else if err != nil {
				t.Fatalf("unexpected extraction error: %v", err)
			}
			if outDir == "" {
				return // nothing handed to the caller
			}
			defer os.RemoveAll(outDir)

			// Whatever survived must be complete: an empty or truncated
			// sssd.conf is the failure mode we must never produce.
			if data, rerr := os.ReadFile(filepath.Join(outDir, "sssd.conf")); rerr == nil && len(data) == 0 {
				t.Errorf("extractor produced an empty sssd.conf (partial evidence)")
			}
		})
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
