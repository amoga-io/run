package cmd

import (
	"fmt"

	"github.com/amoga-io/run/internal/system"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check [deps|requirements]",
	Short: "Check system dependencies and requirements",
	Long:  "Check what system dependencies and requirements are available or missing",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runCheck,
}

func runCheck(cmd *cobra.Command, args []string) error {
	checkType := "all"
	if len(args) > 0 {
		checkType = args[0]
	}

	switch checkType {
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
		return fmt.Errorf("unknown check type: %s. Use: deps, bootstrap, runtime, dev, or all", checkType)
	}
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
	fmt.Println("Checking all system requirements...")

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
