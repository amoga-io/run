package pkg

import (
	"fmt"
	"strings"
	"sync"
)

// PackageResult represents the result of a package operation
type PackageResult struct {
	Name    string
	Success bool
	Error   error
}

// PackageOperation represents a function that operates on a package
type PackageOperation func(packageName string) error

// ExecutePackagesParallel executes package operations in parallel with proper synchronization
func ExecutePackagesParallel(
	manager *Manager,
	packages []string,
	operation PackageOperation,
	operationName string,
) []PackageResult {
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
			if err := AcquirePackageLock(pkgName); err != nil {
				result := PackageResult{
					Name:    pkgName,
					Success: false,
					Error:   fmt.Errorf("failed to acquire lock: %w", err),
				}
				resultChan <- result
				fmt.Printf("✗ %s failed to acquire lock: %v\n", pkgName, err)
				return
			}
			defer ReleasePackageLock(pkgName)

			fmt.Printf("%s %s...\n", operationName, pkgName)

			err := operation(pkgName)

			result := PackageResult{
				Name:    pkgName,
				Success: err == nil,
				Error:   err,
			}

			if result.Success {
				fmt.Printf("✓ %s %s successfully\n", pkgName, operationName)
			} else {
				fmt.Printf("✗ %s failed to %s: %v\n", pkgName, operationName, err)
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

// ShowOperationSummary displays a summary of operation results
func ShowOperationSummary(
	results []PackageResult,
	operationName string,
	retryCommand string,
) {
	var successful, failed []string

	for _, result := range results {
		if result.Success {
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

	if len(failed) > 0 {
		fmt.Printf("✗ Failed to %s (%d): %s\n", operationName, len(failed), strings.Join(failed, ", "))
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
			fmt.Printf("\nTo retry failed packages: %s %s\n", retryCommand, strings.Join(failed, " "))
		}
	}
}
