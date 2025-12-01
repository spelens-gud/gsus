package runner

import (
	"github.com/spelens-gud/gsus/internal/config"
)

// ConfigOptions struct    配置选项.
type ConfigOptions struct {
	// 预留扩展字段
}

// RunAutoConfig function    执行配置操作.
func RunAutoConfig(opts *ConfigOptions) {
	config.ExecuteWithConfig(func(_ config.Option) (err error) {
		cfg, err := config.Get()
		if err != nil {
			return err
		}

		// 配置相关的业务逻辑
		// 目前原实现为空，保留接口
		_ = cfg
		return nil
	})
}
