package cmd

import (
	"fmt"
	"strings"

	pkg "github.com/amoga-io/run/internal/package"
	"github.com/spf13/cobra"
)

var javaVersion string

var installCmd = &cobra.Command{
	Use:   "install [package]",
	Short: "Install packages using installation scripts",
	Long:  "Install one or more packages. Dependencies will be checked and installed automatically.",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runInstall,
}

func init() {
	installCmd.Flags().StringVar(&javaVersion, "version", "", "Java version to install (11, 17, 21)")
}

func runInstall(cmd *cobra.Command, args []string) error {
	// Handle special commands
	if len(args) == 1 && args[0] == "list" {
		return listPackages()
	}

	packageName := args[0]
	if _, exists := pkg.GetPackage(packageName); !exists {
		return fmt.Errorf("package '%s' not found. Run 'run install list' to see available packages", packageName)
	}

	manager, err := pkg.NewManager()
	if err != nil {
		return err
	}

	// If installing java and --version is set, pass it to the script
	if packageName == "java" && javaVersion != "" {
		if err := manager.InstallPackageWithArgs(packageName, []string{"--version", javaVersion}); err != nil {
			return fmt.Errorf("failed to install %s: %w", packageName, err)
		}
		fmt.Printf("✓ %s installed successfully\n\n", packageName)
		return nil
	}

	// Default: install package with no extra args
	if err := manager.InstallPackage(packageName); err != nil {
		return fmt.Errorf("failed to install %s: %w", packageName, err)
	}
	fmt.Printf("✓ %s installed successfully\n\n", packageName)
	return nil
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
