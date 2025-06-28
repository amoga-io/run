package cmd

import (
	"fmt"
	"strings"

	"github.com/amoga-io/run/internal/logger"
	pkg "github.com/amoga-io/run/internal/package"
	"github.com/spf13/cobra"
)

var (
	packageVersion string
	installAll     bool
)

var installCmd = &cobra.Command{
	Use:   "install [package...]",
	Short: "Install packages using installation scripts",
	Long:  "Install one or more packages. Dependencies will be checked and installed automatically.",
	Args:  cobra.ArbitraryArgs,
	RunE:  runInstall,
}

func init() {
	installCmd.Flags().StringVar(&packageVersion, "version", "", "Package version to install (e.g., 18 for node, 3.10 for python)")
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
		// Check if package supports version and version flag is provided
		if pkg.SupportsVersion(packageName) && packageVersion != "" {
			return manager.InstallPackageWithArgs(packageName, []string{"--version", packageVersion})
		}
		return manager.InstallPackage(packageName)
	}

	// Install packages in parallel using generic function
	log.Info("Starting parallel installation: %v", packagesToInstall)
	results := pkg.ExecutePackagesParallel(manager, packagesToInstall, installOperation, "Installing")

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
