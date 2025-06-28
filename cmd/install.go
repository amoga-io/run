package cmd

import (
	"fmt"
	"strings"

	pkg "github.com/amoga-io/run/internal/package"
	"github.com/spf13/cobra"
)

var javaVersion string

var installCmd = &cobra.Command{
	Use:   "install [package...]",
	Short: "Install packages using installation scripts",
	Long:  "Install one or more packages. Dependencies will be checked and installed automatically.",
	Args:  cobra.ArbitraryArgs,
	RunE:  runInstall,
}

func init() {
	installCmd.Flags().StringVar(&javaVersion, "version", "", "Java version to install (11, 17, 21)")
}

func runInstall(cmd *cobra.Command, args []string) error {
	// Handle "list" command
	if len(args) == 1 && args[0] == "list" {
		return listPackages()
	}

	// Show package list and prompt to rerun if no arguments provided
	if len(args) == 0 {
		return showPackageListAndPrompt("install")
	}

	// Validate all packages exist before starting installation
	for _, packageName := range args {
		if _, exists := pkg.GetPackage(packageName); !exists {
			return fmt.Errorf("package '%s' not found. Run 'run install list' to see available packages", packageName)
		}
	}

	manager, err := pkg.NewManager()
	if err != nil {
		return err
	}

	// Install each package
	for _, packageName := range args {
		if err := installSinglePackage(manager, packageName); err != nil {
			return fmt.Errorf("failed to install %s: %w", packageName, err)
		}
	}

	fmt.Printf("✓ All packages installed successfully!\n")
	return nil
}

// installSinglePackage installs a single package with proper version handling
func installSinglePackage(manager *pkg.Manager, packageName string) error {
	fmt.Printf("\nInstalling %s...\n", packageName)

	// If installing java and --version is set, pass it to the script
	if packageName == "java" && javaVersion != "" {
		if err := manager.InstallPackageWithArgs(packageName, []string{"--version", javaVersion}); err != nil {
			return fmt.Errorf("failed to install %s: %w", packageName, err)
		}
		fmt.Printf("✓ %s installed successfully\n", packageName)
		return nil
	}

	// Default: install package with no extra args
	if err := manager.InstallPackage(packageName); err != nil {
		return fmt.Errorf("failed to install %s: %w", packageName, err)
	}
	fmt.Printf("✓ %s installed successfully\n", packageName)
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
