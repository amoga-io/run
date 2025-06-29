package pkg

import (
	"fmt"
	"sync"
)

type Package struct {
	Name              string
	Description       string
	ScriptPath        string
	Dependencies      []string // Required packages/commands before installation
	Commands          []string // Commands to check if package is installed
	Category          string
	VersionSupport    bool     // Whether this package supports version selection
	DefaultVersion    string   // Default version if not specified
	SupportedVersions []string // List of supported versions
	AptPackageName    string   // Name of the package in APT (if different from Name)
}

// Thread-safe package registry
type PackageRegistry struct {
	packages map[string]Package
	mutex    sync.RWMutex
}

// Global registry instance
var globalRegistry = &PackageRegistry{
	packages: map[string]Package{
		"python": {
			Name:              "python",
			Description:       "Python programming language with pip and venv",
			ScriptPath:        "scripts/packages/python.sh",
			Dependencies:      []string{"build-essential", "curl"},
			Commands:          []string{"python3", "pip3"},
			Category:          "development",
			VersionSupport:    true,
			DefaultVersion:    "3.10",
			SupportedVersions: []string{"3.8", "3.9", "3.10", "3.11", "3.12"},
		},
		"node": {
			Name:              "node",
			Description:       "Node.js runtime with npm",
			ScriptPath:        "scripts/packages/node.sh",
			Dependencies:      []string{"curl", "build-essential"},
			Commands:          []string{"node", "npm"},
			Category:          "development",
			VersionSupport:    true,
			DefaultVersion:    "18",
			SupportedVersions: []string{"16", "18", "20", "21"},
		},
		"docker": {
			Name:              "docker",
			Description:       "Docker containerization platform",
			ScriptPath:        "scripts/packages/docker.sh",
			Dependencies:      []string{"ca-certificates", "curl", "gnupg"},
			Commands:          []string{"docker"},
			Category:          "devops",
			VersionSupport:    false,
			DefaultVersion:    "",
			SupportedVersions: []string{},
		},
		"nginx": {
			Name:              "nginx",
			Description:       "High-performance web server",
			ScriptPath:        "scripts/packages/nginx.sh",
			Dependencies:      []string{"curl", "gnupg"},
			Commands:          []string{"nginx"},
			Category:          "web",
			VersionSupport:    true,
			DefaultVersion:    "stable",
			SupportedVersions: []string{"stable", "mainline"},
		},
		"postgres": {
			Name:              "postgres",
			Description:       "PostgreSQL 17 database server",
			ScriptPath:        "scripts/packages/postgres17.sh",
			Dependencies:      []string{"curl", "gnupg", "lsb-release"},
			Commands:          []string{"psql"},
			Category:          "database",
			VersionSupport:    true,
			DefaultVersion:    "17",
			SupportedVersions: []string{"15", "16", "17"},
			AptPackageName:    "postgresql",
		},
		"php": {
			Name:              "php",
			Description:       "PHP 8.3 programming language with FPM",
			ScriptPath:        "scripts/packages/php.sh",
			Dependencies:      []string{"software-properties-common"},
			Commands:          []string{"php"},
			Category:          "development",
			VersionSupport:    true,
			DefaultVersion:    "8.3",
			SupportedVersions: []string{"8.1", "8.2", "8.3"},
		},
		"java": {
			Name:              "java",
			Description:       "OpenJDK Java Development Kit 17",
			ScriptPath:        "scripts/packages/java.sh",
			Dependencies:      []string{},
			Commands:          []string{"java", "javac"},
			Category:          "development",
			VersionSupport:    true,
			DefaultVersion:    "17",
			SupportedVersions: []string{"11", "17", "21"},
		},
		"pm2": {
			Name:              "pm2",
			Description:       "Process manager for Node.js applications",
			ScriptPath:        "scripts/packages/pm2.sh",
			Dependencies:      []string{"node"}, // Requires Node.js to be installed first
			Commands:          []string{"pm2"},
			Category:          "development",
			VersionSupport:    true,
			DefaultVersion:    "latest",
			SupportedVersions: []string{"latest", "5.3.0", "5.4.0", "5.5.0"},
		},
		"essentials": {
			Name:              "essentials",
			Description:       "System essentials and development tools",
			ScriptPath:        "scripts/system/essentials.sh",
			Dependencies:      []string{},
			Commands:          []string{"gcc", "make", "redis-server"},
			Category:          "system",
			VersionSupport:    false,
			DefaultVersion:    "",
			SupportedVersions: []string{},
		},
	},
}

// GetPackage returns a package by name (thread-safe)
func GetPackage(name string) (Package, bool) {
	globalRegistry.mutex.RLock()
	defer globalRegistry.mutex.RUnlock()

	pkg, exists := globalRegistry.packages[name]
	return pkg, exists
}

// ListPackages returns all available packages (thread-safe)
func ListPackages() map[string]Package {
	globalRegistry.mutex.RLock()
	defer globalRegistry.mutex.RUnlock()

	// Create a copy to avoid external modification
	packages := make(map[string]Package)
	for name, pkg := range globalRegistry.packages {
		packages[name] = pkg
	}
	return packages
}

// GetPackagesByCategory returns packages filtered by category (thread-safe)
func GetPackagesByCategory(category string) []Package {
	globalRegistry.mutex.RLock()
	defer globalRegistry.mutex.RUnlock()

	var packages []Package
	for _, pkg := range globalRegistry.packages {
		if pkg.Category == category {
			packages = append(packages, pkg)
		}
	}
	return packages
}

// SupportsVersion checks if a package supports version selection (thread-safe)
func SupportsVersion(packageName string) bool {
	pkg, exists := GetPackage(packageName)
	return exists && pkg.VersionSupport
}

// ValidateVersion validates if a version is supported for a package (thread-safe)
func ValidateVersion(packageName, version string) error {
	pkg, exists := GetPackage(packageName)
	if !exists {
		return fmt.Errorf("package '%s' not found", packageName)
	}

	if !pkg.VersionSupport {
		return fmt.Errorf("package '%s' does not support version selection", packageName)
	}

	if version == "" {
		return nil // Use default version
	}

	// Check if version is supported
	for _, supportedVersion := range pkg.SupportedVersions {
		if supportedVersion == version {
			return nil
		}
	}

	return fmt.Errorf("version '%s' not supported for package '%s'. Supported versions: %v",
		version, packageName, pkg.SupportedVersions)
}

// GetDefaultVersion returns the default version for a package (thread-safe)
func GetDefaultVersion(packageName string) (string, error) {
	pkg, exists := GetPackage(packageName)
	if !exists {
		return "", fmt.Errorf("package '%s' not found", packageName)
	}

	if !pkg.VersionSupport {
		return "", fmt.Errorf("package '%s' does not support version selection", packageName)
	}

	return pkg.DefaultVersion, nil
}

// GetSupportedVersions returns the list of supported versions for a package (thread-safe)
func GetSupportedVersions(packageName string) ([]string, error) {
	pkg, exists := GetPackage(packageName)
	if !exists {
		return nil, fmt.Errorf("package '%s' not found", packageName)
	}

	if !pkg.VersionSupport {
		return nil, fmt.Errorf("package '%s' does not support version selection", packageName)
	}

	return pkg.SupportedVersions, nil
}

// AddPackage adds a new package to the registry (thread-safe)
func AddPackage(pkg Package) error {
	if err := ValidatePackageName(pkg.Name); err != nil {
		return fmt.Errorf("invalid package name: %w", err)
	}

	globalRegistry.mutex.Lock()
	defer globalRegistry.mutex.Unlock()

	globalRegistry.packages[pkg.Name] = pkg
	return nil
}

// RemovePackage removes a package from the registry (thread-safe)
func RemovePackage(name string) error {
	if err := ValidatePackageName(name); err != nil {
		return fmt.Errorf("invalid package name: %w", err)
	}

	globalRegistry.mutex.Lock()
	defer globalRegistry.mutex.Unlock()

	delete(globalRegistry.packages, name)
	return nil
}

// UpdatePackage updates an existing package in the registry (thread-safe)
func UpdatePackage(pkg Package) error {
	if err := ValidatePackageName(pkg.Name); err != nil {
		return fmt.Errorf("invalid package name: %w", err)
	}

	globalRegistry.mutex.Lock()
	defer globalRegistry.mutex.Unlock()

	if _, exists := globalRegistry.packages[pkg.Name]; !exists {
		return fmt.Errorf("package '%s' not found", pkg.Name)
	}

	globalRegistry.packages[pkg.Name] = pkg
	return nil
}
