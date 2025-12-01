package runner

import (
	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/generator"
)

// MountOptions struct    挂载选项.
type MountOptions struct {
	Args []string // 挂载参数
}

func RunAutoMount(opts *MountOptions) {
	config.ExecuteWithConfig(func(_ config.Option) (err error) {
		argsMap := make(map[string]bool, len(opts.Args))
		for _, arg := range opts.Args {
			argsMap[arg] = true
		}
		return generator.Exec(config.Mount{
			Args: opts.Args,
		})
	})
}
