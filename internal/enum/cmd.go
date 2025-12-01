package enum

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use: "enum",
	Run: Run,
}

func Run(cmd *cobra.Command, args []string) {

}
