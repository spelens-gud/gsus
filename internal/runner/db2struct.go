package runner

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/spelens-gud/gsus/apis/db2struct"
	"github.com/spelens-gud/gsus/apis/gengeneric"
	"github.com/spelens-gud/gsus/apis/helpers"
	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/parser"
	"github.com/stoewer/go-strcase"
)

type Db2structOptions struct {
	Tables []string // 指定的表名列表
}

func RunAutoDb2Struct(opts *Db2structOptions) {
	ExecuteWithConfig(func(cfg config.Option) error {
		var genOpts []db2struct.Option
		var db2structConfig = cfg.Db2struct

		if len(db2structConfig.Path) == 0 {
			db2structConfig.Path = "./internal/model"
		}

		pkgName := filepath.Base(db2structConfig.Path)
		genOpts = append(genOpts,
			db2struct.WithPkgName(pkgName),
			db2struct.WithSQLInfo(),
			db2struct.WithCommentOutside(),
			db2struct.WithGormAnnotation(),
		)

		if err := parser.FixFilepathByProjectDir(&db2structConfig.Path); err != nil {
			return fmt.Errorf("failed to fix path: %w", err)
		}

		applyTypeReplacements(&genOpts, db2structConfig.TypeMap)
		applyGenericOptions(&genOpts, db2structConfig.GenericMapTypes, db2structConfig.GenericTemplate)

		dbConfig := buildDBConfig(db2structConfig)

		if len(opts.Tables) <= 0 {
			return db2struct.GenAll(db2structConfig.Path, dbConfig, genOpts...)
		}

		return generateModelsForTables(opts.Tables, db2structConfig.Path, dbConfig, genOpts)
	})
}

func applyTypeReplacements(genOpts *[]db2struct.Option, typeMap map[string]string) {
	for k, v := range typeMap {
		*genOpts = append(*genOpts, db2struct.WithTypeReplace(k, v))
	}
}

func applyGenericOptions(genOpts *[]db2struct.Option, genericMapTypes []string, templatePath string) {
	if len(genericMapTypes) > 0 {
		*genOpts = append(*genOpts, db2struct.WithGenericOption(func(options *gengeneric.Options) {
			options.MapTypes = genericMapTypes
		}))
	}

	if templatePath != "" {
		if tmpl, _, _ := helpers.LoadTemplate(templatePath); tmpl != nil {
			*genOpts = append(*genOpts, db2struct.WithGenericOption(func(options *gengeneric.Options) {
				options.Template = tmpl
			}))
		}
	}
}

func buildDBConfig(config config.Db2struct) db2struct.DBConfig {
	return db2struct.DBConfig{
		User:     config.User,
		Password: config.Password,
		DB:       config.Db,
		Host:     config.Host,
		Port:     config.Port,
	}
}

func generateModelsForTables(tables []string, outputPath string, dbConfig db2struct.DBConfig,
	genOpts []db2struct.Option) error {
	for _, table := range tables {
		bytes, err := db2struct.Gen(table, dbConfig, genOpts...)
		if err != nil {
			log.Printf("generate table [%s] failed: %v", table, err)
			continue
		}

		filename := strcase.SnakeCase(table) + ".go"
		filePath := filepath.Join(outputPath, filename)
		if err = helpers.ImportAndWrite(bytes, filePath); err != nil {
			log.Printf("write file for table [%s] error: %v", table, err)
		}
	}
	return nil
}
