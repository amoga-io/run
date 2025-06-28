package errors

import (
	"fmt"
	"strings"
)

// ErrorType represents the type of error
type ErrorType string

const (
	ValidationError   ErrorType = "validation"
	NotFoundError     ErrorType = "not_found"
	PermissionError   ErrorType = "permission"
	NetworkError      ErrorType = "network"
	InstallationError ErrorType = "installation"
	RemovalError      ErrorType = "removal"
	UpdateError       ErrorType = "update"
	SystemError       ErrorType = "system"
)

// AppError represents a standardized application error
type AppError struct {
	Type    ErrorType
	Message string
	Details map[string]interface{}
	Err     error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithDetail adds a detail to the error
func (e *AppError) WithDetail(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// NewValidationError creates a new validation error
func NewValidationError(message string, err error) *AppError {
	return &AppError{
		Type:    ValidationError,
		Message: message,
		Err:     err,
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource string, name string) *AppError {
	return &AppError{
		Type:    NotFoundError,
		Message: fmt.Sprintf("%s '%s' not found", resource, name),
		Details: map[string]interface{}{
			"resource": resource,
			"name":     name,
		},
	}
}

// NewPermissionError creates a new permission error
func NewPermissionError(operation string, resource string) *AppError {
	return &AppError{
		Type:    PermissionError,
		Message: fmt.Sprintf("insufficient permissions to %s %s", operation, resource),
		Details: map[string]interface{}{
			"operation": operation,
			"resource":  resource,
		},
	}
}

// NewNetworkError creates a new network error
func NewNetworkError(operation string, url string, err error) *AppError {
	return &AppError{
		Type:    NetworkError,
		Message: fmt.Sprintf("network error during %s", operation),
		Details: map[string]interface{}{
			"operation": operation,
			"url":       url,
		},
		Err: err,
	}
}

// NewInstallationError creates a new installation error
func NewInstallationError(packageName string, err error) *AppError {
	return &AppError{
		Type:    InstallationError,
		Message: fmt.Sprintf("failed to install package '%s'", packageName),
		Details: map[string]interface{}{
			"package": packageName,
		},
		Err: err,
	}
}

// NewRemovalError creates a new removal error
func NewRemovalError(packageName string, err error) *AppError {
	return &AppError{
		Type:    RemovalError,
		Message: fmt.Sprintf("failed to remove package '%s'", packageName),
		Details: map[string]interface{}{
			"package": packageName,
		},
		Err: err,
	}
}

// NewUpdateError creates a new update error
func NewUpdateError(component string, err error) *AppError {
	return &AppError{
		Type:    UpdateError,
		Message: fmt.Sprintf("failed to update %s", component),
		Details: map[string]interface{}{
			"component": component,
		},
		Err: err,
	}
}

// NewSystemError creates a new system error
func NewSystemError(operation string, err error) *AppError {
	return &AppError{
		Type:    SystemError,
		Message: fmt.Sprintf("system error during %s", operation),
		Details: map[string]interface{}{
			"operation": operation,
		},
		Err: err,
	}
}

// FormatError formats an error for user display
func FormatError(err error) string {
	if appErr, ok := err.(*AppError); ok {
		// Format AppError with details
		message := appErr.Message
		if len(appErr.Details) > 0 {
			var details []string
			for key, value := range appErr.Details {
				details = append(details, fmt.Sprintf("%s=%v", key, value))
			}
			message += fmt.Sprintf(" (%s)", strings.Join(details, ", "))
		}
		return message
	}

	// Format regular errors
	return err.Error()
}

// IsErrorType checks if an error is of a specific type
func IsErrorType(err error, errorType ErrorType) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == errorType
	}
	return false
}

// GetErrorDetails returns the details of an AppError
func GetErrorDetails(err error) map[string]interface{} {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Details
	}
	return nil
}
