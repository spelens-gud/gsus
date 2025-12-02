package cmd

import (
	"github.com/spelens-gud/gsus/internal/runner"
	"github.com/spf13/cobra"
)

// configCmd var    配置管理命令.
// 该命令用于管理 gsus 项目的配置文件和配置项.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置管理",
	Long:  `管理 gsus 项目配置`,
	Run: func(cmd *cobra.Command, args []string) {
		// 调用 runner 执行配置管理逻辑
		runner.RunAutoConfig(&runner.ConfigOptions{})
	},
}

// init function    初始化 config 命令.
// 将 config 命令注册为根命令的子命令.
func init() {
	rootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
