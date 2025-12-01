package runner

import (
	"fmt"
	"path/filepath"

	"github.com/spelens-gud/gsus/apis/constant"
	"github.com/spelens-gud/gsus/apis/helpers"
	"github.com/spelens-gud/gsus/apis/helpers/executor"
	"github.com/spelens-gud/gsus/apis/httpgen"
	"github.com/spelens-gud/gsus/apis/httpgen/clientgen"
	"github.com/spelens-gud/gsus/basetmpl"
	"github.com/spelens-gud/gsus/internal/config"
)

type ClientOptions struct {
	Args string // 命令参数
}

func RunAutoClient(opts *ClientOptions) {
	executor.ExecuteWithConfig(func(_ config.Option) (err error) {
		if len(opts.Args) == 0 {
			return fmt.Errorf("client path is required")
		}
		clientPath := opts.Args
		if len(clientPath) == 0 {
			return fmt.Errorf("client path is required")
		}

		if err := helpers.FixFilepathByProjectDir(&clientPath); err != nil {
			return err
		}

		svc, err := SearchServices("./")
		if err != nil {
			return err
		}

		apiGroups, err := httpgen.ParseApiFromService(svc)
		if err != nil {
			return err
		}

		templatePath := filepath.Join(clientPath, ".gsus.client_api"+constant.TemplateSuffix)
		apiTemplate, _, err := helpers.InitTemplateAndLoad(templatePath, basetmpl.DefaultHttpClientApiTemplate)
		if err != nil {
			return err
		}

		return clientgen.GenClients(apiGroups, func(option *clientgen.GenOption) {
			option.ClientsPath = clientPath
			option.ApiTemplate = apiTemplate
		})
	})
}
