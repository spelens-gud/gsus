package cmd

/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/

import (
	"github.com/spelens-gud/gsus/internal/runner"
	"github.com/spf13/cobra"
)

// routerCmd represents the router command.
var routerCmd = &cobra.Command{
	Use:   "router [service]",
	Short: "生成 HTTP 路由代码",
	Long:  `根据配置生成 HTTP 路由相关代码`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runner.RunAutoRouter(&runner.RouterOptions{
			Args: args[0],
		})
	},
}

func init() {
	httpCmd.AddCommand(routerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// routerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// routerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
