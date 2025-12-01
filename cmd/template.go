package cmd

/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/

import (
	"github.com/spelens-gud/gsus/internal/runner"
	"github.com/spf13/cobra"
)

var (
	templateGenAll    bool
	templateOverwrite bool
)

// templateCmd represents the template command.
var templateCmd = &cobra.Command{
	Use:   "template [models...]",
	Short: "模板管理",
	Long:  `根据模型生成代码，支持自定义模板`,
	Run: func(cmd *cobra.Command, args []string) {
		runner.RunAutoTemplate(&runner.TemplateOptions{
			Models:    args,
			GenAll:    templateGenAll,
			Overwrite: templateOverwrite,
		})
	},
}

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
