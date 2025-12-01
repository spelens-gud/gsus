package update

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use: "update",
	Run: run,
}

func run(_ *cobra.Command, args []string) {
}
