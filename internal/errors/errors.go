package errors

import (
	"fmt"
	"strings"
)

// Common error types
var (
	ErrInvalidPackageName  = fmt.Errorf("invalid package name")
	ErrPackageNotFound     = fmt.Errorf("package not found")
	ErrVersionNotSupported = fmt.Errorf("version not supported")
	ErrInstallationFailed  = fmt.Errorf("installation failed")
	ErrRemovalFailed       = fmt.Errorf("removal failed")
	ErrValidationFailed    = fmt.Errorf("validation failed")
	ErrDependencyFailed    = fmt.Errorf("dependency failed")
)

// PackageError represents a package-specific error
type PackageError struct {
	PackageName string
	Operation   string
	Message     string
	Err         error
}

func (e *PackageError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s %s: %s (%v)", e.Operation, e.PackageName, e.Message, e.Err)
	}
	return fmt.Sprintf("%s %s: %s", e.Operation, e.PackageName, e.Message)
}

func (e *PackageError) Unwrap() error {
	return e.Err
}

// NewPackageError creates a new package error
func NewPackageError(packageName, operation, message string, err error) *PackageError {
	return &PackageError{
		PackageName: packageName,
		Operation:   operation,
		Message:     message,
		Err:         err,
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Value   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Value != "" {
		return fmt.Sprintf("validation failed for %s '%s': %s", e.Field, e.Value, e.Message)
	}
	return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, value, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// VersionError represents a version-related error
type VersionError struct {
	PackageName string
	Version     string
	Message     string
}

func (e *VersionError) Error() string {
	if e.Version != "" {
		return fmt.Sprintf("version error for %s %s: %s", e.PackageName, e.Version, e.Message)
	}
	return fmt.Sprintf("version error for %s: %s", e.PackageName, e.Message)
}

// NewVersionError creates a new version error
func NewVersionError(packageName, version, message string) *VersionError {
	return &VersionError{
		PackageName: packageName,
		Version:     version,
		Message:     message,
	}
}

// DependencyError represents a dependency-related error
type DependencyError struct {
	PackageName string
	Dependency  string
	Message     string
	Err         error
}

func (e *DependencyError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("dependency error for %s (requires %s): %s (%v)", e.PackageName, e.Dependency, e.Message, e.Err)
	}
	return fmt.Sprintf("dependency error for %s (requires %s): %s", e.PackageName, e.Dependency, e.Message)
}

func (e *DependencyError) Unwrap() error {
	return e.Err
}

// NewDependencyError creates a new dependency error
func NewDependencyError(packageName, dependency, message string, err error) *DependencyError {
	return &DependencyError{
		PackageName: packageName,
		Dependency:  dependency,
		Message:     message,
		Err:         err,
	}
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// IsVersionError checks if an error is a version error
func IsVersionError(err error) bool {
	_, ok := err.(*VersionError)
	return ok
}

// IsPackageError checks if an error is a package error
func IsPackageError(err error) bool {
	_, ok := err.(*PackageError)
	return ok
}

// IsDependencyError checks if an error is a dependency error
func IsDependencyError(err error) bool {
	_, ok := err.(*DependencyError)
	return ok
}

// FormatError formats an error for user display
func FormatError(err error) string {
	if err == nil {
		return ""
	}

	// Handle specific error types
	switch e := err.(type) {
	case *PackageError:
		return e.Error()
	case *ValidationError:
		return e.Error()
	case *VersionError:
		return e.Error()
	case *DependencyError:
		return e.Error()
	default:
		// For generic errors, clean up common patterns
		msg := err.Error()

		// Remove trailing punctuation
		msg = strings.TrimRight(msg, ".")

		// Capitalize first letter
		if len(msg) > 0 {
			msg = strings.ToUpper(msg[:1]) + msg[1:]
		}

		return msg
	}
}

// WrapError wraps an error with additional context
func WrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// CombineErrors combines multiple errors into a single error
func CombineErrors(errors ...error) error {
	var nonNilErrors []error
	for _, err := range errors {
		if err != nil {
			nonNilErrors = append(nonNilErrors, err)
		}
	}

	if len(nonNilErrors) == 0 {
		return nil
	}

	if len(nonNilErrors) == 1 {
		return nonNilErrors[0]
	}

	var messages []string
	for _, err := range nonNilErrors {
		messages = append(messages, err.Error())
	}

	return fmt.Errorf("multiple errors occurred: %s", strings.Join(messages, "; "))
}
