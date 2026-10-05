// cache_test.go
//
// Tests for the caches backing the scanning engine (cache.go): RegexCache
// identity/clearing and FileCache read/has/eviction/clear. Includes a
// concurrency pass for the race detector.
package main

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestRegexCache_GetReturnsSameInstance(t *testing.T) {
	rc := NewRegexCache()
	a := rc.Get(`foo\d+`)
	b := rc.Get(`foo\d+`)
	if a != b {
		t.Errorf("Get must return the cached instance for the same pattern")
	}
	if !a.MatchString("foo123") {
		t.Errorf("cached regex does not match")
	}
}

func TestRegexCache_MustCompileDelegatesToCache(t *testing.T) {
	rc := NewRegexCache()
	if rc.MustCompile(`bar`) != rc.Get(`bar`) {
		t.Errorf("MustCompile must reuse the cached instance")
	}
}

func TestRegexCache_ClearForcesRecompile(t *testing.T) {
	rc := NewRegexCache()
	before := rc.Get(`baz`)
	rc.Clear()
	after := rc.Get(`baz`)
	if before == after {
		t.Errorf("Clear must drop cached patterns (same pointer returned)")
	}
}

func TestRegexCache_ConcurrentGet(t *testing.T) {
	rc := NewRegexCache()
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			re := rc.Get(`concurrent-\d+`)
			if !re.MatchString("concurrent-7") {
				t.Error("bad match under concurrency")
			}
		}()
	}
	wg.Wait()
}

func TestFileCache_GetLinesAndHas(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("one\ntwo\nthree\n"), 0644); err != nil {
		t.Fatal(err)
	}
	fc := NewFileCache(10)
	if fc.Has(dir, "f.txt") {
		t.Errorf("Has must be false before first read")
	}
	lines := fc.GetLines(dir, "f.txt")
	if len(lines) != 3 || lines[0] != "one" || lines[2] != "three" {
		t.Errorf("unexpected cached lines: %q", lines)
	}
	if !fc.Has(dir, "f.txt") {
		t.Errorf("Has must be true after read")
	}
	// Second read must come from cache (same backing array content).
	again := fc.GetLines(dir, "f.txt")
	if len(again) != 3 {
		t.Errorf("cached re-read failed")
	}
}

func TestFileCache_MissingFile(t *testing.T) {
	fc := NewFileCache(10)
	if fc.GetLines(t.TempDir(), "nope.txt") != nil {
		t.Errorf("missing file must yield nil lines")
	}
	if fc.Has(t.TempDir(), "nope.txt") {
		t.Errorf("missing file must not be marked cached")
	}
}

func TestFileCache_DefaultCapacityAndEviction(t *testing.T) {
	fc := NewFileCache(0) // non-positive => default 20
	if fc.maxFiles != 20 {
		t.Errorf("expected default capacity 20, got %d", fc.maxFiles)
	}
	small := NewFileCache(2)
	dir := t.TempDir()
	for _, n := range []string{"a.txt", "b.txt", "c.txt"} {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("x\n"), 0644); err != nil {
			t.Fatal(err)
		}
		small.GetLines(dir, n)
	}
	if small.curFiles > 2 || len(small.buffers) > 2 {
		t.Errorf("cache must respect capacity 2, holds %d", len(small.buffers))
	}
	if !small.Has(dir, "c.txt") {
		t.Errorf("most recently added file must survive eviction")
	}
	small.Clear()
	if small.Has(dir, "c.txt") || small.curFiles != 0 {
		t.Errorf("Clear must empty the cache")
	}
}

// TestFileCache_EvictsOldestNotRandom pins the eviction ORDER. The cache
// documented "evict oldest" but ranged over the map, and Go randomizes map
// iteration: the victim was an arbitrary entry, so a hot file could be dropped
// and re-read and the scan cost varied between otherwise identical runs.
// Insertion order must now decide.
func TestFileCache_EvictsOldestNotRandom(t *testing.T) {
	for run := 0; run < 25; run++ {
		fc := NewFileCache(2)
		dir := t.TempDir()
		for _, n := range []string{"a.txt", "b.txt", "c.txt"} {
			if err := os.WriteFile(filepath.Join(dir, n), []byte("x\n"), 0644); err != nil {
				t.Fatal(err)
			}
			fc.GetLines(dir, n)
		}
		if fc.Has(dir, "a.txt") {
			t.Fatalf("run %d: the OLDEST entry (a.txt) must be evicted first", run)
		}
		if !fc.Has(dir, "b.txt") || !fc.Has(dir, "c.txt") {
			t.Fatalf("run %d: the two newest entries must survive, run %d", run, run)
		}
		if fc.curFiles != 2 {
			t.Errorf("run %d: curFiles=%d, want 2", run, fc.curFiles)
		}
	}
}

// TestCompareConfigsAnonymized_NilProgressFunc: the comparison entry point
// dereferenced progressFunc before the first report, while analyzeData and
// analyzeLogsOnly both treat nil as "no progress reporting". A GUI caller
// passing nil therefore panicked on a nil map/func call instead of running.
func TestCompareConfigsAnonymized_NilProgressFunc(t *testing.T) {
	dirA := setupMockDir(t, map[string]string{
		"sssd.conf": "[domain/example.com]\nid_provider = ad\n",
	})
	defer os.RemoveAll(dirA)
	dirB := setupMockDir(t, map[string]string{
		"sssd.conf": "[domain/example.com]\nid_provider = ad\ncase_sensitive = true\n",
	})
	defer os.RemoveAll(dirB)

	// Must not panic. Empty inputs also exercise the "no findings" path.
	res := CompareConfigsAnonymized(dirA, dirB, false, nil)
	if res.ScoreDelta == 0 && len(res.OnlyInB) == 0 {
		t.Logf("comparison completed (delta=%d, onlyInB=%d)", res.ScoreDelta, len(res.OnlyInB))
	}
}

func TestCreatePooledScanner_Scans(t *testing.T) {
	path := filepath.Join(t.TempDir(), "s.txt")
	if err := os.WriteFile(path, []byte("a\nb\n"), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sc := createPooledScanner(f)
	var got []string
	for sc.Scan() {
		got = append(got, sc.Text())
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("unexpected scan result: %q", got)
	}
}
