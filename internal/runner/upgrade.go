package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/logger"
)

// UpgradeOptions struct    升级选项.
type UpgradeOptions struct {
	Verbose bool // 是否显示详细日志
}

func Upgrade(ctx context.Context, opts *UpgradeOptions) (err error) {
	log := logger.WithPrefix("[upgrade]")
	log.Info("开始执行 upgrade 代码生成")

	vStr := ""
	if opts.Verbose {
		vStr = "-v"
	}

	log.Info("updating gsus from [ %s ] ...", config.GoGetUrl)
	if err = execCommand("go", "get", "-u", vStr, config.GoGetUrl); err == nil {
		return nil
	}

	if !strings.Contains(err.Error(), "uses insecure protocol") {
		log.Error("升级失败")
		return nil
	}

	if err := execCommand("go", "get", "--insecure", "-u", vStr, config.GoGetUrl); err == nil {
		log.Info("升级成功")
		return nil
	}

	return err
}

// RunAutoUpgrade function    执行升级操作.
func RunAutoUpgrade(opts *UpgradeOptions) {
	config.ExecuteWithConfig(func(_ config.Option) (err error) {
		return Upgrade(context.Background(), opts)
	})
}

func execCommand(name string, commands ...string) error {
	var idx int
	for _, c := range commands {
		if len(c) > 0 {
			commands[idx] = c
			idx++
		}
	}
	commands = commands[:idx]
	command := exec.Command(name, commands...)
	command.Stdout = os.Stdout
	bf := new(bytes.Buffer)
	command.Stderr = io.MultiWriter(os.Stderr, bf)
	if err := command.Run(); err != nil {
		return fmt.Errorf("%s: %v", bf.String(), err)
	}
	return nil
}
