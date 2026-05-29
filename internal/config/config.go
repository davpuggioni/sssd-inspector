// Package config provides configuration management and automated environment loading 
// for the SSSD Inspector parsing engine.
package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Global configuration instance exposed for application-wide parameter lookups
var Global *Config

// init automatically executes on application startup to load and validate configurations
func init() {
	var err error
	Global, err = LoadConfig("")
	if err != nil {
		log.Printf("Warning: %v, using defaults", err)
		Global = DefaultConfig()
	}

	// Validate configuration integrity
	if err := Global.Validate(); err != nil {
		log.Printf("Configuration validation error: %v", err)
	}
}

// Config represents the complete application configuration matrix
type Config struct {
	App           AppMetadata         `yaml:"app"`
	Analysis      AnalysisConfig      `yaml:"analysis"`
	Files         FilesConfig         `yaml:"files"`
	GUI           GUIConfig           `yaml:"gui"`
	Anonymization AnonymizationConfig `yaml:"anonymization"`
	KnowledgeBase KnowledgeBaseConfig `yaml:"knowledge_base"`
	Logging       LoggingConfig       `yaml:"logging"`
	Performance   PerformanceConfig   `yaml:"performance"`
	Reports       ReportsConfig       `yaml:"reports"`
	CLI           CLIConfig           `yaml:"cli"`
}

// AppMetadata contains application identity metadata
type AppMetadata struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Author  string `yaml:"author"`
}

// AnalysisConfig contains analysis-related settings
type AnalysisConfig struct {
	MaxFileSize       string   `yaml:"max_file_size"`
	MaxLineLength     string   `yaml:"max_line_length"`
	BufferSize        string   `yaml:"buffer_size"`
	ArchiveFormats    []string `yaml:"archive_formats"`
	Timeout           string   `yaml:"timeout"`
	ExtractionTimeout string   `yaml:"extraction_timeout"`
	ProgressSteps     int      `yaml:"progress_steps"`
}

// FilesConfig contains file-related settings
type FilesConfig struct {
	RelevantFiles []string `yaml:"relevant_files"`
}

// GUIConfig contains GUI-related settings
type GUIConfig struct {
	Window WindowConfig `yaml:"window"`
	Colors ColorConfig  `yaml:"colors"`
}

// WindowConfig contains window settings
type WindowConfig struct {
	Width  int    `yaml:"width"`
	Height int    `yaml:"height"`
	Title  string `yaml:"title"`
}

// ColorConfig contains color settings
type ColorConfig struct {
	Background RGBA `yaml:"background"`
}

// RGBA represents a color with alpha channel
type RGBA struct {
	R uint8 `yaml:"r"`
	G uint8 `yaml:"g"`
	B uint8 `yaml:"b"`
	A uint8 `yaml:"a"`
}

// AnonymizationConfig contains PII anonymization settings
type AnonymizationConfig struct {
	Enabled      bool              `yaml:"enabled"`
	Patterns     map[string]string `yaml:"patterns"`
	Replacements map[string]string `yaml:"replacements"`
}

// KnowledgeBaseConfig contains knowledge base settings
type KnowledgeBaseConfig struct {
	TIDDirectory    string `yaml:"tid_directory"`
	AutoRefresh     bool   `yaml:"auto_refresh"`
	RefreshInterval string `yaml:"refresh_interval"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level  string        `yaml:"level"`
	Format string        `yaml:"format"`
	File   LogFileConfig `yaml:"file"`
}

// LogFileConfig contains log file settings
type LogFileConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Path       string `yaml:"path"`
	MaxSize    string `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	Compress   bool   `yaml:"compress"`
}

// PerformanceConfig contains performance-related settings
type PerformanceConfig struct {
	MaxWorkers int    `yaml:"max_workers"`
	GCPercent  int    `yaml:"gc_percent"`
	ChunkSize  string `yaml:"chunk_size"`
}

// ReportsConfig contains report generation settings
type ReportsConfig struct {
	DefaultFormats  []string          `yaml:"default_formats"`
	OutputDirectory string            `yaml:"output_directory"`
	Templates       map[string]string `yaml:"templates"`
}

// CLIConfig contains CLI-specific settings
type CLIConfig struct {
	DefaultGenerateBothFormats bool   `yaml:"default_generate_both_formats"`
	OutputSuffix               string `yaml:"output_suffix"`
}

// Constants for default values
const (
	DefaultAppName           = "SSSD Inspector"
	DefaultAppVersion        = "0.2.0"
	DefaultMaxFileSize        = "100MB"
	DefaultMaxLineLength     = "1MB"
	DefaultBufferSize        = "64KB"
	DefaultTimeout           = "30m"
	DefaultExtractionTimeout = "30m"
	DefaultProgressSteps     = 10
	DefaultWindowWidth       = 1024
	DefaultWindowHeight      = 768
	DefaultWindowTitle       = "SSSD Inspector"
	DefaultMaxWorkers        = 4
	DefaultGCPercent         = 100
	DefaultChunkSize         = "32KB"
	DefaultOutputSuffix      = "_report"
	DefaultLogLevel          = "info"
	DefaultLogFormat         = "json"
	DefaultMaxBackups        = 5
)

// DefaultConfig returns a default configuration fallback block
func DefaultConfig() *Config {
	return &Config{
		App: AppMetadata{
			Name:    DefaultAppName,
			Version: DefaultAppVersion,
		},
		Analysis: AnalysisConfig{
			MaxFileSize:       DefaultMaxFileSize,
			MaxLineLength:     DefaultMaxLineLength,
			BufferSize:        DefaultBufferSize,
			ArchiveFormats:    []string{"txz", "tar.xz", "tar.gz"},
			Timeout:           DefaultTimeout,
			ExtractionTimeout: DefaultExtractionTimeout,
			ProgressSteps:     DefaultProgressSteps,
		},
		Files: FilesConfig{
			RelevantFiles: []string{
				"nsswitch.conf", "hosts", "nscd.conf", "sssd.conf",
				"systemd.txt", "basic-environment.txt", "updates.txt",
				"y2log.txt", "sssd.txt", "rpm.txt", "etc.txt",
				"network.txt", "ntp.txt", "pam.txt", "fs-diskio.txt",
				"storage.txt", "security-apparmor.txt", "security-selinux.txt",
				"memory.txt", "sar.txt", "messages", "messages.txt", "boot.txt",
			},
		},
		GUI: GUIConfig{
			Window: WindowConfig{
				Width:  DefaultWindowWidth,
				Height: DefaultWindowHeight,
				Title:  DefaultWindowTitle,
			},
			Colors: ColorConfig{
				Background: RGBA{R: 244, G: 244, B: 249, A: 255},
			},
		},
		Anonymization: AnonymizationConfig{
			Enabled: true,
			Patterns: map[string]string{
				"ip_v4": `\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`,
				"ip_v6": `(?i)\b(?:[a-f0-9]{1,4}:){7}[a-f0-9]{1,4}\b|\b(?:[a-f0-9]{1,4}:){1,7}:|\b:(?::[a-f0-9]{1,4}){1,7}\b`,
				"mac":   `(?i)\b(?:[0-9a-f]{2}[:-]){5}[0-9a-f]{2}\b`,
				"email": `(?i)\b[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}\b`,
			},
			Replacements: map[string]string{
				"ip_v4":    "XXX.XXX.XXX.XXX",
				"ip_v6":    "XXXX:XXXX::XXXX",
				"mac":      "XX:XX:XX:XX:XX:XX",
				"email":    "[REDACTED_USER]@example.com",
				"domain":   "example.com",
				"hardware": "[REDACTED]",
			},
		},
		KnowledgeBase: KnowledgeBaseConfig{
			TIDDirectory:    "./kb_articles",
			AutoRefresh:     true,
			RefreshInterval: "24h",
		},
		Logging: LoggingConfig{
			Level:  DefaultLogLevel,
			Format: DefaultLogFormat,
			File: LogFileConfig{
				Enabled:    true,
				Path:       "./logs/sssd-inspector.log",
				MaxSize:    "10MB",
				MaxBackups: DefaultMaxBackups,
				Compress:   true,
			},
		},
		Performance: PerformanceConfig{
			MaxWorkers: DefaultMaxWorkers,
			GCPercent:  DefaultGCPercent,
			ChunkSize:  DefaultChunkSize,
		},
		Reports: ReportsConfig{
			DefaultFormats:  []string{"txt", "html"},
			OutputDirectory: "./reports",
			Templates: map[string]string{
				"html": "./templates/report.html",
				"txt":  "./templates/report.txt",
			},
		},
		CLI: CLIConfig{
			DefaultGenerateBothFormats: true,
			OutputSuffix:               DefaultOutputSuffix,
		},
	}
}

// LoadConfig loads configuration from file or returns default
func LoadConfig(configPath string) (*Config, error) {
	if configPath != "" {
		config := DefaultConfig()
		if err := loadFromFile(configPath, config); err != nil {
			return nil, fmt.Errorf("failed to load config from %s: %w", configPath, err)
		}
		return config, nil
	}

	locations := []string{
		"config.yaml",
		filepath.Join(os.Getenv("HOME"), ".sssd-inspector", "config.yaml"),
		"/etc/sssd-inspector/config.yaml",
	}

	for _, location := range locations {
		if _, err := os.Stat(location); err == nil {
			config := DefaultConfig()
			if err := loadFromFile(location, config); err != nil {
				return nil, fmt.Errorf("failed to load config from %s: %w", location, err)
			}
			return config, nil
		}
	}

	return DefaultConfig(), nil
}

// loadFromFile loads configuration from a YAML file
func loadFromFile(path string, config *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, config)
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

// Validate validates the configuration parameters
func (c *Config) Validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("app name cannot be empty")
	}
	if c.App.Version == "" {
		return fmt.Errorf("app version cannot be empty")
	}

	if _, err := time.ParseDuration(c.Analysis.Timeout); err != nil {
		return fmt.Errorf("invalid timeout format: %s", c.Analysis.Timeout)
	}

	if c.GUI.Window.Width <= 0 || c.GUI.Window.Height <= 0 {
		return fmt.Errorf("window dimensions must be positive")
	}

	if c.Performance.MaxWorkers <= 0 {
		return fmt.Errorf("max_workers must be positive")
	}
	return nil
}

// GetMaxFileSizeBytes returns max file size in bytes
func (c *Config) GetMaxFileSizeBytes() (int64, error) {
	return parseSize(c.Analysis.MaxFileSize)
}

// GetMaxLineLengthBytes returns max line length in bytes
func (c *Config) GetMaxLineLengthBytes() (int64, error) {
	return parseSize(c.Analysis.MaxLineLength)
}

// GetBufferSizeBytes returns buffer size in bytes
func (c *Config) GetBufferSizeBytes() (int64, error) {
	return parseSize(c.Analysis.BufferSize)
}

// GetTimeoutDuration returns timeout as time.Duration
func (c *Config) GetTimeoutDuration() (time.Duration, error) {
	return time.ParseDuration(c.Analysis.Timeout)
}

// parseSize parses size string (e.g., "100MB") to bytes
func parseSize(size string) (int64, error) {
	var multiplier int64
	var numStr string

	for i, r := range size {
		if r >= '0' && r <= '9' || r == '.' {
			continue
		}
		numStr = size[:i]
		unit := size[i:]

		switch unit {
		case "B", "b":
			multiplier = 1
		case "KB", "kb":
			multiplier = 1024
		case "MB", "mb":
			multiplier = 1024 * 1024
		case "GB", "gb":
			multiplier = 1024 * 1024 * 1024
		default:
			return 0, fmt.Errorf("unknown size unit: %s", unit)
		}
		break
	}

	if numStr == "" {
		numStr = size
		multiplier = 1
	}

	var value float64
	_, err := fmt.Sscanf(numStr, "%f", &value)
	if err != nil {
		return 0, fmt.Errorf("invalid size format: %s", size)
	}

	return int64(value * float64(multiplier)), nil
}
