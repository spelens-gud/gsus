package config

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use: "config",
	Run: Run,
}

func Run(cmd *cobra.Command, args []string) {
}
