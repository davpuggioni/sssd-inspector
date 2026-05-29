// Package errors tests for custom error types
package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestNewError(t *testing.T) {
	err := New(ErrFileNotFound, "test file not found")

	if err.Code != ErrFileNotFound {
		t.Errorf("Expected code %s, got %s", ErrFileNotFound, err.Code)
	}

	expectedMsg := "[FILE_NOT_FOUND] test file not found"
	if err.Error() != expectedMsg {
		t.Errorf("Expected message %q, got %q", expectedMsg, err.Error())
	}
}

func TestWrapError(t *testing.T) {
	cause := fmt.Errorf("underlying cause")
	err := Wrap(cause, ErrFileAccess, "cannot access file")

	if err.Code != ErrFileAccess {
		t.Errorf("Expected code %s, got %s", ErrFileAccess, err.Code)
	}

	if !Is(err, ErrFileAccess) {
		t.Errorf("Expected Is to match ErrFileAccess")
	}

	unwrapped := errors.Unwrap(err)
	if unwrapped != cause {
		t.Errorf("Expected unwrapped error to match cause")
	}
}

func TestIsError(t *testing.T) {
	err := New(ErrTimeout, "operation timed out")

	if !Is(err, ErrTimeout) {
		t.Errorf("Expected Is to match ErrTimeout")
	}

	if Is(err, ErrFileNotFound) {
		t.Errorf("Expected Is to NOT match ErrFileNotFound")
	}
}

func TestIsNilError(t *testing.T) {
	if Is(nil, ErrUnknown) {
		t.Errorf("Expected Is(nil, _) to return false")
	}
}

func TestIsWrappedError(t *testing.T) {
	cause := New(ErrPermissionDenied, "permission denied")
	err := Wrap(cause, ErrAnalysisFailed, "analysis failed")

	if !Is(err, ErrPermissionDenied) {
		t.Errorf("Expected Is to find ErrPermissionDenied in wrapped error")
	}

	if !Is(err, ErrAnalysisFailed) {
		t.Errorf("Expected Is to find ErrAnalysisFailed in wrapped error")
	}
}

func TestWithContext(t *testing.T) {
	err := New(ErrConfigNotFound, "config not found").
		WithContext("path", "/etc/config.yaml").
		WithContext("expected", true)

	if err.Context["path"] != "/etc/config.yaml" {
		t.Errorf("Expected path context, got %v", err.Context["path"])
	}

	if err.Context["expected"] != true {
		t.Errorf("Expected expected context to be true, got %v", err.Context["expected"])
	}
}

func TestErrorWithoutCause(t *testing.T) {
	err := New(ErrValidationFailed, "validation failed")
	expectedMsg := "[VALIDATION_FAILED] validation failed"

	if err.Error() != expectedMsg {
		t.Errorf("Expected %q, got %q", expectedMsg, err.Error())
	}
}

func TestErrorWithCause(t *testing.T) {
	cause := fmt.Errorf("underlying error")
	err := Wrap(cause, ErrSystemError, "system error")
	expectedMsg := "[SYSTEM_ERROR] system error: underlying error"

	if err.Error() != expectedMsg {
		t.Errorf("Expected %q, got %q", expectedMsg, err.Error())
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("NewFileNotFound", func(t *testing.T) {
		err := NewFileNotFound("/path/to/file")
		if err.Code != ErrFileNotFound {
			t.Errorf("Expected FILE_NOT_FOUND, got %s", err.Code)
		}
		if err.Context["path"] != "/path/to/file" {
			t.Errorf("Expected path in context")
		}
	})

	t.Run("NewFileAccess", func(t *testing.T) {
		cause := fmt.Errorf("permission denied")
		err := NewFileAccess("/path/to/file", cause)
		if err.Code != ErrFileAccess {
			t.Errorf("Expected FILE_ACCESS, got %s", err.Code)
		}
	})

	t.Run("NewInvalidArchive", func(t *testing.T) {
		cause := fmt.Errorf("corrupt file")
		err := NewInvalidArchive("/path/to/archive.txz", cause)
		if err.Code != ErrInvalidArchive {
			t.Errorf("Expected INVALID_ARCHIVE, got %s", err.Code)
		}
	})

	t.Run("NewPathTraversal", func(t *testing.T) {
		err := NewPathTraversal("../../etc/passwd")
		if err.Code != ErrPathTraversal {
			t.Errorf("Expected PATH_TRAVERSAL, got %s", err.Code)
		}
	})

	t.Run("NewConfigNotFound", func(t *testing.T) {
		err := NewConfigNotFound("/etc/config.yaml")
		if err.Code != ErrConfigNotFound {
			t.Errorf("Expected CONFIG_NOT_FOUND, got %s", err.Code)
		}
	})

	t.Run("NewConfigInvalid", func(t *testing.T) {
		err := NewConfigInvalid("missing required field")
		if err.Code != ErrConfigInvalid {
			t.Errorf("Expected CONFIG_INVALID, got %s", err.Code)
		}
	})

	t.Run("NewPermissionDenied", func(t *testing.T) {
		err := NewPermissionDenied("/etc/krb5.keytab")
		if err.Code != ErrPermissionDenied {
			t.Errorf("Expected PERMISSION_DENIED, got %s", err.Code)
		}
	})

	t.Run("NewTimeout", func(t *testing.T) {
		err := NewTimeout("kerberos authentication")
		if err.Code != ErrTimeout {
			t.Errorf("Expected TIMEOUT, got %s", err.Code)
		}
	})

	t.Run("NewValidationError", func(t *testing.T) {
		err := NewValidationError("invalid configuration syntax")
		if err.Code != ErrValidationFailed {
			t.Errorf("Expected VALIDATION_FAILED, got %s", err.Code)
		}
	})

	t.Run("NewParsingError", func(t *testing.T) {
		cause := fmt.Errorf("unexpected token")
		err := NewParsingError("sssd.conf", cause)
		if err.Code != ErrParsingFailed {
			t.Errorf("Expected PARSING_FAILED, got %s", err.Code)
		}
	})

	t.Run("NewSystemError", func(t *testing.T) {
		cause := fmt.Errorf("out of memory")
		err := NewSystemError("system failure", cause)
		if err.Code != ErrSystemError {
			t.Errorf("Expected SYSTEM_ERROR, got %s", err.Code)
		}
	})
}
