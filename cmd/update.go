package cmd

/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/

import (
	"github.com/spelens-gud/gsus/internal/runner"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command.
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "更新模板和配置",
	Long:  `更新项目中的模板文件和配置`,
	Run: func(cmd *cobra.Command, args []string) {
		runner.RunAutoUpdate(&runner.UpdateOptions{})
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
