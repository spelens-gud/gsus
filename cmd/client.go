package cmd

/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/

import (
	"github.com/spelens-gud/gsus/internal/runner"
	"github.com/spf13/cobra"
)

// clientCmd represents the client command.
var clientCmd = &cobra.Command{
	Use:   "client [service]",
	Short: "生成 HTTP 客户端代码",
	Long:  `根据配置生成 HTTP 客户端相关代码`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runner.RunAutoClient(&runner.ClientOptions{
			Args: args[0],
		})
	},
}

func init() {
	httpCmd.AddCommand(clientCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clientCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// clientCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
