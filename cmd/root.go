// Package cmd 提供 gsus 工具的所有命令行命令定义.
// 本包使用 Cobra 框架构建命令行界面，仅负责命令结构定义和参数绑定，不包含业务逻辑.
package cmd

import (
	"bytes"
	"context"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/x/exp/charmtone"
	"github.com/charmbracelet/x/term"
	"github.com/spelens-gud/gsus/internal/logger"
	"github.com/spelens-gud/gsus/internal/version"
	"github.com/spf13/cobra"
)

// commandName 命令名称.
const commandName = "gsus"

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   commandName,
	Short: "Go 代码生成工具",
	Long: `gsus 是一个强大的 Go 代码生成工具，支持：
- 从数据库表生成结构体 (db2struct)
- 生成 HTTP 客户端代码 (http client)
- 生成 HTTP 路由代码 (http router)
- 生成接口实现代码 (impl)
- 生成枚举类型代码 (enum)`,
}

// versionBit var    版本信息的 ASCII 艺术字样式.
var versionBit = lipgloss.NewStyle().Foreground(charmtone.Zinc).SetString(`
  ___  ____  _  _  ____
 / __)/ ___)/ )( \/ ___)
( (_ \\___ \) \/ (\___ \
 \___/(____/\____/(____/
`)

// defaultVersionTemplate 默认版本信息模板.
const defaultVersionTemplate = `{{with .DisplayName}}{{printf "%s " .}}{{end}}{{printf "version %s" .Version}}

`

// Execute function    执行根命令.
func Execute() {
	// 设置版本模板
	if term.IsTerminal(os.Stdout.Fd()) {
		var b bytes.Buffer
		w := colorprofile.NewWriter(os.Stdout, os.Environ())
		w.Forward = &b
		_, _ = w.WriteString(versionBit.String())
		rootCmd.SetVersionTemplate(b.String() + "\n" + defaultVersionTemplate)
	}

	// 执行命令
	if err := fang.Execute(
		context.Background(),
		rootCmd,
		fang.WithVersion(version.Version),
		fang.WithNotifySignal(os.Interrupt),
	); err != nil {
		logger.Error("command execution failed: %v", err)
		os.Exit(1)
	}
}

// init function    初始化根命令.
// 在此函数中可以定义全局标志和配置.
func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gsus.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
