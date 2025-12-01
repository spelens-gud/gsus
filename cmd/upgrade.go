package cmd

/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/

import (
	"github.com/spelens-gud/gsus/internal/runner"
	"github.com/spf13/cobra"
)

var upgradeVerbose bool

// upgradeCmd represents the upgrade command.
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "升级 gsus 工具",
	Long:  `升级 gsus 工具到最新版本`,
	Run: func(cmd *cobra.Command, args []string) {
		runner.RunAutoUpgrade(&runner.UpgradeOptions{
			Verbose: upgradeVerbose,
		})
	},
}

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
