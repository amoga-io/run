package run

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "run [option]",
	Short: "Run different scripts using subcommands",
	Long:  `Run different scripts based on the subcommand provided.`,
}

func init() {
	// No parent to add to here; will be added in root.go
}
