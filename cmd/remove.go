/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ssk-amoga/devkit/internal"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a package",
	Long:  `Remove a package from your specific method.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check --all flag first
		if allFlag, _ := cmd.Flags().GetBool("all"); allFlag {
			fmt.Println("Removing all packages...")
			for packageName := range internal.RemovePackageRegistry {
				fmt.Printf("Removing package: %s\n", packageName)
				if err := internal.GetScriptAndExecute("install", packageName); err != nil {
					fmt.Printf("Error removing package '%s': %v\n", packageName, err)
				} else {
					fmt.Printf("Successfully removed package: %s\n", packageName)
				}
			}
			return
		}

		// No args provided and --all flag not set
		if len(args) == 0 {
			fmt.Println("Please specify a package to remove or use --all flag to remove all installed packages.")
			return
		}

		// Multiple packages provided
		if len(args) > 1 {
			for _, packageName := range args {
				fmt.Printf("Removing package: %s\n", packageName)
				if err := internal.GetScriptAndExecute("remove", packageName); err != nil {
					fmt.Printf("Error removing package '%s': %v\n", packageName, err)
				} else {
					fmt.Printf("Successfully removed package: %s\n", packageName)
				}
			}
			return
		}

		// Install single package
		packageName := args[0]
		fmt.Printf("Removing package: %s\n", packageName)
		if err := internal.GetScriptAndExecute("remove", packageName); err != nil {
			fmt.Printf("Error removing package '%s': %v\n", packageName, err)
		} else {
			fmt.Printf("Successfully removed package: %s\n", packageName)
		}
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().BoolP("all", "A", false, "remove all packages")
}
