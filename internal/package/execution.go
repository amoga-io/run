package pkg

import (
	"fmt"
	"strings"
)

// PackageResult represents the result of a package operation
type PackageResult struct {
	Name    string
	Success bool
	Skipped bool
	Error   error
	Message string
}

// PackageOperation represents a function that operates on a package
type PackageOperation func(packageName string) error

// ExecutePackagesSequential executes package operations sequentially
func ExecutePackagesSequential(
	manager *Manager,
	packages []string,
	operation PackageOperation,
	operationName string,
) []PackageResult {
	var allResults []PackageResult

	// Process packages sequentially
	for _, packageName := range packages {
		// Acquire lock for this package
		if err := AcquirePackageLock(packageName); err != nil {
			result := PackageResult{
				Name:    packageName,
				Success: false,
				Skipped: false,
				Error:   fmt.Errorf("failed to acquire lock: %w", err),
				Message: "failed to acquire lock",
			}
			allResults = append(allResults, result)
			fmt.Printf("✗ %s failed to acquire lock: %v\n", packageName, err)
			continue
		}
		defer ReleasePackageLock(packageName)

		fmt.Printf("%s %s...\n", operationName, packageName)

		err := operation(packageName)

		// Check if this was a skip (package already installed with same version)
		if IsPackageAlreadyInstalledError(err) {
			result := PackageResult{
				Name:    packageName,
				Success: true,
				Skipped: true,
				Error:   nil,
				Message: err.Error(),
			}
			allResults = append(allResults, result)
			continue
		}

		result := PackageResult{
			Name:    packageName,
			Success: err == nil,
			Skipped: false,
			Error:   err,
			Message: "",
		}

		if result.Success {
			fmt.Printf("✓ %s %s successfully\n", packageName, operationName)
		} else {
			fmt.Printf("✗ %s failed to %s: %v\n", packageName, operationName, err)
		}

		allResults = append(allResults, result)
	}

	return allResults
}

// ShowOperationSummary displays a summary of operation results
func ShowOperationSummary(
	results []PackageResult,
	operationName string,
	retryCommand string,
) {
	var successful, failed, skipped []string

	for _, result := range results {
		if result.Skipped {
			skipped = append(skipped, result.Name)
		} else if result.Success {
			successful = append(successful, result.Name)
		} else {
			failed = append(failed, result.Name)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Printf("%s SUMMARY\n", strings.ToUpper(operationName))
	fmt.Println(strings.Repeat("=", 50))

	if len(successful) > 0 {
		fmt.Printf("✓ Successfully %s (%d): %s\n", operationName, len(successful), strings.Join(successful, ", "))
	}

	if len(skipped) > 0 {
		fmt.Printf("⏭️  Skipped - already installed (%d): %s\n", len(skipped), strings.Join(skipped, ", "))
	}

	if len(failed) > 0 {
		fmt.Printf("✗ Failed to %s (%d): %s\n", operationName, len(failed), strings.Join(failed, ", "))
		fmt.Println("\nFailed packages details:")
		for _, result := range results {
			if !result.Success && !result.Skipped {
				fmt.Printf("  • %s: %v\n", result.Name, result.Error)
			}
		}
	}

	total := len(results)
	if total > 0 {
		fmt.Printf("\nTotal: %d packages processed\n", total)
		successRate := float64(len(successful)+len(skipped)) / float64(total) * 100
		fmt.Printf("Success rate: %.1f%% (%d/%d)\n", successRate, len(successful)+len(skipped), total)

		if len(failed) > 0 {
			fmt.Printf("\nTo retry failed packages: %s %s\n", retryCommand, strings.Join(failed, " "))
		}
	}
}
