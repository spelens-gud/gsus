package cmd

import (
	"github.com/spelens-gud/gsus/internal/runner"
	"github.com/spf13/cobra"
)

// enumCmd var    枚举类型代码生成命令.
// 该命令用于根据配置文件自动生成 Go 枚举类型代码，支持自定义枚举值和相关方法.
var enumCmd = &cobra.Command{
	Use:   "enum",
	Short: "生成枚举类型代码",
	Long:  `根据配置文件生成 Go 枚举类型代码，支持自定义枚举值和方法`,
	Run: func(cmd *cobra.Command, args []string) {
		// 调用 runner 执行枚举类型代码生成逻辑
		runner.RunAutoEnum(&runner.EnumOptions{})
	},
}

// init function    初始化 enum 命令.
// 将 enum 命令注册为根命令的子命令.
func init() {
	rootCmd.AddCommand(enumCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// enumCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// enumCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
