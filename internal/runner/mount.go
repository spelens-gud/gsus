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
// 包含挂载操作所需的参数列表.
type MountOptions struct {
	Args []string // 挂载参数列表
}

// Mount function    执行挂载操作.
// 根据提供的参数执行代码挂载生成.
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

// RunAutoMount function    执行挂载操作（兼容旧接口）.
// 自动加载配置并执行挂载逻辑.
func RunAutoMount(opts *MountOptions) {
	config.ExecuteWithConfig(func(_ config.Option) (err error) {
		return Mount(context.Background(), opts)
	})
}
