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
	"hash/fnv"
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
	root     *acNode
	patterns []string // lowercased, deduplicated patterns
}

// newACMatcher builds the automaton (trie + BFS failure links + merged outputs).
// Patterns are deduplicated: identical literals map to one canonical index.
// Empty patterns are ignored.
func newACMatcher(patterns []string) *ACMatcher {
	root := &acNode{children: make(map[byte]*acNode)}
	m := &ACMatcher{root: root}

	// Deduplicate: canonical index per unique lowercased literal.
	canonical := make(map[string]int)
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

// Match walks text once and invokes fn with the canonical pattern index for
// every pattern occurrence found (including overlapping matches). fn may be
// called multiple times per pattern if it occurs more than once.
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

// Len returns the number of unique patterns registered in the automaton.
func (m *ACMatcher) Len() int { return len(m.patterns) }

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
	items map[uint64]*ACMatcher
}

// acFingerprint computes a 64-bit FNV-1a hash over the joined pattern set.
// Using a fixed-size fingerprint instead of the concatenated pattern string
// keeps the cache key at constant size regardless of how many patterns are
// registered (previously the key duplicated the whole pattern set in memory).
func acFingerprint(patterns []string) uint64 {
	h := fnv.New64a()
	for _, p := range patterns {
		h.Write([]byte(p))
		h.Write([]byte{0})
	}
	return h.Sum64()
}

// NewACCache creates a new ACCache.
func NewACCache() *ACCache {
	return &ACCache{items: make(map[uint64]*ACMatcher)}
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
	c.items = make(map[uint64]*ACMatcher)
}

// globalACCache is the application-wide Aho-Corasick cache.
var globalACCache = NewACCache()
