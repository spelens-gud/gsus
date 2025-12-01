package cmd

/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/

import (
	"fmt"

	"github.com/spf13/cobra"
)

// db2structCmd represents the db2struct command.
var db2structCmd = &cobra.Command{
	Use:   "db2struct",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("db2struct called")
	},
}

func init() {
	rootCmd.AddCommand(db2structCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// db2structCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// db2structCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
