package nginx

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/amoga-io/run/internals/utils"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "nginx",
	Short: "Install Nginx",
	Long:  `Install Nginx on your system. This command will install Nginx using a provided script.`,

	Run: func(cmd *cobra.Command, args []string) {
		println("Nginx installation process started...")

		scriptPath := utils.GetScriptPath("nginx.sh")
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			fmt.Fprintf(cmd.ErrOrStderr(), "Nginx installation script not found at %s\n", scriptPath)
		}

		if err := os.Chmod(scriptPath, 0755); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Failed to make script executable: %v\n", err)
			return
		}

		execCmd := exec.Command("bash", scriptPath)

		execCmd.Stdout = cmd.OutOrStdout()
		execCmd.Stderr = cmd.ErrOrStderr()

		if err := execCmd.Run(); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error executing Nginx installation script: %v\n", err)
			return
		}
	},
}

func init() {
	// No parent to add to here; will be added in root.go
}
