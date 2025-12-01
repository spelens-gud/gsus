package runner

import (
	"fmt"
	"path/filepath"

	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/generator"
	"github.com/spelens-gud/gsus/internal/template"
	"github.com/spelens-gud/gsus/internal/utils"
)

// ImplOptions struct    实现生成选项.
type ImplOptions struct {
	Interface string // 接口名称
	Struct    string // 实现目录
	Prefix    string // 文件目录前缀
}

func RunAutoImpl(opts *ImplOptions) {
	config.ExecuteWithConfig(func(_ config.Option) (err error) {
		syncConfig := &generator.Config{
			SetName:       opts.Interface,
			ImplementsDir: opts.Struct,
			Scope:         "./",
			Prefix:        opts.Prefix,
		}

		if err = utils.FixFilepathByProjectDir(&opts.Struct, &syncConfig.Scope); err != nil {
			return fmt.Errorf("init implement dir error: %v", err)
		}

		templatePath := filepath.Join(opts.Struct, ".gsus.impl"+config.TemplateSuffix)
		temp, _, err := config.InitTemplateAndLoad(templatePath, template.DefaultImplTemplate)
		if err != nil {
			return fmt.Errorf("load implement init template error: %v", err)
		}
		syncConfig.ImplBaseTemplate = temp

		return syncConfig.SyncInterfaceImpls()
	})
}
