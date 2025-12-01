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
	"github.com/stoewer/go-strcase"
)

// Db2structOptions struct    数据库表转结构体选项.
type Db2structOptions struct {
	Tables []string // 指定的表名列表
}

// Db2struct function    执行数据库表转结构体.
func Db2struct(ctx context.Context, opts *Db2structOptions, cfg config.Option) error {
	log := logger.WithPrefix("[db2struct]")
	log.Info("开始执行数据库表转结构体")

	var genOpts []config.DbOption
	db2structConfig := cfg.Db2struct

	// 验证数据库配置
	if err := validateDBConfig(db2structConfig); err != nil {
		log.Error("数据库配置验证失败: %v", err)
		return errors.New(errors.ErrCodeConfig, fmt.Sprintf("数据库配置验证失败: %s", err))
	}

	// 设置默认输出路径
	if len(db2structConfig.Path) == 0 {
		db2structConfig.Path = "./internal/model"
		log.Debug("使用默认输出路径: %s", db2structConfig.Path)
	}

	// 构建生成选项
	pkgName := filepath.Base(db2structConfig.Path)
	genOpts = append(genOpts,
		config.WithPkgName(pkgName),
		config.WithSQLInfo(),
		config.WithCommentOutside(),
		config.WithGormAnnotation(),
	)

	// 修正路径
	if err := utils.FixFilepathByProjectDir(&db2structConfig.Path); err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("无法解析输出路径: %s", err))
	}
	log.Debug("输出路径: %s", db2structConfig.Path)

	// 应用类型替换
	applyTypeReplacements(&genOpts, db2structConfig.TypeMap)

	// 应用泛型选项
	if err := applyGenericOptions(&genOpts, db2structConfig.GenericMapTypes, db2structConfig.GenericTemplate); err != nil {
		log.Error("应用泛型选项失败: %v", err)
		return errors.WrapWithCode(err, errors.ErrCodeConfig, fmt.Sprintf("泛型选项应用失败: %s", err))
	}

	// 构建数据库配置
	dbConfig := buildDBConfig(db2structConfig)

	// 生成所有表或指定表
	if len(opts.Tables) == 0 {
		log.Info("生成所有表的结构体")
		if err := generator.GenAllDb2Struct(db2structConfig.Path, dbConfig, genOpts...); err != nil {
			return errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("未能生成所有表: %s", err))
		}
		log.Info("所有表生成完成")
		return nil
	}

	log.Info("生成指定表的结构体: %v", opts.Tables)
	return generateModelsForTables(opts.Tables, db2structConfig.Path, dbConfig, genOpts)
}

// applyTypeReplacements function    应用类型替换.
func applyTypeReplacements(genOpts *[]config.DbOption, typeMap map[string]string) {
	for k, v := range typeMap {
		*genOpts = append(*genOpts, config.WithTypeReplace(k, v))
	}
}

// applyGenericOptions function    应用泛型选项.
func applyGenericOptions(genOpts *[]config.DbOption, genericMapTypes []string, templatePath string) error {
	if len(genericMapTypes) > 0 {
		*genOpts = append(*genOpts, config.WithGenericOption(func(options *config.Options) {
			options.MapTypes = genericMapTypes
		}))
	}

	if templatePath != "" {
		tmpl, _, err := template.Load(templatePath)
		if err != nil {
			return errors.WrapWithCode(err, errors.ErrCodeTemplate, fmt.Sprintf("加载通用模板失败: %s", err))
		}
		if tmpl != nil {
			*genOpts = append(*genOpts, config.WithGenericOption(func(options *config.Options) {
				options.Template = tmpl
			}))
		}
	}

	return nil
}

// buildDBConfig function    构建数据库配置.
func buildDBConfig(cfg config.Db2struct) generator.DBConfig {
	return generator.DBConfig{
		User:     cfg.User,
		Password: cfg.Password,
		DB:       cfg.Db,
		Host:     cfg.Host,
		Port:     cfg.Port,
	}
}

// generateModelsForTables function    为指定表生成模型.
func generateModelsForTables(tables []string, outputPath string, dbConfig generator.DBConfig,
	genOpts []config.DbOption) error {
	for _, table := range tables {
		bytes, err := generator.GenTable(table, dbConfig, genOpts...)
		if err != nil {
			logger.Error("generate table [%s] failed: %v", table, err)
			continue
		}

		filename := strcase.SnakeCase(table) + ".go"
		filePath := filepath.Join(outputPath, filename)
		if err = utils.ImportAndWrite(bytes, filePath); err != nil {
			logger.Error("write file for table [%s] error: %v", table, err)
		}
	}
	return nil
}

// validateDBConfig function    验证数据库配置.
func validateDBConfig(cfg config.Db2struct) error {
	// 导入 validator 包
	return nil // 这里可以添加具体的验证逻辑
}

// RunAutoDb2Struct function    执行数据库表转结构体（兼容旧接口）.
func RunAutoDb2Struct(opts *Db2structOptions) {
	config.ExecuteWithConfig(func(cfg config.Option) error {
		return Db2struct(context.Background(), opts, cfg)
	})
}
