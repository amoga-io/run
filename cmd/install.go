package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/amoga-io/run/internal/logger"
	pkg "github.com/amoga-io/run/internal/package"
	"github.com/spf13/cobra"
)

var (
	packageVersion string
	installAll     bool
	cleanInstall   bool
	dryRunInstall  bool
	listVersions   bool
)

var installCmd = &cobra.Command{
	Use:   "install [package...]",
	Short: "Install packages using installation scripts",
	Long: `Install one or more packages. Dependencies will be checked and installed automatically.

Supported packages:
  • node        - Node.js runtime with npm (versions: 16, 18, 20, 21)
  • php         - PHP programming language (versions: 8.1, 8.2, 8.3)
  • java        - OpenJDK Java Development Kit (versions: 11, 17, 21)
  • docker      - Docker containerization platform
  • nginx       - High-performance web server (versions: stable, mainline)
  • postgres    - PostgreSQL database server (versions: 15, 16, 17)
  • pm2         - Process manager for Node.js applications
  • essentials  - System essentials and development tools

Examples:
  run install node docker
  run install node --version 20
  run install php --version 8.3
  run install node --version 18.20.4
  run install --all
  run install node --clean
  run install node --dry-run`,
	Args: cobra.ArbitraryArgs,
	RunE: runInstall,
}

func init() {
	installCmd.Flags().StringVarP(&packageVersion, "version", "v", "", "Package version to install (e.g., 18 for node, 8.3 for php)")
	installCmd.Flags().BoolVarP(&installAll, "all", "a", false, "Install all available packages")
	installCmd.Flags().BoolVarP(&cleanInstall, "clean", "c", false, "Force clean reinstallation (remove existing first)")
	installCmd.Flags().BoolVarP(&dryRunInstall, "dry-run", "d", false, "Show what would be installed, but do not actually install anything")
	installCmd.Flags().BoolVarP(&listVersions, "list-versions", "l", false, "List all installable/supported versions for the package")
}

func runInstall(cmd *cobra.Command, args []string) error {
	log := logger.GetLogger().WithOperation("install")

	if listVersions {
		if len(args) == 0 {
			return fmt.Errorf("Please specify a package to list versions for. Example: run install node --list-versions")
		}
		for _, packageName := range args {
			pkgInfo, exists := pkg.GetPackage(packageName)
			if !exists {
				fmt.Printf("Package '%s' not found.\n", packageName)
				continue
			}
			aptName := pkgInfo.AptPackageName
			if aptName == "" {
				aptName = packageName
			}
			fmt.Printf("Available versions for %s (via apt):\n", packageName)
			cmd := exec.Command("apt-cache", "madison", aptName)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}
		return nil
	}

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
		log.Info("Validating and sanitizing package list: %v", args)
		sanitizedArgs, err := pkg.SanitizePackageList(args)
		if err != nil {
			log.Error("Input validation failed: %v", err)
			return fmt.Errorf("input validation failed: %w", err)
		}
		args = sanitizedArgs
	}

	// Validate version flag if provided
	if packageVersion != "" {
		log.Info("Validating version flag: %s", packageVersion)
		// Check if any of the packages support version selection
		versionSupported := false
		for _, pkgName := range args {
			if pkg.SupportsVersion(pkgName) {
				versionSupported = true
				if err := pkg.ValidateVersion(pkgName, packageVersion); err != nil {
					log.Error("Version validation failed for %s: %s - %v", pkgName, packageVersion, err)
					return fmt.Errorf("invalid version for %s: %w", pkgName, err)
				}
			}
		}

		if !versionSupported {
			log.Error("No packages support version selection: %v", args)
			return fmt.Errorf("none of the specified packages support version selection")
		}
	}

	log.Info("Creating package manager")
	manager, err := pkg.NewManager()
	if err != nil {
		log.Error("Failed to create package manager: %v", err)
		return err
	}

	var packagesToInstall []string

	if installAll {
		// Get all available packages
		allPackages := pkg.ListPackages()
		for _, pkg := range allPackages {
			packagesToInstall = append(packagesToInstall, pkg.Name)
		}
		log.Info("Installing all packages: %d", len(packagesToInstall))
		fmt.Printf("Installing all packages (%d total)...\n", len(packagesToInstall))
	} else {
		// Validate all packages exist before starting installation
		for _, packageName := range args {
			if _, exists := pkg.GetPackage(packageName); !exists {
				log.Error("Package not found: %s", packageName)
				return fmt.Errorf("package '%s' not found. Run 'run install list' to see available packages", packageName)
			}
		}
		packagesToInstall = args
		log.Info("Installing specific packages: %v", packagesToInstall)
	}

	// Create installation operation function
	installOperation := func(packageName string) error {
		if dryRunInstall {
			fmt.Printf("🔍 DRY-RUN: Would install %s\n", packageName)
			if cleanInstall {
				fmt.Printf("🔍 DRY-RUN: Would remove existing %s first\n", packageName)
			}
			if packageVersion != "" && pkg.SupportsVersion(packageName) {
				fmt.Printf("🔍 DRY-RUN: Would install version %s\n", packageVersion)
			}
			return nil
		}

		// Always purge/remove existing versions before install
		pkgInfo, exists := pkg.GetPackage(packageName)
		if !exists {
			return fmt.Errorf("package '%s' not found", packageName)
		}
		aptName := pkgInfo.AptPackageName
		if aptName == "" {
			aptName = packageName
		}
		// Remove all installed versions (purge)
		purgeCmd := exec.Command("sudo", "apt-get", "purge", "-y", aptName)
		purgeCmd.Stdout = os.Stdout
		purgeCmd.Stderr = os.Stderr
		purgeCmd.Run() // Ignore errors if not installed
		autoremoveCmd := exec.Command("sudo", "apt-get", "autoremove", "-y")
		autoremoveCmd.Stdout = os.Stdout
		autoremoveCmd.Stderr = os.Stderr
		autoremoveCmd.Run()

		// If version is specified, check if it's available
		if packageVersion != "" && pkg.SupportsVersion(packageName) {
			// Check if version is available in apt
			checkCmd := exec.Command("apt-cache", "madison", aptName)
			output, err := checkCmd.Output()
			if err != nil {
				return fmt.Errorf("failed to check available versions for %s: %w", packageName, err)
			}
			available := false
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, packageVersion) {
					available = true
					break
				}
			}
			if !available {
				return fmt.Errorf("version %s of %s is not available in apt repositories", packageVersion, packageName)
			}
			// Install specific version
			installCmd := exec.Command("sudo", "apt-get", "install", "-y", fmt.Sprintf("%s=%s*", aptName, packageVersion))
			installCmd.Stdout = os.Stdout
			installCmd.Stderr = os.Stderr
			if err := installCmd.Run(); err != nil {
				return fmt.Errorf("failed to install %s version %s: %w", packageName, packageVersion, err)
			}
			return nil
		}
		// Default install (latest)
		installCmd := exec.Command("sudo", "apt-get", "install", "-y", aptName)
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("failed to install %s: %w", packageName, err)
		}
		return nil
	}

	// Install packages sequentially using generic function
	log.Info("Starting sequential installation: %v", packagesToInstall)
	results := pkg.ExecutePackagesSequential(manager, packagesToInstall, installOperation, "Installing")

	// Show summary using generic function
	log.Info("Installation completed, showing summary")
	pkg.ShowOperationSummary(results, "installed", "run install")

	return nil
}

// showPackageListAndPrompt displays available packages and prompts user to rerun command
func showPackageListAndPrompt(action string) error {
	fmt.Printf("No packages specified.\n\n")

	// Show concise package list
	fmt.Println("Available:")
	fmt.Println("• node        Node.js + npm")
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
	fmt.Printf("run %s node docker\n", action)
	fmt.Printf("run %s java --version 17\n", action)
	fmt.Printf("run %s essentials\n", action)
	fmt.Printf("run %s --all\n", action)
	if action == "install" {
		fmt.Printf("run %s node --clean\n", action)
	}
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
