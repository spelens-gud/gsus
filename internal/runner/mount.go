package runner

import (
	"context"
	"fmt"

	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/errors"
	"github.com/spelens-gud/gsus/internal/generator"
	"github.com/spelens-gud/gsus/internal/logger"
)

// MountOptions struct    挂载选项.
type MountOptions struct {
	Args []string // 挂载参数
}

// Mount function.
func Mount(ctx context.Context, opts *MountOptions) error {
	log := logger.WithPrefix("[mount]")
	log.Info("开始执行 mount 代码生成")

	argsMap := make(map[string]bool, len(opts.Args))
	for _, arg := range opts.Args {
		argsMap[arg] = true
	}
	if err := generator.Exec(config.Mount{
		Args: opts.Args,
	}); err != nil {
		log.Error("mount 生成错误")
		return errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("mount 生成错误: %v", err))
	}

	log.Info("mount 模板生成成功")
	return nil
}
func RunAutoMount(opts *MountOptions) {
	config.ExecuteWithConfig(func(_ config.Option) (err error) {
		return Mount(context.Background(), opts)
	})
}
