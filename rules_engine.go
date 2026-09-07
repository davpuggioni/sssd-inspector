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
// ones. Invalid rules are skipped with a warning (fail-safe).
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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

// loadAnalysisRules discovers and parses rules files. Looked up (in order):
//   - "<exe dir>/rules.yaml", "<exe dir>/rules/*.yaml"
//   - "./rules.yaml", "./rules/*.yaml"
//
// A missing rules directory is NOT an error: rules are optional.
func loadAnalysisRules() []AnalysisRule {
	var candidates []string
	if exePath, err := os.Executable(); err == nil {
		base := filepath.Dir(exePath)
		candidates = append(candidates,
			filepath.Join(base, "rules.yaml"),
			filepath.Join(base, "rules"),
		)
	}
	candidates = append(candidates, "rules.yaml", "rules")

	var rules []AnalysisRule
	for _, c := range candidates {
		info, err := os.Stat(c)
		if err != nil {
			continue
		}
		var files []string
		if info.IsDir() {
			matches, _ := filepath.Glob(filepath.Join(c, "*.yaml"))
			matches2, _ := filepath.Glob(filepath.Join(c, "*.yml"))
			files = append(files, matches...)
			files = append(files, matches2...)
		} else {
			files = []string{c}
		}
		for _, f := range files {
			content, err := os.ReadFile(f)
			if err != nil {
				continue
			}
			var doc analysisRulesFile
			if err := yaml.Unmarshal(content, &doc); err != nil {
				fmt.Printf("Warning: invalid rules file %s: %v\n", f, err)
				continue
			}
			for _, r := range doc.Rules {
				if _, ok := ruleSeverity(r.Severity); !ok {
					fmt.Printf("Warning: rule %q in %s has invalid severity %q, skipped\n", r.Name, f, r.Severity)
					continue
				}
				if len(r.Patterns) == 0 || r.Message == "" {
					fmt.Printf("Warning: rule %q in %s misses patterns or message, skipped\n", r.Name, f)
					continue
				}
				rules = append(rules, r)
			}
		}
	}
	return rules
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
	addConfigFinding(report, sev, rule.Category, msg, strings.Join(files, ","), "rule:"+rule.Name, 0, evidenceLine)
}
