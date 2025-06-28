package cmd

import (
	"fmt"

	pkg "github.com/amoga-io/run/internal/package"
	"github.com/spf13/cobra"
)

var removeAll bool
var forceRemove bool
var dryRunRemove bool

var removeCmd = &cobra.Command{
	Use:   "remove [package...]",
	Short: "Remove packages completely from the system",
	Long:  "Remove one or more packages completely, including all configuration files and traces.",
	Args:  cobra.ArbitraryArgs,
	RunE:  runRemove,
}

func init() {
	removeCmd.Flags().BoolVar(&removeAll, "all", false, "Remove all available packages")
	removeCmd.Flags().BoolVar(&forceRemove, "force", false, "Force removal of system-critical packages (DANGEROUS)")
	removeCmd.Flags().BoolVar(&dryRunRemove, "dry-run", false, "Show what would be removed, but do not actually remove anything")
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
		result, _ := manager.SafeRemovePackage(packageName, forceRemove, dryRunRemove)
		results = append(results, result)
	}

	pkg.ShowRemovalSummary(results)
	return nil
}
