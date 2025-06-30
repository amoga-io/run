/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "run",
	Short: "Run is a CLI tool to manage your development environment",
	Long:  `Run is a command-line tool for managing development tools and packages using the apt package manager. It supports installing, removing, listing, and searching packages.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, display help
		if len(args) == 0 {
			cmd.Help()
			return
		}

		// TODO: Implement correct version logic
		if versionFlag, _ := cmd.Flags().GetBool("version"); versionFlag {
			cmd.Println("Run version 1.0.0")
			return
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// verifyCmd represents the verify command for installation verification
var verifyCmd = &cobra.Command{
	Use:    "verify",
	Short:  "Verify installation",
	Hidden: true, // Hide from help menu
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("run CLI is installed and working correctly")
	},
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("Run version 1.0.0")
	},
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().BoolP("version", "v", false, "Display run version")

	// Add subcommands to root command
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(versionCmd)
}
