package runner

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/generator"
	"github.com/spelens-gud/gsus/internal/utils"
	"github.com/stoewer/go-strcase"
)

type Db2structOptions struct {
	Tables []string // 指定的表名列表
}

func RunAutoDb2Struct(opts *Db2structOptions) {
	config.ExecuteWithConfig(func(cfg config.Option) error {
		var genOpts []config.DbOption
		var db2structConfig = cfg.Db2struct

		if len(db2structConfig.Path) == 0 {
			db2structConfig.Path = "./internal/model"
		}

		pkgName := filepath.Base(db2structConfig.Path)
		genOpts = append(genOpts,
			config.WithPkgName(pkgName),
			config.WithSQLInfo(),
			config.WithCommentOutside(),
			config.WithGormAnnotation(),
		)

		if err := utils.FixFilepathByProjectDir(&db2structConfig.Path); err != nil {
			return fmt.Errorf("failed to fix path: %w", err)
		}

		applyTypeReplacements(&genOpts, db2structConfig.TypeMap)
		applyGenericOptions(&genOpts, db2structConfig.GenericMapTypes, db2structConfig.GenericTemplate)

		dbConfig := buildDBConfig(db2structConfig)

		if len(opts.Tables) <= 0 {
			return generator.GenAllDb2Struct(db2structConfig.Path, dbConfig, genOpts...)
		}

		return generateModelsForTables(opts.Tables, db2structConfig.Path, dbConfig, genOpts)
	})
}

func applyTypeReplacements(genOpts *[]config.DbOption, typeMap map[string]string) {
	for k, v := range typeMap {
		*genOpts = append(*genOpts, config.WithTypeReplace(k, v))
	}
}

func applyGenericOptions(genOpts *[]config.DbOption, genericMapTypes []string, templatePath string) {
	if len(genericMapTypes) > 0 {
		*genOpts = append(*genOpts, config.WithGenericOption(func(options *config.Options) {
			options.MapTypes = genericMapTypes
		}))
	}

	if templatePath != "" {
		if tmpl, _, _ := config.LoadTemplate(templatePath); tmpl != nil {
			*genOpts = append(*genOpts, config.WithGenericOption(func(options *config.Options) {
				options.Template = tmpl
			}))
		}
	}
}

func buildDBConfig(config config.Db2struct) generator.DBConfig {
	return generator.DBConfig{
		User:     config.User,
		Password: config.Password,
		DB:       config.Db,
		Host:     config.Host,
		Port:     config.Port,
	}
}

func generateModelsForTables(tables []string, outputPath string, dbConfig generator.DBConfig,
	genOpts []config.DbOption) error {
	for _, table := range tables {
		bytes, err := generator.GenTable(table, dbConfig, genOpts...)
		if err != nil {
			log.Printf("generate table [%s] failed: %v", table, err)
			continue
		}

		filename := strcase.SnakeCase(table) + ".go"
		filePath := filepath.Join(outputPath, filename)
		if err = utils.ImportAndWrite(bytes, filePath); err != nil {
			log.Printf("write file for table [%s] error: %v", table, err)
		}
	}
	return nil
}
