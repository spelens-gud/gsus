package cmd

import (
	"github.com/spelens-gud/gsus/internal/runner"
	"github.com/spf13/cobra"
)

var (
	// templateGenAll var    是否生成所有模型.
	templateGenAll bool
	// templateOverwrite var    是否覆盖已存在的文件.
	templateOverwrite bool
)

// templateCmd var    模板管理命令.
// 该命令用于根据模型定义生成代码，支持使用自定义模板.
// 可以指定特定模型，也可以通过 --all 标志生成所有模型.
var templateCmd = &cobra.Command{
	Use:   "template [models...]",
	Short: "模板管理",
	Long:  `根据模型生成代码，支持自定义模板`,
	Run: func(cmd *cobra.Command, args []string) {
		// 调用 runner 执行模板代码生成逻辑
		runner.RunAutoTemplate(&runner.TemplateOptions{
			Models:    args,
			GenAll:    templateGenAll,
			Overwrite: templateOverwrite,
		})
	},
}

// init function    初始化 template 命令.
// 将 template 命令注册为根命令的子命令，并定义命令标志.
func init() {
	rootCmd.AddCommand(templateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// templateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// templateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	templateCmd.Flags().BoolVar(&templateGenAll, "all", false, "generate all model")
	templateCmd.Flags().BoolVar(&templateOverwrite, "overwrite", false, "overwrite files")
}
