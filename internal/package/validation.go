package pkg

import (
	"regexp"
	"strings"

	"github.com/amoga-io/run/internal/errors"
)

// Package name validation regex - only allow alphanumeric, hyphens, and underscores
var packageNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// ValidatePackageName validates and sanitizes package names
func ValidatePackageName(name string) error {
	if name == "" {
		return errors.NewValidationError("package name cannot be empty", nil)
	}

	if len(name) > 50 {
		return errors.NewValidationError("package name too long", nil).WithDetail("max_length", 50).WithDetail("actual_length", len(name))
	}

	if !packageNameRegex.MatchString(name) {
		return errors.NewValidationError("package name contains invalid characters", nil).WithDetail("allowed_chars", "alphanumeric, hyphens, underscores")
	}

	// Check for reserved names
	reservedNames := map[string]bool{
		"list":     true,
		"help":     true,
		"version":  true,
		"update":   true,
		"check":    true,
		"remove":   true,
		"install":  true,
		"internal": true,
	}

	if reservedNames[name] {
		return errors.NewValidationError("package name is reserved", nil).WithDetail("reserved_name", name)
	}

	return nil
}

// ValidateVersionString validates version strings
func ValidateVersionString(version string) error {
	if version == "" {
		return errors.NewValidationError("version cannot be empty", nil)
	}

	// Allow common version patterns
	versionRegex := regexp.MustCompile(`^[0-9]+(\.[0-9]+)*$`)
	if !versionRegex.MatchString(version) {
		return errors.NewValidationError("invalid version format", nil).WithDetail("expected_format", "X.Y.Z")
	}

	return nil
}

// SanitizePackageList sanitizes a list of package names
func SanitizePackageList(packages []string) ([]string, error) {
	if len(packages) == 0 {
		return nil, errors.NewValidationError("no packages specified", nil)
	}

	var sanitized []string
	seen := make(map[string]bool)

	for _, pkg := range packages {
		pkg = strings.TrimSpace(pkg)
		if pkg == "" {
			continue
		}

		if err := ValidatePackageName(pkg); err != nil {
			return nil, errors.NewValidationError("invalid package name", err).WithDetail("package", pkg)
		}

		if seen[pkg] {
			return nil, errors.NewValidationError("duplicate package name", nil).WithDetail("package", pkg)
		}

		seen[pkg] = true
		sanitized = append(sanitized, pkg)
	}

	if len(sanitized) == 0 {
		return nil, errors.NewValidationError("no valid packages specified", nil)
	}

	return sanitized, nil
}
