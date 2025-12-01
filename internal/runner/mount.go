package runner

import (
	"github.com/spelens-gud/gsus/apis/helpers/executor"
	"github.com/spelens-gud/gsus/apis/mount"
	"github.com/spelens-gud/gsus/internal/config"
)

// MountOptions struct    挂载选项.
type MountOptions struct {
	Args []string // 挂载参数
}

func RunAutoMount(opts *MountOptions) {
	executor.ExecuteWithConfig(func(_ config.Option) (err error) {
		argsMap := make(map[string]bool, len(opts.Args))
		for _, arg := range opts.Args {
			argsMap[arg] = true
		}
		return mount.Exec(mount.Config{
			Args: opts.Args,
		})
	})
}
