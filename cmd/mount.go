package cmd

/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/

import (
	"github.com/spelens-gud/gsus/internal/runner"
	"github.com/spf13/cobra"
)

// mountCmd represents the mount command.
var mountCmd = &cobra.Command{
	Use:   "mount [args...]",
	Short: "挂载相关操作",
	Long:  `执行挂载相关的代码生成操作`,
	Run: func(cmd *cobra.Command, args []string) {
		runner.RunAutoMount(&runner.MountOptions{
			Args: args,
		})
	},
}

func init() {
	rootCmd.AddCommand(mountCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mountCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// mountCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
