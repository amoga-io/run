package install

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/amoga-io/run/internals/utils"
	"github.com/spf13/cobra"
)

var DockerInstallCmd = &cobra.Command{
	Use:   "docker",
	Short: "Install Docker",
	Long:  `Install Docker on your system. This command will guide you through the installation process.`,
	Run: func(cmd *cobra.Command, args []string) {
		println("Docker installation process started...")

		scriptPath := utils.GetScriptPath("docker.sh")
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			fmt.Fprintf(cmd.ErrOrStderr(), "Docker installation script not found at %s\n", scriptPath)
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
	Cmd.AddCommand(DockerInstallCmd)
}
