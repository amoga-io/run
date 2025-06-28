package cmd

import (
	"fmt"
	"strings"

	pkg "github.com/amoga-io/run/internal/package"
	"github.com/amoga-io/run/internal/system"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check [packages|deps|requirements]",
	Short: "Check installed packages and system dependencies",
	Long:  "Check what packages are installed and what system dependencies are available or missing",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runCheck,
}

func runCheck(cmd *cobra.Command, args []string) error {
	checkType := "packages"
	if len(args) > 0 {
		checkType = args[0]
	}

	switch checkType {
	case "packages", "pkgs":
		return checkPackages()
	case "deps", "dependencies":
		return checkDependencies()
	case "bootstrap":
		return checkBootstrap()
	case "runtime":
		return checkRuntime()
	case "dev", "development":
		return checkDevelopment()
	case "all", "requirements":
		return checkAll()
	default:
		return fmt.Errorf("unknown check type: %s. Use: packages, deps, bootstrap, runtime, dev, or all", checkType)
	}
}

// checkPackages checks which packages from our registry are installed
func checkPackages() error {
	fmt.Println("Checking installed packages...")
	fmt.Println()

	allPackages := pkg.ListPackages()
	var installed, notInstalled []pkg.Package

	// Check each package
	for _, pkg := range allPackages {
		if isPackageInstalled(pkg) {
			installed = append(installed, pkg)
		} else {
			notInstalled = append(notInstalled, pkg)
		}
	}

	// Show installed packages
	if len(installed) > 0 {
		fmt.Printf("✓ Installed packages (%d):\n", len(installed))
		for _, pkg := range installed {
			fmt.Printf("  • %-12s - %s\n", pkg.Name, pkg.Description)
		}
		fmt.Println()
	}

	// Show not installed packages
	if len(notInstalled) > 0 {
		fmt.Printf("❌ Not installed (%d):\n", len(notInstalled))
		for _, pkg := range notInstalled {
			fmt.Printf("  • %-12s - %s\n", pkg.Name, pkg.Description)
		}
		fmt.Println()
	}

	// Show summary
	total := len(allPackages)
	installedCount := len(installed)
	notInstalledCount := len(notInstalled)

	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("PACKAGE STATUS SUMMARY")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Total packages: %d\n", total)
	fmt.Printf("Installed: %d (%.1f%%)\n", installedCount, float64(installedCount)/float64(total)*100)
	fmt.Printf("Not installed: %d (%.1f%%)\n", notInstalledCount, float64(notInstalledCount)/float64(total)*100)
	fmt.Println()

	// Show installation suggestions
	if len(notInstalled) > 0 {
		fmt.Println("To install missing packages:")
		var packageNames []string
		for _, pkg := range notInstalled {
			packageNames = append(packageNames, pkg.Name)
		}
		fmt.Printf("  run install %s\n", strings.Join(packageNames, " "))
		fmt.Printf("  run install --all\n")
	}

	return nil
}

// isPackageInstalled checks if a package is installed by checking its commands
func isPackageInstalled(pkg pkg.Package) bool {
	for _, cmd := range pkg.Commands {
		if !system.CommandExists(cmd) {
			return false
		}
	}
	return len(pkg.Commands) > 0 // Only return true if there are commands to check
}

func checkBootstrap() error {
	fmt.Println("Checking bootstrap requirements...")
	missing, err := system.CheckSystemRequirements(system.Bootstrap)
	if err != nil {
		return err
	}
	return reportMissing("Bootstrap", missing)
}

func checkRuntime() error {
	fmt.Println("Checking runtime requirements...")
	missing, err := system.CheckSystemRequirements(system.Runtime)
	if err != nil {
		return err
	}
	return reportMissing("Runtime", missing)
}

func checkDevelopment() error {
	fmt.Println("Checking development requirements...")
	missing, err := system.CheckSystemRequirements(system.Development)
	if err != nil {
		return err
	}
	return reportMissing("Development", missing)
}

func checkDependencies() error {
	fmt.Println("Checking legacy dependencies...")
	missing, err := system.CheckAllDependencies()
	if err != nil {
		return err
	}

	if len(missing) == 0 {
		fmt.Println("✓ All legacy dependencies are available")
		return nil
	}

	fmt.Printf("❌ Missing %d legacy dependencies:\n", len(missing))
	for _, dep := range missing {
		fmt.Printf("  - %s: %s (install: %s)\n", dep.Command, dep.Description, dep.Package)
	}
	return nil
}

func checkAll() error {
	fmt.Println("Checking all system requirements and packages...")
	fmt.Println()

	// Check packages first
	if err := checkPackages(); err != nil {
		return err
	}

	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("SYSTEM DEPENDENCIES")
	fmt.Println(strings.Repeat("-", 50))

	categories := []system.RequirementCategory{
		system.Bootstrap,
		system.Runtime,
		system.Development,
		system.Optional,
	}

	allGood := true

	for _, category := range categories {
		missing, err := system.CheckSystemRequirements(category)
		if err != nil {
			return err
		}

		if len(missing) > 0 {
			allGood = false
			fmt.Printf("\n%s requirements:\n", string(category))
			for _, req := range missing {
				status := "❌"
				if !req.Critical {
					status = "⚠️"
				}
				fmt.Printf("  %s %s: %s\n", status, req.Name, req.Description)
			}
		}
	}

	if allGood {
		fmt.Println("✓ All system requirements are satisfied")
	} else {
		fmt.Println("\n❌ = Critical missing")
		fmt.Println("⚠️ = Optional missing")
	}

	return nil
}

func reportMissing(category string, missing []system.SystemRequirement) error {
	if len(missing) == 0 {
		fmt.Printf("✓ All %s requirements are available\n", category)
		return nil
	}

	fmt.Printf("❌ Missing %d %s requirements:\n", len(missing), category)
	for _, req := range missing {
		status := "required"
		if !req.Critical {
			status = "optional"
		}
		fmt.Printf("  - %s: %s (%s)\n", req.Name, req.Description, status)
	}
	return nil
}
