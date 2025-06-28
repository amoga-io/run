package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/amoga-io/run/internal/logger"
)

// InstallationType represents how a package is installed
type InstallationType string

const (
	InstallTypeAPT          InstallationType = "apt"
	InstallTypeManual       InstallationType = "manual"
	InstallTypeUser         InstallationType = "user"
	InstallTypeVersion      InstallationType = "version_manager"
	InstallTypeAlternatives InstallationType = "alternatives"
	InstallTypeUnknown      InstallationType = "unknown"
)

// RemovalResult represents the result of a package removal operation
type RemovalResult struct {
	PackageName      string
	InstallationType InstallationType
	Success          bool
	Warning          string
	Error            error
	RemovedPaths     []string
}

// Critical packages that should never be removed
var criticalPackages = map[string]string{
	"systemd":        "System service manager - critical for system operation",
	"ubuntu-desktop": "Ubuntu desktop environment - critical for GUI",
	"glibc":          "GNU C Library - critical for system operation",
	"sudo":           "Superuser privileges - critical for system administration",
	"bash":           "Bourne Again Shell - critical for system operation",
	"coreutils":      "Core utilities - critical for system operation",
	"netplan":        "Network configuration - critical for networking",
	"gnome":          "GNOME desktop environment - critical for GUI",
	"xorg":           "X Window System - critical for GUI",
	"python3":        "System Python - critical for Ubuntu system tools",
	"nodejs":         "System Node.js - critical for system tools",
	"php":            "System PHP - critical for system tools",
	"apt":            "Package manager - critical for system operation",
	"dpkg":           "Package manager - critical for system operation",
	"systemctl":      "System control - critical for system operation",
	"init":           "System initialization - critical for system operation",
	"login":          "User login - critical for system operation",
	"passwd":         "Password management - critical for system operation",
	"useradd":        "User management - critical for system operation",
	"groupadd":       "Group management - critical for system operation",
}

// SafeRemovePackage safely removes a package with comprehensive safety checks
func (m *Manager) SafeRemovePackage(packageName string, force bool, dryRun bool) (*RemovalResult, error) {
	log := logger.GetLogger().WithOperation("safe_remove_package").WithPackage(packageName)

	result := &RemovalResult{
		PackageName: packageName,
		Success:     false,
	}

	// Check if package is critical
	if isCritical, reason := m.isCriticalPackage(packageName); isCritical && !force {
		result.Warning = fmt.Sprintf("Skipped removal of '%s' (system-critical: %s)", packageName, reason)
		log.Warn("Skipped critical package removal: %s - %s", packageName, reason)
		return result, nil
	}

	// Detect installation type
	installType, err := m.detectInstallationType(packageName)
	if err != nil {
		result.Error = fmt.Errorf("failed to detect installation type: %w", err)
		return result, result.Error
	}
	result.InstallationType = installType

	log.Info("Detected installation type: %s for package: %s", installType, packageName)

	// Perform removal based on installation type
	switch installType {
	case InstallTypeAPT:
		return m.removeAPT(packageName, dryRun)
	case InstallTypeManual:
		return m.removeManual(packageName, dryRun)
	case InstallTypeUser:
		return m.removeUser(packageName, dryRun)
	case InstallTypeVersion:
		return m.removeVersionManager(packageName, dryRun)
	case InstallTypeAlternatives:
		return m.removeAlternatives(packageName, dryRun)
	case InstallTypeUnknown:
		result.Warning = fmt.Sprintf("Package '%s' not found or not installed", packageName)
		return result, nil
	default:
		result.Error = fmt.Errorf("unknown installation type: %s", installType)
		return result, result.Error
	}
}

// isCriticalPackage checks if a package is critical for system operation
func (m *Manager) isCriticalPackage(packageName string) (bool, string) {
	// Check exact match
	if reason, exists := criticalPackages[packageName]; exists {
		return true, reason
	}

	// Check prefix matches for critical packages
	for critical, reason := range criticalPackages {
		if strings.HasPrefix(packageName, critical) {
			return true, reason
		}
	}

	return false, ""
}

// detectInstallationType detects how a package is installed
func (m *Manager) detectInstallationType(packageName string) (InstallationType, error) {
	// Check if installed via APT
	if m.isInstalledViaAPT(packageName) {
		return InstallTypeAPT, nil
	}

	// Check if installed manually in /usr/local
	if m.isInstalledManually(packageName) {
		return InstallTypeManual, nil
	}

	// Check if installed via version manager
	if m.isInstalledViaVersionManager(packageName) {
		return InstallTypeVersion, nil
	}

	// Check if installed in user directory
	if m.isInstalledInUserDir(packageName) {
		return InstallTypeUser, nil
	}

	// Check if installed via alternatives
	if m.isInstalledViaAlternatives(packageName) {
		return InstallTypeAlternatives, nil
	}

	return InstallTypeUnknown, nil
}

// isInstalledViaAPT checks if package is installed via APT
func (m *Manager) isInstalledViaAPT(packageName string) bool {
	// Check dpkg -l
	cmd := exec.Command("dpkg", "-l", packageName)
	if err := cmd.Run(); err == nil {
		return true
	}

	// Check apt-cache policy
	cmd = exec.Command("apt-cache", "policy", packageName)
	output, err := cmd.Output()
	if err == nil {
		outputStr := string(output)
		// If package is installed, it will show "Installed: <version>"
		if strings.Contains(outputStr, "Installed:") && !strings.Contains(outputStr, "(none)") {
			return true
		}
	}

	return false
}

// isInstalledManually checks if package is installed manually in /usr/local
func (m *Manager) isInstalledManually(packageName string) bool {
	paths := []string{
		filepath.Join("/usr/local/bin", packageName),
		filepath.Join("/usr/local/bin", packageName+"*"),
		filepath.Join("/usr/local/lib", packageName+"*"),
		filepath.Join("/usr/local/include", packageName+"*"),
	}

	for _, path := range paths {
		if matches, _ := filepath.Glob(path); len(matches) > 0 {
			return true
		}
	}

	return false
}

// isInstalledViaVersionManager checks if package is installed via version manager
func (m *Manager) isInstalledViaVersionManager(packageName string) bool {
	return IsPackageInstalledViaVersionManager(packageName)
}

// isInstalledInUserDir checks if package is installed in user directory
func (m *Manager) isInstalledInUserDir(packageName string) bool {
	home := os.Getenv("HOME")

	userPaths := []string{
		filepath.Join(home, ".local/bin", packageName),
		filepath.Join(home, ".local/lib", packageName+"*"),
		filepath.Join(home, ".config", packageName+"*"),
		filepath.Join(home, "."+packageName),
		filepath.Join(home, ".cache", packageName+"*"),
	}

	for _, path := range userPaths {
		if matches, _ := filepath.Glob(path); len(matches) > 0 {
			return true
		}
	}

	return false
}

// isInstalledViaAlternatives checks if package is installed via alternatives
func (m *Manager) isInstalledViaAlternatives(packageName string) bool {
	// Check if the package is managed by update-alternatives
	cmd := exec.Command("update-alternatives", "--list", packageName)
	if err := cmd.Run(); err == nil {
		return true
	}

	// Check if the package has alternatives configuration
	cmd = exec.Command("update-alternatives", "--display", packageName)
	if err := cmd.Run(); err == nil {
		return true
	}

	return false
}

// removeAPT removes a package installed via APT
func (m *Manager) removeAPT(packageName string, dryRun bool) (*RemovalResult, error) {
	result := &RemovalResult{
		PackageName:      packageName,
		InstallationType: InstallTypeAPT,
		Success:          false,
	}

	if dryRun {
		result.Success = true
		result.RemovedPaths = []string{
			fmt.Sprintf("APT package: %s", packageName),
			"System-wide configuration files",
		}
		return result, nil
	}

	// Show warning for system packages
	fmt.Printf("⚠️  Removing APT package '%s' - this may affect system stability\n", packageName)
	fmt.Printf("   Consider using '--force' if you're sure this is safe\n")

	commands := [][]string{
		{"sudo", "apt", "remove", "--purge", packageName, "-y"},
		{"sudo", "apt", "autoremove", "-y"},
	}

	var errors []string
	for _, cmd := range commands {
		execCmd := exec.Command(cmd[0], cmd[1:]...)
		if err := execCmd.Run(); err != nil {
			errors = append(errors, fmt.Sprintf("command failed: %v", cmd))
		}
	}

	if len(errors) > 0 {
		result.Error = fmt.Errorf("APT removal failed: %s", strings.Join(errors, "; "))
		return result, result.Error
	}

	result.Success = true
	result.RemovedPaths = []string{
		fmt.Sprintf("APT package: %s", packageName),
		"System-wide configuration files",
	}

	return result, nil
}

// removeManual removes a manually installed package
func (m *Manager) removeManual(packageName string, dryRun bool) (*RemovalResult, error) {
	result := &RemovalResult{
		PackageName:      packageName,
		InstallationType: InstallTypeManual,
		Success:          false,
	}

	paths := []string{
		filepath.Join("/usr/local/bin", packageName+"*"),
		filepath.Join("/usr/local/lib", packageName+"*"),
		filepath.Join("/usr/local/include", packageName+"*"),
		filepath.Join("/usr/local/share", packageName+"*"),
		filepath.Join("/usr/local/etc", packageName+"*"),
	}

	var foundPaths []string
	for _, path := range paths {
		if matches, _ := filepath.Glob(path); len(matches) > 0 {
			foundPaths = append(foundPaths, matches...)
		}
	}

	if len(foundPaths) == 0 {
		result.Warning = fmt.Sprintf("No manual installation found for '%s'", packageName)
		return result, nil
	}

	if dryRun {
		result.Success = true
		result.RemovedPaths = foundPaths
		return result, nil
	}

	// Remove found paths
	for _, path := range foundPaths {
		cmd := exec.Command("sudo", "rm", "-rf", path)
		if err := cmd.Run(); err != nil {
			result.Error = fmt.Errorf("failed to remove %s: %w", path, err)
			return result, result.Error
		}
	}

	result.Success = true
	result.RemovedPaths = foundPaths

	return result, nil
}

// removeUser removes a user-installed package
func (m *Manager) removeUser(packageName string, dryRun bool) (*RemovalResult, error) {
	result := &RemovalResult{
		PackageName:      packageName,
		InstallationType: InstallTypeUser,
		Success:          false,
	}

	home := os.Getenv("HOME")
	paths := []string{
		filepath.Join(home, ".local/bin", packageName+"*"),
		filepath.Join(home, ".local/lib", packageName+"*"),
		filepath.Join(home, ".config", packageName+"*"),
		filepath.Join(home, "."+packageName+"*"),
		filepath.Join(home, ".cache", packageName+"*"),
	}

	var foundPaths []string
	for _, path := range paths {
		if matches, _ := filepath.Glob(path); len(matches) > 0 {
			foundPaths = append(foundPaths, matches...)
		}
	}

	if len(foundPaths) == 0 {
		result.Warning = fmt.Sprintf("No user installation found for '%s'", packageName)
		return result, nil
	}

	if dryRun {
		result.Success = true
		result.RemovedPaths = foundPaths
		return result, nil
	}

	// Remove found paths (no sudo needed for user directories)
	for _, path := range foundPaths {
		cmd := exec.Command("rm", "-rf", path)
		if err := cmd.Run(); err != nil {
			result.Error = fmt.Errorf("failed to remove %s: %w", path, err)
			return result, result.Error
		}
	}

	result.Success = true
	result.RemovedPaths = foundPaths

	return result, nil
}

// removeVersionManager removes a package installed via version manager
func (m *Manager) removeVersionManager(packageName string, dryRun bool) (*RemovalResult, error) {
	result := &RemovalResult{
		PackageName:      packageName,
		InstallationType: InstallTypeVersion,
		Success:          false,
	}

	// Get version manager info using centralized utility
	info, err := GetVersionManagerInfo(packageName)
	if err != nil {
		result.Warning = fmt.Sprintf("No version manager installation found for '%s'", packageName)
		return result, nil
	}

	if len(info.Paths) == 0 {
		result.Warning = fmt.Sprintf("No version manager installation found for '%s'", packageName)
		return result, nil
	}

	if dryRun {
		result.Success = true
		result.RemovedPaths = info.Paths
		if len(info.Commands) > 0 {
			fmt.Printf("Would execute: %s\n", strings.Join(info.Commands, "; "))
		}
		return result, nil
	}

	// Execute version manager commands first
	for _, cmdStr := range info.Commands {
		fmt.Printf("Executing: %s\n", cmdStr)
		if err := ExecuteVersionManagerCommand(cmdStr); err != nil {
			logger.Warn("Version manager command failed: %s - %v", cmdStr, err)
			fmt.Printf("Warning: %s failed (continuing with file removal)\n", cmdStr)
		}
	}

	// Remove found paths as fallback
	for _, path := range info.Paths {
		cmd := exec.Command("rm", "-rf", path)
		if err := cmd.Run(); err != nil {
			result.Error = fmt.Errorf("failed to remove %s: %w", path, err)
			return result, result.Error
		}
	}

	result.Success = true
	result.RemovedPaths = info.Paths

	return result, nil
}

// removeAlternatives removes a package installed via alternatives
func (m *Manager) removeAlternatives(packageName string, dryRun bool) (*RemovalResult, error) {
	result := &RemovalResult{
		PackageName:      packageName,
		InstallationType: InstallTypeAlternatives,
		Success:          false,
	}

	if dryRun {
		result.Success = true
		result.RemovedPaths = []string{
			fmt.Sprintf("Alternatives configuration for: %s", packageName),
		}
		return result, nil
	}

	// Show warning for alternatives packages
	fmt.Printf("⚠️  Removing alternatives configuration for '%s' - this may affect system stability\n", packageName)
	fmt.Printf("   Consider using '--force' if you're sure this is safe\n")

	// Remove the alternatives configuration
	cmd := exec.Command("sudo", "update-alternatives", "--remove-all", packageName)
	if err := cmd.Run(); err != nil {
		result.Error = fmt.Errorf("failed to remove alternatives configuration: %w", err)
		return result, result.Error
	}

	result.Success = true
	result.RemovedPaths = []string{
		fmt.Sprintf("Alternatives configuration for: %s", packageName),
	}

	return result, nil
}

// ShowRemovalSummary displays a summary of removal results
func ShowRemovalSummary(results []*RemovalResult) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("REMOVAL SUMMARY")
	fmt.Println(strings.Repeat("=", 60))

	var successful, failed, warnings []string

	for _, result := range results {
		if result.Success {
			successful = append(successful, fmt.Sprintf("%s (%s)", result.PackageName, result.InstallationType))
		} else if result.Error != nil {
			failed = append(failed, fmt.Sprintf("%s: %v", result.PackageName, result.Error))
		} else if result.Warning != "" {
			warnings = append(warnings, result.Warning)
		}
	}

	if len(successful) > 0 {
		fmt.Printf("✓ Successfully removed (%d): %s\n", len(successful), strings.Join(successful, ", "))
	}

	if len(warnings) > 0 {
		for _, warning := range warnings {
			fmt.Printf("⚠️  %s\n", warning)
		}
	}

	if len(failed) > 0 {
		fmt.Printf("✗ Failed to remove (%d): %s\n", len(failed), strings.Join(failed, ", "))
	}

	total := len(results)
	if total > 0 {
		fmt.Printf("\nTotal: %d packages processed\n", total)
		successRate := float64(len(successful)) / float64(total) * 100
		fmt.Printf("Success rate: %.1f%% (%d/%d)\n", successRate, len(successful), total)
	}
}
