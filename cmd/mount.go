package cmd

import (
	"github.com/spelens-gud/gsus/internal/runner"
	"github.com/spf13/cobra"
)

// mountCmd var    挂载相关操作命令.
// 该命令用于执行挂载相关的代码生成操作.
// 支持传入多个参数以指定挂载的具体内容.
var mountCmd = &cobra.Command{
	Use:   "mount [args...]",
	Short: "挂载相关操作",
	Long:  `执行挂载相关的代码生成操作`,
	Run: func(cmd *cobra.Command, args []string) {
		// 调用 runner 执行挂载操作逻辑
		runner.RunAutoMount(&runner.MountOptions{
			Args: args,
		})
	},
}

// init function    初始化 mount 命令.
// 将 mount 命令注册为根命令的子命令.
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
