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

var arg string

// helloCmd represents the hello command
var helloCmd = &cobra.Command{
	Use:   "hello",
	Short: "Run a simple hello world script",

	Run: func(cmd *cobra.Command, args []string) {
		scriptPath := GetScriptPath("hello.sh")

		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			fmt.Fprintf(cmd.ErrOrStderr(), "Script not found: %s\n", scriptPath)
			return
		}

		if err := os.Chmod(scriptPath, 0755); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error setting script permissions: %v\n", err)
			return
		}

		var execCmd *exec.Cmd

		if arg != "" {
			execCmd = exec.Command("bash", scriptPath, arg)
		} else {
			execCmd = exec.Command("bash", scriptPath)
		}

		execCmd.Stdout = cmd.OutOrStdout()
		execCmd.Stderr = cmd.OutOrStderr()

		fmt.Println("Running hello script...")

		if err := execCmd.Run(); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error executing script: %v\n", err)
			return
		}

		fmt.Println("Hello script executed successfully!")
	},
}

func init() {
	runCmd.AddCommand(helloCmd)

	// Add -a / --arg flag
	helloCmd.Flags().StringVarP(&arg, "arg", "a", "", "Argument to pass to hello script")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// helloCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// helloCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
