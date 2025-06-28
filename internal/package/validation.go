package pkg

import (
	"fmt"
	"regexp"
	"strings"
)

// Package name validation regex - only allow alphanumeric, hyphens, and underscores
var packageNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// ValidatePackageName validates and sanitizes package names
func ValidatePackageName(name string) error {
	if name == "" {
		return fmt.Errorf("package name cannot be empty")
	}

	if len(name) > 50 {
		return fmt.Errorf("package name too long (max 50 characters)")
	}

	if !packageNameRegex.MatchString(name) {
		return fmt.Errorf("package name contains invalid characters (only alphanumeric, hyphens, and underscores allowed)")
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
		return fmt.Errorf("package name '%s' is reserved", name)
	}

	return nil
}

// ValidateVersionString validates version strings
func ValidateVersionString(version string) error {
	if version == "" {
		return fmt.Errorf("version cannot be empty")
	}

	// Allow common version patterns
	versionRegex := regexp.MustCompile(`^[0-9]+(\.[0-9]+)*$`)
	if !versionRegex.MatchString(version) {
		return fmt.Errorf("invalid version format (expected format: X.Y.Z)")
	}

	return nil
}

// SanitizePackageList sanitizes a list of package names
func SanitizePackageList(packages []string) ([]string, error) {
	if len(packages) == 0 {
		return nil, fmt.Errorf("no packages specified")
	}

	var sanitized []string
	seen := make(map[string]bool)

	for _, pkg := range packages {
		pkg = strings.TrimSpace(pkg)
		if pkg == "" {
			continue
		}

		if err := ValidatePackageName(pkg); err != nil {
			return nil, fmt.Errorf("invalid package name '%s': %w", pkg, err)
		}

		if seen[pkg] {
			return nil, fmt.Errorf("duplicate package name: %s", pkg)
		}

		seen[pkg] = true
		sanitized = append(sanitized, pkg)
	}

	if len(sanitized) == 0 {
		return nil, fmt.Errorf("no valid packages specified")
	}

	return sanitized, nil
}
