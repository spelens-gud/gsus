package client

import (
	"path/filepath"

	"github.com/spelens-gud/gsus/apis/constant"
	"github.com/spelens-gud/gsus/apis/helpers"
	"github.com/spelens-gud/gsus/apis/helpers/executor"
	"github.com/spelens-gud/gsus/apis/httpgen"
	"github.com/spelens-gud/gsus/apis/httpgen/clientgen"
	"github.com/spelens-gud/gsus/basetmpl"
	"github.com/spelens-gud/gsus/internal/fileconfig"
	"github.com/spelens-gud/gsus/internal/http/services"
	"github.com/spf13/cobra"
)

func Run(cmd *cobra.Command, args []string) {
	executor.ExecuteWithConfig(func(_ fileconfig.Config) (err error) {
		clientPath := args[0]

		if err = helpers.FixFilepathByProjectDir(&clientPath); err != nil {
			return
		}

		svc, err := services.SearchServices("./")
		if err != nil {
			return
		}

		apiGroups, err := httpgen.ParseApiFromService(svc)
		if err != nil {
			return
		}

		templatePath := filepath.Join(clientPath, ".gsus.client_api"+constant.TemplateSuffix)

		apiTemplate, _, err := helpers.InitTemplateAndLoad(templatePath, basetmpl.DefaultHttpClientApiTemplate)
		if err != nil {
			return
		}

		return clientgen.GenClients(apiGroups, func(option *clientgen.GenOption) {
			option.ClientsPath = clientPath
			option.ApiTemplate = apiTemplate
		})
	})
}
