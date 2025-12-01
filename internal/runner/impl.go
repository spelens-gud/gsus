package runner

import (
	"fmt"
	"path/filepath"

	"github.com/spelens-gud/gsus/apis/constant"
	"github.com/spelens-gud/gsus/apis/helpers"
	"github.com/spelens-gud/gsus/apis/helpers/executor"
	"github.com/spelens-gud/gsus/apis/syncimpl"
	"github.com/spelens-gud/gsus/basetmpl"
	"github.com/spelens-gud/gsus/internal/config"
)

// ImplOptions struct    实现生成选项.
type ImplOptions struct {
	Interface string // 接口名称
	Struct    string // 实现目录
	Prefix    string // 文件目录前缀
}

func RunAutoImpl(opts *ImplOptions) {
	executor.ExecuteWithConfig(func(_ config.Option) (err error) {
		syncConfig := &syncimpl.Config{
			SetName:       opts.Interface,
			ImplementsDir: opts.Struct,
			Scope:         "./",
			Prefix:        opts.Prefix,
		}

		if err = helpers.FixFilepathByProjectDir(&opts.Struct, &syncConfig.Scope); err != nil {
			return fmt.Errorf("init implement dir error: %v", err)
		}

		templatePath := filepath.Join(opts.Struct, ".gsus.impl"+constant.TemplateSuffix)
		temp, _, err := helpers.InitTemplateAndLoad(templatePath, basetmpl.DefaultImplTemplate)
		if err != nil {
			return fmt.Errorf("load implement init template error: %v", err)
		}
		syncConfig.ImplBaseTemplate = temp

		return syncConfig.SyncInterfaceImpls()
	})
}
