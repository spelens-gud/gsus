package runner

import (
	"github.com/spelens-gud/gsus/internal/config"
)

// EnumOptions struct    枚举生成选项.
type EnumOptions struct {
	// 预留扩展字段
}

// RunAutoEnum function    执行枚举生成.
func RunAutoEnum(opts *EnumOptions) {
	config.ExecuteWithConfig(func(_ config.Option) (err error) {
		cfg, err := config.Get()
		if err != nil {
			return err
		}

		// 枚举生成的业务逻辑
		_ = cfg
		return nil
	})
}
