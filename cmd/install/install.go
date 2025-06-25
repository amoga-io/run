package install

import (
	"fmt"

	"github.com/amoga-io/run/cmd/install/docker"
	"github.com/amoga-io/run/cmd/install/nginx"
	"github.com/amoga-io/run/cmd/install/node"
	"github.com/amoga-io/run/cmd/install/php"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{

	Use:   "install [option]",
	Short: "Install different tools",
	Long:  `Install different tools or dependencies based on the subcommand provided.`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Please specify a tool to install. Use 'install --help' for more information.")
	},
}

func init() {
	Cmd.AddCommand(docker.Cmd)
	Cmd.AddCommand(nginx.Cmd)
	Cmd.AddCommand(node.Cmd)
	Cmd.AddCommand(php.Cmd)

	// No parent to add to here; will be added in root.go
}
