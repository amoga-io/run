package cmd

import (
	"fmt"

	pkg "github.com/amoga-io/run/internal/package"
	"github.com/spf13/cobra"
)

var removeAll bool
var dryRunRemove bool

var removeCmd = &cobra.Command{
	Use:   "remove [package...]",
	Short: "Remove packages completely from the system",
	Long: `Remove one or more packages completely, including all configuration files and traces.

The remove command will:
  • Stop and disable associated services
  • Remove all package files and configurations
  • Clean up dependencies where safe
  • Provide detailed removal summary

Safety features:
  • Critical system packages are protected by default
  • Use --dry-run to preview what would be removed

Examples:
  run remove node python
  run remove docker
  run remove node --dry-run
  run remove --all`,
	Args: cobra.ArbitraryArgs,
	RunE: runRemove,
}

func init() {
	removeCmd.Flags().BoolVarP(&removeAll, "all", "a", false, "Remove all available packages")
	removeCmd.Flags().BoolVarP(&dryRunRemove, "dry-run", "d", false, "Show what would be removed, but do not actually remove anything")
}

func runRemove(cmd *cobra.Command, args []string) error {
	// Show package list and prompt to rerun if no arguments provided
	if len(args) == 0 && !removeAll {
		return showPackageListAndPrompt("remove")
	}

	// Validate and sanitize input
	if !removeAll {
		sanitizedArgs, err := pkg.SanitizePackageList(args)
		if err != nil {
			return fmt.Errorf("input validation failed: %w", err)
		}
		args = sanitizedArgs
	}

	manager, err := pkg.NewManager()
	if err != nil {
		return err
	}

	var packagesToRemove []string

	if removeAll {
		// Get all available packages
		allPackages := pkg.ListPackages()
		for _, pkg := range allPackages {
			packagesToRemove = append(packagesToRemove, pkg.Name)
		}
		fmt.Printf("Removing all packages (%d total)...\n", len(packagesToRemove))
	} else {
		// Validate packages exist before starting removal
		for _, packageName := range args {
			if _, exists := pkg.GetPackage(packageName); !exists {
				return fmt.Errorf("package '%s' not found. Run 'run install list' to see available packages", packageName)
			}
		}
		packagesToRemove = args
	}

	var results []*pkg.RemovalResult
	for _, packageName := range packagesToRemove {
		result, _ := manager.SafeRemovePackage(packageName, false, dryRunRemove)
		results = append(results, result)
	}

	pkg.ShowRemovalSummary(results)
	return nil
}
