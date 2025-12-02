package cmd

import (
	"github.com/spelens-gud/gsus/internal/runner"
	"github.com/spf13/cobra"
)

// db2structCmd var    数据库表转结构体命令.
// 该命令用于根据数据库表结构自动生成对应的 Go 结构体代码.
// 支持指定特定表名，如果不指定则生成所有表的结构体.
var db2structCmd = &cobra.Command{
	Use:   "db2struct [tables...]",
	Short: "从数据库表生成 Go 结构体",
	Long:  `根据配置文件中的数据库连接信息，将数据库表转换为 Go 结构体代码。如果不指定表名，则生成所有表的结构体。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 调用 runner 执行数据库表转结构体逻辑
		runner.RunAutoDb2Struct(&runner.Db2structOptions{
			Tables: args,
		})
	},
}

// init function    初始化 db2struct 命令.
// 将 db2struct 命令注册为根命令的子命令.
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
