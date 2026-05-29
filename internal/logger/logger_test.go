// Package logger tests for structured logging
package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestLoggerDebug(t *testing.T) {
	var buf bytes.Buffer
	log := New(&Config{
		Level:  DebugLevel,
		Output: &buf,
	})

	log.Debug("test debug message", Fields{"key": "value"})

	output := buf.String()
	if !strings.Contains(output, "[DEBUG]") {
		t.Errorf("Expected [DEBUG] in output, got: %s", output)
	}
	if !strings.Contains(output, "test debug message") {
		t.Errorf("Expected message in output, got: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("Expected fields in output, got: %s", output)
	}
}

func TestLoggerInfo(t *testing.T) {
	var buf bytes.Buffer
	log := New(&Config{
		Level:  InfoLevel,
		Output: &buf,
	})

	log.Info("test info message", nil)

	output := buf.String()
	if !strings.Contains(output, "[INFO]") {
		t.Errorf("Expected [INFO] in output, got: %s", output)
	}
	if !strings.Contains(output, "test info message") {
		t.Errorf("Expected message in output, got: %s", output)
	}
}

func TestLoggerWarn(t *testing.T) {
	var buf bytes.Buffer
	log := New(&Config{
		Level:  WarnLevel,
		Output: &buf,
	})

	log.Warn("test warning", Fields{"warning": "test"})

	output := buf.String()
	if !strings.Contains(output, "[WARN]") {
		t.Errorf("Expected [WARN] in output, got: %s", output)
	}
}

func TestLoggerError(t *testing.T) {
	var buf bytes.Buffer
	log := New(&Config{
		Level:  ErrorLevel,
		Output: &buf,
	})

	log.Error("test error", nil, nil)

	output := buf.String()
	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("Expected [ERROR] in output, got: %s", output)
	}
	if !strings.Contains(output, "test error") {
		t.Errorf("Expected message in output, got: %s", output)
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	log := New(&Config{
		Level:  ErrorLevel, // Only show errors
		Output: &buf,
	})

	log.Debug("debug message", nil)
	log.Info("info message", nil)
	log.Warn("warn message", nil)

	// Debug, Info, Warn should not appear when Level is Error
	output := buf.String()
	if strings.Contains(output, "debug message") {
		t.Errorf("Debug message should be filtered out at Error level")
	}
	if strings.Contains(output, "info message") {
		t.Errorf("Info message should be filtered out at Error level")
	}
	if strings.Contains(output, "warn message") {
		t.Errorf("Warn message should be filtered out at Error level")
	}
}

func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer
	log := New(&Config{
		Level:  InfoLevel,
		Output: &buf,
	})

	loggerWithFields := log.WithFields(Fields{"component": "test"})
	loggerWithFields.Info("message with fields", Fields{"action": "test"})

	output := buf.String()
	if !strings.Contains(output, "component=test") {
		t.Errorf("Expected component field in output, got: %s", output)
	}
	if !strings.Contains(output, "action=test") {
		t.Errorf("Expected action field in output, got: %s", output)
	}
}

func TestLoggerErrorWithCause(t *testing.T) {
	var buf bytes.Buffer
	log := New(&Config{
		Level:  ErrorLevel,
		Output: &buf,
	})

	err := &testError{"something went wrong"}
	log.Error("operation failed", err, Fields{"operation": "test"})

	output := buf.String()
	if !strings.Contains(output, "operation failed") {
		t.Errorf("Expected message in output, got: %s", output)
	}
	if !strings.Contains(output, "something went wrong") {
		t.Errorf("Expected error cause in output, got: %s", output)
	}
}

func TestLoggerSetLevel(t *testing.T) {
	var buf bytes.Buffer
	log := New(&Config{
		Level:  InfoLevel,
		Output: &buf,
	})

	log.SetLevel(DebugLevel)
	log.Debug("debug after level change", nil)

	output := buf.String()
	if !strings.Contains(output, "[DEBUG]") {
		t.Errorf("Expected debug message after level change, got: %s", output)
	}
}

func TestGlobalLogger(t *testing.T) {
	// Reset to default for this test
	oldLogger := GetGlobalLogger()
	defer SetGlobalLogger(oldLogger)

	var buf bytes.Buffer
	testLogger := New(&Config{
		Level:  InfoLevel,
		Output: &buf,
	})
	SetGlobalLogger(testLogger)

	Info("global info test", Fields{"global": "true"})
	Warn("global warn test", nil)

	output := buf.String()
	if !strings.Contains(output, "global info test") {
		t.Errorf("Expected global info message in output, got: %s", output)
	}
	if !strings.Contains(output, "global warn test") {
		t.Errorf("Expected global warn message in output, got: %s", output)
	}
}

func TestLoggerWithPrefix(t *testing.T) {
	var buf bytes.Buffer
	log := New(&Config{
		Level:  InfoLevel,
		Output: &buf,
		Prefix: "[TEST]",
	})

	log.Info("prefixed message", nil)

	output := buf.String()
	if !strings.Contains(output, "[TEST]") {
		t.Errorf("Expected prefix in output, got: %s", output)
	}
}

// testError implements the error interface for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
