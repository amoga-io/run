package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ssk-amoga/devkit/internal"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a package",
	Long:  `Install a package in your specific method.`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		// Check --all flag first
		if allFlag, _ := cmd.Flags().GetBool("all"); allFlag {
			fmt.Println("Installing all packages...")
			for packageName := range internal.InstallPackageRegistry {
				fmt.Printf("Installing package: %s\n", packageName)
				if err := internal.GetScriptAndExecute("install", packageName); err != nil {
					fmt.Printf("Error installing package '%s': %v\n", packageName, err)
				} else {
					fmt.Printf("Successfully installed package: %s\n", packageName)
				}
			}
			return
		}

		// No args provided and --all flag not set
		if len(args) == 0 {
			fmt.Println("Please specify a package to install or use --all flag to install all packages.")
			return
		}

		// Multiple packages provided
		if len(args) > 1 {
			for _, packageName := range args {
				fmt.Printf("Installing package: %s\n", packageName)
				if err := internal.GetScriptAndExecute("install", packageName); err != nil {
					fmt.Printf("Error installing package '%s': %v\n", packageName, err)
				} else {
					fmt.Printf("Successfully installed package: %s\n", packageName)
				}
			}
			return
		}

		// Install single package
		packageName := args[0]
		fmt.Printf("Installing package: %s\n", packageName)
		if err := internal.GetScriptAndExecute("install", packageName); err != nil {
			fmt.Printf("Error installing package '%s': %v\n", packageName, err)
		} else {
			fmt.Printf("Successfully installed package: %s\n", packageName)
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().BoolP("all", "a", false, "install all packages")
}
