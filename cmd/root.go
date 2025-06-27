/*
Copyright © 2025 Amoga
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "run",
	Short:   "A Git-based CLI for Ubuntu systems",
	Long:    "A simple Git-based CLI tool that can update itself from GitHub repository.",
	Version: "1.0.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Only add the update command for now
	rootCmd.AddCommand(updateCmd)
}
