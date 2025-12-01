package template

import (
	"os"
	"strings"

	"github.com/spelens-gud/gsus/apis/helpers"
	"github.com/spelens-gud/gsus/apis/helpers/executor"
	"github.com/spelens-gud/gsus/apis/templates"
	"github.com/spelens-gud/gsus/internal/fileconfig"
	"github.com/spf13/cobra"
	"github.com/stoewer/go-strcase"
)

var (
	Cmd = &cobra.Command{
		Use:   "template",
		Short: "generate codes from table model by custom templates",
		Run:   run,
	}

	genAll, overwrite *bool
)

func init() {
	genAll = Cmd.Flags().Bool("all", false, "generate all model")
	overwrite = Cmd.Flags().Bool("overwrite", false, "overwrite files")
}

func run(_ *cobra.Command, args []string) {
	executor.ExecuteWithConfig(func(cfg fileconfig.Config) (err error) {
		if len(cfg.Templates.ModelPath) == 0 {
			cfg.Templates.ModelPath = cfg.Db2struct.Path
		}

		if err = helpers.FixFilepathByProjectDir(&cfg.Templates.ModelPath); err != nil {
			return
		}

		if len(args) == 0 && *genAll {
			info, _ := os.ReadDir(cfg.Templates.ModelPath)
			for _, i := range info {
				if i.IsDir() || !strings.HasSuffix(i.Name(), ".go") {
					continue
				}
				args = append(args, strings.TrimSuffix(i.Name(), ".go"))
			}
		}

		for _, arg := range args {
			if err = templates.Gen(templates.Config{
				ModelPath: cfg.Templates.ModelPath,
				ModelName: strcase.SnakeCase(arg),
				Templates: cfg.Templates.Templates,
				Overwrite: *overwrite,
			}, cfg); err != nil {
				return
			}
		}
		return
	})
}
