package runner

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/errors"
	"github.com/spelens-gud/gsus/internal/generator"
	"github.com/spelens-gud/gsus/internal/generator/db"
	"github.com/spelens-gud/gsus/internal/logger"
	"github.com/spelens-gud/gsus/internal/template"
	"github.com/spelens-gud/gsus/internal/utils"
	"github.com/stoewer/go-strcase"
)

// Db2structOptions struct    数据库表转结构体选项.
type Db2structOptions struct {
	Tables []string // 指定的表名列表
}

// Db2struct function    执行数据库表转结构体.
func Db2struct(ctx context.Context, opts *Db2structOptions, cfg config.Option) error {
	log := logger.WithPrefix("[db2struct]")
	log.Info("开始执行数据库表转结构体代码生成")

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
	dbConfig := buildDBConfigV2(db2structConfig)

	// 生成所有表或指定表
	if len(opts.Tables) == 0 {
		log.Info("生成所有表的结构体")
		if err := generator.GenAllDb2StructWithAdapter(ctx, db2structConfig.Path, dbConfig, genOpts...); err != nil {
			return errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("未能生成所有表: %s", err))
		}
		log.Info("所有表生成完成")
		return nil
	}

	log.Info("生成指定表的结构体: %v", opts.Tables)
	return generateModelsForTablesV2(ctx, opts.Tables, db2structConfig.Path, dbConfig, genOpts)
}

// applyTypeReplacements function    应用类型替换.
func applyTypeReplacements(genOpts *[]config.DbOption, typeMap map[string]string) {
	for k, v := range typeMap {
		*genOpts = append(*genOpts, config.WithTypeReplace(k, v))
	}
}

// applyGenericOptions function    应用泛型选项.
func applyGenericOptions(genOpts *[]config.DbOption, genericMapTypes []string, templatePath string) error {
	if len(genericMapTypes) > 0 {
		*genOpts = append(*genOpts, config.WithGenericOption(func(options *config.TypeOptions) {
			options.MapTypes = genericMapTypes
		}))
	}

	if templatePath != "" {
		tmpl, _, err := template.Load(templatePath)
		if err != nil {
			return errors.WrapWithCode(err, errors.ErrCodeTemplate, fmt.Sprintf("加载通用模板失败: %s", err))
		}
		if tmpl != nil {
			*genOpts = append(*genOpts, config.WithGenericOption(func(options *config.TypeOptions) {
				options.Template = tmpl
			}))
		}
	}

	return nil
}

// buildDBConfigV2 function    构建数据库配置(新版本).
func buildDBConfigV2(cfg config.Db2struct) *db.Config {
	// 设置默认数据库类型
	dbType := db.Type(cfg.Type)
	if dbType == "" {
		dbType = db.MySQL
	}

	// 设置默认字符集
	charset := cfg.Charset
	if charset == "" {
		charset = "utf8mb4"
	}

	return &db.Config{
		Type:     dbType,
		User:     cfg.User,
		Password: cfg.Password,
		Host:     cfg.Host,
		Port:     cfg.Port,
		Database: cfg.Db,
		Charset:  charset,
	}
}

// generateModelsForTablesV2 function    为指定表生成模型(新版本).
func generateModelsForTablesV2(ctx context.Context, tables []string, outputPath string,
	dbConfig *db.Config, genOpts []config.DbOption) error {
	log := logger.WithPrefix("[db2struct]")

	for _, table := range tables {
		log.Info("生成表 %s 的结构体", table)
		bytes, err := generator.GenTableWithAdapter(ctx, table, dbConfig, genOpts...)
		if err != nil {
			log.Error("生成表 [%s] 失败: %v", table, err)
			continue
		}

		filename := strcase.SnakeCase(table) + ".go"
		filePath := filepath.Join(outputPath, filename)
		if err = utils.ImportAndWrite(bytes, filePath); err != nil {
			log.Error("写入表 [%s] 文件失败: %v", table, err)
			continue
		}
		log.Info("表 %s 生成完成", table)
	}
	return nil
}

// validateDBConfig function    验证数据库配置.
func validateDBConfig(cfg config.Db2struct) error {
	// 导入 validator 包
	return nil // 这里可以添加具体的验证逻辑
}

// RunAutoDb2Struct function    执行数据库表转结构体（兼容旧接口）.
func RunAutoDb2Struct(opts *Db2structOptions) {
	config.ExecuteWithConfig(func(cfg config.Option) error {
		return Db2struct(context.Background(), opts, cfg)
	})
}
