package cmd

import (
	"github.com/spelens-gud/gsus/internal/runner"
	"github.com/spf13/cobra"
)

// initCmd var    项目初始化命令.
// 该命令用于在当前项目中初始化 gsus 配置文件和模板目录.
// 执行后会创建必要的配置文件和目录结构.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化 gsus 项目配置",
	Long:  `在当前项目中初始化 gsus 配置文件和模板目录`,
	Run: func(cmd *cobra.Command, args []string) {
		// 调用 runner 执行项目初始化逻辑
		runner.RunAutoInit(&runner.InitOptions{})
	},
}

// init function    初始化 init 命令.
// 将 init 命令注册为根命令的子命令.
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
