package cmd

import (
	"github.com/spelens-gud/gsus/internal/runner"
	"github.com/spf13/cobra"
)

var implPrefix string

// implCmd represents the impl command.
var implCmd = &cobra.Command{
	Use:   "impl [interface] [struct]",
	Short: "生成接口实现代码",
	Long:  `根据接口定义自动生成实现代码骨架，支持自定义文件目录前缀`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		runner.RunAutoImpl(&runner.ImplOptions{
			Interface: args[0],
			Struct:    args[1],
			Prefix:    implPrefix,
		})
	},
}

func init() {
	rootCmd.AddCommand(implCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// implCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// implCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	implCmd.Flags().StringVarP(&implPrefix, "prefix", "p", "", "实现文件目录前缀")
}
