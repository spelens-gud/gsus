package runner

import (
	"context"
	"path/filepath"

	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/errors"
	"github.com/spelens-gud/gsus/internal/generator"
	"github.com/spelens-gud/gsus/internal/template"
	"github.com/spelens-gud/gsus/internal/utils"
	"github.com/spelens-gud/gsus/internal/validator"
)

// ImplOptions struct    接口实现生成选项.
type ImplOptions struct {
	Interface string // 接口名称
	Struct    string // 实现目录
	Prefix    string // 文件目录前缀
}

// Impl function    执行接口实现代码生成.
func Impl(ctx context.Context, opts *ImplOptions) error {
	// 验证参数
	if err := validator.ValidateRequired(opts.Interface, "interface name"); err != nil {
		return err
	}
	if err := validator.ValidateRequired(opts.Struct, "struct directory"); err != nil {
		return err
	}

	// 构建生成配置
	syncConfig := &generator.Config{
		SetName:       opts.Interface,
		ImplementsDir: opts.Struct,
		Scope:         "./",
		Prefix:        opts.Prefix,
	}

	// 修正路径
	if err := utils.FixFilepathByProjectDir(&opts.Struct, &syncConfig.Scope); err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeFile, "failed to resolve implement directory")
	}

	// 加载模板
	templatePath := filepath.Join(opts.Struct, ".gsus.impl"+config.TemplateSuffix)
	temp, _, err := template.InitAndLoad(templatePath, template.DefaultImplTemplate)
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeTemplate, "failed to load implement template")
	}
	syncConfig.ImplBaseTemplate = temp

	// 同步接口实现
	if err := syncConfig.SyncInterfaceImpls(); err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeGenerate, "failed to sync interface implementations")
	}

	return nil
}

// RunAutoImpl function    执行接口实现代码生成（兼容旧接口）.
func RunAutoImpl(opts *ImplOptions) {
	config.ExecuteWithConfig(func(_ config.Option) error {
		return Impl(context.Background(), opts)
	})
}
