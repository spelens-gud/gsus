package cmd

import (
	"github.com/spelens-gud/gsus/internal/runner"
	"github.com/spf13/cobra"
)

// updateCmd var    更新模板和配置命令.
// 该命令用于更新项目中的模板文件和配置文件到最新版本.
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "更新模板和配置",
	Long:  `更新项目中的模板文件和配置`,
	Run: func(cmd *cobra.Command, args []string) {
		// 调用 runner 执行更新逻辑
		runner.RunAutoUpdate(&runner.UpdateOptions{})
	},
}

// init function    初始化 update 命令.
// 将 update 命令注册为根命令的子命令.
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
