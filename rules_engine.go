// rules_engine.go
// Phase 5: data-driven analysis rules.
//
// The built-in detectors are compiled into the binary; this engine instead
// loads check definitions from YAML files so support engineers can add new
// site-specific or case-specific detectors WITHOUT recompiling:
//
//	rules:
//	  - name: "legacy-rc4-enctype"
//	    severity: warning        # critical | error | warning
//	    category: "crypto"
//	    files: ["sssd.conf", "krb5.conf"]
//	    message: "Legacy RC4 enctype found; modern AD rejects it."
//	    patterns:
//	      - "rc4-hmac"
//	    match: any               # any (default) | all
//	    pattern_type: literal    # literal (default) | regex
//
// Rules are additive: they can only ADD findings, never suppress built-in
// ones. Invalid rules are skipped (fail-safe) and the skip is reported as a
// Diagnostic with file and line — never silently swallowed.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"sssd-inspector/config"

	"gopkg.in/yaml.v3"
)

// AnalysisRule is one data-driven detector definition.
type AnalysisRule struct {
	Name        string   `yaml:"name" json:"name"`
	Severity    string   `yaml:"severity" json:"severity"`
	Category    string   `yaml:"category" json:"category"`
	Files       []string `yaml:"files" json:"files"`
	Patterns    []string `yaml:"patterns" json:"patterns"`
	Match       string   `yaml:"match" json:"match"`               // any (default) | all
	PatternType string   `yaml:"pattern_type" json:"pattern_type"` // literal (default) | regex
	Message     string   `yaml:"message" json:"message"`
}

// analysisRulesFile models one rules YAML document.
type analysisRulesFile struct {
	Rules []AnalysisRule `yaml:"rules"`
}

// Definition discovery roots. Package-level vars so tests can isolate the
// machine's real /etc/sssd-inspector and ~/.sssd-inspector directories
// (TestMain points them at nothing; individual tests re-point them as needed).
var (
	userDefinitionsRoot   = config.UserDefinitionsRoot
	systemDefinitionsRoot = config.SystemDefinitionsRoot
)

// ruleSeverity maps a rule severity string to the typed Severity.
func ruleSeverity(s string) (Severity, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "critical":
		return SevCritical, true
	case "error":
		return SevError, true
	case "warning", "warn":
		return SevWarning, true
	default:
		return SevWarning, false
	}
}

// yamlErrorLine extracts the 1-based line from a yaml.v3 parse error
// ("yaml: line 7: could not find expected ':'"). Returns 0 when unknown.
func yamlErrorLine(err error) int {
	if err == nil {
		return 0
	}
	m := regexp.MustCompile(`line (\d+)`).FindStringSubmatch(err.Error())
	if m == nil {
		return 0
	}
	var line int
	if _, scanErr := fmt.Sscanf(m[1], "%d", &line); scanErr != nil {
		return 0
	}
	return line
}

// ruleSequenceLines returns the source line of each entry of the top-level
// "rules" sequence, aligned by index with the decoded rule slice (both come
// from the same document, so positions match 1:1). Returns nil when the
// document cannot be walked (e.g. parse failed before this point).
func ruleSequenceLines(content []byte) []int {
	var root yaml.Node
	if err := yaml.Unmarshal(content, &root); err != nil {
		return nil
	}
	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return nil
	}
	doc := root.Content[0]
	if doc.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(doc.Content); i += 2 {
		key, val := doc.Content[i], doc.Content[i+1]
		if key.Value == "rules" && val.Kind == yaml.SequenceNode {
			lines := make([]int, 0, len(val.Content))
			for _, item := range val.Content {
				lines = append(lines, item.Line)
			}
			return lines
		}
	}
	return nil
}

// ruleCandidatePaths returns every rules file/directory location to scan,
// in discovery order: next to the executable, the working directory (both
// historical), then the system-wide and per-user definition roots that match
// the config.yaml search convention. Duplicate locations (the executable's
// own directory when it IS the working directory) are collapsed by absolute
// path so one file is never loaded twice.
func ruleCandidatePaths() []string {
	var raw []string
	if exePath, err := os.Executable(); err == nil {
		base := filepath.Dir(exePath)
		raw = append(raw,
			filepath.Join(base, "rules.yaml"),
			filepath.Join(base, "rules"),
		)
	}
	raw = append(raw, "rules.yaml", "rules")
	for _, root := range []string{systemDefinitionsRoot(), userDefinitionsRoot()} {
		if root == "" {
			continue
		}
		raw = append(raw,
			filepath.Join(root, "rules.yaml"),
			filepath.Join(root, "rules"),
		)
	}

	seen := make(map[string]bool, len(raw))
	paths := make([]string, 0, len(raw))
	for _, p := range raw {
		key := p
		if abs, err := filepath.Abs(p); err == nil {
			key = abs
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		paths = append(paths, p)
	}
	return paths
}

// RuleInfo pairs one parsed rule with the definition file it came from.
// Provenance is what makes "my rule is not firing" answerable: the Studio
// shows which file, line and scope a rule was loaded from, instead of
// leaving the user to guess where the inspector looked.
type RuleInfo struct {
	Rule  AnalysisRule    `json:"rule"`
	File  string          `json:"file"`
	Scope DefinitionScope `json:"scope"`
	Line  int             `json:"line,omitempty"`
}

// yamlKindName renders a yaml.Node kind for human-readable diagnostics.
func yamlKindName(k yaml.Kind) string {
	switch k {
	case yaml.DocumentNode:
		return "a document"
	case yaml.SequenceNode:
		return "a sequence"
	case yaml.MappingNode:
		return "a mapping"
	case yaml.ScalarNode:
		return "a scalar value"
	case yaml.AliasNode:
		return "an alias"
	default:
		return "an unknown node"
	}
}

// yamlTypeNames rewrites the Go type names yaml.v3 puts in unmarshal errors
// into the vocabulary of a rules file: "cannot unmarshal !!seq into
// main.analysisRulesFile" tells a user nothing about what to fix.
var yamlTypeNames = strings.NewReplacer(
	"!!seq", "a list",
	"!!map", "a mapping",
	"!!str", "text",
	"!!int", "a whole number",
	"!!float", "a decimal number",
	"!!bool", "true/false",
	"into main.analysisRulesFile", "into a rules document (expected a top-level \"rules\" list)",
	"into []main.AnalysisRule", "into a list of rules",
	"into main.AnalysisRule", "into a rule mapping",
	"into []string", "into a list of lines of text",
	"into string", "into a single line of text",
	"into int", "into a whole number",
	"into bool", "into true/false",
)

// friendlyYAMLError renders a yaml.v3 error without leaking Go type names,
// keeping the "line N:" prefix the user needs to find the problem.
func friendlyYAMLError(err error) string {
	if err == nil {
		return ""
	}
	return yamlTypeNames.Replace(err.Error())
}

// rulesDocumentCheck walks the raw YAML before it is decoded into structs and
// returns a human-readable issue when the document cannot possibly yield rules
// ("" when it is acceptable), plus any syntax error found while parsing.
//
// This is what turns the two mistakes real users make into actionable
// messages instead of a silent zero-rule load or a Go type name:
//
//   - a top-level sequence (the "- name: x" entries written without the
//     "rules:" header), and
//   - a mapping that used another key ("checks:", "rule:", ...).
//
// A document whose content is entirely commented out is a template, not a
// mistake (the shipped rules/example.yaml is written that way), so it is
// deliberately NOT flagged.
func rulesDocumentCheck(content []byte) (string, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(content, &root); err != nil {
		return "", err
	}
	if len(root.Content) == 0 {
		return "", nil
	}
	doc := root.Content[0]
	if doc.Kind != yaml.MappingNode {
		return fmt.Sprintf("expected a mapping with a top-level \"rules\" list, found %s", yamlKindName(doc.Kind)), nil
	}

	var otherKeys []string
	var rulesNode *yaml.Node
	for i := 0; i+1 < len(doc.Content); i += 2 {
		if key, val := doc.Content[i], doc.Content[i+1]; key.Value == "rules" {
			rulesNode = val
		} else {
			otherKeys = append(otherKeys, key.Value)
		}
	}
	if rulesNode == nil {
		if len(otherKeys) == 0 {
			return "", nil // an empty mapping declares no rules: valid, just empty
		}
		return fmt.Sprintf("no top-level \"rules\" list found (top-level keys: %s)", strings.Join(otherKeys, ", ")), nil
	}
	if rulesNode.Kind != yaml.SequenceNode {
		return fmt.Sprintf("the top-level \"rules\" value must be a list, found %s", yamlKindName(rulesNode.Kind)), nil
	}
	for i, item := range rulesNode.Content {
		// Aliases (*anchor) are resolved by the decoder, so they are accepted.
		if item.Kind != yaml.MappingNode && item.Kind != yaml.AliasNode {
			return fmt.Sprintf("rule #%d (line %d) must be a mapping with name, severity, message and patterns; found %s",
				i+1, item.Line, yamlKindName(item.Kind)), nil
		}
	}
	return "", nil
}

// parseRulesDocument parses and validates ONE rules YAML document. It is the
// single validation path shared by the file loader, the Definitions Studio
// editor (Validate button) and the CLI -validate-rules gate, so the same
// bytes can never be accepted by the editor and skipped by the analysis.
//
// label is the file path (or an editor placeholder) reported in diagnostics.
func parseRulesDocument(content []byte, label string) ([]RuleInfo, []Diagnostic) {
	structuralIssue, syntaxErr := rulesDocumentCheck(content)
	if syntaxErr != nil {
		return nil, []Diagnostic{{
			File:     label,
			Line:     yamlErrorLine(syntaxErr),
			Message:  fmt.Sprintf("invalid rules file: %s", friendlyYAMLError(syntaxErr)),
			Severity: SevWarning,
		}}
	}
	if structuralIssue != "" {
		return nil, []Diagnostic{{File: label, Message: structuralIssue, Severity: SevWarning}}
	}

	var doc analysisRulesFile
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return nil, []Diagnostic{{
			File:     label,
			Line:     yamlErrorLine(err),
			Message:  fmt.Sprintf("invalid rules file: %s", friendlyYAMLError(err)),
			Severity: SevWarning,
		}}
	}
	if len(doc.Rules) == 0 {
		return nil, nil
	}

	lines := ruleSequenceLines(content)
	var infos []RuleInfo
	var diags []Diagnostic
	seenNames := make(map[string]bool, len(doc.Rules))

	for i, r := range doc.Rules {
		line := 0
		if i < len(lines) {
			line = lines[i]
		}
		if strings.TrimSpace(r.Name) == "" {
			diags = append(diags, Diagnostic{
				File:     label,
				Line:     line,
				Message:  fmt.Sprintf("rule #%d: missing name, skipped (a nameless rule cannot be traced in the report)", i+1),
				Severity: SevWarning,
			})
			continue
		}
		if _, ok := ruleSeverity(r.Severity); !ok {
			diags = append(diags, Diagnostic{
				File:     label,
				Line:     line,
				Message:  fmt.Sprintf("rule %q: invalid severity %q (expected critical, error or warning), skipped", r.Name, r.Severity),
				Severity: SevWarning,
			})
			continue
		}
		if len(r.Patterns) == 0 || r.Message == "" {
			diags = append(diags, Diagnostic{
				File:     label,
				Line:     line,
				Message:  fmt.Sprintf("rule %q: missing patterns or message, skipped", r.Name),
				Severity: SevWarning,
			})
			continue
		}
		if seenNames[r.Name] {
			diags = append(diags, Diagnostic{
				File:     label,
				Line:     line,
				Message:  fmt.Sprintf("rule %q: duplicate name, both rules load and their findings cannot be told apart (rename one)", r.Name),
				Severity: SevWarning,
			})
		}
		seenNames[r.Name] = true
		infos = append(infos, RuleInfo{Rule: r, File: label, Line: line})
	}
	return infos, diags
}

// rulesFilesIn expands one candidate path into the concrete rules files it
// stands for: the file itself, or every *.yaml/*.yml inside a directory.
// A missing path yields nothing, so discovery and validation agree on what
// "there are rules here" means.
func rulesFilesIn(candidate string) []string {
	info, err := os.Stat(candidate)
	if err != nil {
		return nil
	}
	if !info.IsDir() {
		return []string{candidate}
	}
	matches, _ := filepath.Glob(filepath.Join(candidate, "*.yaml"))
	matches2, _ := filepath.Glob(filepath.Join(candidate, "*.yml"))
	return append(matches, matches2...)
}

// loadRulesAt parses every rules file behind one candidate path (the file
// itself, or all *.yaml/*.yml of a directory), tagging each rule with its
// provenance. A candidate without rules yields nothing.
func loadRulesAt(candidate string) ([]RuleInfo, []Diagnostic) {
	var infos []RuleInfo
	var diags []Diagnostic
	for _, f := range rulesFilesIn(candidate) {
		content, err := os.ReadFile(f)
		if err != nil {
			diags = append(diags, Diagnostic{
				File:     absPath(f),
				Message:  fmt.Sprintf("cannot read rules file: %v", err),
				Severity: SevWarning,
			})
			continue
		}
		fileInfos, fileDiags := parseRulesDocument(content, absPath(f))
		diags = append(diags, fileDiags...)
		for _, ri := range fileInfos {
			ri.Scope = definitionScopeOf(f)
			infos = append(infos, ri)
		}
	}
	return infos, diags
}

// loadAnalysisRuleInfos discovers and parses rules files from
// ruleCandidatePaths(), keeping per-rule provenance (file, scope, line).
// A missing rules directory is NOT an error: rules are optional. Every
// skipped file or rule is returned as a Diagnostic (with the offending line
// when known) so the caller can surface it in the report — the historical
// fmt.Printf output was invisible to GUI users.
func loadAnalysisRuleInfos() ([]RuleInfo, []Diagnostic) {
	var infos []RuleInfo
	var diags []Diagnostic

	// Rules are additive across files, so the same name can appear in more than
	// one of them — typically a copy of ./rules.yaml dropped into the user
	// scope. Both load (they are separate documents, and the report keeps both
	// findings), but their findings are indistinguishable, so it is worth
	// saying. Rules in the SAME document are handled by parseRulesDocument.
	seen := make(map[string]string, 0)
	for _, c := range ruleCandidatePaths() {
		cInfos, cDiags := loadRulesAt(c)
		for _, ri := range cInfos {
			if prev, ok := seen[ri.Rule.Name]; ok && prev != ri.File {
				diags = append(diags, Diagnostic{
					File:     ri.File,
					Line:     ri.Line,
					Message:  fmt.Sprintf("rule %q is also defined in %s; both load and their findings cannot be told apart (rename one)", ri.Rule.Name, prev),
					Severity: SevWarning,
				})
			}
			seen[ri.Rule.Name] = ri.File
			infos = append(infos, ri)
		}
		diags = append(diags, cDiags...)
	}
	return infos, diags
}

// loadAnalysisRules returns the parsed rules of loadAnalysisRuleInfos without
// provenance: the entry point used by the analysis pipeline.
func loadAnalysisRules() ([]AnalysisRule, []Diagnostic) {
	infos, diags := loadAnalysisRuleInfos()
	if len(infos) == 0 {
		return nil, diags
	}
	rules := make([]AnalysisRule, 0, len(infos))
	for _, ri := range infos {
		rules = append(rules, ri.Rule)
	}
	return rules, diags
}

// ruleMatcher tracks one rule's evaluation state inside the unified
// single-pass scan (see applyAnalysisRules).
type ruleMatcher struct {
	rule      AnalysisRule
	sev       Severity
	files     []string
	matchAll  bool
	regexes   []*regexp.Regexp // nil when the rule matches literals
	hits      []bool           // per-pattern hits on the current line
	hitCount  int
	lineNo    int64 // last line these hits belong to
	satisfied bool
	evidence  string
	inDirty   bool // queued for end-of-line evaluation
}

// defaultRuleFiles is the target set of a rule that declares no "files".
var defaultRuleFiles = []string{"sssd.conf", "sssd.txt", "messages", "messages.txt"}

// applyAnalysisRules evaluates the data-driven rules against the target
// supportconfig directory and appends findings to the report.
//
// All rules that share a target-file set are evaluated in ONE scan of those
// files: their literal patterns feed a single Aho-Corasick automaton whose
// canonical index fans out through a posting list to (rule, pattern position)
// pairs, and the group's regex rules run in the same per-line callback. This
// replaces the former one-scan-per-rule loop — O(rules × lines) file reads —
// with O(distinct file sets × lines), which is a single read for the common
// case where every rule uses the default file set.
//
// Semantics are unchanged: match:all still requires every pattern on the SAME
// line, evidence is still the first satisfying line, a satisfied rule stops
// being evaluated, and findings are emitted in input-rule order.
func applyAnalysisRules(dirPath string, report *ReportData, rules []AnalysisRule) {
	if len(rules) == 0 {
		return
	}

	// 1. Build one matcher per rule (severity, file set, compiled regexes).
	matchers := make([]*ruleMatcher, 0, len(rules))
	for _, rule := range rules {
		sev, _ := ruleSeverity(rule.Severity)
		files := rule.Files
		if len(files) == 0 {
			files = defaultRuleFiles
		}
		m := &ruleMatcher{
			rule:     rule,
			sev:      sev,
			files:    files,
			matchAll: strings.EqualFold(strings.TrimSpace(rule.Match), "all"),
			hits:     make([]bool, len(rule.Patterns)),
		}
		if strings.EqualFold(strings.TrimSpace(rule.PatternType), "regex") {
			m.regexes = make([]*regexp.Regexp, 0, len(rule.Patterns))
			for _, p := range rule.Patterns {
				m.regexes = append(m.regexes, globalRegexCache.Get("(?i)"+p))
			}
		}
		matchers = append(matchers, m)
	}

	// 2. Group matchers by file set; each group becomes ONE scan pass.
	groups := make(map[string][]*ruleMatcher)
	groupFiles := make(map[string][]string)
	var groupOrder []string
	for _, m := range matchers {
		key := strings.Join(m.files, "\x00")
		if _, ok := groups[key]; !ok {
			groupOrder = append(groupOrder, key)
			groupFiles[key] = m.files
		}
		groups[key] = append(groups[key], m)
	}

	var lineNo int64
	for _, key := range groupOrder {
		scanRuleGroup(dirPath, groupFiles[key], groups[key], &lineNo)
	}

	// 3. Emit in input order so the report stays deterministic.
	for _, m := range matchers {
		if !m.satisfied {
			continue
		}
		msg := fmt.Sprintf("[RULE: %s] %s", m.rule.Name, m.rule.Message)
		addConfigFindingEx(report, m.sev, m.rule.Category, msg,
			strings.Join(m.files, ","), "rule:"+m.rule.Name, 0, m.evidence,
			"rule:"+m.rule.Name, ConfidenceHeuristic, "")
	}
}

// rulePostingEntry fans one canonical AC match out to every (rule, position)
// that registered the literal. Precomputed once per rule group so the
// per-line callback is a map lookup, not a scan.
type rulePostingEntry struct {
	m   *ruleMatcher
	pos int
}

// ruleEvalState is the reusable per-line evaluation state for a set of
// data-driven rules. It holds the compiled literal automaton (+ posting
// list), the regex rules, and the transient per-line bookkeeping that
// scanRuleGroup used to keep locally. Extracting it here lets TWO callers
// share the logic: the standalone applyAnalysisRules scan AND the unified
// diagnostic single-pass scan (which feeds the same lines the diagnostic
// engine already reads, so the large log files are read once, not twice).
// Semantics are identical in both paths: match:all requires every pattern
// on the SAME line, evidence is the first satisfying line, a satisfied rule
// stops being evaluated.
type ruleEvalState struct {
	group   []*ruleMatcher
	ac      *ACMatcher
	posting map[int][]rulePostingEntry
	dirty   []*ruleMatcher
	lineNo  int64
}

// newRuleEvalState compiles the literal automaton and posting list for the
// given rule group. A nil automaton means the group has no literal patterns
// (regex-only group).
func newRuleEvalState(group []*ruleMatcher) *ruleEvalState {
	st := &ruleEvalState{group: group}
	var literal []*ruleMatcher
	var allPatterns []string
	for _, m := range group {
		if m.regexes == nil {
			literal = append(literal, m)
			allPatterns = append(allPatterns, m.rule.Patterns...)
		}
	}
	if len(allPatterns) > 0 {
		st.ac = globalACCache.Get(allPatterns)
		st.posting = make(map[int][]rulePostingEntry, len(allPatterns))
		for _, m := range literal {
			for pos, p := range m.rule.Patterns {
				ci, ok := st.ac.CanonicalIndex(p)
				if !ok {
					continue // empty pattern: never matches (as before)
				}
				st.posting[ci] = append(st.posting[ci], rulePostingEntry{m: m, pos: pos})
			}
		}
	}
	return st
}

// ruleFileSet returns true when the rule wants lines from the given file:
// either the rule targets that file explicitly, or it uses the default file
// set and the file belongs to it.
func ruleFileSet(m *ruleMatcher, file string) bool {
	for _, f := range m.files {
		if f == file {
			return true
		}
	}
	return false
}

// feedLine evaluates one line (from the given file) against every matcher of
// the group. ln is the running line counter shared across files, preserved
// so match:all keeps its same-line semantics and evidence order is stable.
func (st *ruleEvalState) feedLine(file, line string, ln int64) {
	st.lineNo = ln
	lowered := strings.ToLower(line)
	if st.ac != nil {
		st.ac.Match(lowered, func(ci int) {
			for _, e := range st.posting[ci] {
				if !ruleFileSet(e.m, file) {
					continue
				}
				st.mark(e.m, e.pos, ln)
			}
		})
	}
	for _, m := range st.group {
		if m.regexes == nil || m.satisfied {
			continue
		}
		if !ruleFileSet(m, file) {
			continue
		}
		for i, re := range m.regexes {
			if re.MatchString(line) {
				st.mark(m, i, ln)
			}
		}
	}

	// End-of-line: a rule fires when this line completed it (any: at
	// least one pattern; all: every pattern on this same line).
	for _, m := range st.dirty {
		m.inDirty = false
		if m.satisfied {
			continue
		}
		ok := m.hitCount > 0
		if m.matchAll {
			ok = m.hitCount == len(m.rule.Patterns)
		}
		if ok {
			m.satisfied = true
			m.evidence = strings.TrimSpace(line)
		}
	}
	st.dirty = st.dirty[:0]
}

// mark records a (rule, pattern-position) hit on line ln, resetting the
// per-line hit vector when a new line starts.
func (st *ruleEvalState) mark(m *ruleMatcher, pos int, ln int64) {
	if m.satisfied {
		return
	}
	if m.lineNo != ln {
		m.lineNo = ln
		m.hitCount = 0
		for i := range m.hits {
			m.hits[i] = false
		}
	}
	if !m.hits[pos] {
		m.hits[pos] = true
		m.hitCount++
	}
	if !m.inDirty {
		m.inDirty = true
		st.dirty = append(st.dirty, m)
	}
}

// scanRuleGroup streams the group's files once, evaluating every matcher of
// the group against each line with a single automaton pass (literals) plus
// the group's regex rules. Per-line work is delegated to ruleEvalState so
// the unified diagnostic scan reuses the exact same semantics with the file
// name known per line.
func scanRuleGroup(dirPath string, files []string, group []*ruleMatcher, lineNo *int64) {
	st := newRuleEvalState(group)
	scanFilesWithName(dirPath, files, func(file, line string) {
		*lineNo++
		st.feedLine(file, line, *lineNo)
	})
}
