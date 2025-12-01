package router

import (
	"fmt"
	"path/filepath"

	"github.com/spelens-gud/gsus/apis/constant"
	"github.com/spelens-gud/gsus/apis/helpers"
	"github.com/spelens-gud/gsus/apis/helpers/executor"
	"github.com/spelens-gud/gsus/apis/httpgen"
	"github.com/spelens-gud/gsus/apis/httpgen/routergen"
	"github.com/spelens-gud/gsus/basetmpl"
	"github.com/spelens-gud/gsus/internal/fileconfig"
	"github.com/spelens-gud/gsus/internal/http/services"
	"github.com/spf13/cobra"
)

const defaultTemplateFileName = ".gsus.router" + constant.TemplateSuffix

func Run(cmd *cobra.Command, args []string) {
	executor.ExecuteWithConfig(func(_ fileconfig.Config) (err error) {
		routerPath := args[0]
		if len(routerPath) == 0 {
			routerPath = "./api"
		}

		if err = helpers.FixFilepathByProjectDir(&routerPath); err != nil {
			return
		}

		svc, err := services.SearchServices("./")
		if err != nil {
			return
		}

		apiGroups, err := httpgen.ParseApiFromService(svc)
		if err != nil {
			return fmt.Errorf("parse http api annotation error: %v", err)
		}

		templatePath := filepath.Join(routerPath, defaultTemplateFileName)

		customTemplate, hash, err := helpers.InitTemplateAndLoad(templatePath, basetmpl.DefaultHttpRouterTemplate)
		if err != nil {
			return fmt.Errorf("init router template error: %v", err)
		}

		return routergen.GenApiRouterGroups(apiGroups, routerPath, func(options *routergen.GenOptions) {
			options.Template = customTemplate
			options.TemplateHash = hash
		})
	})
}
