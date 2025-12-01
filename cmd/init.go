package cmd

/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/

import (
	"github.com/spelens-gud/gsus/internal/runner"
	"github.com/spf13/cobra"
)

// initCmd represents the init command.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化 gsus 项目配置",
	Long:  `在当前项目中初始化 gsus 配置文件和模板目录`,
	Run: func(cmd *cobra.Command, args []string) {
		runner.RunAutoInit(&runner.InitOptions{})
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
