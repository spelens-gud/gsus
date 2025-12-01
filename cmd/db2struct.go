package cmd

/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/

import (
	"github.com/spelens-gud/gsus/internal/runner"
	"github.com/spf13/cobra"
)

// db2structCmd represents the db2struct command.
var db2structCmd = &cobra.Command{
	Use:   "db2struct [tables...]",
	Short: "从数据库表生成 Go 结构体",
	Long:  `根据配置文件中的数据库连接信息，将数据库表转换为 Go 结构体代码`,
	Run: func(cmd *cobra.Command, args []string) {
		runner.RunAutoDb2Struct(&runner.Db2structOptions{
			Tables: args,
		})
	},
}

func init() {
	rootCmd.AddCommand(db2structCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// db2structCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// db2structCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
