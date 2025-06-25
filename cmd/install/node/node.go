package node

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/amoga-io/run/internals/utils"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "node",
	Short: "Install Node.js",
	Long:  `Install Node.js on your system. This command will install Node.js using a provided script.`,

	Run: func(cmd *cobra.Command, args []string) {
		println("Node.js installation process started...")

		scriptPath := utils.GetScriptPath("node.sh")
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			fmt.Fprintf(cmd.ErrOrStderr(), "Node.js installation script not found at %s\n", scriptPath)
		}

		if err := os.Chmod(scriptPath, 0755); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Failed to make script executable: %v\n", err)
			return
		}

		execCmd := exec.Command("bash", scriptPath)

		execCmd.Stdout = cmd.OutOrStdout()
		execCmd.Stderr = cmd.ErrOrStderr()

		if err := execCmd.Run(); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error executing Node.js installation script: %v\n", err)
			return
		}
	},
}

func init() {
	// No parent to add to here; will be added in install.go
}
