// ahocorasick.go - Aho-Corasick multi-pattern string matching automaton.
//
// ~97%+ of the scanner's patterns are literal strings, not regular
// expressions. Matching them through one giant regex alternation forces RE2
// to explore hundreds of alternation branches per line and pushes against
// its program-size limits. An Aho-Corasick trie instead matches ALL literal
// patterns in a single O(n) pass over each line, regardless of how many
// patterns are registered — the cost per character is constant.
//
// The automaton is case-insensitive (patterns and text are folded to
// lowercase at build/scan time) and works at the byte level, which is safe
// because all registered patterns are ASCII literals.

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"sync"
)

// acNode is one state of the Aho-Corasick goto automaton.
type acNode struct {
	children map[byte]*acNode
	fail     *acNode // failure link (longest proper suffix that is also a prefix)
	out      []int   // canonical indices of patterns ending at this state
}

// ACMatcher is a compiled Aho-Corasick automaton over a set of literal patterns.
type ACMatcher struct {
	root      *acNode
	patterns  []string       // lowercased, deduplicated patterns (canonical order)
	canonical map[string]int // lowercased literal -> canonical index
}

// newACMatcher builds the automaton (trie + BFS failure links + merged outputs).
// Patterns are deduplicated: identical literals map to one canonical index.
// Empty patterns are ignored.
func newACMatcher(patterns []string) *ACMatcher {
	root := &acNode{children: make(map[byte]*acNode)}
	m := &ACMatcher{
		root:      root,
		canonical: make(map[string]int, len(patterns)),
	}

	// Deduplicate: canonical index per unique lowercased literal.
	canonical := m.canonical
	for _, p := range patterns {
		lp := strings.ToLower(p)
		if lp == "" {
			continue
		}
		if _, ok := canonical[lp]; ok {
			continue
		}
		idx := len(m.patterns)
		canonical[lp] = idx
		m.patterns = append(m.patterns, lp)

		node := root
		for i := 0; i < len(lp); i++ {
			c := lp[i]
			next, ok := node.children[c]
			if !ok {
				next = &acNode{children: make(map[byte]*acNode)}
				node.children[c] = next
			}
			node = next
		}
		node.out = append(node.out, idx)
	}

	// BFS to build failure links; outputs from failure targets are merged
	// into each node so every match is reported without following fail
	// chains during the scan itself.
	queue := make([]*acNode, 0, 64)
	for _, child := range root.children {
		child.fail = root
		queue = append(queue, child)
	}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		for c, child := range n.children {
			queue = append(queue, child)
			f := n.fail
			for {
				if next, ok := f.children[c]; ok && next != child {
					child.fail = next
					break
				}
				if f == root {
					child.fail = root
					break
				}
				f = f.fail
			}
			child.out = append(child.out, child.fail.out...)
		}
	}
	return m
}

// MatchBoundSelective behaves like Match but applies a word-boundary check
// only to the patterns selected by needBound (canonical index -> true means
// "this pattern requires boundaries"). It exists so a single automaton can
// host both long diagnostic literals (substring semantics, e.g. "[28][No
// space left on device]") and short ones ("kdc", "ad ") that would otherwise
// match inside unrelated words ("backupkdc", "systemd").
//
// The boundary rule is the same as MatchWordBound: the byte before the match
// start and the byte after the match end must both be absent or non-word
// bytes (word byte = ASCII letter, digit or '_'). needBound==nil means no
// pattern requires boundaries (identical to Match); the check costs two
// byte loads per selected hit, so scan throughput is unchanged.
func (m *ACMatcher) MatchBoundSelective(text string, needBound func(patternIndex int) bool, fn func(patternIndex int)) {
	if needBound == nil {
		m.Match(text, fn)
		return
	}
	node := m.root
	for i := 0; i < len(text); i++ {
		c := lowerASCII(text[i])
		for node != m.root && node.children[c] == nil {
			node = node.fail
		}
		if next, ok := node.children[c]; ok {
			node = next
		} else {
			node = m.root
		}
		for _, idx := range node.out {
			if !needBound(idx) {
				fn(idx)
				continue
			}
			patLen := len(m.patterns[idx])
			start := i - patLen + 1
			if start > 0 && isWordByte(text[start-1]) {
				continue
			}
			end := i + 1
			if end < len(text) && isWordByte(text[end]) {
				continue
			}
			fn(idx)
		}
	}
}

// needsWordBoundary reports whether a literal pattern is short enough that a
// substring match would be unreliable: trimmed length <= 4 (e.g. "kdc",
// "ad ", "tgt "). Such patterns must only fire at word boundaries; longer
// literals keep substring semantics so existing diagnostics are unaffected.
func needsWordBoundary(pattern string) bool {
	t := strings.TrimSpace(strings.ToLower(pattern))
	return t != "" && len(t) <= 4
}
func (m *ACMatcher) Match(text string, fn func(patternIndex int)) {
	node := m.root
	for i := 0; i < len(text); i++ {
		c := lowerASCII(text[i])
		for {
			if next, ok := node.children[c]; ok {
				node = next
				break
			}
			if node == m.root {
				break
			}
			node = node.fail
		}
		if len(node.out) > 0 {
			for _, idx := range node.out {
				fn(idx)
			}
		}
	}
}

// isWordByte reports whether b can appear inside a word. The boundary
// definition is deliberately conservative: letters, digits and '_' keep a
// token together, everything else ('.', '-', ' ', '[', ... and any
// non-ASCII byte) separates words. This mirrors how the SSSD log patterns
// were authored ("tgt failed" must not match "intgt_failed").
func isWordByte(b byte) bool {
	return b >= '0' && b <= '9' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b == '_'
}

// MatchWordBound behaves like Match but additionally applies a word-boundary
// check on the matched occurrence: the byte immediately before and after the
// match must NOT be a word byte (or must be absent).
//
// This is the stage-2 verifier proposed for short, collision-prone literals
// ("ad ", "tgt ", "kdc", "srv"): a plain substring match would report
// "login failed_retry" as a KDC hit. The check costs two byte loads per
// CANDIDATE match — it never runs on lines that produce no AC hit, so the
// single-pass scanning cost is unchanged.
//
// fn receives the canonical index exactly as Match does; a filtered-out
// candidate is silently dropped.
func (m *ACMatcher) MatchWordBound(text string, fn func(patternIndex int)) {
	node := m.root
	for i := 0; i < len(text); i++ {
		c := lowerASCII(text[i])
		for {
			if next, ok := node.children[c]; ok {
				node = next
				break
			}
			if node == m.root {
				break
			}
			node = node.fail
		}
		if len(node.out) == 0 {
			continue
		}
		for _, idx := range node.out {
			start := i - len(m.patterns[idx]) + 1
			if start > 0 && isWordByte(text[start-1]) {
				continue
			}
			end := i + 1
			if end < len(text) && isWordByte(text[end]) {
				continue
			}
			fn(idx)
		}
	}
}

// Len returns the number of unique patterns registered in the automaton.
func (m *ACMatcher) Len() int { return len(m.patterns) }

// CanonicalIndex returns the canonical index of a registered literal.
//
// Match reports CANONICAL indices: identical literals (or literals differing
// only by case) share one single index, so the index returned to the caller is
// NOT necessarily the position the literal occupied in the slice passed to
// newACMatcher. Callers that keep a parallel metadata table MUST map through
// this method instead of indexing that table with the canonical index:
// otherwise every literal registered after the first duplicate is attributed
// to the wrong metadata (wrong description, wrong handler, wrong evidence).
func (m *ACMatcher) CanonicalIndex(literal string) (int, bool) {
	idx, ok := m.canonical[strings.ToLower(literal)]
	return idx, ok
}

// lowerASCII folds an ASCII byte to lowercase, leaving other bytes untouched.
func lowerASCII(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}

// ACCache provides thread-safe caching of compiled Aho-Corasick automatons,
// mirroring RegexCache. The pattern set for a given scan is static per
// process, so each automaton is compiled exactly once.
type ACCache struct {
	mu    sync.Mutex
	items map[string]*ACMatcher
}

// acFingerprint computes a stable cache key for a pattern set.
//
// The key is the hex SHA-256 of the length-prefixed, lowercased, SORTED
// pattern set. Two properties matter for a forensic tool:
//
//   - Order independence: the same set registered in a different order must
//     hit the same cache entry instead of recompiling a duplicate automaton
//     (the previous FNV-1a key was sensitive to ordering).
//   - No silent aliasing: a 64-bit unseeded hash could collide, and a
//     collision would hand back an automaton built from a DIFFERENT pattern
//     set — wrong verdicts attributed to the right rules. SHA-256 removes the
//     risk; pattern sets are tiny, so the extra hashing cost is negligible.
func acFingerprint(patterns []string) string {
	folded := make([]string, 0, len(patterns))
	for _, p := range patterns {
		folded = append(folded, strings.ToLower(p))
	}
	sort.Strings(folded)

	h := sha256.New()
	var lenBuf [8]byte
	for _, p := range folded {
		n := len(p)
		for i := 0; i < 8; i++ {
			lenBuf[i] = byte(n >> (8 * i))
		}
		h.Write(lenBuf[:])
		h.Write([]byte(p))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// NewACCache creates a new ACCache.
func NewACCache() *ACCache {
	return &ACCache{items: make(map[string]*ACMatcher)}
}

// Get returns a compiled automaton for the given pattern set, building it on
// first access. Thread-safe: multiple goroutines can call Get concurrently.
func (c *ACCache) Get(patterns []string) *ACMatcher {
	key := acFingerprint(patterns)

	c.mu.Lock()
	defer c.mu.Unlock()
	if m, ok := c.items[key]; ok {
		return m
	}
	m := newACMatcher(patterns)
	c.items[key] = m
	return m
}

// Clear removes all cached automatons, freeing memory.
func (c *ACCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*ACMatcher)
}

// globalACCache is the application-wide Aho-Corasick cache.
var globalACCache = NewACCache()
