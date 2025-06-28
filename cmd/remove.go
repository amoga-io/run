package cmd

import (
	"fmt"
	"strings"
	"sync"

	pkg "github.com/amoga-io/run/internal/package"
	"github.com/spf13/cobra"
)

var removeAll bool

var removeCmd = &cobra.Command{
	Use:   "remove [package...]",
	Short: "Remove packages completely from the system",
	Long:  "Remove one or more packages completely, including all configuration files and traces.",
	Args:  cobra.ArbitraryArgs,
	RunE:  runRemove,
}

func init() {
	removeCmd.Flags().BoolVar(&removeAll, "all", false, "Remove all available packages")
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

	// Remove packages in parallel with proper synchronization
	results := removePackagesParallel(manager, packagesToRemove)

	// Show summary
	showRemoveSummary(results)

	return nil
}

// removePackagesParallel removes multiple packages in parallel with proper locking
func removePackagesParallel(manager *pkg.Manager, packages []string) []PackageResult {
	var wg sync.WaitGroup
	resultChan := make(chan PackageResult, len(packages))

	// Use a mutex to protect shared state
	var resultMutex sync.Mutex
	var allResults []PackageResult

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

			fmt.Printf("Removing %s...\n", pkgName)

			err := manager.RemovePackage(pkgName)

			result := PackageResult{
				Name:    pkgName,
				Success: err == nil,
				Error:   err,
			}

			if result.Success {
				fmt.Printf("✓ %s removed successfully\n", pkgName)
			} else {
				fmt.Printf("✗ %s failed to remove: %v\n", pkgName, err)
			}

			resultChan <- result
		}(packageName)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results with proper synchronization
	for result := range resultChan {
		resultMutex.Lock()
		allResults = append(allResults, result)
		resultMutex.Unlock()
	}

	return allResults
}

// showRemoveSummary displays a summary of removal results
func showRemoveSummary(results []PackageResult) {
	var successful, failed []string

	for _, result := range results {
		if result.Success {
			successful = append(successful, result.Name)
		} else {
			failed = append(failed, result.Name)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("REMOVAL SUMMARY")
	fmt.Println(strings.Repeat("=", 50))

	if len(successful) > 0 {
		fmt.Printf("✓ Successfully removed (%d): %s\n", len(successful), strings.Join(successful, ", "))
	}

	if len(failed) > 0 {
		fmt.Printf("✗ Failed to remove (%d): %s\n", len(failed), strings.Join(failed, ", "))
		fmt.Println("\nFailed packages details:")
		for _, result := range results {
			if !result.Success {
				fmt.Printf("  • %s: %v\n", result.Name, result.Error)
			}
		}
	}

	total := len(results)
	if total > 0 {
		fmt.Printf("\nTotal: %d packages processed\n", total)
		fmt.Printf("Success rate: %.1f%% (%d/%d)\n", float64(len(successful))/float64(total)*100, len(successful), total)

		if len(failed) > 0 {
			fmt.Printf("\nTo retry failed packages: run remove %s\n", strings.Join(failed, " "))
		}
	}
}
