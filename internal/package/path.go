package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SafePath represents a validated file path
type SafePath struct {
	Path string
}

// ValidatePath validates and sanitizes file paths to prevent path traversal attacks
func ValidatePath(path string) (*SafePath, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	// Clean the path to resolve any .. or . components
	cleanPath := filepath.Clean(path)

	// Check for path traversal attempts
	if strings.Contains(cleanPath, "..") {
		return nil, fmt.Errorf("path traversal not allowed: %s", path)
	}

	// Ensure path doesn't start with / (absolute path)
	if filepath.IsAbs(cleanPath) {
		return nil, fmt.Errorf("absolute paths not allowed: %s", path)
	}

	// Check for suspicious patterns
	suspiciousPatterns := []string{
		"~", "~user", "~root",
		"/etc", "/var", "/usr", "/bin", "/sbin",
		"*", "?", "[", "]",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(cleanPath, pattern) {
			return nil, fmt.Errorf("path contains suspicious pattern: %s", pattern)
		}
	}

	return &SafePath{Path: cleanPath}, nil
}

// Join safely joins path components
func (sp *SafePath) Join(components ...string) (*SafePath, error) {
	if len(components) == 0 {
		return sp, nil
	}

	// Validate each component
	for _, component := range components {
		if strings.Contains(component, "..") || strings.Contains(component, "/") {
			return nil, fmt.Errorf("invalid path component: %s", component)
		}
	}

	joinedPath := filepath.Join(sp.Path, filepath.Join(components...))
	return ValidatePath(joinedPath)
}

// Resolve resolves the path relative to a base directory
func (sp *SafePath) Resolve(baseDir string) (string, error) {
	// Validate base directory
	if !filepath.IsAbs(baseDir) {
		return "", fmt.Errorf("base directory must be absolute: %s", baseDir)
	}

	// Ensure base directory exists
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return "", fmt.Errorf("base directory does not exist: %s", baseDir)
	}

	// Join and validate final path
	finalPath := filepath.Join(baseDir, sp.Path)

	// Ensure the final path is within the base directory
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base directory: %w", err)
	}

	absFinal, err := filepath.Abs(finalPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve final path: %w", err)
	}

	// Check if final path is within base directory
	if !strings.HasPrefix(absFinal, absBase) {
		return "", fmt.Errorf("path traversal detected: %s", finalPath)
	}

	return finalPath, nil
}

// GetRepoPath safely gets the repository path
func GetRepoPath() (*SafePath, error) {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return nil, fmt.Errorf("HOME environment variable is not set")
	}

	// Validate home directory
	if !filepath.IsAbs(homeDir) {
		return nil, fmt.Errorf("HOME directory is not absolute: %s", homeDir)
	}

	// Create safe path for .run directory
	return ValidatePath(".run")
}
