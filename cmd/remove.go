package cmd

import (
	"fmt"

	pkg "github.com/amoga-io/run/internal/package"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [package...]",
	Short: "Remove packages completely from the system",
	Long:  "Remove one or more packages completely, including all configuration files and traces.",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runRemove,
}

func runRemove(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return listPackages()
	}
	// Validate packages exist before starting removal
	for _, packageName := range args {
		if _, exists := pkg.GetPackage(packageName); !exists {
			return fmt.Errorf("package '%s' not found. Run 'run install list' to see available packages", packageName)
		}
	}

	// Create package manager
	manager, err := pkg.NewManager()
	if err != nil {
		return err
	}

	// Remove each package
	for _, packageName := range args {
		if err := manager.RemovePackage(packageName); err != nil {
			return fmt.Errorf("failed to remove %s: %w", packageName, err)
		}
		fmt.Printf("✓ %s removed successfully\n\n", packageName)
	}

	fmt.Printf("✓ All packages removed successfully!\n")
	return nil
}
