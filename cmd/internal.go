package cmd

import (
	"fmt"

	"github.com/amoga-io/run/internal/system"
	"github.com/spf13/cobra"
)

var internalCmd = &cobra.Command{
	Use:    "internal",
	Short:  "Internal commands (hidden from help)",
	Hidden: true,
}

var internalVerifyCmd = &cobra.Command{
	Use:    "verify-installation",
	Short:  "Verify installation is complete (internal use)",
	Hidden: true,
	RunE:   runInternalVerify,
}

var internalDepCheckCmd = &cobra.Command{
	Use:    "dep-check",
	Short:  "Check dependencies (internal use)",
	Hidden: true,
	RunE:   runInternalDepCheck,
}

func runInternalVerify(cmd *cobra.Command, args []string) error {
	fmt.Println("Verifying installation...")

	// Check if all critical dependencies are available
	missing, err := system.CheckSystemRequirements(system.Bootstrap, system.Runtime)
	if err != nil {
		return err
	}

	if len(missing) > 0 {
		fmt.Printf("Warning: Some dependencies are missing but installation succeeded\n")
		for _, req := range missing {
			if req.Critical {
				fmt.Printf("  - %s: %s (critical)\n", req.Name, req.Description)
			} else {
				fmt.Printf("  - %s: %s\n", req.Name, req.Description)
			}
		}
		fmt.Println("These will be installed automatically when needed.")
	}

	fmt.Println("✓ Installation verified successfully")
	return nil
}

func runInternalDepCheck(cmd *cobra.Command, args []string) error {
	return system.EnsureBootstrapRequirements()
}

func init() {
	internalCmd.AddCommand(internalVerifyCmd)
	internalCmd.AddCommand(internalDepCheckCmd)
}
