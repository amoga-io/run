package python

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/amoga-io/run/internals/utils"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "python",
	Short: "Install python",
	Long:  `Install python on your system. This command will install Python using a provided script.`,

	Run: func(cmd *cobra.Command, args []string) {
		println("python installation process started...")

		// Get the script path for python installation
		scriptPath := utils.GetScriptPath("python.sh")
		if _, err := os.Stat(scriptPath); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "python installation script not found at %s\n", scriptPath)
			return
		}

		// Make the script executable
		if err := os.Chmod(scriptPath, 0755); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Failed to make script executable: %v\n", err)
			return
		}

		// Execute the script
		execCmd := exec.Command("bash", scriptPath)

		// Set the output and error streams for the command
		execCmd.Stdout = cmd.OutOrStdout()
		execCmd.Stderr = cmd.ErrOrStderr()

		// Run the command and check for errors
		if err := execCmd.Run(); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error executing python installation script: %v\n", err)
			return
		}
	},
}

func init() {}
