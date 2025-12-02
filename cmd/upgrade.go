package cmd

import (
	"github.com/spelens-gud/gsus/internal/runner"
	"github.com/spf13/cobra"
)

// upgradeVerbose var    是否输出详细日志.
var upgradeVerbose bool

// upgradeCmd var    升级工具命令.
// 该命令用于将 gsus 工具升级到最新版本.
// 支持通过 -v 标志输出详细的升级日志.
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "升级 gsus 工具",
	Long:  `升级 gsus 工具到最新版本`,
	Run: func(cmd *cobra.Command, args []string) {
		// 调用 runner 执行升级逻辑
		runner.RunAutoUpgrade(&runner.UpgradeOptions{
			Verbose: upgradeVerbose,
		})
	},
}

// init function    初始化 upgrade 命令.
// 将 upgrade 命令注册为根命令的子命令，并定义命令标志.
func init() {
	rootCmd.AddCommand(upgradeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// upgradeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// upgradeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	upgradeCmd.Flags().BoolVarP(&upgradeVerbose, "verbose", "v", false, "log more detail")
}
