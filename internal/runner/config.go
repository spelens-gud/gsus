package runner

import (
	"context"

	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/logger"
)

// ConfigOptions struct    配置选项.
type ConfigOptions struct {
	// 预留扩展字段
}

func Config(ctx context.Context, opts *ConfigOptions) error {
	log := logger.WithPrefix("[config]")
	log.Info("开始执行 config 代码生成")

	cfg, err := config.Get()
	if err != nil {
		return err
	}

	// 配置相关的业务逻辑
	// 目前原实现为空，保留接口
	_ = cfg

	log.Info("config 模版生成成功")
	return nil
}

// RunAutoConfig function    执行配置操作.
func RunAutoConfig(opts *ConfigOptions) {
	config.ExecuteWithConfig(func(_ config.Option) (err error) {
		return Config(context.Background(), opts)
	})
}
