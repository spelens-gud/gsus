package cmd

import (
	"github.com/spf13/cobra"
)

// httpCmd var    HTTP 相关代码生成命令.
// 该命令是一个父命令，包含 client 和 router 两个子命令.
// 用于生成 HTTP 客户端或路由相关代码.
var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "HTTP 相关代码生成",
	Long:  `生成 HTTP 客户端或路由相关代码`,
}

// init function    初始化 http 命令.
// 将 http 命令注册为根命令的子命令.
func init() {
	rootCmd.AddCommand(httpCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// httpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// httpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
