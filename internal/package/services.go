package pkg

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/amoga-io/run/internal/logger"
	"github.com/amoga-io/run/internal/system"
)

// PackageInstaller handles package installation operations
type PackageInstaller struct {
	manager *Manager
}

// NewPackageInstaller creates a new package installer
func NewPackageInstaller(manager *Manager) *PackageInstaller {
	return &PackageInstaller{manager: manager}
}

// InstallPackage installs a package with full rollback support
func (pi *PackageInstaller) InstallPackage(packageName string) error {
	return pi.manager.InstallPackage(packageName)
}

// InstallPackageWithArgs installs a package with arguments and rollback support
func (pi *PackageInstaller) InstallPackageWithArgs(packageName string, extraArgs []string) error {
	return pi.manager.InstallPackageWithArgs(packageName, extraArgs)
}

// PackageRemover handles package removal operations
type PackageRemover struct {
	manager *Manager
}

// NewPackageRemover creates a new package remover
func NewPackageRemover(manager *Manager) *PackageRemover {
	return &PackageRemover{manager: manager}
}

// RemovePackage removes a package safely
func (pr *PackageRemover) RemovePackage(packageName string) error {
	return pr.manager.RemovePackage(packageName)
}

// DependencyManager handles dependency operations
type DependencyManager struct {
	manager *Manager
}

// NewDependencyManager creates a new dependency manager
func NewDependencyManager(manager *Manager) *DependencyManager {
	return &DependencyManager{manager: manager}
}

// InstallDependencies installs package dependencies
func (dm *DependencyManager) InstallDependencies(pkg Package) error {
	return dm.manager.installDependencies(pkg)
}

// CheckDependencies checks if dependencies are satisfied
func (dm *DependencyManager) CheckDependencies(pkg Package) ([]string, error) {
	var missingPackages []string

	for _, dep := range pkg.Dependencies {
		// Check if dependency is a package in our registry
		if depPkg, exists := GetPackage(dep); exists {
			if !dm.manager.IsPackageInstalled(depPkg) {
				missingPackages = append(missingPackages, dep)
			}
		} else {
			// Check if it's a system command/package
			if !system.CommandExists(dep) {
				missingPackages = append(missingPackages, dep)
			}
		}
	}

	return missingPackages, nil
}

// ScriptExecutor handles script execution operations
type ScriptExecutor struct {
	repoPath string
}

// NewScriptExecutor creates a new script executor
func NewScriptExecutor(repoPath string) *ScriptExecutor {
	return &ScriptExecutor{repoPath: repoPath}
}

// ExecuteScript executes an installation script
func (se *ScriptExecutor) ExecuteScript(pkg Package) error {
	log := logger.GetLogger().WithOperation("execute_script").WithPackage(pkg.Name)

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

	resolvedScriptPath, err := safePath.Resolve(se.repoPath)
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

	// Execute script
	cmd := exec.Command(resolvedScriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = se.repoPath

	if err := cmd.Run(); err != nil {
		log.Error("Installation script failed %s: %v", resolvedScriptPath, err)
		return fmt.Errorf("installation script failed: %w", err)
	}

	log.Info("Installation script completed successfully")
	fmt.Printf("✓ %s installed successfully\n", pkg.Name)
	return nil
}

// ExecuteScriptWithArgs executes an installation script with arguments
func (se *ScriptExecutor) ExecuteScriptWithArgs(pkg Package, extraArgs []string) error {
	log := logger.GetLogger().WithOperation("execute_script_with_args").WithPackage(pkg.Name)

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

	resolvedScriptPath, err := safePath.Resolve(se.repoPath)
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
	cmd.Dir = se.repoPath

	if err := cmd.Run(); err != nil {
		log.Error("Installation script failed %s %v: %v", resolvedScriptPath, extraArgs, err)
		return fmt.Errorf("installation script failed: %w", err)
	}

	log.Info("Installation script completed successfully")
	fmt.Printf("✓ %s installed successfully\n", pkg.Name)
	return nil
}

// PackageChecker handles package checking operations
type PackageChecker struct {
	manager *Manager
}

// NewPackageChecker creates a new package checker
func NewPackageChecker(manager *Manager) *PackageChecker {
	return &PackageChecker{manager: manager}
}

// IsPackageInstalled checks if a package is installed
func (pc *PackageChecker) IsPackageInstalled(pkg Package) bool {
	return pc.manager.IsPackageInstalled(pkg)
}

// GetSystemVersion gets the system version of a package
func (pc *PackageChecker) GetSystemVersion(packageName string) string {
	return pc.manager.GetSystemVersion(packageName)
}

// SuggestionProvider handles package suggestions
type SuggestionProvider struct {
	manager *Manager
}

// NewSuggestionProvider creates a new suggestion provider
func NewSuggestionProvider(manager *Manager) *SuggestionProvider {
	return &SuggestionProvider{manager: manager}
}

// ProvideSuggestions provides smart suggestions for package installation
func (sp *SuggestionProvider) ProvideSuggestions(pkg Package) {
	sp.manager.provideSuggestions(pkg)
}
