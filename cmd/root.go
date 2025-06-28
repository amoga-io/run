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
	Use:   "run",
	Short: "Ubuntu Server Package Manager",
	Long: `A safe and intelligent CLI tool for managing developer and system packages on Ubuntu systems.

Features:
  • Install development tools (Node.js, Python, Java, PHP)
  • Manage system services (Docker, Nginx, PostgreSQL)
  • Version management for supported packages
  • Automatic dependency resolution
  • Safe package removal with rollback support

Examples:
  run install node python docker
  run install node --version 20
  run check --all
  run remove docker --force`,
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
	// Add all commands to root
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(internalCmd)
}
