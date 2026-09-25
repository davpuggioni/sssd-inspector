# Configuration Guide

## Overview

SSSD Inspector uses a YAML-based configuration system that allows customization of analysis parameters, file paths, and application behavior without code changes.

## Configuration File Location

The application searches for `config.yaml` in the following order:
1. Current working directory
2. User's home directory (`~/.sssd-inspector/config.yaml`)
3. System-wide configuration (`/etc/sssd-inspector/config.yaml`)

## Configuration Structure

```yaml
# Application metadata
app:
  name: "SSSD Inspector"
  version: "0.2.3"
  author: "Davide Michele Puggioni"
  
# Analysis parameters
analysis:
  # File processing limits
  max_file_size: "100MB"
  max_line_length: "1MB"
  buffer_size: "64KB"
  
  # Supported archive formats
  archive_formats:
    - "txz"
    - "tar.xz"
    - "tar.gz"
  
  # Analysis timeouts
  timeout: "30m"
  
# File patterns and locations
files:
  # Relevant files for analysis
  relevant_files:
    - "nsswitch.conf"
    - "hosts"
    - "nscd.conf"
    - "sssd.conf"
    - "systemd.txt"
    - "basic-environment.txt"
    - "updates.txt"
    - "y2log.txt"
    - "sssd.txt"
    - "rpm.txt"
    - "etc.txt"
    - "network.txt"
    - "ntp.txt"
    - "pam.txt"
    - "fs-diskio.txt"
    - "storage.txt"
    - "security-apparmor.txt"
    - "security-selinux.txt"
    - "memory.txt"
    - "sar.txt"
    - "messages"
    - "messages.txt"
    - "boot.txt"

# GUI settings
gui:
  window:
    width: 1024
    height: 768
    title: "SSSD Inspector"
  
  colors:
    background:
      r: 244
      g: 244
      b: 249
      a: 255

# Anonymization settings
anonymization:
  enabled: true
  
  # PII patterns
  patterns:
    ip_v4: '\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b'
    ip_v6: '(?i)\b(?:[a-f0-9]{1,4}:){7}[a-f0-9]{1,4}\b|\b(?:[a-f0-9]{1,4}:){1,7}:|\b:(?::[a-f0-9]{1,4}){1,7}\b'
    mac: '(?i)\b(?:[0-9a-f]{2}[:-]){5}[0-9a-f]{2}\b'
    email: '(?i)\b[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}\b'
  
  # Replacement strings
  replacements:
    ip_v4: "XXX.XXX.XXX.XXX"
    ip_v6: "XXXX:XXXX::XXXX"
    mac: "XX:XX:XX:XX:XX:XX"
    email: "[REDACTED_USER]@example.com"
    domain: "example.com"
    hardware: "[REDACTED]"

# Knowledge base settings
knowledge_base:
  # TID articles directory
  tid_directory: "./kb_articles"
  
  # Auto-refresh settings
  auto_refresh: true
  refresh_interval: "24h"

# Logging settings
logging:
  level: "info"  # debug, info, warn, error
  format: "json"  # json, text
  
  # Log file settings
  file:
    enabled: true
    path: "./logs/sssd-inspector.log"
    max_size: "10MB"
    max_backups: 5
    compress: true

# Performance settings
performance:
  # Concurrent processing
  max_workers: 4
  
  # Memory management
  gc_percent: 100
  
  # Streaming settings
  chunk_size: "32KB"
  
# Report settings
reports:
  # Default formats
  default_formats:
    - "txt"
    - "html"
  
  # Output directory
  output_directory: "./reports"
  
  # Template settings
  templates:
    html: "./templates/report.html"
    txt: "./templates/report.txt"
```

## Environment Variables

Configuration can be overridden using environment variables:

```bash
# Override configuration file location
SSSD_INSPECTOR_CONFIG="/path/to/custom/config.yaml"

# Override specific settings
SSSD_INSPECTOR_LOG_LEVEL="debug"
SSSD_INSPECTOR_GUI_WIDTH="1280"
SSSD_INSPECTOR_ANALYSIS_TIMEOUT="60m"
```

## Validation

The configuration is validated on startup. Invalid configurations will result in error messages and fallback to default values.

### Common Validation Errors

1. **Invalid time format**: Use Go duration format (e.g., "30m", "1h", "24h")
2. **Invalid size format**: Use size format with units (e.g., "64KB", "100MB", "1GB")
3. **Invalid color values**: RGBA values must be 0-255
4. **Missing required fields**: All required sections must be present

## Default Configuration

If no configuration file is found, the application uses built-in defaults. These defaults are optimized for typical usage scenarios.

## Configuration Migration

When upgrading from versions without configuration support, the application will automatically generate a default configuration file on first run.
