# SSSD Inspector Documentation

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Configuration](#configuration)
- [API Reference](#api-reference)
- [Development Guide](#development-guide)
- [Testing](#testing)
- [Deployment](#deployment)

## Overview

SSSD Inspector is a specialized diagnostic tool for analyzing SSSD (System Security Services Daemon) configurations and logs from SUSE Linux Enterprise Server (SLES) supportconfig archives.

### Key Features

- **Hybrid Mode**: Both CLI and GUI interfaces
- **Streaming Analysis**: Efficient processing of large log files with minimal memory footprint
- **Pattern Matching**: Advanced pattern detection based on SSSD source code
- **Knowledge Base Integration**: Automatic matching with SUSE TID articles
- **PII Protection**: Built-in anonymization capabilities
- **Multi-format Reports**: TXT, HTML, and PDF output

### Architecture

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

## Configuration

The application uses a YAML configuration file (`config.yaml`) for customizable settings. See [Configuration Guide](./configuration.md) for detailed options.

## API Reference

### Backend Methods

- `Analyze(targetPath string, anonymize bool) (ReportData, error)`
- `OpenFileBrowser() (string, error)`
- `SavePDF(b64 string) (string, error)`
- `SaveTXT(report ReportData) (string, error)`

See [API Documentation](./api.md) for complete reference.

## Development Guide

### Prerequisites

- Go 1.23+
- Node.js & npm
- Wails CLI v2

### Setup

```bash
# Install dependencies
go mod tidy
cd frontend && npm install

# Run in development mode
wails dev

# Build for production
wails build -platform linux/amd64 -ldflags "-w -s" -clean
```

### Code Structure

```
sssd-inspector/
├── main.go              # Application entry point
├── app.go               # Wails app structure
├── types.go             # Data structures
├── config/              # Configuration management
├── analyzer/            # Analysis modules
│   ├── core.go          # Core analysis logic
│   ├── auth.go          # Authentication analysis
│   ├── logs.go          # Log parsing
│   └── system.go        # System diagnostics
├── utils/               # Utility functions
├── report/              # Report generation
└── frontend/            # Web interface
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./analyzer/...
```

### Test Coverage

- Core analysis logic: 85%+
- Utility functions: 90%+
- Report generation: 80%+

## Deployment

### Building Binaries

```bash
# Linux
wails build -platform linux/amd64

# Windows
wails build -platform windows/amd64

# macOS
wails build -platform darwin/amd64
```

### Distribution

Built binaries include:
- All dependencies embedded
- No external runtime requirements
- Cross-platform compatibility
