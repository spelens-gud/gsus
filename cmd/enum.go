package cmd

import (
	"github.com/spelens-gud/gsus/internal/runner"
	"github.com/spf13/cobra"
)

// enumCmd represents the enum command.
var enumCmd = &cobra.Command{
	Use:   "enum",
	Short: "生成枚举类型代码",
	Long:  `根据配置文件生成 Go 枚举类型代码，支持自定义枚举值和方法`,
	Run: func(cmd *cobra.Command, args []string) {
		runner.RunAutoEnum(&runner.EnumOptions{})
	},
}

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
