package cmd

import (
	"github.com/spelens-gud/gsus/internal/runner"
	"github.com/spf13/cobra"
)

// routerCmd var    HTTP 路由代码生成命令.
// 该命令用于根据服务接口定义自动生成 HTTP 路由注册代码.
// 需要提供服务路径作为参数.
var routerCmd = &cobra.Command{
	Use:   "router [service]",
	Short: "生成 HTTP 路由代码",
	Long:  `根据服务定义生成 HTTP 路由注册代码`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 调用 runner 执行路由代码生成逻辑
		runner.RunAutoRouter(&runner.RouterOptions{
			RouterPath: args[0],
		})
	},
}

// init function    初始化 router 命令.
// 将 router 命令注册为 http 命令的子命令.
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
