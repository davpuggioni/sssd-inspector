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
├── main_cli.go             # CLI entry point (build tag: cli)
├── main_gui.go             # GUI entry point (build tag: !cli)
├── app.go                  # Wails app structure and API methods (GUI mode)
├── shared.go               # Shared CLI/GUI logic (runCLI, runLogDirAnalyze)
├── types.go                # Root-level type definitions (ReportData, etc.)
├── utils.go                # Legacy wrapper functions (backward compatibility)
├── legacy_test_helpers.go  # Bridge functions for existing tests
├── config/                 # Configuration management
│   └── config.go           # Configuration structures and loading
├── constants/              # Application constants
│   └── constants.go        # All constant definitions
├── errors/                 # Custom error types and helpers
│   └── errors.go           # Error wrapping, context, helpers
├── logger/                 # Logging infrastructure
│   └── logger.go           # Structured logger with levels
├── pkg/
│   ├── analysis/           # Core analysis engine
│   │   ├── core.go         # Core analysis orchestration, AnonymizeReport
│   │   ├── context.go      # AnalyzerContext (file I/O, cache, scanning)
│   │   ├── auth.go         # Authentication analysis
│   │   ├── logs.go         # Log parsing and analysis
│   │   ├── kb.go           # Knowledge base article matching
│   │   ├── singlepass.go   # Single-pass log scanning engine
│   │   └── system.go       # System diagnostics
│   ├── extract/            # Archive extraction
│   │   └── extract.go      # XZ/tar extraction with path traversal protection
│   ├── fileutil/           # File I/O utilities
│   │   ├── cache.go        # Regex cache, file cache, scanner pool
│   │   ├── filter.go       # File relevance filtering
│   │   └── scanner.go      # File scanning utilities
│   ├── report/             # Report generation
│   │   └── report.go       # Text and HTML report building functions
│   └── types/              # Shared data structures
│       └── types.go        # ReportData, TIDArticle, TimelineEvent, SSSDLogError
├── frontend/               # Wails frontend (JavaScript/CSS)
│   ├── src/                # Source files
│   ├── wailsjs/            # Wails generated bindings
│   └── package.json        # Frontend dependencies
├── docs/                   # Documentation
├── kb_articles/            # Knowledge base articles (JSON)
├── config.yaml             # Default configuration file
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
└── wails.json              # Wails configuration
```

### Build Tags

The project uses Go build tags to separate CLI and GUI entry points:

- **CLI build**: `go build -tags cli` — uses `main_cli.go`, excludes `app.go`
- **GUI build**: `wails build` or `wails dev` — uses `main_gui.go`, includes `app.go`
- **Test**: `go test ./...` — runs all tests regardless of build tags

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
   - Update frontend JavaScript to call new methods
   - Update API documentation

#### Frontend Changes (JavaScript)

1. **UI modifications**:
   - Edit `frontend/src/main.js`
   - Update CSS styles as needed
   - Test with both CLI and GUI modes

2. **Build process**:
   ```bash
   cd frontend
   npm run build
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
go test ./pkg/analysis/...
go test ./pkg/fileutil/...

# Run tests with verbose output
go test -v ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### Writing Tests

```go
// Example test file: pkg/analysis/core_test.go
package analysis

import (
    "testing"
)

func TestGetSSSDVersion(t *testing.T) {
    tests := []struct {
        name       string
        packages   []string
        wantMajor  int
        wantMinor  int
    }{
        {
            name:      "valid version",
            packages:  []string{"sssd-2.6.2-1.x86_64"},
            wantMajor: 2,
            wantMinor: 6,
        },
        {
            name:      "invalid version",
            packages:  []string{"other-package-1.0"},
            wantMajor: 0,
            wantMinor: 0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            major, minor := GetSSSDVersion(tt.packages)
            if major != tt.wantMajor || minor != tt.wantMinor {
                t.Errorf("GetSSSDVersion() = (%d, %d), want (%d, %d)",
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
4. **Build release binaries**:
   ```bash
   wails build -platform linux/amd64,windows/amd64,darwin/amd64 -clean
   ```
5. **Test binaries** on target platforms
6. **Create release tag**:
   ```bash
   git tag -a v0.2.0 -m "Release version 0.2.0"
   git push origin v0.2.0
   ```
7. **Upload binaries** to release platform
