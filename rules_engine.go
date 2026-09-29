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
	"context"
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

// applyAnalysisRules evaluates the data-driven rules against the target
// supportconfig directory and appends findings to the report. Literal
// patterns run through the Aho-Corasick engine; regex patterns through RE2.
func applyAnalysisRules(dirPath string, report *ReportData, rules []AnalysisRule) {
	for _, rule := range rules {
		sev, _ := ruleSeverity(rule.Severity)
		files := rule.Files
		if len(files) == 0 {
			files = []string{"sssd.conf", "sssd.txt", "messages", "messages.txt"}
		}

		regexMode := strings.EqualFold(strings.TrimSpace(rule.PatternType), "regex")
		matchAll := strings.EqualFold(strings.TrimSpace(rule.Match), "all")

		var matchLine func(line string, hit func(patternIdx int))
		if regexMode {
			regexes := make([]*regexp.Regexp, 0, len(rule.Patterns))
			for _, p := range rule.Patterns {
				regexes = append(regexes, globalRegexCache.Get("(?i)"+p))
			}
			matchLine = func(line string, hit func(patternIdx int)) {
				for i, r := range regexes {
					if r.MatchString(line) {
						hit(i)
					}
				}
			}
		} else {
			ac := globalACCache.Get(rule.Patterns)
			matchLine = func(line string, hit func(patternIdx int)) {
				ac.Match(strings.ToLower(line), hit)
			}
		}

		applyRuleScan(dirPath, files, report, rule, sev, matchLine, matchAll)
	}
}

// applyRuleScan streams the rule's target files once and evaluates the
// matcher callback per line, honoring any/all semantics. The first matching
// line is kept as evidence; evaluation stops as soon as the rule is
// satisfied (any mode) to keep the cost minimal.
func applyRuleScan(
	dirPath string,
	files []string,
	report *ReportData,
	rule AnalysisRule,
	sev Severity,
	matchLine func(line string, hit func(patternIdx int)),
	matchAll bool,
) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultFileScanTimeout)
	defer cancel()

	satisfied := false
	var evidenceLine string

	scanFilesWithContext(ctx, dirPath, files, func(line string) {
		if satisfied && !matchAll {
			return
		}
		hits := make([]bool, len(rule.Patterns))
		hitCount := 0
		matchLine(line, func(idx int) {
			if !hits[idx] {
				hits[idx] = true
				hitCount++
			}
		})
		ok := hitCount > 0
		if matchAll {
			ok = hitCount == len(rule.Patterns)
		}
		if ok {
			satisfied = true
			if evidenceLine == "" {
				evidenceLine = strings.TrimSpace(line)
			}
		}
	})

	if !satisfied {
		return
	}
	msg := fmt.Sprintf("[RULE: %s] %s", rule.Name, rule.Message)
	addConfigFindingEx(report, sev, rule.Category, msg, strings.Join(files, ","), "rule:"+rule.Name, 0, evidenceLine, "rule:"+rule.Name, ConfidenceHeuristic, "")
}
