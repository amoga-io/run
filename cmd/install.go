package cmd

import (
	"fmt"
	"strings"
	"sync"

	"github.com/amoga-io/run/internal/logger"
	pkg "github.com/amoga-io/run/internal/package"
	"github.com/spf13/cobra"
)

var (
	javaVersion string
	installAll  bool
)

var installCmd = &cobra.Command{
	Use:   "install [package...]",
	Short: "Install packages using installation scripts",
	Long:  "Install one or more packages. Dependencies will be checked and installed automatically.",
	Args:  cobra.ArbitraryArgs,
	RunE:  runInstall,
}

func init() {
	installCmd.Flags().StringVar(&javaVersion, "version", "", "Java version to install (11, 17, 21)")
	installCmd.Flags().BoolVar(&installAll, "all", false, "Install all available packages")
}

func runInstall(cmd *cobra.Command, args []string) error {
	log := logger.GetLogger().WithOperation("install")

	// Handle "list" command
	if len(args) == 1 && args[0] == "list" {
		log.Info("Listing available packages")
		return listPackages()
	}

	// Show package list and prompt to rerun if no arguments provided
	if len(args) == 0 && !installAll {
		log.Info("No packages specified, showing package list")
		return showPackageListAndPrompt("install")
	}

	// Validate and sanitize input
	if !installAll {
		log.Info("Validating and sanitizing package list", "packages", args)
		sanitizedArgs, err := pkg.SanitizePackageList(args)
		if err != nil {
			log.Error("Input validation failed", "error", err)
			return fmt.Errorf("input validation failed: %w", err)
		}
		args = sanitizedArgs
	}

	// Validate version flag if provided
	if javaVersion != "" {
		log.Info("Validating version flag", "version", javaVersion)
		// Check if any of the packages support version selection
		versionSupported := false
		for _, pkgName := range args {
			if pkg.SupportsVersion(pkgName) {
				versionSupported = true
				if err := pkg.ValidateVersion(pkgName, javaVersion); err != nil {
					log.Error("Version validation failed", "package", pkgName, "version", javaVersion, "error", err)
					return fmt.Errorf("invalid version for %s: %w", pkgName, err)
				}
			}
		}

		if !versionSupported {
			log.Error("No packages support version selection", "packages", args)
			return fmt.Errorf("none of the specified packages support version selection")
		}
	}

	log.Info("Creating package manager")
	manager, err := pkg.NewManager()
	if err != nil {
		log.Error("Failed to create package manager", "error", err)
		return err
	}

	var packagesToInstall []string

	if installAll {
		// Get all available packages
		allPackages := pkg.ListPackages()
		for _, pkg := range allPackages {
			packagesToInstall = append(packagesToInstall, pkg.Name)
		}
		log.Info("Installing all packages", "count", len(packagesToInstall))
		fmt.Printf("Installing all packages (%d total)...\n", len(packagesToInstall))
	} else {
		// Validate all packages exist before starting installation
		for _, packageName := range args {
			if _, exists := pkg.GetPackage(packageName); !exists {
				log.Error("Package not found", "package", packageName)
				return fmt.Errorf("package '%s' not found. Run 'run install list' to see available packages", packageName)
			}
		}
		packagesToInstall = args
		log.Info("Installing specific packages", "packages", packagesToInstall)
	}

	// Install packages in parallel
	log.Info("Starting parallel installation", "packages", packagesToInstall)
	results := installPackagesParallel(manager, packagesToInstall)

	// Show summary
	log.Info("Installation completed, showing summary")
	showInstallSummary(results)

	return nil
}

// PackageResult represents the result of a package operation
type PackageResult struct {
	Name    string
	Success bool
	Error   error
}

// installPackagesParallel installs multiple packages in parallel with proper locking
func installPackagesParallel(manager *pkg.Manager, packages []string) []PackageResult {
	var wg sync.WaitGroup
	resultChan := make(chan PackageResult, len(packages))

	// Start goroutines for each package
	for _, packageName := range packages {
		wg.Add(1)
		go func(pkgName string) {
			defer wg.Done()

			// Acquire lock for this package
			if err := pkg.AcquirePackageLock(pkgName); err != nil {
				result := PackageResult{
					Name:    pkgName,
					Success: false,
					Error:   fmt.Errorf("failed to acquire lock: %w", err),
				}
				resultChan <- result
				fmt.Printf("✗ %s failed to acquire lock: %v\n", pkgName, err)
				return
			}
			defer pkg.ReleasePackageLock(pkgName)

			fmt.Printf("Installing %s...\n", pkgName)

			var err error
			// Check if package supports version and version flag is provided
			if pkg.SupportsVersion(pkgName) && javaVersion != "" {
				err = manager.InstallPackageWithArgs(pkgName, []string{"--version", javaVersion})
			} else {
				err = manager.InstallPackage(pkgName)
			}

			result := PackageResult{
				Name:    pkgName,
				Success: err == nil,
				Error:   err,
			}

			if result.Success {
				fmt.Printf("✓ %s installed successfully\n", pkgName)
			} else {
				fmt.Printf("✗ %s failed to install: %v\n", pkgName, err)
			}

			resultChan <- result
		}(packageName)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var allResults []PackageResult
	for result := range resultChan {
		allResults = append(allResults, result)
	}

	return allResults
}

// showInstallSummary displays a summary of installation results
func showInstallSummary(results []PackageResult) {
	var successful, failed []string

	for _, result := range results {
		if result.Success {
			successful = append(successful, result.Name)
		} else {
			failed = append(failed, result.Name)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("INSTALLATION SUMMARY")
	fmt.Println(strings.Repeat("=", 50))

	if len(successful) > 0 {
		fmt.Printf("✓ Successfully installed (%d): %s\n", len(successful), strings.Join(successful, ", "))
	}

	if len(failed) > 0 {
		fmt.Printf("✗ Failed to install (%d): %s\n", len(failed), strings.Join(failed, ", "))
		fmt.Println("\nFailed packages details:")
		for _, result := range results {
			if !result.Success {
				fmt.Printf("  • %s: %v\n", result.Name, result.Error)
			}
		}
	}

	total := len(results)
	fmt.Printf("\nTotal: %d packages processed\n", total)
	fmt.Printf("Success rate: %.1f%% (%d/%d)\n", float64(len(successful))/float64(total)*100, len(successful), total)

	if len(failed) > 0 {
		fmt.Printf("\nTo retry failed packages: run install %s\n", strings.Join(failed, " "))
	}
}

// showPackageListAndPrompt displays available packages and prompts user to rerun command
func showPackageListAndPrompt(action string) error {
	fmt.Printf("No packages specified.\n\n")

	// Show concise package list
	fmt.Println("Available:")
	fmt.Println("• node        Node.js + npm")
	fmt.Println("• python      Python + pip + venv")
	fmt.Println("• php         PHP 8.3 + FPM")
	fmt.Println("• java        OpenJDK 17")
	fmt.Println("• pm2         Process manager for Node.js")
	fmt.Println("• essentials  System tools")
	fmt.Println("• docker      Docker platform")
	fmt.Println("• nginx       Web server")
	fmt.Println("• postgres    PostgreSQL 17")
	fmt.Println()

	// Show usage examples
	fmt.Println("Examples:")
	fmt.Printf("run %s node python\n", action)
	fmt.Printf("run %s java --version 17\n", action)
	fmt.Printf("run %s essentials\n", action)
	fmt.Printf("run %s --all\n", action)
	fmt.Println()
	fmt.Printf("Run run %s list to see all.\n", action)

	return fmt.Errorf("please specify packages to %s", action)
}

// toTitle converts first character to uppercase (replacement for deprecated strings.Title)
func toTitle(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func listPackages() error {
	packages := pkg.ListPackages()

	fmt.Println("Available packages:")
	fmt.Println()

	categories := make(map[string][]pkg.Package)
	for _, p := range packages {
		categories[p.Category] = append(categories[p.Category], p)
	}

	for category, pkgs := range categories {
		fmt.Printf("%s:\n", toTitle(category))
		for _, p := range pkgs {
			deps := ""
			if len(p.Dependencies) > 0 {
				deps = fmt.Sprintf(" (deps: %s)", strings.Join(p.Dependencies, ", "))
			}
			fmt.Printf("  %-12s - %s%s\n", p.Name, p.Description, deps)
		}
		fmt.Println()
	}

	return nil
}
