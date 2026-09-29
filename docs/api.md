# API Documentation

## Overview

SSSD Inspector exposes a set of methods through the Wails framework for communication between the frontend (JavaScript) and backend (Go). This document provides comprehensive documentation for all available APIs.

## Backend API Methods

### App Structure

The main application logic is encapsulated in the `App` struct, which provides the following methods:

---

### Analyze

**Signature**: `Analyze(targetPath string, anonymize bool) (ReportData, error)`

**Description**: Analyzes a supportconfig archive or directory and generates a comprehensive diagnostic report.

**Parameters**:
- `targetPath` (string): Path to the supportconfig archive (.txz, .tar.xz) or directory
- `anonymize` (bool): Whether to redact PII (Personally Identifiable Information) from the report

**Returns**:
- `ReportData`: Comprehensive analysis results
- `error`: Error if analysis fails

**Example Usage**:
```javascript
try {
    const result = await Analyze("/path/to/supportconfig.txz", true);
    console.log("Analysis completed:", result);
} catch (error) {
    console.error("Analysis failed:", error);
}
```

**Progress Events**: Emits `analyze-progress` events during processing:
```javascript
EventsOn("analyze-progress", (message, percentage) => {
    console.log(`Progress: ${percentage}% - ${message}`);
});
```

---

### OpenFileBrowser

**Signature**: `OpenFileBrowser() (string, error)`

**Description**: Opens a native file browser dialog for selecting supportconfig archives.

**Returns**:
- `string`: Selected file path, or empty string if cancelled
- `error`: Error if dialog fails to open

**Example Usage**:
```javascript
try {
    const filePath = await OpenFileBrowser();
    if (filePath) {
        console.log("Selected file:", filePath);
    }
} catch (error) {
    console.error("File browser error:", error);
}
```

**File Filters**: 
- Supportconfig Archives (*.txz, *.tar.xz)
- All Files (*.*)

---

### SavePDF

**Signature**: `SavePDF(b64 string) (string, error)`

**Description**: Saves a base64-encoded PDF report to a user-selected location.

**Parameters**:
- `b64` (string): Base64-encoded PDF data (with data URL prefix)

**Returns**:
- `string`: Path where file was saved, or "cancelled" if user cancelled
- `error`: Error if save operation fails

**Example Usage**:
```javascript
try {
    const pdfData = "data:application/pdf;base64,JVBERi0xLjQK...";
    const savedPath = await SavePDF(pdfData);
    if (savedPath !== "cancelled") {
        console.log("PDF saved to:", savedPath);
    }
} catch (error) {
    console.error("PDF save error:", error);
}
```

---

### SaveTXT

**Signature**: `SaveTXT(report ReportData) (string, error)`

**Description**: Saves a text report based on the provided ReportData structure.

**Parameters**:
- `report` (ReportData): Analysis results to convert to text format

**Returns**:
- `string`: Path where file was saved, or "cancelled" if user cancelled
- `error`: Error if save operation fails

**Example Usage**:
```javascript
try {
    const savedPath = await SaveTXT(reportData);
    if (savedPath !== "cancelled") {
        console.log("Text report saved to:", savedPath);
    }
} catch (error) {
    console.error("Text save error:", error);
}
```

---

## Definitions Studio

The seven methods behind the second GUI view. Every one of them is a thin
wrapper over `definitions_service.go`, the same implementation the
`-definitions-info`, `-validate-rules` and `-rules-test` flags use — so the
Studio and the CLI cannot disagree about what a definition file does. The
frontend reaches them through `frontend/src/api/backend.ts`, the only importer
of the generated bindings.

### ListDefinitions

**Signature**: `ListDefinitions() DefinitionsInventory`

**Description**: The full discovery inventory: every location the analysis
inspects for rules, KB articles and a catalog override, what was found there,
whether the current user may write there, plus the loaded rules with their
provenance and the diagnostics for skipped input. Apart from the writability
probe (a real create-and-remove, not permission bits) it is read-only.

```javascript
const inv = await ListDefinitions();
console.log(inv.rule_count, inv.article_count);
inv.files.filter(f => f.kind === "catalog" && f.exists && !f.writable)
   .forEach(f => console.warn(`${f.path} is not writable for you`));
```

### ValidateRuleYAML

**Signature**: `ValidateRuleYAML(content string) RuleValidationResult`

**Description**: Validates a rules document **without saving it**, through
`parseRulesDocument` — the loader's own validator, so the editor cannot
disagree with the analysis. A malformed document is not an error: the verdict
travels in `Valid`/`Diagnostics` and the editor renders it inline.

```javascript
const verdict = await ValidateRuleYAML(editorText);
if (!verdict.valid) {
    verdict.diagnostics.forEach(d => console.warn(`${d.file}:${d.line} ${d.message}`));
}
```

### ReadRuleYAML

**Signature**: `ReadRuleYAML(scope string) (RuleDocument, error)`

**Description**: The `rules.yaml` of a scope (`"user"` or `"system"`), so the
editor opens what is actually in effect instead of an empty buffer. A scope
without one is the normal state of a fresh installation, not an error:
`Exists=false`, `Content=""`.

```go
type RuleDocument struct {
    Path    string `json:"path"`
    Scope   string `json:"scope"`
    Exists  bool   `json:"exists"`
    Bytes   int    `json:"bytes"`
    Content string `json:"content"`
}
```

### SaveRuleYAML

**Signature**: `SaveRuleYAML(content string, scope string) (DefinitionSaveResult, error)`

**Description**: Validates and writes a rules document into the user or system
scope. **Invalid documents are refused** (`Saved=false`, verdict in
`Validation`): the analysis skips bad rules fail-safe, so saving them would
leave the user believing in definitions that silently do nothing. The write is
atomic (temp file + rename) and the previous content is kept as
`rules.yaml.bak`. A system-scope failure that needs root returns an actionable
hint.

### TestRulesAgainst

**Signature**: `TestRulesAgainst(targetPath string) (RuleTestResult, error)`

**Description**: Dry-runs the rules the analysis would load against a
supportconfig directory or archive, through `applyAnalysisRules` — the real
matcher, on a throwaway report. Nothing is written and no report is produced.
With `match: all` a rule fires only when every pattern is present in the scanned
files, otherwise the first matching line wins; `pattern_type: regex` compiles
through RE2 with case-insensitive matching.

```javascript
const dry = await TestRulesAgainst("/tmp/supportconfig.txz");
dry.outcomes.filter(o => o.matched)
   .forEach(o => console.log(`${o.rule.name} fired on: ${o.evidence}`));
```

### GetCatalogInfo

**Signature**: `GetCatalogInfo() CatalogInfo`

**Description**: The option catalog in effect. Since M3 a `catalog.json` in a
definitions root overrides the embedded copy (per-user before per-system), so
this reports the resolution, not just the built-in one.

```go
type CatalogInfo struct {
    Source         string   `json:"source"`
    Version        string   `json:"version"`
    Generated      string   `json:"generated"`
    Sources        []string `json:"sources,omitempty"`
    OptionCount    int      `json:"option_count"`
    SectionCount   int      `json:"section_count"`
    Available      bool     `json:"available"`
    Error          string   `json:"error,omitempty"`
    Effective      string   `json:"effective"`        // "embedded" or the override path
    UsingOverride  bool     `json:"using_override"`
    OverridePaths  []string `json:"override_paths,omitempty"`
    Diagnostics    []Diagnostic `json:"diagnostics,omitempty"`
}
```

`Diagnostics` carries overrides that were skipped (unreadable, malformed, or an
empty options map). A skipped override never changes what the analysis does: it
falls back to the embedded catalog, and the report's `CatalogProvenance` states
the release the claims are based on, with the override path appended when one
is in effect.

### InstallCatalog

**Signature**: `InstallCatalog(path string, scope string) (DefinitionSaveResult, error)`
plus `OpenCatalogFile() (string, error)` (the OS chooser).

**Description**: Installs a generated `catalog.json` as the override for a
scope, so replacing the catalog does not mean copying a file by hand. The
document is decoded with `decodeCatalog` — the loader's own validation — so a
file that would be skipped on load is **refused** and nothing is written; the
write is atomic and the previous catalog is kept as `catalog.json.bak`. The
override cache is keyed by content, so the next analysis picks it up without a
restart.

### OpenDefinitionsRoot

**Signature**: `OpenDefinitionsRoot(scope string) error`

**Description**: Creates (if needed) and opens a definitions folder in the
desktop file manager, so a user can drop definition files in without knowing
the path. It is the only method that needs the Wails runtime; without an
attached frontend it returns an error instead of panicking.

---

## Data Structures

### ReportData

The main data structure containing analysis results:

```go
type ReportData struct {
    // Metadata
    AppVersion    string `json:"app_version"`
    Timestamp     string `json:"timestamp"`
    SupportCaseID string `json:"support_case_id"`
    
    // System Information
    KernelVersion string `json:"kernel_version"`
    SLESRlease    string `json:"sles_release"`
    SCCStatus     string `json:"scc_status"`
    
    // Hardware & Virtualization
    HardwareManufacturer string `json:"hardware_manufacturer"`
    HardwareModel        string `json:"hardware_model"`
    Hypervisor           string `json:"hypervisor"`
    VirtualIdentity      string `json:"virtual_identity"`
    MACType              string `json:"mac_type"`
    
    // Services & Packages
    SssdInstalled   bool     `json:"sssd_installed"`
    SssdConfigFound bool     `json:"sssd_config_found"`
    SssdService     string   `json:"sssd_service"`
    WinbindService  string   `json:"winbind_service"`
    NscdStatus      string   `json:"nscd_status"`
    SSSDPackages    []string `json:"sssd_packages"`
    
    // Authentication
    NsswitchValid     bool     `json:"nsswitch_valid"`
    PamSssInstalled   bool     `json:"pam_sss_installed"`
    PamGDPRRestricted bool     `json:"pam_gdpr_restricted"`
    HostsIssues       []string `json:"hosts_issues"`
    HostsFileStatus   string   `json:"hosts_file_status"` // present / missing_on_host / not_collected
    Nameservers       []string `json:"nameservers"`
    SearchDomain      string   `json:"search_domain"`
    TimeService       string   `json:"time_service"`
    KerberosRealm     string   `json:"kerberosRealm"`
    KeytabFound       bool     `json:"keytab_found"`
    
    // SSSD Configuration
    ADProviderMode bool `json:"ad_provider_mode"`
    EnumerateIssue bool `json:"enumerate_issue"`
    UseFQDNSet     bool `json:"use_fqdn_set"`
    
    // Analysis Results
    SSSDLogErrors     []SSSDLogError `json:"sssd_log_errors"`
    SSSDConfigSnippet string         `json:"sssd_config_snippet"`
    MACDenialExamples []string       `json:"mac_denial_examples"`
    Problems          []string       `json:"problems"`
    Warnings          []string       `json:"warnings"`
    MatchedTIDs       []TIDArticle   `json:"matched_tids"`
    Timeline          []TimelineEvent `json:"timeline"`
}
```

### SSSDLogError

Represents a detected SSSD log error:

```go
type SSSDLogError struct {
    Description string   `json:"description"`
    Examples    []string `json:"examples"`
}
```

### TIDArticle

Represents a knowledge base article:

```go
type TIDArticle struct {
    TIDID          string   `json:"tid_id"`
    Title          string   `json:"title"`
    URL            string   `json:"url"`
    Description    string   `json:"description"`
    LogPatterns    []string `json:"log_patterns"`
    ConfigPatterns []string `json:"config_patterns"`
    Evidence       []string `json:"evidence,omitempty"` // log lines that triggered the match
}
```

`Evidence` holds the log lines that made the article match (capped at three by
the matcher). It was `json:"-"` until M3, which is why the "Evidence Found"
block in the GUI could never render. It is an ordinary log excerpt and
`anonymizeReport` scrubs it exactly like `SSSDLogError.Examples` and
`TimelineEvent.RawLog`.

### TimelineEvent

Represents a chronological log event:

```go
type TimelineEvent struct {
    Timestamp string `json:"timestamp"`
    Message   string `json:"message"`
    RawLog    string `json:"raw_log"`
}
```

---

## Error Handling

All API methods follow Go error handling conventions:

1. **Success**: Returns expected result and `nil` error
2. **Failure**: Returns zero/default value and descriptive error

Common error types:
- **File not found**: When target path doesn't exist
- **Invalid format**: When archive format is not supported
- **Permission denied**: When lacking file access permissions
- **Parse error**: When configuration files contain syntax errors
- **Memory error**: When processing large files exceeds limits

### Error Handling Example

```javascript
try {
    const result = await Analyze(path, anonymize);
    // Handle success
} catch (error) {
    if (error.message.includes("not found")) {
        // Handle file not found
    } else if (error.message.includes("permission")) {
        // Handle permission error
    } else {
        // Handle other errors
    }
}
```

---

## Event System

The application uses Wails event system for real-time progress updates:

### Available Events

- `analyze-progress`: Emitted during analysis with progress information

### Event Data Format

```javascript
{
    "message": "Scanning Hardware & OS Data...",
    "percentage": 10
}
```

### Event Handling

```javascript
// Listen to progress events
EventsOn("analyze-progress", (message, percentage) => {
    updateProgressBar(percentage);
    updateStatusMessage(message);
});
```

---

## Performance Considerations

- **Large Files**: Use streaming for files > 100MB
- **Memory Management**: Automatic cleanup of temporary files
- **Concurrent Operations**: Single analysis at a time
- **Progress Updates**: Throttled to prevent UI blocking

---

## Security Considerations

- **File Access**: Limited to user-specified paths
- **PII Protection**: Built-in anonymization available
- **Temporary Files**: Automatically cleaned up
- **Input Validation**: All file paths validated
