package runner

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/errors"
	"github.com/spelens-gud/gsus/internal/generator"
	"github.com/spelens-gud/gsus/internal/logger"
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
	log := logger.WithPrefix("[impl]")
	log.Info("开始执行 impl 代码生成")

	// 验证参数
	if err := validator.ValidateRequired(opts.Interface, "interface name"); err != nil {
		log.Error("验证接口名称失败")
		return errors.WrapWithCode(err, errors.ErrCodeConfig, fmt.Sprintf("验证接口名称失败: %s", err))
	}
	if err := validator.ValidateRequired(opts.Struct, "struct directory"); err != nil {
		log.Error("验证实现目录失败")
		return errors.WrapWithCode(err, errors.ErrCodeConfig, fmt.Sprintf("验证实现目录失败: %s", err))
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
		log.Error("无法解析实现目录")
		return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("无法解析实现目录: %s", err))
	}

	// 加载模板
	templatePath := filepath.Join(opts.Struct, ".gsus.impl"+config.TemplateSuffix)
	temp, _, err := template.InitAndLoad(templatePath, template.DefaultImplTemplate)
	if err != nil {
		log.Error("加载实现模板失败")
		return errors.WrapWithCode(err, errors.ErrCodeTemplate, fmt.Sprintf("加载实现模板失败: %s", err))
	}
	syncConfig.ImplBaseTemplate = temp

	// 同步接口实现
	if err := syncConfig.SyncInterfaceImpls(); err != nil {
		log.Error("同步接口实现失败")
		return errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("同步接口实现失败: %s", err))
	}

	log.Info("接口实现代码生成完成")
	return nil
}

// RunAutoImpl function    执行接口实现代码生成（兼容旧接口）.
func RunAutoImpl(opts *ImplOptions) {
	config.ExecuteWithConfig(func(_ config.Option) error {
		return Impl(context.Background(), opts)
	})
}
