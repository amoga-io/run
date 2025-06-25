package docker

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/amoga-io/run/internals/utils"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "docker",
	Short: "Install Docker",
	Long:  `Install Docker on your system. This command will install Docker using a provided script.`,
	Run: func(cmd *cobra.Command, args []string) {
		println("Docker installation process started...")

		scriptPath := utils.GetScriptPath("docker.sh")
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			fmt.Fprintf(cmd.ErrOrStderr(), "Docker installation script not found at %s\n", scriptPath)
			return
		}

		if err := os.Chmod(scriptPath, 0755); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Failed to make script executable: %v\n", err)
			return
		}

		execCmd := exec.Command("bash", scriptPath)

		execCmd.Stdout = cmd.OutOrStdout()
		execCmd.Stderr = cmd.ErrOrStderr()

		if err := execCmd.Run(); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error executing Docker installation script: %v\n", err)
			return
		}
	},
}

func init() {
	// No parent to add to here; will be added in root.go
	// This command is a subcommand of the install command
}
