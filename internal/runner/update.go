package runner

import (
	"github.com/spelens-gud/gsus/apis/helpers/executor"
	"github.com/spelens-gud/gsus/internal/config"
)

// UpdateOptions struct    更新选项.
type UpdateOptions struct {
	// 预留扩展字段
}

// RunAutoUpdate function    执行更新操作.
func RunAutoUpdate(opts *UpdateOptions) {
	executor.ExecuteWithConfig(func(_ config.Option) (err error) {
		cfg, err := config.Get()
		if err != nil {
			return err
		}

		// 更新操作的业务逻辑
		_ = cfg
		return nil
	})
}
