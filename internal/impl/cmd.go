package impl

import (
	"fmt"
	"path/filepath"

	"github.com/spelens-gud/gsus/apis/constant"
	"github.com/spelens-gud/gsus/apis/helpers"
	"github.com/spelens-gud/gsus/apis/helpers/executor"
	"github.com/spelens-gud/gsus/apis/syncimpl"
	"github.com/spelens-gud/gsus/basetmpl"
	"github.com/spelens-gud/gsus/internal/fileconfig"
	"github.com/spf13/cobra"
)

func Run(cmd *cobra.Command, args []string) {
	executor.ExecuteWithConfig(func(_ fileconfig.Config) (err error) {
		syncConfig := &syncimpl.Config{
			SetName:       args[0],
			ImplementsDir: args[1],
			Scope:         "./",
			Prefix:        cmd.Flag("prefix").Value.String(),
		}

		if err = helpers.FixFilepathByProjectDir(&args[1], &syncConfig.Scope); err != nil {
			return fmt.Errorf("init implement dir error: %v", err)
		}

		templatePath := filepath.Join(args[1], ".gsus.impl"+constant.TemplateSuffix)
		temp, _, err := helpers.InitTemplateAndLoad(templatePath, basetmpl.DefaultImplTemplate)
		if err != nil {
			return fmt.Errorf("load implement init template error: %v", err)
		}
		syncConfig.ImplBaseTemplate = temp

		return syncConfig.SyncInterfaceImpls()
	})
}
