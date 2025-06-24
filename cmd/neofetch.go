/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// neofetchCmd represents the neofetch command
var neofetchCmd = &cobra.Command{
	Use:   "neofetch",
	Short: "Display system information using Neofetch",

	Run: func(cmd *cobra.Command, args []string) {
		scriptPath := GetScriptPath("neofetch.sh")

		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			fmt.Fprintf(cmd.ErrOrStderr(), "Script not found: %s\n", scriptPath)
			return
		}

		if err := os.Chmod(scriptPath, 0755); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error setting script permissions: %v\n", err)
			return
		}

		execCmd := exec.Command("bash", scriptPath)

		execCmd.Stdout = cmd.OutOrStdout()
		execCmd.Stderr = cmd.OutOrStderr()

		fmt.Println("Running Neofetch script...")

		if err := execCmd.Run(); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error executing script: %v\n", err)
			return
		}

		fmt.Println("Neofetch executed successfully!")
	},
}

func init() {
	installCmd.AddCommand(neofetchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// neofetchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// neofetchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
