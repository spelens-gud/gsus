package runner

import (
	"context"

	"github.com/spelens-gud/gsus/internal/config"
)

// EnumOptions struct    枚举生成选项.
type EnumOptions struct {
	// 预留扩展字段
}

// Enum function    执行枚举生成.
func Enum(ctx context.Context, opts *EnumOptions, cfg config.Option) error {
	// TODO: 实现枚举生成的业务逻辑
	_ = cfg
	return nil
}

// RunAutoEnum function    执行枚举生成（兼容旧接口）.
func RunAutoEnum(opts *EnumOptions) {
	config.ExecuteWithConfig(func(cfg config.Option) error {
		return Enum(context.Background(), opts, cfg)
	})
}
