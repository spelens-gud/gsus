package db2struct

import (
	"log"
	"path/filepath"

	"github.com/spelens-gud/gsus/apis/db2struct"
	"github.com/spelens-gud/gsus/apis/gengeneric"
	"github.com/spelens-gud/gsus/apis/helpers"
	"github.com/spelens-gud/gsus/apis/helpers/executor"
	"github.com/spelens-gud/gsus/internal/fileconfig"
	"github.com/spf13/cobra"
	"github.com/stoewer/go-strcase"
)

func Run(_ *cobra.Command, args []string) {
	executor.ExecuteWithConfig(func(cfg fileconfig.Config) (err error) {
		if len(cfg.Db2struct.Path) == 0 {
			cfg.Db2struct.Path = "./internal/model"
		}

		var (
			db2structConfig = cfg.Db2struct
			_               = helpers.FixFilepathByProjectDir(&db2structConfig.Path)
			dbConfig        = db2struct.DBConfig{
				User:     db2structConfig.User,
				Password: db2structConfig.Password,
				DB:       db2structConfig.Db,
				Host:     db2structConfig.Host,
				Port:     db2structConfig.Port,
			}
			opts []db2struct.Option
		)

		opts = append(opts,
			db2struct.WithPkgName(filepath.Base(cfg.Db2struct.Path)),
			db2struct.WithSQLInfo(),
			db2struct.WithCommentOutside(),
			db2struct.WithGormAnnotation(),
		)

		for k, v := range db2structConfig.TypeMap {
			opts = append(opts, db2struct.WithTypeReplace(k, v))
		}

		if len(db2structConfig.GenericMapTypes) > 0 {
			opts = append(opts, db2struct.WithGenericOption(func(options *gengeneric.Options) {
				options.MapTypes = db2structConfig.GenericMapTypes
			}))
		}

		if len(db2structConfig.GenericTemplate) > 0 {
			if tmpl, _, _ := helpers.LoadTemplate(db2structConfig.GenericTemplate); tmpl != nil {
				opts = append(opts, db2struct.WithGenericOption(func(options *gengeneric.Options) {
					options.Template = tmpl
				}))
			}
		}

		if len(args) > 0 {
			for _, table := range args {
				if bytes, err := db2struct.Gen(table, dbConfig, opts...); err == nil {
					if err = helpers.ImportAndWrite(bytes, filepath.Join(db2structConfig.Path, strcase.SnakeCase(table)+".go")); err != nil {
						log.Printf("generate table [ %s ] model error: %v", table, err)
					}
				}
			}
			return
		}

		return db2struct.GenAll(db2structConfig.Path, dbConfig, opts...)
	})
}
