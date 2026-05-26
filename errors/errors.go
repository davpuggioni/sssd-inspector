// Package errors provides custom error types for SSSD Inspector
package errors

import (
	"fmt"
)

// ErrorCode represents error categories
type ErrorCode string

const (
	// ErrUnknown is the default error code
	ErrUnknown ErrorCode = "UNKNOWN"

	// File operation errors
	ErrFileNotFound   ErrorCode = "FILE_NOT_FOUND"
	ErrFileAccess     ErrorCode = "FILE_ACCESS"
	ErrFileCorrupt    ErrorCode = "FILE_CORRUPT"
	ErrFileTooLarge   ErrorCode = "FILE_TOO_LARGE"
	ErrInvalidArchive ErrorCode = "INVALID_ARCHIVE"
	ErrPathTraversal  ErrorCode = "PATH_TRAVERSAL"

	// Analysis errors
	ErrAnalysisFailed   ErrorCode = "ANALYSIS_FAILED"
	ErrParsingFailed    ErrorCode = "PARSING_FAILED"
	ErrValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrTimeout          ErrorCode = "TIMEOUT"

	// Configuration errors
	ErrConfigNotFound ErrorCode = "CONFIG_NOT_FOUND"
	ErrConfigInvalid  ErrorCode = "CONFIG_INVALID"
	ErrConfigMissing  ErrorCode = "CONFIG_MISSING"

	// Permission errors
	ErrPermissionDenied ErrorCode = "PERMISSION_DENIED"

	// System errors
	ErrSystemError       ErrorCode = "SYSTEM_ERROR"
	ErrNetworkError      ErrorCode = "NETWORK_ERROR"
	ErrResourceExhausted ErrorCode = "RESOURCE_EXHAUSTED"
)

// AnalysisError represents an error that occurred during analysis
type AnalysisError struct {
	Code    ErrorCode
	Message string
	Cause   error
	Context map[string]interface{}
}

// Error implements the error interface
func (e *AnalysisError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause error
func (e *AnalysisError) Unwrap() error {
	return e.Cause
}

// WithContext adds context to the error
func (e *AnalysisError) WithContext(key string, value interface{}) *AnalysisError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// New creates a new AnalysisError
func New(code ErrorCode, message string) *AnalysisError {
	return &AnalysisError{
		Code:    code,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(cause error, code ErrorCode, message string) *AnalysisError {
	return &AnalysisError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// Is checks if the error matches a specific error code
// It traverses the entire error chain to find a match
func Is(err error, code ErrorCode) bool {
	if err == nil {
		return false
	}

	// Check if this error is an AnalysisError with the matching code
	if ae, ok := err.(*AnalysisError); ok {
		if ae.Code == code {
			return true
		}
		// If not, check the cause recursively
		if ae.Cause != nil {
			return Is(ae.Cause, code)
		}
		return false
	}

	// Check if it supports the Unwrap interface (e.g., standard errors)
	if unwrapper, ok := err.(interface{ Unwrap() error }); ok {
		return Is(unwrapper.Unwrap(), code)
	}

	return false
}

// Helper functions for common error scenarios

// NewFileNotFound creates a FILE_NOT_FOUND error
func NewFileNotFound(path string) *AnalysisError {
	return New(ErrFileNotFound, fmt.Sprintf("file not found: %s", path)).
		WithContext("path", path)
}

// NewFileAccess creates a FILE_ACCESS error
func NewFileAccess(path string, cause error) *AnalysisError {
	return Wrap(cause, ErrFileAccess, fmt.Sprintf("cannot access file: %s", path)).
		WithContext("path", path)
}

// NewInvalidArchive creates an INVALID_ARCHIVE error
func NewInvalidArchive(path string, cause error) *AnalysisError {
	return Wrap(cause, ErrInvalidArchive, fmt.Sprintf("invalid or corrupted archive: %s", path)).
		WithContext("path", path)
}

// NewPathTraversal creates a PATH_TRAVERSAL error
func NewPathTraversal(path string) *AnalysisError {
	return New(ErrPathTraversal, fmt.Sprintf("path traversal attempt detected: %s", path)).
		WithContext("path", path)
}

// NewConfigNotFound creates a CONFIG_NOT_FOUND error
func NewConfigNotFound(path string) *AnalysisError {
	return New(ErrConfigNotFound, fmt.Sprintf("configuration file not found: %s", path)).
		WithContext("path", path)
}

// NewConfigInvalid creates a CONFIG_INVALID error
func NewConfigInvalid(message string) *AnalysisError {
	return New(ErrConfigInvalid, fmt.Sprintf("invalid configuration: %s", message))
}

// NewPermissionDenied creates a PERMISSION_DENIED error
func NewPermissionDenied(resource string) *AnalysisError {
	return New(ErrPermissionDenied, fmt.Sprintf("permission denied: %s", resource)).
		WithContext("resource", resource)
}

// NewTimeout creates a TIMEOUT error
func NewTimeout(operation string) *AnalysisError {
	return New(ErrTimeout, fmt.Sprintf("operation timed out: %s", operation)).
		WithContext("operation", operation)
}

// NewValidationError creates a VALIDATION_FAILED error
func NewValidationError(message string) *AnalysisError {
	return New(ErrValidationFailed, fmt.Sprintf("validation failed: %s", message))
}

// NewParsingError creates a PARSING_FAILED error
func NewParsingError(context string, cause error) *AnalysisError {
	return Wrap(cause, ErrParsingFailed, fmt.Sprintf("parsing failed: %s", context)).
		WithContext("context", context)
}

// NewSystemError creates a SYSTEM_ERROR error
func NewSystemError(message string, cause error) *AnalysisError {
	return Wrap(cause, ErrSystemError, message)
}
