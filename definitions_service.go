// definitions_service.go
//
// The definitions service is the single implementation behind the Definitions
// Studio (GUI) and the definitions CLI flags (-definitions-info,
// -validate-rules, -rules-test).
//
// Why it exists: rules and KB articles have always been loaded fail-safe (bad
// input is skipped) but silently, from a set of search paths a user could not
// discover. Everything in this file answers the three questions a support
// engineer actually asks — WHERE does the inspector look, IS my file valid,
// DID my rule fire — and it answers them through the same code the analysis
// runs, so the answers cannot drift from reality:
//
//   - ListDefinitions  → "where does it look, what did it find there?"
//   - ValidateRuleYAML → parseRulesDocument, the loader's own validator
//   - TestRulesAgainst → applyAnalysisRules, the loader's own matcher
//   - SaveRuleYAML     → refuses invalid YAML, atomic, keeps a .bak
//   - CatalogInfo      → read-only view of the embedded option catalog
//
// The GUI bindings live in app_definitions.go (Wails) and the CLI front-end in
// definitions_cli.go; both are thin wrappers so the two front-ends can never
// disagree about what a definition file does.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DefinitionScope identifies which search location a definition file came
// from. For the KB corpus the order below is also the override precedence
// (a later scope wins per article key).
type DefinitionScope string

// Definition scopes, from lowest to highest KB override precedence.
const (
	// ScopeExeDir is the directory holding the running executable.
	ScopeExeDir DefinitionScope = "exe-dir"
	// ScopeWorkDir is the process working directory (the historical default).
	ScopeWorkDir DefinitionScope = "working-dir"
	// ScopeSystem is the system-wide root (/etc/sssd-inspector).
	ScopeSystem DefinitionScope = "system"
	// ScopeUser is the per-user root ($HOME/.sssd-inspector).
	ScopeUser DefinitionScope = "user"
	// ScopeEmbedded is content compiled into the binary (KB corpus, catalog).
	ScopeEmbedded DefinitionScope = "embedded"
	// ScopeEditor is content that has not been saved yet (Studio editor).
	ScopeEditor DefinitionScope = "editor"
)

// Label renders a scope for humans (CLI output and GUI badges).
func (s DefinitionScope) Label() string {
	switch s {
	case ScopeExeDir:
		return "next to the executable"
	case ScopeWorkDir:
		return "working directory"
	case ScopeSystem:
		return "system-wide"
	case ScopeUser:
		return "user"
	case ScopeEmbedded:
		return "embedded in the binary"
	case ScopeEditor:
		return "unsaved editor content"
	default:
		return string(s)
	}
}

// DefinitionKind distinguishes the two definition families the inspector
// loads from the filesystem.
type DefinitionKind string

// Definition kinds.
const (
	// KindRules is a rules YAML file/directory.
	KindRules DefinitionKind = "rules"
	// KindKBArticles is a knowledge-base article directory.
	KindKBArticles DefinitionKind = "kb-articles"
	// KindCatalog is an option-catalog override file.
	KindCatalog DefinitionKind = "catalog"
)

// DefinitionFileInfo describes one discovery location for one definition kind:
// whether anything is there, whether the user may write there, and what the
// inspector found in it. Read-only locations (embedded) report Writable=false.
type DefinitionFileInfo struct {
	Path         string          `json:"path"`
	Kind         DefinitionKind  `json:"kind"`
	Scope        DefinitionScope `json:"scope"`
	Exists       bool            `json:"exists"`
	IsDir        bool            `json:"is_dir"`
	Writable     bool            `json:"writable"`
	RuleCount    int             `json:"rule_count"`
	ArticleCount int             `json:"article_count"`
	// OptionCount is the number of sssd.conf options in a catalog override
	// (the only kind that carries options).
	OptionCount int          `json:"option_count,omitempty"`
	SizeBytes   int64        `json:"size_bytes,omitempty"`
	ModTime     string       `json:"mod_time,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
}

// DefinitionsInventory answers "where does the inspector look for my
// definitions, and what did it find there?". Files lists every candidate
// location in discovery order; Rules carries the loaded rules with their
// provenance; Diagnostics is the same skipped-input report the analysis
// produces, deduplicated.
// nonNilSlice returns s, or an empty slice when s is nil.
//
// encoding/json renders a nil slice as JSON null, which contradicts the
// generated front-end models: they declare a plain array (rules: RuleInfo[]), so
// a null is a lie the UI has to defend against at every read. The definitions
// service is the boundary that owns these payloads, so it is where the promise
// is kept: slice fields tagged without omitempty always leave here non-nil.
// Fields tagged omitempty are simply absent and need no help.
func nonNilSlice[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

type DefinitionsInventory struct {
	UserRoot     string               `json:"user_root"`
	SystemRoot   string               `json:"system_root"`
	Files        []DefinitionFileInfo `json:"files"`
	Rules        []RuleInfo           `json:"rules"`
	RuleCount    int                  `json:"rule_count"`
	ArticleCount int                  `json:"article_count"`
	Diagnostics  []Diagnostic         `json:"diagnostics,omitempty"`
}

// CatalogInfo summarises the embedded SSSD option catalog. It is read-only on
// purpose: regenerating the catalog is a maintainer operation (-gen-catalog)
// that reads upstream man pages and writes into the repository.
type CatalogInfo struct {
	Source       string   `json:"source"`
	Version      string   `json:"version"`
	Generated    string   `json:"generated"`
	Sources      []string `json:"sources,omitempty"`
	OptionCount  int      `json:"option_count"`
	SectionCount int      `json:"section_count"`
	Available    bool     `json:"available"`
	Error        string   `json:"error,omitempty"`

	// Effective names the catalog actually in use: "embedded" or the override
	// file path. Without it the Studio would describe the embedded release
	// while the analysis validated against an override.
	Effective string `json:"effective"`
	// UsingOverride is Effective != "embedded".
	UsingOverride bool `json:"using_override"`
	// OverridePaths are the locations searched for an override, in precedence
	// order, so a user can be told where to drop one.
	OverridePaths []string `json:"override_paths,omitempty"`
	// Diagnostics carries overrides that were skipped (unreadable, malformed,
	// empty). A skipped override is reported, never silent.
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
}

// RuleValidationResult is the verdict on one rules document. Valid is false
// when the document would be skipped or partially skipped by the loader; the
// Diagnostics carry the reason and the line.
type RuleValidationResult struct {
	Label       string       `json:"label"`
	Valid       bool         `json:"valid"`
	RuleCount   int          `json:"rule_count"`
	Rules       []RuleInfo   `json:"rules,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
}

// DefinitionSaveResult reports the outcome of SaveRuleYAML. It is returned
// even when the save was refused, so the caller can render Validation inline
// instead of having to parse an error string.
type DefinitionSaveResult struct {
	Path       string               `json:"path"`
	Scope      DefinitionScope      `json:"scope"`
	Saved      bool                 `json:"saved"`
	Backup     string               `json:"backup,omitempty"`
	Bytes      int                  `json:"bytes"`
	Validation RuleValidationResult `json:"validation"`
}

// RuleTestOutcome is one rule's dry-run verdict against a target
// supportconfig.
type RuleTestOutcome struct {
	Rule     AnalysisRule    `json:"rule"`
	File     string          `json:"file"`
	Scope    DefinitionScope `json:"scope"`
	Line     int             `json:"line,omitempty"`
	Matched  bool            `json:"matched"`
	Evidence string          `json:"evidence,omitempty"`
	Message  string          `json:"message,omitempty"`
}

// RuleTestResult is the dry-run report: which of the loaded rules would fire
// against a given supportconfig, and the evidence line that made it fire.
type RuleTestResult struct {
	TargetPath  string            `json:"target_path"`
	Total       int               `json:"total"`
	Matched     int               `json:"matched"`
	Outcomes    []RuleTestOutcome `json:"outcomes"`
	Diagnostics []Diagnostic      `json:"diagnostics,omitempty"`
}

// sameOrUnder reports whether path is root itself or nested inside it.
// Empty roots never match.
func sameOrUnder(path, root string) bool {
	if root == "" {
		return false
	}
	if path == root {
		return true
	}
	return strings.HasPrefix(path, strings.TrimSuffix(root, string(os.PathSeparator))+string(os.PathSeparator))
}

// definitionScopeOf classifies a definition path by the discovery location it
// belongs to. The dedicated definition roots are checked first: when the
// working directory IS the home directory, "$HOME/.sssd-inspector/rules.yaml"
// must report as the user scope, not as a lucky working-directory hit.
func definitionScopeOf(path string) DefinitionScope {
	abs := absPath(path)
	if sameOrUnder(abs, systemDefinitionsRoot()) {
		return ScopeSystem
	}
	if sameOrUnder(abs, userDefinitionsRoot()) {
		return ScopeUser
	}
	if exe, err := os.Executable(); err == nil && sameOrUnder(abs, filepath.Dir(exe)) {
		return ScopeExeDir
	}
	return ScopeWorkDir
}

// formatModTime renders a file timestamp for the inventory, empty when zero.
func formatModTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// scopeRoot resolves a user-facing scope name ("user"/"system") to the
// definitions root that backs it.
func scopeRoot(scope string) (string, DefinitionScope, error) {
	switch DefinitionScope(strings.ToLower(strings.TrimSpace(scope))) {
	case ScopeUser:
		root := userDefinitionsRoot()
		if root == "" {
			return "", ScopeUser, fmt.Errorf("cannot determine the per-user definitions directory (no HOME)")
		}
		return root, ScopeUser, nil
	case ScopeSystem:
		root := systemDefinitionsRoot()
		if root == "" {
			return "", ScopeSystem, fmt.Errorf("system definitions directory is not configured")
		}
		return root, ScopeSystem, nil
	default:
		return "", "", fmt.Errorf("unknown definitions scope %q (expected \"user\" or \"system\")", scope)
	}
}

// ListDefinitions builds the full discovery inventory: every location the
// loader inspects for rules and KB articles, whether anything is there,
// whether the current user could write there, and what was found. The
// per-location counts come from the same parsers the analysis uses, and the
// authoritative totals from the real aggregate loaders (so overrides are
// already applied). Apart from the writability probe it is read-only.
func ListDefinitions() DefinitionsInventory {
	inv := DefinitionsInventory{
		UserRoot:   userDefinitionsRoot(),
		SystemRoot: systemDefinitionsRoot(),
	}

	var perLocation []Diagnostic
	for _, p := range ruleCandidatePaths() {
		fi := describeDefinitionPath(p, KindRules)
		inv.Files = append(inv.Files, fi)
		perLocation = append(perLocation, fi.Diagnostics...)
	}
	for _, d := range externalKBDirs() {
		fi := describeDefinitionPath(d, KindKBArticles)
		inv.Files = append(inv.Files, fi)
		perLocation = append(perLocation, fi.Diagnostics...)
	}
	// The catalog override locations belong here too: "where do I drop a
	// catalog for a newer SSSD release?" is the same question the other rows
	// answer, and a broken override is a report the user must see.
	for _, p := range catalogOverrideCandidates() {
		fi := describeOverrideCatalog(p)
		inv.Files = append(inv.Files, fi)
		perLocation = append(perLocation, fi.Diagnostics...)
	}
	embedded := embeddedDefinitionInfo()
	inv.Files = append(inv.Files, embedded)

	// The authoritative numbers: exactly what the analysis will load, with
	// KB override precedence already applied.
	infos, ruleDiags := loadAnalysisRuleInfos()
	articles, kbDiags := loadKBArticlesDiag()
	inv.Rules = nonNilSlice(infos)
	inv.RuleCount = len(infos)
	inv.ArticleCount = len(articles)

	// One deduplicated diagnostic list for the whole inventory: the same file
	// is reported by the per-location scan and by the aggregate loaders, and
	// the user must see it once.
	var report ReportData
	report.AddDiagnostics(perLocation)
	report.AddDiagnostics(embedded.Diagnostics)
	report.AddDiagnostics(ruleDiags)
	report.AddDiagnostics(kbDiags)
	inv.Diagnostics = report.Diagnostics
	return inv
}

// describeDefinitionPath inspects one candidate location and reports what the
// inspector would find there.
func describeDefinitionPath(path string, kind DefinitionKind) DefinitionFileInfo {
	fi := DefinitionFileInfo{
		Path:     absPath(path),
		Kind:     kind,
		Scope:    definitionScopeOf(path),
		Writable: definitionWritable(path),
	}

	info, err := os.Stat(path)
	if err != nil {
		return fi // candidate that does not exist (yet): Writable says if it could
	}
	fi.Exists = true
	fi.IsDir = info.IsDir()
	if !info.IsDir() {
		fi.SizeBytes = info.Size()
		fi.ModTime = formatModTime(info.ModTime())
	}

	switch kind {
	case KindRules:
		infos, diags := loadRulesAt(path)
		fi.RuleCount = len(infos)
		fi.Diagnostics = diags
	case KindKBArticles:
		if fi.IsDir {
			articles, diags := loadKBArticlesFromDir(path)
			fi.ArticleCount = len(articles)
			fi.Diagnostics = diags
		}
	}
	return fi
}

// definitionWritable reports whether this process can write to path. Go's
// standard library has no access(2) and permission bits lie about non-owners
// (a 0755 /etc/sssd-inspector is not writable by a normal user), so the only
// honest answer is a probe: open an existing file without truncating it, or
// create-and-remove a dot file in the target directory. The probe file is
// removed immediately, and this is called only from the explicit "where do my
// definitions live?" request — never during an analysis.
func definitionWritable(path string) bool {
	info, err := os.Stat(path)
	if err == nil && !info.IsDir() {
		f, err := os.OpenFile(path, os.O_WRONLY, 0)
		if err != nil {
			return false
		}
		f.Close()
		return true
	}
	if err != nil {
		// The path does not exist yet: probe the nearest existing ancestor,
		// which is exactly the directory SaveRuleYAML would create it in.
		dir := filepath.Dir(path)
		for {
			if _, statErr := os.Stat(dir); statErr == nil {
				break
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				return false
			}
			dir = parent
		}
		path = dir
	}

	f, err := os.CreateTemp(path, ".sssd-inspector-probe-*")
	if err != nil {
		return false
	}
	name := f.Name()
	f.Close()
	os.Remove(name)
	return true
}

// embeddedDefinitionInfo describes the definition content compiled into the
// binary (the KB corpus). Its Path uses the "embedded://" scheme: there is no
// file to open, and the front-ends must not try.
func embeddedDefinitionInfo() DefinitionFileInfo {
	articles, diags := loadEmbeddedKBArticles()
	return DefinitionFileInfo{
		Path:         "embedded://kb_articles",
		Kind:         KindKBArticles,
		Scope:        ScopeEmbedded,
		Exists:       true,
		ArticleCount: len(articles),
		Writable:     false,
		Diagnostics:  diags,
	}
}

// ValidateRuleYAML validates rules YAML coming from the Studio editor. It
// touches neither the filesystem nor the analysis: the verdict is in the
// result (Valid + Diagnostics) so the front-end can render it inline.
func ValidateRuleYAML(content string) RuleValidationResult {
	return validateRuleYAMLAs(content, "<editor>")
}

// validateRuleYAMLAs is the shared validation entry point: the Studio passes
// the "<editor>" label, the CLI gate passes the real file path so the
// diagnostics point at a file the user can open.
func validateRuleYAMLAs(content, label string) RuleValidationResult {
	infos, diags := parseRulesDocument([]byte(content), label)
	return RuleValidationResult{
		Label:       label,
		Valid:       len(diags) == 0,
		RuleCount:   len(infos),
		Rules:       infos,
		Diagnostics: diags,
	}
}

// ValidateDiscoveredRules validates every rules file the analysis would load,
// from the very bytes the analysis would read, and returns one result per file
// plus the loader's own diagnostics. The CLI gate (-validate-rules) uses it; an
// empty result means "no definitions installed", which is not an error.
func ValidateDiscoveredRules() ([]RuleValidationResult, []Diagnostic) {
	var results []RuleValidationResult
	var diags []Diagnostic
	for _, c := range ruleCandidatePaths() {
		for _, f := range rulesFilesIn(c) {
			content, err := os.ReadFile(f)
			if err != nil {
				diags = append(diags, Diagnostic{
					File:     absPath(f),
					Message:  fmt.Sprintf("cannot read rules file: %v", err),
					Severity: SevWarning,
				})
				continue
			}
			results = append(results, validateRuleYAMLAs(string(content), absPath(f)))
		}
	}
	return results, diags
}

// ValidateDiscoveredKBArticles loads the KB corpus exactly as the analysis
// does and returns only the diagnostics: the article files are JSON, checked
// structurally by the loader itself. Used by the -validate-rules gate so a
// broken article file is reported by the same command as a broken rule.
func ValidateDiscoveredKBArticles() []Diagnostic {
	_, diags := loadKBArticlesDiag()
	return diags
}

// SaveRuleYAML validates content and writes it to the rules.yaml of the given
// scope ("user" or "system").
//
// Refusing to write an invalid document is deliberate: the analysis skips bad
// rules fail-safe, so saving a broken file would leave the user believing in
// definitions that silently do nothing — the exact failure mode this service
// exists to remove. The verdict travels back in Result.Validation, so the
// caller can show the same diagnostics the editor displays.
//
// The write is atomic (temp file in the same directory, then rename) so an
// interrupted save can never truncate a working rules file, and the previous
// content is preserved as rules.yaml.bak.
func SaveRuleYAML(content, scope string) (DefinitionSaveResult, error) {
	root, resolved, err := scopeRoot(scope)
	if err != nil {
		return DefinitionSaveResult{Validation: validateRuleYAMLAs(content, "<editor>")}, err
	}

	res := DefinitionSaveResult{
		Path:       filepath.Join(root, "rules.yaml"),
		Scope:      resolved,
		Validation: validateRuleYAMLAs(content, filepath.Join(root, "rules.yaml")),
	}
	if !res.Validation.Valid {
		return res, nil // refused: the caller renders res.Validation
	}

	if err := os.MkdirAll(root, 0o755); err != nil {
		return res, fmt.Errorf("cannot create %s: %w%s", root, err, writeHint(root))
	}
	if data, readErr := os.ReadFile(res.Path); readErr == nil {
		backup := res.Path + ".bak"
		if err := os.WriteFile(backup, data, 0o644); err != nil {
			return res, fmt.Errorf("cannot write backup %s: %w%s", backup, err, writeHint(root))
		}
		res.Backup = backup
	}

	if err := writeFileAtomic(res.Path, []byte(content)); err != nil {
		return res, err
	}
	res.Saved = true
	res.Bytes = len(content)
	return res, nil
}

// RuleDocument is the current content of one rules document, as the editor
// loads it. Exists is false when the scope has no rules.yaml yet: the Studio
// then starts from a template instead of showing an error.
type RuleDocument struct {
	Path    string          `json:"path"`
	Scope   DefinitionScope `json:"scope"`
	Exists  bool            `json:"exists"`
	Bytes   int             `json:"bytes"`
	Content string          `json:"content"`
}

// ReadRuleYAML returns the rules.yaml of the given scope ("user" or "system")
// so the Studio opens what is actually in effect instead of an empty editor.
//
// A missing file is not an error — that is the normal state of a fresh
// installation, and the caller needs a document to start editing from. Only a
// real read failure (permissions, I/O) is reported, and it names the path.
func ReadRuleYAML(scope string) (RuleDocument, error) {
	root, resolved, err := scopeRoot(scope)
	if err != nil {
		return RuleDocument{}, err
	}
	doc := RuleDocument{Path: filepath.Join(root, "rules.yaml"), Scope: resolved}
	data, readErr := os.ReadFile(doc.Path)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return doc, nil
		}
		return doc, fmt.Errorf("cannot read %s: %w", doc.Path, readErr)
	}
	doc.Exists = true
	doc.Bytes = len(data)
	doc.Content = string(data)
	return doc, nil
}

// InstallCatalog copies a generated catalog.json into a definitions root so it
// becomes the catalog the analysis validates sssd.conf against.
//
// It is the same write as any other definition file: the document is decoded
// and accepted only if it is usable (decodeCatalog, the same validation the
// loader applies), the write is atomic and the previous content is kept as
// catalog.json.bak. A file that would be skipped on load is refused here too —
// installing an override that silently does nothing is the failure mode this
// service exists to prevent — and because the cache is keyed by content, the
// next analysis picks the new catalog up without a restart.
func InstallCatalog(path, scope string) (DefinitionSaveResult, error) {
	clean := strings.TrimSpace(path)
	if clean == "" {
		return DefinitionSaveResult{}, fmt.Errorf("no catalog file selected")
	}
	raw, err := os.ReadFile(clean)
	if err != nil {
		return DefinitionSaveResult{}, fmt.Errorf("cannot read %s: %w", clean, err)
	}
	if _, err := decodeCatalog(raw, clean); err != nil {
		return DefinitionSaveResult{}, fmt.Errorf("%s cannot be used as a catalog override: %w", clean, err)
	}

	root, resolved, err := scopeRoot(scope)
	if err != nil {
		return DefinitionSaveResult{}, err
	}
	res := DefinitionSaveResult{Path: filepath.Join(root, CatalogOverrideFileName), Scope: resolved}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return res, fmt.Errorf("cannot create %s: %w%s", root, err, writeHint(root))
	}
	if previous, readErr := os.ReadFile(res.Path); readErr == nil {
		backup := res.Path + ".bak"
		if err := os.WriteFile(backup, previous, 0o644); err != nil {
			return res, fmt.Errorf("cannot write backup %s: %w%s", backup, err, writeHint(root))
		}
		res.Backup = backup
	}
	if err := writeFileAtomic(res.Path, raw); err != nil {
		return res, err
	}
	res.Saved = true
	res.Bytes = len(raw)
	return res, nil
}

// writeHint appends an actionable hint when a write into the system-wide root
// failed: it is the one failure a normal user cannot diagnose from "permission
// denied" alone.
func writeHint(root string) string {
	if root == systemDefinitionsRoot() && os.Geteuid() != 0 {
		return "; system-wide definitions need root (use sudo, or save to the user scope instead)"
	}
	return ""
}

// TestRulesAgainst dry-runs the rules the analysis would load against a
// supportconfig directory or archive and reports, per rule, whether it fired
// and on which evidence line. It calls applyAnalysisRules — the very matcher
// the real analysis runs — on a throwaway ReportData, so a rule that fires
// here fires there, and nothing is written anywhere.
//
// Reminder of the semantics the dry-run inherits: with match: all a rule fires
// only when both patterns are present in the scanned files, otherwise the
// first matching line wins; pattern_type: regex compiles through RE2 with
// case-insensitive matching.
func TestRulesAgainst(targetPath string) (RuleTestResult, error) {
	dir, cleanup, err := resolveDefinitionTarget(targetPath)
	if err != nil {
		return RuleTestResult{}, err
	}
	defer cleanup()

	infos, diags := loadAnalysisRuleInfos()
	res := RuleTestResult{
		TargetPath:  targetPath,
		Total:       len(infos),
		Outcomes:    []RuleTestOutcome{},
		Diagnostics: diags,
	}
	if len(infos) == 0 {
		return res, nil
	}

	rules := make([]AnalysisRule, 0, len(infos))
	for _, ri := range infos {
		rules = append(rules, ri.Rule)
	}

	var probe ReportData
	applyAnalysisRules(dir, &probe, rules)

	// Findings are keyed by RuleID ("rule:<name>"). Two rules sharing a name
	// collapse onto one key, which is exactly why parseRulesDocument reports a
	// duplicate-name diagnostic: the ambiguity is surfaced, never silent.
	findings := make(map[string]ConfigFinding, len(probe.ConfigFindings))
	for _, f := range probe.ConfigFindings {
		if strings.HasPrefix(f.RuleID, "rule:") {
			findings[f.RuleID] = f
		}
	}

	res.Outcomes = make([]RuleTestOutcome, 0, len(infos))
	for _, ri := range infos {
		outcome := RuleTestOutcome{Rule: ri.Rule, File: ri.File, Scope: ri.Scope, Line: ri.Line}
		if f, ok := findings["rule:"+ri.Rule.Name]; ok {
			outcome.Matched = true
			outcome.Evidence = f.Evidence
			outcome.Message = f.Message
			res.Matched++
		}
		res.Outcomes = append(res.Outcomes, outcome)
	}
	return res, nil
}

// resolveDefinitionTarget accepts a supportconfig directory or archive and
// returns a directory to scan plus a cleanup function (a no-op for a directory
// the caller already owns).
func resolveDefinitionTarget(path string) (string, func(), error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", func() {}, fmt.Errorf("cannot access %q: %w", path, err)
	}
	if info.IsDir() {
		return path, func() {}, nil
	}
	dir, err := extractArchiveToTemp(path, func(string, int) {})
	if err != nil {
		return "", func() {}, fmt.Errorf("cannot extract %q: %w", path, err)
	}
	return dir, func() { os.RemoveAll(dir) }, nil
}

// GetCatalogInfo reports the embedded SSSD option catalog: which upstream
// release it was generated from, when, and how many options and sections it
// carries. The Studio shows it so a finding can be traced to a catalog
// version. Read-only by design — regenerating the catalog is a maintainer
// operation (-gen-catalog) that reads upstream man pages and writes into the
// repository, so it is deliberately not offered as a user-facing action.
func GetCatalogInfo() CatalogInfo {
	// The same resolution the analysis uses, so the Studio answers "which
	// release am I validated against?" with the answer that was applied, not
	// with the answer the embedded copy implies.
	cat, res, diags := loadOptionCatalog()
	info := CatalogInfo{
		Effective:     res.effective(),
		UsingOverride: !res.Embedded && res.Path != "",
		OverridePaths: catalogOverrideCandidates(),
		Diagnostics:   diags,
	}
	if cat == nil {
		info.Available = false
		if len(diags) == 0 {
			info.Error = "no option catalog available (the embedded copy is broken)"
		}
		return info
	}
	info.Source = cat.Source
	info.Version = cat.Version
	info.Generated = cat.Generated
	info.Sources = cat.Sources
	info.OptionCount = len(cat.Options)
	info.SectionCount = len(cat.Sections)
	info.Available = true
	return info
}

// --- Text rendering (CLI front-end) ---------------------------------------

// diagnosticLocation renders the "file:line" form of a diagnostic, shared by
// the CLI renderer and the stderr banner so the two can never disagree.
func diagnosticLocation(d Diagnostic) string {
	if d.Line > 0 {
		return fmt.Sprintf("%s:%d", d.File, d.Line)
	}
	return d.File
}

// RenderDiagnostics renders diagnostics as one line per problem, and returns
// "" when there is nothing to report.
func RenderDiagnostics(diags []Diagnostic) string {
	if len(diags) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Problems found while loading definitions (%d):\n", len(diags)))
	for _, d := range diags {
		sb.WriteString(fmt.Sprintf("  [!] %s: %s\n", diagnosticLocation(d), d.Message))
	}
	return sb.String()
}

// indentLines prefixes every line of text, used to nest diagnostics under the
// file they belong to.
func indentLines(text, prefix string) string {
	if text == "" {
		return ""
	}
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n") + "\n"
}

// definitionStatus summarises what one location holds.
func definitionStatus(f DefinitionFileInfo) string {
	switch {
	case !f.Exists:
		return "not present"
	case f.Kind == KindRules:
		return fmt.Sprintf("%d rule(s)", f.RuleCount)
	case f.Kind == KindCatalog:
		if len(f.Diagnostics) > 0 {
			return "unusable"
		}
		return fmt.Sprintf("%d option(s)", f.OptionCount)
	default:
		return fmt.Sprintf("%d article(s)", f.ArticleCount)
	}
}

// writableLabel renders the writability probe result.
func writableLabel(writable bool) string {
	if writable {
		return "writable"
	}
	return "read-only"
}

// orNone renders an empty path as a placeholder instead of an empty column.
func orNone(s string) string {
	if s == "" {
		return "(unavailable)"
	}
	return s
}

// RenderDefinitionsInventory renders the inventory as plain text
// (-definitions-info). It is deliberately explicit about every path and its
// writability: making the search paths discoverable is the entire point.
func RenderDefinitionsInventory(inv DefinitionsInventory) string {
	var sb strings.Builder
	sb.WriteString("Definition search paths (discovery order; for KB articles a later\n")
	sb.WriteString("directory overrides an earlier one):\n")
	sb.WriteString(fmt.Sprintf("  user scope:   %s\n", orNone(inv.UserRoot)))
	sb.WriteString(fmt.Sprintf("  system scope: %s\n", orNone(inv.SystemRoot)))
	sb.WriteString("\n")
	for _, f := range inv.Files {
		// The path comes last so a long path cannot push the status columns
		// around: the left columns stay aligned on every machine.
		sb.WriteString(fmt.Sprintf("  [%-9s] %-11s %s (%s, %s)\n",
			f.Scope, f.Kind, f.Path, definitionStatus(f), writableLabel(f.Writable)))
	}
	sb.WriteString(fmt.Sprintf("\nLoaded: %d rule(s), %d KB article(s)\n", inv.RuleCount, inv.ArticleCount))
	// Name the catalog in effect: it decides what "valid option" means for
	// every configuration finding in the report.
	if cat, res, _ := loadOptionCatalog(); cat != nil {
		version := cat.Version
		if version == "" {
			version = "unknown"
		}
		origin := "embedded in the binary"
		if !res.Embedded {
			origin = "override"
		}
		sb.WriteString(fmt.Sprintf("Option catalog: SSSD %s, %d option(s) (%s)\n", version, len(cat.Options), origin))
	}
	if len(inv.Rules) > 0 {
		sb.WriteString("\nRules loaded:\n")
		for _, ri := range inv.Rules {
			sb.WriteString(fmt.Sprintf("  %-28s %-8s %s:%d\n", ri.Rule.Name, ri.Rule.Severity, ri.File, ri.Line))
		}
	}
	if diagText := RenderDiagnostics(inv.Diagnostics); diagText != "" {
		sb.WriteString("\n" + diagText)
	}
	return sb.String()
}

// RenderRuleValidation renders the -validate-rules gate. ok is false when the
// process should exit non-zero: a file that loads fewer rules than it declares
// means the user believes in definitions the analysis will never use.
func RenderRuleValidation(results []RuleValidationResult, kbDiags []Diagnostic) (string, bool) {
	var sb strings.Builder
	ok := true
	ruleCount := 0

	if len(results) == 0 {
		sb.WriteString("No rules files found: the inspector runs with its built-in detectors only.\n")
	}
	for _, r := range results {
		ruleCount += r.RuleCount
		status := "ok"
		if !r.Valid {
			status = "PROBLEMS"
			ok = false
		}
		sb.WriteString(fmt.Sprintf("[%s] %s: %d rule(s) loaded\n", status, r.Label, r.RuleCount))
		if !r.Valid {
			sb.WriteString(indentLines(RenderDiagnostics(r.Diagnostics), "    "))
		}
	}
	if len(results) > 0 {
		sb.WriteString(fmt.Sprintf("%d file(s), %d rule(s) loaded.\n", len(results), ruleCount))
	}
	if len(kbDiags) > 0 {
		ok = false
		sb.WriteString("\n" + RenderDiagnostics(kbDiags))
	}
	return sb.String(), ok
}

// RenderRuleTest renders the -rules-test dry-run: one line per rule, with the
// evidence that made it fire (or an explicit no-match).
func RenderRuleTest(res RuleTestResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Rule dry-run against %s\n", res.TargetPath))
	sb.WriteString(fmt.Sprintf("%d rule(s) loaded, %d would fire.\n\n", res.Total, res.Matched))
	for _, o := range res.Outcomes {
		mark := "no "
		if o.Matched {
			mark = "YES"
		}
		sb.WriteString(fmt.Sprintf("  [%s] %-28s %-8s %s:%d (%s)\n",
			mark, o.Rule.Name, o.Rule.Severity, o.File, o.Line, o.Scope))
		if o.Matched {
			sb.WriteString(fmt.Sprintf("        %s\n", o.Message))
			if o.Evidence != "" {
				sb.WriteString(fmt.Sprintf("        evidence: %s\n", o.Evidence))
			}
		}
	}
	if diagText := RenderDiagnostics(res.Diagnostics); diagText != "" {
		sb.WriteString("\n" + diagText)
	}
	return sb.String()
}

func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".rules-*.tmp")
	if err != nil {
		return fmt.Errorf("cannot create a temporary file in %s: %w%s", dir, err, writeHint(dir))
	}
	tmpName := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tmpName) // no-op once the rename succeeded
	}()

	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("cannot write %s: %w", path, err)
	}
	if err := tmp.Chmod(0o644); err != nil {
		return fmt.Errorf("cannot set permissions on %s: %w", path, err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("cannot flush %s: %w", path, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("cannot close %s: %w", path, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("cannot replace %s: %w%s", path, err, writeHint(dir))
	}
	return nil
}
