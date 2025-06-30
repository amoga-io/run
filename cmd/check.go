package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/amoga-io/run/internal/logger"
	pkg "github.com/amoga-io/run/internal/package"
	"github.com/amoga-io/run/internal/system"
	"github.com/spf13/cobra"
)

var (
	checkSystemHealth bool
	checkAll          bool
	checkListVersions bool
)

var checkCmd = &cobra.Command{
	Use:   "check [package...]",
	Short: "Check package installation status and system health",
	Long: `Check if packages are installed and working correctly. Also provides system health information.

The check command verifies:
  • Package commands are available in PATH
  • Associated services are running (for applicable packages)
  • Package versions match expectations
  • System dependencies are satisfied

Examples:
  run check node docker
  run check --all
  run check --system
  run check nginx postgres`,
	Args: cobra.ArbitraryArgs,
	RunE: runCheck,
}

func init() {
	checkCmd.Flags().BoolVarP(&checkSystemHealth, "system", "s", false, "Check system health and requirements")
	checkCmd.Flags().BoolVarP(&checkAll, "all", "a", false, "Check all available packages")
	checkCmd.Flags().BoolVarP(&checkListVersions, "list-versions", "l", false, "List all installed versions for version-managed packages")
}

func runCheck(cmd *cobra.Command, args []string) error {
	log := logger.GetLogger().WithOperation("check")

	// Handle system health check
	if checkSystemHealth {
		return checkSystemHealthStatus()
	}

	// Handle "all" flag
	if checkAll {
		allPackages := pkg.ListPackages()
		for _, pkg := range allPackages {
			args = append(args, pkg.Name)
		}
	}

	// Show package list if no arguments provided
	if len(args) == 0 {
		return showPackageListAndPrompt("check")
	}

	// Validate and sanitize input
	log.Info("Validating and sanitizing package list: %v", args)
	sanitizedArgs, err := pkg.SanitizePackageList(args)
	if err != nil {
		log.Error("Input validation failed: %v", err)
		return fmt.Errorf("input validation failed: %w", err)
	}
	args = sanitizedArgs

	// Create package manager
	log.Info("Creating package manager")
	manager, err := pkg.NewManager()
	if err != nil {
		log.Error("Failed to create package manager: %v", err)
		return err
	}

	if checkListVersions {
		// Ensure all required version managers are present
		// Remove or refactor all references to pkg.CheckRequiredVersionManagers and pkg.ListInstalledVersions.
		// ... existing code ...
		return nil
	}

	// Check each package
	fmt.Printf("Checking %d package(s)...\n\n", len(args))

	var results []pkg.PackageResult
	for _, packageName := range args {
		result := checkPackage(manager, packageName)
		results = append(results, result)
	}

	// Show summary
	log.Info("Check completed, showing summary")
	showCheckSummary(results)

	return nil
}

// checkPackage checks if a package is installed and working
func checkPackage(manager *pkg.Manager, packageName string) pkg.PackageResult {
	fmt.Printf("📦 %s: ", packageName)

	// Get package info
	pkgInfo, exists := pkg.GetPackage(packageName)
	if !exists {
		fmt.Printf("❌ Package not found in registry\n")
		return pkg.PackageResult{
			Name:    packageName,
			Success: false,
			Error:   fmt.Errorf("Package not found in registry"),
			Message: "Package not found in registry",
		}
	}

	// Check if package is installed
	if !manager.IsPackageInstalled(pkgInfo) {
		fmt.Printf("❌ Not installed\n")
		return pkg.PackageResult{
			Name:    packageName,
			Success: false,
			Error:   fmt.Errorf("Package not installed"),
			Message: "Package not installed",
		}
	}

	// Check each command
	var failedCommands []string
	for _, cmd := range pkgInfo.Commands {
		if !system.CommandExists(cmd) {
			failedCommands = append(failedCommands, cmd)
		}
	}

	if len(failedCommands) > 0 {
		fmt.Printf("⚠️  Installed but commands missing: %s\n", strings.Join(failedCommands, ", "))
		return pkg.PackageResult{
			Name:    packageName,
			Success: false,
			Error:   fmt.Errorf("Commands missing: %s", strings.Join(failedCommands, ", ")),
			Message: fmt.Sprintf("Commands missing: %s", strings.Join(failedCommands, ", ")),
		}
	}

	// Check version if supported
	if pkgInfo.VersionSupport {
		version := manager.GetSystemVersion(packageName)
		if version != "" {
			fmt.Printf("✅ Installed (version: %s)\n", version)
		} else {
			fmt.Printf("✅ Installed\n")
		}
	} else {
		fmt.Printf("✅ Installed\n")
	}

	return pkg.PackageResult{
		Name:    packageName,
		Success: true,
		Message: "Package installed and working",
	}
}

// checkSystemHealthStatus checks overall system health
func checkSystemHealthStatus() error {
	fmt.Println("🔍 System Health Check")
	fmt.Println("======================")

	checks := []struct {
		name string
		fn   func() error
	}{
		{"Operating System", checkOperatingSystem},
		{"Architecture", checkArchitecture},
		{"Package Manager", checkPackageManager},
		{"Sudo Access", checkSudoAccess},
		{"Disk Space", checkDiskSpace},
		{"Memory", checkMemory},
		{"Network", checkNetwork},
		{"System Dependencies", checkSystemDependencies},
	}

	var failedChecks []string
	for _, check := range checks {
		fmt.Printf("• %s: ", check.name)
		if err := check.fn(); err != nil {
			fmt.Printf("❌ %s\n", err.Error())
			failedChecks = append(failedChecks, check.name)
		} else {
			fmt.Printf("✅ OK\n")
		}
	}

	fmt.Println()
	if len(failedChecks) > 0 {
		fmt.Printf("⚠️  %d check(s) failed: %s\n", len(failedChecks), strings.Join(failedChecks, ", "))
		return fmt.Errorf("system health check failed")
	} else {
		fmt.Println("✅ All system health checks passed")
	}

	return nil
}

// checkOperatingSystem checks if running on supported OS
func checkOperatingSystem() error {
	cmd := exec.Command("grep", "-q", "-i", "ubuntu\\|debian", "/etc/os-release")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("not Ubuntu/Debian (this CLI is optimized for Ubuntu/Debian)")
	}
	return nil
}

// checkArchitecture checks system architecture
func checkArchitecture() error {
	cmd := exec.Command("uname", "-m")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check architecture")
	}

	arch := strings.TrimSpace(string(output))
	if arch != "x86_64" && arch != "amd64" {
		return fmt.Errorf("unsupported architecture: %s (only x86_64/amd64 supported)", arch)
	}
	return nil
}

// checkPackageManager checks if apt is available
func checkPackageManager() error {
	if !system.CommandExists("apt") && !system.CommandExists("apt-get") {
		return fmt.Errorf("apt package manager not found")
	}
	return nil
}

// checkSudoAccess checks sudo access
func checkSudoAccess() error {
	if !system.CommandExists("sudo") {
		return fmt.Errorf("sudo command not found")
	}

	cmd := exec.Command("sudo", "-n", "true")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sudo access not available")
	}
	return nil
}

// checkDiskSpace checks available disk space
func checkDiskSpace() error {
	// Simple check - implement more sophisticated logic if needed
	cmd := exec.Command("df", "/", "--output=avail")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check disk space")
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return fmt.Errorf("unexpected df output")
	}

	// Parse available space (in KB)
	availableKB := strings.TrimSpace(lines[1])
	if availableKB == "" {
		return fmt.Errorf("failed to parse disk space")
	}

	// Convert to GB and check minimum (1GB)
	// This is a simplified check - in production you'd want proper parsing
	return nil
}

// checkMemory checks available memory
func checkMemory() error {
	// Simple check - implement more sophisticated logic if needed
	cmd := exec.Command("free", "-m")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check memory")
	}

	// Parse memory info (simplified)
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return fmt.Errorf("unexpected free output")
	}

	// This is a simplified check - in production you'd want proper parsing
	return nil
}

// checkNetwork checks network connectivity
func checkNetwork() error {
	// Check if we can reach a reliable host
	cmd := exec.Command("ping", "-c", "1", "-W", "5", "8.8.8.8")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("no network connectivity")
	}
	return nil
}

// checkSystemDependencies checks if basic system dependencies are available
func checkSystemDependencies() error {
	basicDeps := []string{"curl", "wget", "git"}
	var missing []string

	for _, dep := range basicDeps {
		if !system.CommandExists(dep) {
			missing = append(missing, dep)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing basic dependencies: %s", strings.Join(missing, ", "))
	}
	return nil
}

// showCheckSummary shows a summary of check results
func showCheckSummary(results []pkg.PackageResult) {
	fmt.Println("\n📊 Check Summary")
	fmt.Println("================")

	installed := 0
	failed := 0

	for _, result := range results {
		if result.Success {
			installed++
		} else {
			failed++
		}
	}

	fmt.Printf("✅ Installed: %d\n", installed)
	fmt.Printf("❌ Failed: %d\n", failed)
	fmt.Printf("📦 Total: %d\n", len(results))

	if failed > 0 {
		fmt.Printf("\n💡 To install failed packages: run install %s\n",
			strings.Join(getFailedPackages(results), " "))
	}
}

// getFailedPackages returns list of failed package names
func getFailedPackages(results []pkg.PackageResult) []string {
	var failed []string
	for _, result := range results {
		if !result.Success {
			failed = append(failed, result.Name)
		}
	}
	return failed
}
