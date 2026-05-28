# SSSD Inspector 🔍

**SSSD Inspector** is a native diagnostic utility designed to parse supportconfig archives generated on SLES or OpenSUSE and identify complex identity management failures. Unlike generic log viewers, this tool uses a specialized pattern-matching engine derived directly from the SSSD C source code to provide human-readable explanations for cryptic error messages.

---

## ✨ Features

### Core Analysis Engine
- **Single-Pass Log Scanning** — Analyzes all log files (sssd.txt, messages) in one pass, extracting errors, warnings, timeline events, and KB article evidence simultaneously — up to **10× faster** than traditional multi-scan approaches
- **350+ Pattern Matching** — Comprehensive database of SSSD error patterns mapped to human-readable descriptions, covering:
  - **Kerberos** — Clock skew, encryption type mismatches, KDC unreachable, FAST tunnel failures
  - **LDAP/AD** — TLS handshake failures, SASL bind errors, USN rollbacks, LDAP size limits
  - **PAM/NSS** — Offline authentication blocks, shell vetoes, negative cache rejections
  - **SSSD Service** — Watchdog terminations, configuration errors, file permission issues, cache corruption
  - **IPA/FreeIPA** — HBAC rule evaluation, SELinux user mapping, cross-forest AD trusts
  - **OAuth2/OIDC** — Identity Provider configuration validation
  - **And many more** — Sudo, SSH, InfoPipe, PAC, proxy providers
- **Knowledge Base Integration** — Dynamically loaded JSON articles (TIDs) correlated with log evidence for precise troubleshooting
- **PII Anonymization** — Built-in redaction of IP addresses, MAC addresses, email addresses, domain names, and Kerberos realms for safe report sharing

### Performance Optimizations
- **Regex Cache** — Thread-safe compilation cache, compiles each regex pattern once per application lifetime
- **File Cache** — LRU-based file content cache reads each file only once per analysis
- **Scanner Pool** — Reusable buffer allocations via `sync.Pool` to reduce GC pressure
- **Context-Aware Scanning** — Timeout-based cancellation prevents hangs on corrupted or NFS-mounted files
- **Parallel Analysis** — Independent analysis phases execute concurrently using worker pools
- **Multi-Core Support** — Parallel file scanning and batch processing for large supportconfig archives

### Output Formats
- **Text Report** — Formatted terminal output with problem severity indicators
- **HTML Report** — Rich, styled HTML with collapsible sections and visual hierarchy
- **PDF Export** — Print-ready PDF generation via browser (GUI mode)

### User Interface
- **Native GUI** — Built with Wails (Go + WebKit), providing a fast, lightweight desktop experience
- **Dark Mode** — Automatic system theme detection with manual toggle
- **Drag & Drop** — Supportconfig files can be dragged directly into the application
- **Progress Tracking** — Real-time progress bar with detailed status messages during analysis
- **Export Controls** — One-click export to TXT and PDF
- **Zoom Controls** — Adjust report text size for comfortable reading
- **Keyboard Shortcuts** — Ctrl+O (open), Ctrl+Enter (analyze), Ctrl+P (export PDF), Ctrl+S (export TXT)

---

## 📋 Requirements

### Runtime
- **Linux** (SLES 12+, OpenSUSE Leap 15+, or any modern Linux distribution)
- Supportconfig archive (.txz or .tar.xz) from SLES or OpenSUSE

### Build Prerequisites
| Dependency | Version | Purpose |
|-----------|---------|---------|
| Go | 1.23+ | Backend compilation |
| Node.js | 18+ | Frontend asset build |
| npm | 9+ | Frontend dependency management |
| Wails CLI | 2.11+ | GUI framework (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`) |

---

## 🚀 Installation

### Setup

```bash
# Install dependencies
go mod tidy
cd frontend && npm install

# Run in development mode
wails dev
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/your-repo/sssd-inspector.git
cd sssd-inspector

# Install dependencies
npm install --prefix frontend

# Build the CLI binary
wails build -tags cli -platform linux/amd64 -ldflags "-w -s" -clean

# Build the GUI binary
wails build -platform linux/amd64 -ldflags "-w -s" -clean

or

# Build the GUI binary using the last webkit2_41 
wails build -platform linux/amd64 -tags webkit2_41 -ldflags "-w -s" -clean

```

---

## 💻 CLI Usage

### Basic Analysis

```bash
# Analyze a supportconfig archive
sssd-inspector /path/to/supportconfig.txz

# Analyze an already-extracted directory
sssd-inspector /path/to/supportconfig-directory/

# Generate both TXT and HTML reports
sssd-inspector /path/to/supportconfig.txz -txt -html
```

### Options

| Flag | Description |
|------|-------------|
| `-v, --version` | Print program version |
| `-analyze <path>` | Path to supportconfig directory or archive |
| `-txt` | Generate a TXT report (default: both formats) |
| `-html` | Generate an HTML report (default: both formats) |
| `-anonymize` | Redact PII (IPs, domains, emails) from the report |

### Examples

```bash
# Full analysis with PII redaction and all report formats
sssd-inspector /tmp/supportconfig-abc123.txz -txt -html -anonymize

# Quick analysis with default TXT output
sssd-inspector /var/log/supportconfig/

# Version check
sssd-inspector -v
```

### Output

The CLI produces:
1. **Console output** — Color-coded summary of all findings directly in the terminal
2. **TXT report** — `supportconfig-abc123_report.txt` (if `-txt` is specified)
3. **HTML report** — `supportconfig-abc123_report.html` (if `-html` is specified)

---

## 🖥️ GUI Usage

### Building and Running

```bash
# Build the GUI
# Linux
wails build -platform linux/amd64

# Windows
wails build -platform windows/amd64

# macOS
wails build -platform darwin/amd64

# Run (or launch the binary)
./sssd-inspector
```

### Workflow

1. **Open the application** — A native window appears with the SSSD Inspector interface
2. **Select a file** — Click "Browse..." or drag a supportconfig.txz file into the input field
3. **Configure options** — Check "Anonymize PII" if you need to redact sensitive information
4. **Start analysis** — Click "Analyze" and watch the progress bar fill in real-time
5. **Review results** — Browse through system info, problems, warnings, error details, timeline, and KB articles
6. **Export** — Click "Export PDF" for print-ready output or "Export TXT" for a plain text file

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+O` | Open file browser |
| `Ctrl+Enter` | Start analysis |
| `Ctrl+P` | Export PDF report |
| `Ctrl+S` | Export TXT report |

---

## 📊 Report Sections

### System Information
- OS release and kernel version
- Hardware manufacturer and model
- Virtualization platform (VMware, KVM, Xen, etc.)
- SCC registration status

### Authentication Services
- SSSD installation status and package versions
- SSSD, Winbind, and NSCD service states
- NSSwitch configuration validity
- PAM module presence

### AD / Kerberos Integration
- DNS nameserver configuration and connectivity
- Time synchronization service status (chronyd/ntpd)
- Kerberos realm detection and keytab verification
- krb5.conf analysis (encryption types, domain realm mapping)

### SSSD Configuration Deep Dive
- AD provider detection
- Enumerate flag warning (performance risk)
- FQDN usage verification
- File permission analysis (sssd.conf ownership and mode)
- /var/lib/sss/ ownership verification
- Configuration conflict detection (duplicate parameters)

### Log Analysis
- **Critical Errors** — Kerberos failures, LDAP connection issues, TLS handshake errors, cache corruption
- **Error Timeline** — Chronological view of all error events with timestamps
- **MAC Security Denials** — AppArmor/SELinux denials related to SSSD

### Knowledge Base Articles
- TID articles matched against actual log evidence
- Direct links to SUSE Technical Information Database
- Relevant configuration suggestions

---

## 🏗️ Architecture

```
sssd-inspector/
├── main_cli.go              # CLI entry point (build tag: cli)
├── main_gui.go              # GUI entry point (Wails framework)
├── app.go                   # Application structure & Wails bindings
├── shared.go                # Shared CLI/GUI logic & configuration
│
├── analyzer_core.go         # Main analysis orchestrator (analyzeData)
├── analyzer_singlepass.go   # Single-pass log scanner (mega-regex engine)
├── analyzer_parallel.go     # Parallel analysis phases
├── analyzer_system.go       # OS, hardware, services analysis
├── analyzer_auth.go         # DNS, Kerberos, PAM, NSS analysis
├── analyzer_logs.go         # SSSD log pattern definitions (350+ patterns)
│
├── utils.go                 # File scanning, section extraction
├── utils_parallel.go        # Parallel file/batch processing
├── cache.go                 # RegexCache, FileCache, Scanner Pool
├── interfaces.go            # Core interfaces for DI
│
├── loaders.go               # Secure archive extraction (.txz)
├── kb.go                    # Knowledge base article matching
├── report.go                # Text & HTML report generation
├── types.go                 # Data structures (ReportData, TIDArticle, etc.)
│
├── config/                  # YAML-based configuration
├── constants/               # Application constants
├── errors/                  # Custom error types
├── logger/                  # Structured logging
│
├── frontend/                # Web UI (Vite + Vanilla JS)
│   ├── src/
│   │   ├── main.js          # Application entry point
│   │   ├── components/      # Reusable UI components
│   │   ├── styles/          # Modular CSS (layout, components, report)
│   │   ├── config/          # Frontend configuration & constants
│   │   └── utils/           # Validation & helper utilities
│   └── wailsjs/             # Generated Wails bindings
│
├── kb_articles/             # Knowledge base articles (JSON)
└── docs/                    # Documentation
```


```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend       │    │   File System  │
│   (JavaScript)  │◄──►│   (Go/Wails)    │◄──►│   Archives      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   Knowledge     │
                       │   Base (TIDs)   │
                       └─────────────────┘
```



### Single-Pass Scanning Engine

SSSD Inspector uses a single-pass log scanning engine. Instead of reading the same log files ~29 times during analysis, the engine:

1. **Combines all patterns** — 350+ error patterns + quick checks + KB article patterns → one mega-regex
2. **Pre-filters lines** — Fast keyword check (`sssd`, `krb5`, `ldap`, `pam`, etc.) skips ~90% of unrelated syslog lines
3. **Single pass** — Each line is matched once against the combined regex, extracting errors, keytab info, watchdog alerts, crypto bugs, evidence, and timestamps simultaneously
4. **Parallel phases** — Independent analysis steps (PAM, NSSwitch, hosts, packages, services) execute concurrently

### Caching System

- **RegexCache** — Thread-safe `sync.RWMutex`-protected map of compiled regexp, compile-once per pattern per application lifetime
- **FileCache** — LRU-like cache (max 20 files) with double-checked locking, reads each file only once per analysis
- **ScannerPool** — `sync.Pool` of 64KB byte slices for scanner buffers, reducing GC pressure

---

## 🧪 Testing

```bash
# Run all tests
cd sssd-inspector
go test -count=1 ./...

# Run with verbose output
go test -count=1 -v ./...

# Run with coverage
go test -cover ./..

# Run benchmarks
go test -bench=. -benchmem -count=1 ./...

# Run specific test
go test -count=1 -run TestAnonymizeReport_DeepPII -v
```

The test suite includes:
- **Unit tests** — Individual analyzer functions with mock supportconfig files
- **Integration tests** — End-to-end analysis pipeline with asset embedding verification
- **Configuration tests** — YAML config loading, validation, and defaults
- **Parallel processing tests** — Worker pool correctness, batch processing, edge cases
- **Benchmark tests** — Performance comparison between sequential and parallel scanning
- **Error handling tests** — Custom error types, wrapping, and context propagation
- **Logger tests** — Structured logging levels, fields, and filtering
- **UI component tests** — Button, ProgressBar, StatusMessage, FileInput lifecycle

---

## ⚙️ Configuration

The application supports YAML-based configuration in multiple locations (checked in order):

1. `config.yaml` in the current directory
2. `~/.sssd-inspector/config.yaml`
3. `/etc/sssd-inspector/config.yaml`
4. Built-in defaults (if no file found)

### Configuration File Example

```yaml
# SSSD Inspector Configuration
app:
  name: "SSSD Inspector"
  version: "0.2.0"

analysis:
  max_file_size: "100MB"
  max_line_length: "1MB"
  buffer_size: "64KB"
  timeout: "30m"

gui:
  window:
    width: 1024
    height: 768
    title: "SSSD Inspector"

anonymization:
  enabled: true
  patterns:
    ip_v4: '\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b'
    email: '(?i)\b[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}\b'

knowledge_base:
  tid_directory: "./kb_articles"
  auto_refresh: true
  refresh_interval: "24h"

performance:
  max_workers: 4
  gc_percent: 100

cli:
  default_generate_both_formats: true
```

---

## 🛡️ Security Features

- **Path Traversal Prevention** — Archive extraction validates all paths against directory boundaries
- **Zip/Tar Bomb Protection** — Maximum file size limit (2GB) during extraction
- **PII Redaction** — Automatic masking of IPv4/IPv6, MAC, email, and domain information
- **Safe File Handling** — Streaming file processing with near-zero RAM usage
- **Timeout Protection** — Context-aware operations prevent hangs on corrupted files

---

## 🤝 Contributing

Contributions are welcome! Here's how you can help:

1. **Add new error patterns** — Extend `buildErrorPatterns()` in `analyzer_logs.go` with SSSD error messages from real-world troubleshooting
2. **Create KB articles** — Add JSON files to `kb_articles/` with log patterns and configuration checks
3. **Improve tests** — Add test cases for new scenarios and edge cases
4. **Report issues** — Open GitHub issues with supportconfig samples (anonymized) showing missing detections

### Adding a New Error Pattern

```go
// In analyzer_logs.go, add to the errorPatterns map:
"Your new SSSD error message": "Human-readable explanation of what this means and how to fix it",
```

### Adding a KB Article

```json
{
  "tid_id": "TID-000000001",
  "title": "Your KB Article Title",
  "url": "https://www.suse.com/support/kb/doc/?id=000000001",
  "description": "Detailed description of the issue and resolution steps.",
  "log_patterns": ["Exact log line pattern to match"],
  "config_patterns": ["Relevant sssd.conf setting"]
}
```

---

## 📄 License

This program is free software; you can redistribute it and/or modify it under the terms of the **GNU General Public License version 3** as published by the Free Software Foundation.

---

## ⚠️ Disclaimer

This tool is intended for diagnostic purposes. It parses logs based on patterns found in the SSSD source code; however, always verify system configurations manually before applying changes in production environments.

---

## 👤 Author

**Davide M. Puggioni** — SUSE Technical Support

---

## 🙏 Acknowledgments

- Built with [Wails](https://wails.io/) — Native Go + WebKit desktop applications
- Pattern engine inspired by SSSD C source code analysis
- Knowledge Base articles sourced from SUSE Technical Information Database (TID)
