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
    Evidence       []string `json:"-"` // Not exported to JSON
}
```

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
