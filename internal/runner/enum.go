package runner

import (
	"context"

	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/logger"
)

// EnumOptions struct    枚举生成选项.
type EnumOptions struct {
	// 预留扩展字段
}

// Enum function    执行枚举生成.
func Enum(ctx context.Context, opts *EnumOptions, cfg config.Option) error {
	log := logger.WithPrefix("[enum]")
	log.Info("开始执行 enum 代码生成")

	// 配置相关的业务逻辑
	_ = cfg

	log.Info("enum 代码生成完成")
	return nil
}

// RunAutoEnum function    执行枚举生成（兼容旧接口）.
func RunAutoEnum(opts *EnumOptions) {
	config.ExecuteWithConfig(func(cfg config.Option) error {
		return Enum(context.Background(), opts, cfg)
	})
}
