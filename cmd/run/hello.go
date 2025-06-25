package run

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/amoga-io/run/internals/utils"
	"github.com/spf13/cobra"
)

var arg string

var HelloCmd = &cobra.Command{
	Use:   "hello",
	Short: "Run a simple hello world script",
	Run: func(cmd *cobra.Command, args []string) {
		scriptPath := utils.GetScriptPath("hello.sh")
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
		if err := execCmd.Run(); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error executing script: %v\n", err)
			return
		}
	},
}

func init() {
	HelloCmd.Flags().StringVarP(&arg, "arg", "a", "", "Argument to pass to hello script")
	Cmd.AddCommand(HelloCmd)
}
