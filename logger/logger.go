// Package logger provides structured logging capabilities for SSSD Inspector
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"
)

// Level represents log severity levels
type Level int

const (
	// DebugLevel for detailed diagnostic information
	DebugLevel Level = iota
	// InfoLevel for general operational information
	InfoLevel
	// WarnLevel for warning conditions
	WarnLevel
	// ErrorLevel for error conditions
	ErrorLevel
)

// String returns the string representation of a log level
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Fields represents structured log fields
type Fields map[string]interface{}

// Logger defines the interface for structured logging
type Logger interface {
	// Debug logs a debug message
	Debug(msg string, fields Fields)
	// Info logs an info message
	Info(msg string, fields Fields)
	// Warn logs a warning message
	Warn(msg string, fields Fields)
	// Error logs an error message
	Error(msg string, err error, fields Fields)
	// WithFields returns a new logger with additional fields
	WithFields(fields Fields) Logger
	// SetLevel sets the minimum log level
	SetLevel(level Level)
}

// logger implements the Logger interface
type logger struct {
	level  Level
	output io.Writer
	fields Fields
	prefix string
}

// Config holds logger configuration
type Config struct {
	Level  Level
	Output io.Writer
	Prefix string
}

// DefaultConfig returns a default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:  InfoLevel,
		Output: os.Stderr,
		Prefix: "",
	}
}

// New creates a new structured logger
func New(config *Config) Logger {
	if config == nil {
		config = DefaultConfig()
	}
	if config.Output == nil {
		config.Output = os.Stderr
	}

	return &logger{
		level:  config.Level,
		output: config.Output,
		fields: make(Fields),
		prefix: config.Prefix,
	}
}

// globalLogger is the default global logger
var globalLogger Logger = New(DefaultConfig())

// Global logger functions for backward compatibility

// Debug logs a debug message using the global logger
func Debug(msg string, fields Fields) {
	globalLogger.Debug(msg, fields)
}

// Info logs an info message using the global logger
func Info(msg string, fields Fields) {
	globalLogger.Info(msg, fields)
}

// Warn logs a warning message using the global logger
func Warn(msg string, fields Fields) {
	globalLogger.Warn(msg, fields)
}

// Error logs an error message using the global logger
func Error(msg string, err error, fields Fields) {
	globalLogger.Error(msg, err, fields)
}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(l Logger) {
	globalLogger = l
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() Logger {
	return globalLogger
}

// Implementation of Logger interface

func (l *logger) Debug(msg string, fields Fields) {
	if l.level <= DebugLevel {
		l.log(DebugLevel, msg, fields)
	}
}

func (l *logger) Info(msg string, fields Fields) {
	if l.level <= InfoLevel {
		l.log(InfoLevel, msg, fields)
	}
}

func (l *logger) Warn(msg string, fields Fields) {
	if l.level <= WarnLevel {
		l.log(WarnLevel, msg, fields)
	}
}

func (l *logger) Error(msg string, err error, fields Fields) {
	if l.level <= ErrorLevel {
		if err != nil {
			if fields == nil {
				fields = make(Fields)
			}
			fields["error"] = err.Error()
		}
		l.log(ErrorLevel, msg, fields)
	}
}

func (l *logger) WithFields(fields Fields) Logger {
	newFields := make(Fields)
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &logger{
		level:  l.level,
		output: l.output,
		fields: newFields,
		prefix: l.prefix,
	}
}

func (l *logger) SetLevel(level Level) {
	l.level = level
}

func (l *logger) log(level Level, msg string, fields Fields) {
	// Merge logger fields with message fields
	allFields := make(Fields)
	for k, v := range l.fields {
		allFields[k] = v
	}
	for k, v := range fields {
		allFields[k] = v
	}

	// Get caller information
	_, file, line, _ := runtime.Caller(3) // Skip runtime.Caller, log, and this function

	// Format timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	// Build log message
	logMsg := fmt.Sprintf("%s [%s] %s", timestamp, level.String(), msg)
	if l.prefix != "" {
		logMsg = fmt.Sprintf("%s %s", l.prefix, logMsg)
	}

	// Add file location for debug level
	if level == DebugLevel {
		logMsg = fmt.Sprintf("%s (%s:%d)", logMsg, file, line)
	}

	// Add fields
	if len(allFields) > 0 {
		fieldStr := ""
		for k, v := range allFields {
			if fieldStr != "" {
				fieldStr += " "
			}
			fieldStr += fmt.Sprintf("%s=%v", k, v)
		}
		logMsg = fmt.Sprintf("%s %s", logMsg, fieldStr)
	}

	// Write to output
	fmt.Fprintln(l.output, logMsg)
}

// Legacy compatibility functions for gradual migration

// Printf provides backward compatibility with log.Printf
func Printf(format string, v ...interface{}) {
	globalLogger.Info(fmt.Sprintf(format, v...), nil)
}

// Fatalf provides backward compatibility with log.Fatalf
func Fatalf(format string, v ...interface{}) {
	globalLogger.Error(fmt.Sprintf(format, v...), nil, nil)
	log.Fatalf(format, v...)
}

// PrintWarning provides backward compatibility for warnings
func PrintWarning(format string, v ...interface{}) {
	globalLogger.Warn(fmt.Sprintf(format, v...), nil)
}
