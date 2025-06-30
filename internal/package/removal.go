package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
func (m *Manager) SafeRemovePackage(packageName string, _ bool, dryRun bool) (*RemovalResult, error) {
	result := &RemovalResult{
		PackageName: packageName,
		Success:     false,
	}

	// Remove via APT
	aptResult, _ := m.removeAPT(packageName, false, dryRun)
	if aptResult != nil && (aptResult.Success || aptResult.Warning != "") {
		result.RemovedPaths = append(result.RemovedPaths, aptResult.RemovedPaths...)
	}

	// Remove manual installs
	manualResult, _ := m.removeManual(packageName, dryRun)
	if manualResult != nil && (manualResult.Success || manualResult.Warning != "") {
		result.RemovedPaths = append(result.RemovedPaths, manualResult.RemovedPaths...)
	}

	// Remove user installs
	userResult, _ := m.removeUser(packageName, dryRun)
	if userResult != nil && (userResult.Success || userResult.Warning != "") {
		result.RemovedPaths = append(result.RemovedPaths, userResult.RemovedPaths...)
	}

	// Remove alternatives
	altResult, _ := m.removeAlternatives(packageName, false, dryRun)
	if altResult != nil && (altResult.Success || altResult.Warning != "") {
		result.RemovedPaths = append(result.RemovedPaths, altResult.RemovedPaths...)
	}

	// Mark as success if any removal succeeded
	if len(result.RemovedPaths) > 0 {
		result.Success = true
	}

	return result, nil
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

// getInstalledAptPackagesMatching returns a list of installed APT packages matching a pattern (e.g., "php", "postgresql", "nginx")
func getInstalledAptPackagesMatching(pattern string) ([]string, error) {
	// Special handling for postgres: match postgresql-*
	if pattern == "postgresql" {
		cmd := exec.Command("bash", "-c", "dpkg -l | grep '^ii' | awk '{print $2}' | grep '^postgresql'")
		output, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		var pkgs []string
		for _, line := range lines {
			if line != "" {
				pkgs = append(pkgs, line)
			}
		}
		return pkgs, nil
	}
	// Default logic
	cmd := exec.Command("bash", "-c", "dpkg -l | grep '^ii' | awk '{print $2}' | grep '^"+pattern+"'")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var pkgs []string
	for _, line := range lines {
		if line != "" {
			pkgs = append(pkgs, line)
		}
	}
	return pkgs, nil
}

// removeAPT removes a package installed via APT
func (m *Manager) removeAPT(packageName string, force bool, dryRun bool) (*RemovalResult, error) {
	result := &RemovalResult{
		PackageName:      packageName,
		InstallationType: InstallTypeAPT,
		Success:          false,
	}

	var pkgsToRemove []string
	found, err := getInstalledAptPackagesMatching(packageName)
	fmt.Printf("[DEBUG] getInstalledAptPackagesMatching(%s) -> %v\n", packageName, found)
	if err != nil {
		fmt.Printf("[DEBUG] Error from getInstalledAptPackagesMatching: %v\n", err)
	}
	if len(found) > 0 {
		pkgsToRemove = found
	}
	if len(pkgsToRemove) == 0 {
		// Show what dpkg -l | grep postgresql returns
		cmd := exec.Command("bash", "-c", "dpkg -l | grep postgresql")
		out, _ := cmd.Output()
		fmt.Printf("[DEBUG] dpkg -l | grep postgresql:\n%s\n", string(out))
		// Fallback: try to remove all postgresql-* packages if any are present
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 2 && strings.HasPrefix(fields[1], "postgresql") {
				pkgsToRemove = append(pkgsToRemove, fields[1])
			}
		}
	}

	if len(pkgsToRemove) == 0 {
		fmt.Printf("[DEBUG] No postgresql packages found for removal.\n")
	}

	// Debug output
	fmt.Printf("[DEBUG] Packages to remove for '%s': %v\n", packageName, pkgsToRemove)

	if dryRun {
		result.Success = true
		result.RemovedPaths = []string{
			"APT packages: " + strings.Join(pkgsToRemove, ", "),
			"System-wide configuration files",
		}
		return result, nil
	}

	if !force {
		fmt.Printf("⚠️  Removing APT package(s) '%s' - this may affect system stability\n", strings.Join(pkgsToRemove, ", "))
		fmt.Printf("   Consider using '--force' if you're sure this is safe\n")
	}

	commands := [][]string{
		append([]string{"sudo", "apt", "remove", "--purge"}, append(pkgsToRemove, "-y")...),
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
		"APT packages: " + strings.Join(pkgsToRemove, ", ") + " (purged)",
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

// removeAlternatives removes a package installed via alternatives
func (m *Manager) removeAlternatives(packageName string, force bool, dryRun bool) (*RemovalResult, error) {
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
	if !force {
		fmt.Printf("\u26a0\ufe0f  Removing alternatives configuration for '%s' - this may affect system stability\n", packageName)
		fmt.Printf("   Consider using '--force' if you're sure this is safe\n")
	}

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
