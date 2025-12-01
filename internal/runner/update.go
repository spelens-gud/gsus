package runner

import (
	"context"

	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/logger"
)

// UpdateOptions struct    更新选项.
type UpdateOptions struct {
	// 预留扩展字段
}

func Update(ctx context.Context, opts *UpdateOptions) error {
	log := logger.WithPrefix("[update]")
	log.Info("开始执行 update 代码生成")

	cfg, err := config.Get()
	if err != nil {
		return err
	}

	// 更新操作的业务逻辑
	_ = cfg

	log.Info("更新操作完成")
	return nil
}

// RunAutoUpdate function    执行更新操作.
func RunAutoUpdate(opts *UpdateOptions) {
	config.ExecuteWithConfig(func(_ config.Option) (err error) {
		return Update(context.Background(), opts)
	})
}
