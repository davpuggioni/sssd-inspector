# Development Guide

## Overview

This guide provides comprehensive information for developers working on SSSD Inspector, including setup procedures, coding standards, and development workflows.

## Prerequisites

### Required Tools

- **Go 1.23+**: Main programming language
- **Node.js & npm**: Frontend development
- **Wails CLI v2**: Desktop application framework

### Installation

```bash
# Install Go
# Visit https://golang.org/dl/ for installation instructions

# Install Node.js
# Visit https://nodejs.org/ for installation instructions

# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Verify installation
wails version
```

## Project Structure

```
sssd-inspector/
├── main.go              # Application entry point
├── app.go               # Wails app structure and API methods
├── types.go             # Data structures and type definitions
├── config/              # Configuration management
│   └── config.go        # Configuration structures and loading
├── constants/           # Application constants
│   └── constants.go     # All constant definitions
├── utils/               # Utility functions
│   └── utils.go         # File processing utilities
├── analyzer/            # Analysis modules
│   ├── core.go          # Core analysis orchestration
│   ├── auth.go          # Authentication analysis
│   ├── logs.go          # Log parsing and analysis
│   └── system.go        # System diagnostics
├── report/              # Report generation
│   └── report.go        # Report building functions
├── kb.go                # Knowledge base management
├── loaders.go           # Archive extraction utilities
├── frontend/            # Web interface
│   ├── src/             # Source files
│   ├── dist/            # Built files
│   └── package.json     # Frontend dependencies
├── docs/                # Documentation
├── kb_articles/         # Knowledge base articles
├── config.yaml          # Default configuration file
├── go.mod               # Go module definition
└── wails.json           # Wails configuration
```

## Development Workflow

### 1. Setup Development Environment

```bash
# Clone the repository
git clone <repository-url>
cd sssd-inspector

# Install Go dependencies
go mod tidy

# Install frontend dependencies
cd frontend
npm install
cd ..

# Run in development mode
wails dev
```

### 2. Making Changes

#### Backend Changes (Go)

1. **Add new functionality**:
   - Create appropriate types in `types.go`
   - Implement logic in relevant analyzer files
   - Add constants if needed
   - Write tests

2. **Configuration changes**:
   - Update `config.yaml` with new settings
   - Modify `config/config.go` structures
   - Add validation if needed

3. **API changes**:
   - Add methods to `App` struct in `app.go`
   - Regenerate the bindings: `wails generate module`
   - Expose them in `frontend/src/api/backend.ts` and use them from a component
   - Update API documentation

#### Frontend changes (React + TypeScript)

1. **UI modifications**:
   - Edit `frontend/src/App.tsx` and the component under `frontend/src/components/`
   - Put new state in a hook under `frontend/src/hooks/`
   - Update CSS styles as needed
   - Test with both CLI and GUI modes

2. **Build process**:
   ```bash
   cd frontend
   npm test         # vitest: type-safe component tests
   npm run build    # tsc --noEmit && vite build
   cd ..
   wails build
   ```

### 3. Testing

#### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./analyzer/...

# Run tests with verbose output
go test -v ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### Writing Tests

```go
// Example test file: analyzer_core_test.go
package main

import (
    "testing"
    "sssd-inspector/constants"
)

func TestGetSSSDVersion(t *testing.T) {
    tests := []struct {
        name     string
        packages []string
        wantMajor int
        wantMinor int
    }{
        {
            name: "valid version",
            packages: []string{"sssd-2.6.2-1.x86_64"},
            wantMajor: 2,
            wantMinor: 6,
        },
        {
            name: "invalid version",
            packages: []string{"other-package-1.0"},
            wantMajor: 0,
            wantMinor: 0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            major, minor := getSSSDVersion(tt.packages)
            if major != tt.wantMajor || minor != tt.wantMinor {
                t.Errorf("getSSSDVersion() = (%d, %d), want (%d, %d)",
                    major, minor, tt.wantMajor, tt.wantMinor)
            }
        })
    }
}
```

### 4. Building

#### Development Build

```bash
wails dev
```

#### Production Build

```bash
# Linux
wails build -platform linux/amd64 -ldflags "-w -s" -clean

# Windows
wails build -platform windows/amd64

# macOS
wails build -platform darwin/amd64
```

#### Cross-Platform Build

```bash
# Build for all platforms
wails build -platform linux/amd64,windows/amd64,darwin/amd64
```

## Coding Standards

### Go Code Style

1. **Use gofmt**: Format all Go code with `gofmt`
2. **GoDoc comments**: Document all exported functions and types
3. **Error handling**: Use proper error wrapping with `%w`
4. **Constants**: Use the constants package for all magic numbers
5. **Configuration**: Make features configurable through config.yaml

#### Example

```go
// NewFileProcessor creates a new FileProcessor with default settings
func NewFileProcessor() *FileProcessor {
    return &FileProcessor{
        bufferSize:    constants.DefaultBufferSize,
        maxLineLength: constants.DefaultMaxLineLength,
    }
}

// ProcessFile processes a file with the given options
// It returns the processed content or an error if processing fails.
func (fp *FileProcessor) ProcessFile(path string, options ProcessOptions) (string, error) {
    if path == "" {
        return "", fmt.Errorf("file path cannot be empty")
    }
    
    // Implementation...
    
    if err != nil {
        return "", fmt.Errorf("failed to process file %s: %w", path, err)
    }
    
    return result, nil
}
```

### JavaScript Code Style

1. **Use modern ES6+ features**
2. **Document functions with JSDoc**
3. **Handle errors gracefully**
4. **Use constants for magic strings**

#### Example

```javascript
/**
 * Updates the progress bar and status message
 * @param {number} percentage - Progress percentage (0-100)
 * @param {string} message - Status message to display
 */
function updateProgress(percentage, message) {
    const progressBar = document.getElementById('progressBar');
    const statusMessage = document.getElementById('statusMessage');
    
    if (progressBar) {
        progressBar.style.width = `${percentage}%`;
    }
    
    if (statusMessage) {
        statusMessage.textContent = message;
    }
}
```

## Configuration Management

### Adding New Configuration Options

1. **Update config.yaml**:
   ```yaml
   new_section:
     new_option: "default_value"
     another_option: 42
   ```

2. **Update config/config.go**:
   ```go
   type NewSectionConfig struct {
       NewOption    string `yaml:"new_option"`
       AnotherOption int   `yaml:"another_option"`
   }

   type Config struct {
       // ... existing fields ...
       NewSection NewSectionConfig `yaml:"new_section"`
   }
   ```

3. **Add validation**:
   ```go
   func (c *Config) Validate() error {
       // ... existing validation ...
       
       if c.NewSection.NewOption == "" {
           return fmt.Errorf("new_option cannot be empty")
       }
       
       return nil
   }
   ```

4. **Add constants** (if needed):
   ```go
   const (
       DefaultNewOption = "default_value"
       DefaultAnotherOption = 42
   )
   ```

## Adding New Diagnostic Patterns

SSSD Inspector matches diagnostic signals across SSSD log files, configuration files (`sssd.conf`), and Kerberos/DNS state. Depending on what you are adding and whether you can recompile the binary, there are three main extension points.

### 1. Built-in SSSD Log Patterns (Requires Recompilation)

This is the primary way to detect diagnostic messages emitted by SSSD components (`sssd_be`, `krb5_child`, `ldap_child`, responders).

#### Step A: Extract the Literal Signature from SSSD C Source

Look at the SSSD C source code (typically `DEBUG(SSSDBG_CRIT_FAILURE, ...)`, `SSSDBG_OP_FAILURE`, or `SSSDBG_ERR`) in `src/providers/` or `src/responder/`.

Example log line from SSSD:
```text
(2026-01-01 10:00:00) [sssd[be[example.com]]] [sdap_async] (0x0020): Failed to set the LDAP referrer option
```

**Rule for patterns:**
- Use a **contiguous substring** of the log message.
- Stop **before format specifiers** (`%s`, `%d`, etc.) because the log line will contain the substituted value at runtime.
- Matching is case-insensitive.

#### Step B: Add the Pattern to `analyzer_logs.go` (`buildErrorPatterns`)

Open `analyzer_logs.go` and add an entry inside `buildErrorPatterns()`:

```go
"Failed to set the LDAP referrer option": "LDAP Referral Error: LDAP/AD server rejected the referral chase option. Searches relying on referral chasing may fail. Check ldap_referrals and DC ACLs.",
```

The key is the substring to match, and the value is a formatted description:
`"<Category Title>: <explanation and remediation advice>"`.

#### Step C: Register the Category Prefix in `logPatternCategory`

In `analyzer_logs.go`, register the descriptive prefix in `logPatternCategory`:

```go
var logPatternCategory = map[string]string{
    // ...
    "LDAP Referral Error": "ldap",
}
```

This prefix is matched case-insensitively by `categoryFor()` (`analyzer_logs.go`). It binds the log finding to one of the root-cause domains (`dns`, `srv`, `net`, `time`, `join`, `keytab`, `krb5`, `crypto`, `tls`, `sasl`, `gpo`, `idmap`, `db`, `cache`, `enum`, `access`, `offline`).

If you introduce a brand new category, also configure its relative weight in `logCategorySeverity` in `analysis_scoring.go` (weights range from 1 to 3).

#### Step D: Verify the Pre-Filter in `analyzer_singlepass.go`

SSSD Inspector uses an Aho-Corasick automaton to discard ~90% of unrelated syslog lines before detailed scanning. Check `prefilterKeywords` in `analyzer_singlepass.go`:

```go
var prefilterKeywords = []string{
    "sssd", "krb5", "ldap", "keytab", "winbind", "ad ", "gpo", "pam",
    "nss", "hbac", "ipa", "kdc", "tgt ", "tls", "gssapi", "library",
    "dlopen", "shared",
}
```

If your message comes from an external child process or a log where none of these keywords appear on the line, append a relevant keyword to `prefilterKeywords`.

#### Step E: Regular Expression Patterns (Optional)

By default, patterns run through the fast O(n) Aho-Corasick multi-pattern trie (`ahocorasick.go`). If your pattern requires regex syntax, it must contain one of `regexMetaTokens` (`.*`, `\d`, `\s`, `\w`, `+?`, `(?i)`) to be routed to RE2 in `analyzer_singlepass.go`.

---

### 2. Site-Specific Rules via `rules.yaml` (No Recompilation)

For customer-specific or custom diagnostic checks without modifying Go code, create `rules.yaml` (or drop `.yaml` files into a `rules/` directory next to the binary or current working directory).

Template:
```yaml
rules:
  - name: "custom-ldap-referral-check"
    severity: "warning"          # critical | error | warning
    category: "ldap"             # category string
    files: ["sssd.txt", "messages", "messages.txt"]
    patterns:
      - "Failed to set the LDAP referrer option"
    pattern_type: "literal"      # literal (default) | regex
    match: "any"                 # any (default) | all (all patterns on the same line)
    message: "LDAP referral option was rejected by the domain controller."
```

Rules are additive: they generate `ConfigFinding` entries in the diagnostic report.

---

### 3. SUSE Knowledge Base Articles via `kb_articles/*.json`

When an error pattern should be linked to an official resolution article, create a JSON file in `kb_articles/`:

```json
{
  "tid_id": "TID-000000000",
  "title": "LDAP referral option rejected by Active Directory",
  "url": "https://www.suse.com/support/kb/doc/?id=000000000",
  "description": "Active Directory domain controllers reject the LDAP referral chasing option.",
  "log_patterns": [
    "Failed to set the LDAP referrer option"
  ],
  "config_patterns": [
    "ldap_referrals"
  ]
}
```

The single-pass scanner will extract evidence lines and correlate them during the KB matching phase.

---

### Testing and Hygiene Guidelines

When adding new patterns or test fixtures:
1. **Never use real company names, live hostnames, or production IP addresses.** Always use RFC 2606 reserved domains (e.g., `example.com`, `example.org`) and test IPs (e.g., `192.0.2.10`, `198.51.100.1`). `test_fixture_hygiene_test.go` enforces this rule.
2. Add a unit test in `analyzer_logs_test.go` or `analysis_scoring_test.go` confirming that your pattern matches and is categorized as expected.
3. Verify test suite with:
   ```bash
   go test ./... && go test -tags cli ./...
   ```


## Debugging

### Backend Debugging

1. **Use logging**:
   ```go
   log.Printf("Debug: processing file %s", filename)
   ```

2. **Use Delve debugger**:
   ```bash
   go install github.com/go-delve/delve/cmd/dlv@latest
   dlv debug
   ```

3. **Add test breakpoints**:
   ```go
   if debug {
       fmt.Printf("Debug point: %+v\n", data)
   }
   ```

### Frontend Debugging

1. **Use browser developer tools**
2. **Add console.log statements**:
   ```javascript
   console.log('Debug:', data);
   ```

3. **Use Wails debug mode**:
   ```bash
   wails dev -debug
   ```

## Performance Considerations

### Memory Management

1. **Use streaming for large files**: The `FileProcessor` struct provides efficient streaming
2. **Clean up resources**: Use `defer` for file handles and temporary directories
3. **Limit buffer sizes**: Configure appropriate buffer sizes in config.yaml

### Performance Profiling

```bash
# CPU profiling
go tool pprof http://localhost:6060/debug/pprof/profile

# Memory profiling
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profiling
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

## Contributing

### Pull Request Process

1. **Create feature branch**:
   ```bash
   git checkout -b feature/new-feature
   ```

2. **Make changes** following coding standards
3. **Add tests** for new functionality
4. **Ensure all tests pass**:
   ```bash
   go test ./...
   ```

5. **Update documentation** as needed
6. **Submit pull request** with clear description

### Code Review Checklist

- [ ] Code follows project style guidelines
- [ ] All tests pass
- [ ] Documentation is updated
- [ ] Configuration is properly handled
- [ ] Error handling is robust
- [ ] No hardcoded magic numbers
- [ ] GoDoc comments are present
- [ ] Security considerations are addressed

## Troubleshooting

### Common Issues

1. **Build fails with missing dependencies**:
   ```bash
   go mod tidy
   ```

2. **Frontend not updating**:
   ```bash
   cd frontend
   npm run build
   cd ..
   ```

3. **Configuration not loading**:
   - Check config.yaml syntax
   - Verify file permissions
   - Check error logs

4. **Tests failing**:
   - Ensure all dependencies are installed
   - Check test data files
   - Verify configuration

### Getting Help

1. **Check logs**: Look for error messages in console output
2. **Review documentation**: Check relevant sections in docs/
3. **Debug systematically**: Use breakpoints and logging
4. **Ask for help**: Contact the development team

## Release Process

1. **Update version** in `constants/constants.go` and `config.yaml`
2. **Update CHANGELOG.md** with new features and fixes
3. **Run full test suite**:
   ```bash
   go test ./...
   ```
4. **Build and package the release**:
   ```bash
   ./build_release.sh
   ```
   The script runs `wails build -platform linux/amd64 -tags webkit2_41 -ldflags "-w -s" -clean`,
   checks the built binary's `-v` output against `constants.AppVersion`, and
   writes `dist/sssd-inspector-<version>-linux-amd64.tar.gz` (binary +
   `kb_articles/` + `LICENSE` + `README.md`, plus a `.sha256`).
5. **Test binaries** on target platforms
6. **Create release tag**:
   ```bash
   git tag -a v0.2.3 -m "Release version 0.2.3"
   git push origin v0.2.3
   ```
7. **Upload binaries** to release platform
