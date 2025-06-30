package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/amoga-io/run/internal/logger"
	"github.com/amoga-io/run/internal/system"
)

type Manager struct {
	repoPath string
}

// PackageAlreadyInstalledError represents when a package is already installed with the same version
type PackageAlreadyInstalledError struct {
	PackageName string
	Version     string
}

func (e *PackageAlreadyInstalledError) Error() string {
	if e.Version != "" {
		return fmt.Sprintf("package %s already installed with version %s", e.PackageName, e.Version)
	}
	return fmt.Sprintf("package %s already installed", e.PackageName)
}

// IsPackageAlreadyInstalledError checks if an error is a PackageAlreadyInstalledError
func IsPackageAlreadyInstalledError(err error) bool {
	_, ok := err.(*PackageAlreadyInstalledError)
	return ok
}

// NewTestManager creates a manager for testing purposes
// This bypasses the repository requirement for unit tests
func NewTestManager() *Manager {
	return &Manager{
		repoPath: "/tmp/test-repo", // Use a temporary path for tests
	}
}

// NewManager creates a new manager instance
func NewManager() (*Manager, error) {
	repoPath, err := GetRepoPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository path: %w", err)
	}

	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return nil, fmt.Errorf("HOME environment variable is not set")
	}

	resolvedPath, err := repoPath.Resolve(homeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve repository path: %w", err)
	}

	// Only check for repository existence in non-test environments
	if os.Getenv("TESTING") != "true" {
		if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("repository not found at %s. Please reinstall CLI", resolvedPath)
		}
	}

	return &Manager{repoPath: resolvedPath}, nil
}

// InstallPackage installs a package
func (m *Manager) InstallPackage(packageName string) error {
	return m.InstallPackageWithVersion(packageName, "")
}

// InstallPackageWithVersion installs a package with a specific version
func (m *Manager) InstallPackageWithVersion(packageName string, targetVersion string) error {
	pkg, err := m.validatePackage(packageName)
	if err != nil {
		return err
	}

	rollbackPoint, err := m.setupRollback(packageName)
	if err != nil {
		return err
	}
	defer m.cleanupRollback(rollbackPoint, err)

	if err := m.installDependencies(pkg); err != nil {
		return m.handleDependencyError(err, rollbackPoint)
	}

	// Check if package is already installed with same version
	skipInstallation, err := m.handleExistingInstallation(pkg, targetVersion)
	if err != nil {
		return err
	}
	if skipInstallation {
		return nil // Package already installed with same version
	}

	return m.executeInstallation(pkg, rollbackPoint)
}

// validatePackage validates and retrieves package information
func (m *Manager) validatePackage(packageName string) (Package, error) {
	pkg, exists := GetPackage(packageName)
	if !exists {
		return Package{}, fmt.Errorf("package '%s' not found", packageName)
	}

	fmt.Printf("Installing %s (%s)...\n", pkg.Name, pkg.Description)
	return pkg, nil
}

// setupRollback creates rollback manager and rollback point
func (m *Manager) setupRollback(packageName string) (*RollbackPoint, error) {
	rollbackManager, err := NewRollbackManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create rollback manager: %w", err)
	}

	rollbackPoint, err := rollbackManager.CreateRollbackPoint(packageName, "install")
	if err != nil {
		return nil, fmt.Errorf("failed to create rollback point: %w", err)
	}

	return rollbackPoint, nil
}

// cleanupRollback cleans up rollback point on success
func (m *Manager) cleanupRollback(rollbackPoint *RollbackPoint, err error) {
	if err == nil && rollbackPoint != nil {
		rollbackManager, _ := NewRollbackManager()
		if rollbackManager != nil {
			rollbackManager.CleanupRollbackPoint(rollbackPoint.ID)
		}
	}
}

// handleDependencyError handles dependency installation errors with rollback
func (m *Manager) handleDependencyError(err error, rollbackPoint *RollbackPoint) error {
	if rollbackPoint != nil {
		rollbackPoint.ExecuteRollback()
	}
	return fmt.Errorf("failed to install dependencies: %w", err)
}

// checkPackageVersion checks if a package is already installed with the same version
func (m *Manager) checkPackageVersion(pkg Package, targetVersion string) (bool, string, error) {
	// If no version specified, just check if package is installed
	if targetVersion == "" {
		if m.IsPackageInstalled(pkg) {
			return true, "already installed", nil
		}
		return false, "", nil
	}

	// Check if package supports version selection
	if !pkg.VersionSupport {
		// For packages without version support, just check if installed
		if m.IsPackageInstalled(pkg) {
			return true, "already installed", nil
		}
		return false, "", nil
	}

	// Get current system version
	currentVersion := m.GetSystemVersion(pkg.Name)
	if currentVersion == "" {
		// Package not installed
		return false, "", nil
	}

	// Compare versions
	if currentVersion == targetVersion {
		return true, fmt.Sprintf("already installed (version %s)", currentVersion), nil
	}

	// Different version installed
	return false, fmt.Sprintf("different version installed (current: %s, target: %s)", currentVersion, targetVersion), nil
}

// handleExistingInstallation handles existing package installation with version checking
func (m *Manager) handleExistingInstallation(pkg Package, targetVersion string) (bool, error) {
	isSameVersion, message, err := m.checkPackageVersion(pkg, targetVersion)
	if err != nil {
		return false, err
	}

	if isSameVersion {
		fmt.Printf("✓ %s %s - skipping installation\n", pkg.Name, message)
		return true, &PackageAlreadyInstalledError{
			PackageName: pkg.Name,
			Version:     targetVersion,
		}
	}

	if m.IsPackageInstalled(pkg) {
		// For python, require manual removal for safety
		if pkg.Name == "python" {
			fmt.Printf("Package %s is already installed with a different version. Please remove it first using the CLI remove command.\n", pkg.Name)
			return false, fmt.Errorf("package %s is already installed with a different version", pkg.Name)
		}
		// For all other packages, auto-remove existing version (including APT)
		fmt.Printf("Removing existing %s before installing new version...\n", pkg.Name)
		_, err := m.SafeRemovePackage(pkg.Name, true, false)
		if err != nil {
			return false, fmt.Errorf("failed to remove existing %s: %w", pkg.Name, err)
		}
	}

	return false, nil // Proceed with installation
}

// executeInstallation executes the actual installation
func (m *Manager) executeInstallation(pkg Package, rollbackPoint *RollbackPoint) error {
	if err := m.executeInstallScript(pkg); err != nil {
		if rollbackPoint != nil {
			rollbackPoint.ExecuteRollback()
		}
		return fmt.Errorf("installation failed: %w", err)
	}
	return nil
}

// InstallPackageWithArgs installs a package and passes extra arguments to the install script
func (m *Manager) InstallPackageWithArgs(packageName string, extraArgs []string) error {
	pkg, exists := GetPackage(packageName)
	if !exists {
		return fmt.Errorf("package '%s' not found", packageName)
	}

	fmt.Printf("Installing %s (%s)...\n", pkg.Name, pkg.Description)

	// Extract version from extraArgs if present
	targetVersion := ""
	if len(extraArgs) >= 2 && extraArgs[0] == "--version" {
		targetVersion = extraArgs[1]
	}

	// Create rollback manager
	rollbackManager, err := NewRollbackManager()
	if err != nil {
		return fmt.Errorf("failed to create rollback manager: %w", err)
	}

	// Create rollback point
	rollbackPoint, err := rollbackManager.CreateRollbackPoint(packageName, "install")
	if err != nil {
		return fmt.Errorf("failed to create rollback point: %w", err)
	}

	// Defer rollback cleanup on success
	defer func() {
		if err == nil {
			rollbackManager.CleanupRollbackPoint(rollbackPoint.ID)
		}
	}()

	// Validate dependencies before installation
	if err := ValidateDependencies(); err != nil {
		return fmt.Errorf("dependency validation failed: %w", err)
	}

	// Smart suggestions before installation
	m.provideSuggestions(pkg)

	// Step 1: Check and install dependencies
	if err := m.installDependencies(pkg); err != nil {
		// Execute rollback on dependency failure
		rollbackPoint.ExecuteRollback()
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	// Step 2: Check if package is already installed with same version
	skipInstallation, err := m.handleExistingInstallation(pkg, targetVersion)
	if err != nil {
		return err
	}
	if skipInstallation {
		return nil // Package already installed with same version
	}

	// Step 3: Execute installation script with extra arguments
	if err := m.executeInstallScriptWithArgs(pkg, extraArgs); err != nil {
		// Execute rollback on installation failure
		rollbackPoint.ExecuteRollback()
		return fmt.Errorf("installation failed: %w", err)
	}

	return nil
}

// provideSuggestions provides smart suggestions based on package being installed
func (m *Manager) provideSuggestions(pkg Package) {
	essentialsPkg, essentialsExists := GetPackage("essentials")
	if !essentialsExists {
		return
	}

	isEssentialsInstalled := m.IsPackageInstalled(essentialsPkg)

	// Suggest essentials for development packages
	if pkg.Category == "development" && !isEssentialsInstalled {
		fmt.Printf("💡 Tip: Installing 'essentials' first provides build tools helpful for %s\n", pkg.Name)
		fmt.Printf("💡 Run: run install essentials %s\n\n", pkg.Name)
	}

	// Suggest essentials for packages that commonly need build tools
	buildIntensivePackages := map[string]bool{
		"python": true,
		"node":   true,
		"php":    true,
	}

	if buildIntensivePackages[pkg.Name] && !isEssentialsInstalled {
		fmt.Printf("💡 Recommended: '%s' benefits from development tools in 'essentials'\n", pkg.Name)
		fmt.Printf("💡 Consider: run install essentials %s\n\n", pkg.Name)
	}

	// Suggest related packages
	relatedSuggestions := map[string][]string{
		"nginx":    {"php", "node"},
		"postgres": {"python", "node", "java"},
		"docker":   {"node", "python"},
		"node":     {"pm2"},
	}

	if suggestions, exists := relatedSuggestions[pkg.Name]; exists {
		var availableSuggestions []string
		for _, suggestion := range suggestions {
			if _, exists := GetPackage(suggestion); exists {
				if !m.IsPackageInstalled(Package{Name: suggestion}) {
					availableSuggestions = append(availableSuggestions, suggestion)
				}
			}
		}
		if len(availableSuggestions) > 0 {
			fmt.Printf("💡 Commonly used with %s: %s\n", pkg.Name, strings.Join(availableSuggestions, ", "))
			fmt.Printf("💡 Install together: run install %s %s\n\n", pkg.Name, strings.Join(availableSuggestions, " "))
		}
	}
}

// installDependencies checks and installs required dependencies
func (m *Manager) installDependencies(pkg Package) error {
	if len(pkg.Dependencies) == 0 {
		fmt.Printf("No dependencies required for %s\n", pkg.Name)
		return nil
	}

	fmt.Printf("Checking dependencies for %s: %s\n", pkg.Name, strings.Join(pkg.Dependencies, ", "))

	var missingPackages []string

	for _, dep := range pkg.Dependencies {
		// Check if dependency is a package in our registry
		if depPkg, exists := GetPackage(dep); exists {
			if !m.IsPackageInstalled(depPkg) {
				fmt.Printf("Required package %s is not installed\n", dep)
				// Recursively install package dependencies
				if err := m.InstallPackage(dep); err != nil {
					return fmt.Errorf("failed to install required package %s: %w", dep, err)
				}
			}
		} else {
			// Check if it's a system command/package
			if !system.CommandExists(dep) {
				missingPackages = append(missingPackages, dep)
			}
		}
	}

	// Install missing system packages
	if len(missingPackages) > 0 {
		fmt.Printf("Installing system dependencies: %s\n", strings.Join(missingPackages, ", "))
		if err := system.InstallSystemPackages(missingPackages); err != nil {
			return fmt.Errorf("failed to install system dependencies: %w", err)
		}
	}

	fmt.Printf("✓ All dependencies satisfied for %s\n", pkg.Name)
	return nil
}

// IsPackageInstalled checks if a package is installed by checking its commands and services
func (m *Manager) IsPackageInstalled(pkg Package) bool {
	// Check commands first
	for _, cmd := range pkg.Commands {
		if !system.CommandExists(cmd) {
			return false
		}
	}

	// Check service status for relevant packages
	switch pkg.Name {
	case "docker":
		return m.checkDockerService()
	case "nginx":
		return m.checkNginxService()
	case "postgres":
		return m.checkPostgresService()
	case "php":
		return m.checkPHPService()
	}

	return len(pkg.Commands) > 0
}

// checkDockerService checks if Docker service is running
func (m *Manager) checkDockerService() bool {
	// Check if docker command exists
	if !system.CommandExists("docker") {
		return false
	}

	// Check if Docker service is active
	cmd := exec.Command("systemctl", "is-active", "docker")
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

// checkNginxService checks if Nginx service is running
func (m *Manager) checkNginxService() bool {
	// Check if nginx command exists
	if !system.CommandExists("nginx") {
		return false
	}

	// Check if Nginx service is active
	cmd := exec.Command("systemctl", "is-active", "nginx")
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

// checkPostgresService checks if PostgreSQL service is running
func (m *Manager) checkPostgresService() bool {
	// Check if psql command exists
	if !system.CommandExists("psql") {
		return false
	}

	// Check if PostgreSQL service is active
	cmd := exec.Command("systemctl", "is-active", "postgresql")
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

// checkPHPService checks if PHP-FPM service is running
func (m *Manager) checkPHPService() bool {
	// Check if php command exists
	if !system.CommandExists("php") {
		return false
	}

	// Check if PHP-FPM service is active (try different versions)
	phpVersions := []string{"php8.3-fpm", "php8.2-fpm", "php8.1-fpm", "php-fpm"}
	for _, version := range phpVersions {
		cmd := exec.Command("systemctl", "is-active", version)
		if err := cmd.Run(); err == nil {
			return true
		}
	}

	// If no FPM service is running, but CLI works, consider it installed
	return true
}

// executeInstallScript executes the installation script for a package
func (m *Manager) executeInstallScript(pkg Package) error {
	log := logger.GetLogger().WithOperation("execute_install_script").WithPackage(pkg.Name)

	// Validate script path
	if err := ValidatePackageName(pkg.Name); err != nil {
		log.Error("Invalid package name: %v", err)
		return fmt.Errorf("invalid script path: %w", err)
	}

	// Resolve script path
	safePath, err := ValidatePath(pkg.ScriptPath)
	if err != nil {
		log.Error("Invalid script path %s: %v", pkg.ScriptPath, err)
		return fmt.Errorf("invalid script path: %w", err)
	}

	resolvedScriptPath, err := safePath.Resolve(m.repoPath)
	if err != nil {
		log.Error("Failed to resolve script path %s: %v", pkg.ScriptPath, err)
		return fmt.Errorf("failed to resolve script path: %w", err)
	}

	// Check if script exists
	if _, err := os.Stat(resolvedScriptPath); os.IsNotExist(err) {
		log.Error("Script file does not exist: %s", resolvedScriptPath)
		return fmt.Errorf("installation script not found: %s", resolvedScriptPath)
	}

	// Make script executable
	if err := os.Chmod(resolvedScriptPath, 0755); err != nil {
		log.Error("Failed to make script executable %s: %v", resolvedScriptPath, err)
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	log.Info("Executing installation script: %s", resolvedScriptPath)
	fmt.Printf("Executing installation script for %s...\n", pkg.Name)

	// Execute script with enhanced environment
	cmd := exec.Command(resolvedScriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = m.repoPath

	// Set enhanced environment variables for non-interactive operation
	cmd.Env = append(os.Environ(),
		"DEBIAN_FRONTEND=noninteractive",
		"APT_LISTCHANGES_FRONTEND=none",
		"NEEDRESTART_MODE=a",
	)

	if err := cmd.Run(); err != nil {
		log.Error("Installation script failed %s: %v", resolvedScriptPath, err)
		return fmt.Errorf("installation script failed: %w", err)
	}

	log.Info("Installation script completed successfully")
	fmt.Printf("✓ %s installed successfully\n", pkg.Name)
	return nil
}

// executeInstallScriptWithArgs executes the installation script with extra arguments
func (m *Manager) executeInstallScriptWithArgs(pkg Package, extraArgs []string) error {
	log := logger.GetLogger().WithOperation("execute_install_script_with_args").WithPackage(pkg.Name)

	// Validate script path
	if err := ValidatePackageName(pkg.Name); err != nil {
		log.Error("Invalid package name: %v", err)
		return fmt.Errorf("invalid script path: %w", err)
	}

	// Resolve script path
	safePath, err := ValidatePath(pkg.ScriptPath)
	if err != nil {
		log.Error("Invalid script path %s: %v", pkg.ScriptPath, err)
		return fmt.Errorf("invalid script path: %w", err)
	}

	resolvedScriptPath, err := safePath.Resolve(m.repoPath)
	if err != nil {
		log.Error("Failed to resolve script path %s: %v", pkg.ScriptPath, err)
		return fmt.Errorf("failed to resolve script path: %w", err)
	}

	// Check if script exists
	if _, err := os.Stat(resolvedScriptPath); os.IsNotExist(err) {
		log.Error("Script file does not exist: %s", resolvedScriptPath)
		return fmt.Errorf("installation script not found: %s", resolvedScriptPath)
	}

	// Make script executable
	if err := os.Chmod(resolvedScriptPath, 0755); err != nil {
		log.Error("Failed to make script executable %s: %v", resolvedScriptPath, err)
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	log.Info("Executing installation script with arguments: %s %v", resolvedScriptPath, extraArgs)
	fmt.Printf("Executing installation script for %s with arguments...\n", pkg.Name)

	// Execute script with arguments
	cmd := exec.Command(resolvedScriptPath, extraArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = m.repoPath

	if err := cmd.Run(); err != nil {
		log.Error("Installation script failed %s %v: %v", resolvedScriptPath, extraArgs, err)
		return fmt.Errorf("installation script failed: %w", err)
	}

	log.Info("Installation script completed successfully")
	fmt.Printf("✓ %s installed successfully\n", pkg.Name)
	return nil
}

// RemovePackage is deprecated. Use SafeRemovePackage instead.
func (m *Manager) RemovePackage(packageName string) error {
	return fmt.Errorf("RemovePackage is deprecated. Use SafeRemovePackage instead")
}

// GetSystemVersion gets the system-installed version of a package
func (m *Manager) GetSystemVersion(packageName string) string {
	// Try to get the version using the package's command (e.g., node --version, php -v, java -version, etc.)
	switch packageName {
	case "node":
		if out, err := exec.Command("node", "--version").Output(); err == nil {
			return strings.TrimSpace(string(out))
		}
	case "php":
		if out, err := exec.Command("php", "-v").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) > 0 {
				return lines[0]
			}
		}
	case "java":
		if out, err := exec.Command("java", "-version").CombinedOutput(); err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) > 0 {
				return lines[0]
			}
		}
	case "docker":
		if out, err := exec.Command("docker", "--version").Output(); err == nil {
			return strings.TrimSpace(string(out))
		}
	case "nginx":
		if out, err := exec.Command("nginx", "-v").CombinedOutput(); err == nil {
			return strings.TrimSpace(string(out))
		}
	case "postgres":
		if out, err := exec.Command("psql", "--version").Output(); err == nil {
			return strings.TrimSpace(string(out))
		}
	case "pm2":
		if out, err := exec.Command("pm2", "--version").Output(); err == nil {
			return strings.TrimSpace(string(out))
		}
	}
	return ""
}
