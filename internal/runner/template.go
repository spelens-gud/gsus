package runner

import (
	"fmt"
	"os"
	"strings"

	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/errors"
	"github.com/spelens-gud/gsus/internal/generator"
	"github.com/spelens-gud/gsus/internal/utils"
	"github.com/stoewer/go-strcase"
)

// TemplateOptions struct    模板操作选项.
type TemplateOptions struct {
	Models    []string // 模型名称列表
	GenAll    bool     // 是否生成所有模型
	Overwrite bool     // 是否覆盖已存在的文件
}

// RunAutoTemplate function    执行模板操作.
func RunAutoTemplate(opts *TemplateOptions) {
	config.ExecuteWithConfig(func(cfg config.Option) (err error) {
		cfg, err = config.Get()
		if err != nil {
			return errors.WrapWithCode(err, errors.ErrCodeConfig, fmt.Sprintf("获取配置文件错误: %s", err))
		}

		if len(cfg.Templates.ModelPath) == 0 {
			cfg.Templates.ModelPath = cfg.Db2struct.Path
		}

		if err = utils.FixFilepathByProjectDir(&cfg.Templates.ModelPath); err != nil {
			return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("按项目目录修复文件路径: %s", err))
		}

		models := opts.Models
		if len(models) == 0 && opts.GenAll {
			models, err = collectModelsFromPath(cfg.Templates.ModelPath)
			if err != nil {
				return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("无法解析模型路径: %s", err))
			}
		}

		for _, model := range models {
			if err = processModel(model, cfg, opts); err != nil {
				return errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("生成模板代码失败: %s", err))
			}
		}
		return nil
	})
}

func collectModelsFromPath(modelPath string) ([]string, error) {
	models := make([]string, 0)
	info, err := os.ReadDir(modelPath)
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("无法读取模型目录: %s", err))
	}

	for _, i := range info {
		if i.IsDir() || !strings.HasSuffix(i.Name(), ".go") {
			continue
		}
		models = append(models, strings.TrimSuffix(i.Name(), ".go"))
	}
	return models, nil
}

func processModel(model string, cfg config.Option, opts *TemplateOptions) error {
	return generator.GenTemplate(generator.TemplateConfig{
		ModelPath: cfg.Templates.ModelPath,
		ModelName: strcase.SnakeCase(model),
		Templates: cfg.Templates.Templates,
		Overwrite: opts.Overwrite,
	}, cfg)
}
