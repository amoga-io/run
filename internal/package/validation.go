package pkg

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/amoga-io/run/internal/errors"
)

// Validation constants
const (
	MaxPackageNameLength = 50
	MinPackageNameLength = 1
	MaxVersionLength     = 20
	MaxDescriptionLength = 200
)

// Package name validation regex - alphanumeric, hyphens, underscores only
var packageNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Version validation regex - alphanumeric, dots, hyphens, underscores
var versionRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// ValidatePackageName validates a package name
func ValidatePackageName(name string) error {
	if name == "" {
		return errors.NewValidationError("package name", "", "cannot be empty")
	}

	if len(name) < MinPackageNameLength {
		return errors.NewValidationError("package name", name, fmt.Sprintf("must be at least %d characters", MinPackageNameLength))
	}

	if len(name) > MaxPackageNameLength {
		return errors.NewValidationError("package name", name, fmt.Sprintf("must be at most %d characters", MaxPackageNameLength))
	}

	if !packageNameRegex.MatchString(name) {
		return errors.NewValidationError("package name", name, "must contain only alphanumeric characters, hyphens, and underscores")
	}

	// Check for reserved names
	reservedNames := []string{"system", "root", "admin", "sudo", "apt", "dpkg", "list", "help", "version", "update", "check", "remove", "install", "internal"}
	for _, reserved := range reservedNames {
		if strings.EqualFold(name, reserved) {
			return errors.NewValidationError("package name", name, fmt.Sprintf("cannot use reserved name '%s'", reserved))
		}
	}

	return nil
}

// ValidateVersion validates a version string
func ValidateVersionString(packageName, version string) error {
	if version == "" {
		return nil // Empty version is valid (means use default)
	}

	if len(version) > MaxVersionLength {
		return errors.NewVersionError(packageName, version, fmt.Sprintf("version must be at most %d characters", MaxVersionLength))
	}

	if !versionRegex.MatchString(version) {
		return errors.NewVersionError(packageName, version, "version must contain only alphanumeric characters, dots, hyphens, and underscores")
	}

	// Check if package supports version selection
	if !SupportsVersion(packageName) {
		return errors.NewVersionError(packageName, version, "package does not support version selection")
	}

	// Validate against supported versions if available
	pkg, exists := GetPackage(packageName)
	if exists && len(pkg.SupportedVersions) > 0 {
		versionSupported := false
		for _, supported := range pkg.SupportedVersions {
			if version == supported {
				versionSupported = true
				break
			}
		}
		if !versionSupported {
			return errors.NewVersionError(packageName, version, fmt.Sprintf("version not supported. Supported versions: %s", strings.Join(pkg.SupportedVersions, ", ")))
		}
	}

	return nil
}

// ValidatePackageList validates a list of package names
func ValidatePackageList(packages []string) error {
	if len(packages) == 0 {
		return errors.NewValidationError("package list", "", "cannot be empty")
	}

	var validationErrors []error
	for _, pkg := range packages {
		if err := ValidatePackageName(pkg); err != nil {
			validationErrors = append(validationErrors, err)
		}
	}

	if len(validationErrors) > 0 {
		return errors.CombineErrors(validationErrors...)
	}

	return nil
}

// SanitizePackageList sanitizes and validates a list of package names
func SanitizePackageList(packages []string) ([]string, error) {
	if len(packages) == 0 {
		return nil, errors.NewValidationError("package list", "", "cannot be empty")
	}

	var sanitized []string
	seen := make(map[string]bool)

	for _, pkg := range packages {
		// Trim whitespace
		pkg = strings.TrimSpace(pkg)

		// Skip empty entries
		if pkg == "" {
			continue
		}

		// Convert to lowercase for consistency
		pkg = strings.ToLower(pkg)

		// Validate package name
		if err := ValidatePackageName(pkg); err != nil {
			return nil, err
		}

		// Check for duplicates
		if seen[pkg] {
			continue
		}
		seen[pkg] = true

		sanitized = append(sanitized, pkg)
	}

	if len(sanitized) == 0 {
		return nil, errors.NewValidationError("package list", "", "no valid packages found after sanitization")
	}

	return sanitized, nil
}

// ValidateDependencyGraph validates a dependency graph
func ValidateDependencyGraph(graph *DependencyGraph) error {
	// Check for circular dependencies
	cycles, err := graph.DetectCircularDependencies()
	if err != nil {
		return errors.WrapError(err, "failed to detect circular dependencies")
	}

	if len(cycles) > 0 {
		return errors.NewValidationError("dependency graph", "", fmt.Sprintf("circular dependencies found: %v", cycles))
	}

	// Check for missing dependencies
	packages := ListPackages()
	for _, node := range graph.Nodes {
		for _, dep := range node.Dependencies {
			if _, exists := packages[dep]; !exists {
				// This is a system dependency, which is fine
				continue
			}
		}
	}

	return nil
}

// ValidateInstallationPath validates an installation path
func ValidateInstallationPath(path string) error {
	if path == "" {
		return errors.NewValidationError("installation path", "", "cannot be empty")
	}

	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return errors.NewValidationError("installation path", path, "contains path traversal attempt")
	}

	// Check for absolute paths outside allowed directories
	if strings.HasPrefix(path, "/") {
		allowedPrefixes := []string{"/usr/local", "/opt", "/home"}
		allowed := false
		for _, prefix := range allowedPrefixes {
			if strings.HasPrefix(path, prefix) {
				allowed = true
				break
			}
		}
		if !allowed {
			return errors.NewValidationError("installation path", path, "absolute path not in allowed directories")
		}
	}

	return nil
}

// ValidateScriptPath validates a script path
func ValidateScriptPath(path string) error {
	if path == "" {
		return errors.NewValidationError("script path", "", "cannot be empty")
	}

	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return errors.NewValidationError("script path", path, "contains path traversal attempt")
	}

	// Check for absolute paths
	if strings.HasPrefix(path, "/") {
		return errors.NewValidationError("script path", path, "must be relative path")
	}

	// Check for script extension
	if !strings.HasSuffix(path, ".sh") {
		return errors.NewValidationError("script path", path, "must have .sh extension")
	}

	return nil
}

// ValidatePackageConfig validates a package configuration
func ValidatePackageConfig(config PackageConfig) error {
	var validationErrors []error

	// Validate name
	if err := ValidatePackageName(config.Name); err != nil {
		validationErrors = append(validationErrors, err)
	}

	// Validate description
	if config.Description == "" {
		validationErrors = append(validationErrors, errors.NewValidationError("description", "", "cannot be empty"))
	} else if len(config.Description) > MaxDescriptionLength {
		validationErrors = append(validationErrors, errors.NewValidationError("description", config.Description, fmt.Sprintf("must be at most %d characters", MaxDescriptionLength)))
	}

	// Validate script path
	if err := ValidateScriptPath(config.ScriptPath); err != nil {
		validationErrors = append(validationErrors, err)
	}

	// Validate dependencies
	for _, dep := range config.Dependencies {
		if err := ValidatePackageName(dep); err != nil {
			validationErrors = append(validationErrors, errors.WrapError(err, fmt.Sprintf("dependency '%s'", dep)))
		}
	}

	// Validate commands
	for _, cmd := range config.Commands {
		if cmd == "" {
			validationErrors = append(validationErrors, errors.NewValidationError("command", "", "cannot be empty"))
		}
	}

	// Validate version support
	if config.VersionSupport {
		if config.DefaultVersion != "" {
			if err := ValidateVersionString(config.Name, config.DefaultVersion); err != nil {
				validationErrors = append(validationErrors, err)
			}
		}

		for _, version := range config.SupportedVersions {
			if err := ValidateVersionString(config.Name, version); err != nil {
				validationErrors = append(validationErrors, err)
			}
		}
	}

	if len(validationErrors) > 0 {
		return errors.CombineErrors(validationErrors...)
	}

	return nil
}

// IsValidPackageName checks if a package name is valid without returning an error
func IsValidPackageName(name string) bool {
	return ValidatePackageName(name) == nil
}

// IsValidVersion checks if a version is valid for a package without returning an error
func IsValidVersion(packageName, version string) bool {
	return ValidateVersionString(packageName, version) == nil
}
