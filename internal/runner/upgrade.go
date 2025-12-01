package runner

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spelens-gud/gsus/apis/constant"
	"github.com/spelens-gud/gsus/apis/helpers/executor"
	"github.com/spelens-gud/gsus/internal/config"
)

// UpgradeOptions struct    升级选项.
type UpgradeOptions struct {
	Verbose bool // 是否显示详细日志
}

// RunAutoUpgrade function    执行升级操作.
func RunAutoUpgrade(opts *UpgradeOptions) {
	executor.ExecuteWithConfig(func(_ config.Option) (err error) {
		vStr := ""
		if opts.Verbose {
			vStr = "-v"
		}

		log.Printf("updating gsus from [ %s ] ...", constant.GoGetUrl)
		if err = execCommand("go", "get", "-u", vStr, constant.GoGetUrl); err == nil {
			return nil
		}

		if !strings.Contains(err.Error(), "uses insecure protocol") {
			return nil
		}

		if err = execCommand("go", "get", "--insecure", "-u", vStr, constant.GoGetUrl); err == nil {
			return nil
		}

		return err
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
