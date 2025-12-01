package runner

import (
	"fmt"
	"path/filepath"

	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/generator"
	"github.com/spelens-gud/gsus/internal/parser"
	"github.com/spelens-gud/gsus/internal/template"
	"github.com/spelens-gud/gsus/internal/utils"
)

type ClientOptions struct {
	Args string // 命令参数
}

func RunAutoClient(opts *ClientOptions) {
	config.ExecuteWithConfig(func(_ config.Option) (err error) {
		if len(opts.Args) == 0 {
			return fmt.Errorf("client path is required")
		}
		clientPath := opts.Args
		if len(clientPath) == 0 {
			return fmt.Errorf("client path is required")
		}

		if err := utils.FixFilepathByProjectDir(&clientPath); err != nil {
			return err
		}

		svc, err := SearchServices("./")
		if err != nil {
			return err
		}

		apiGroups, err := parser.ParseApiFromService(svc)
		if err != nil {
			return err
		}

		templatePath := filepath.Join(clientPath, ".gsus.client_api"+config.TemplateSuffix)
		apiTemplate, _, err := config.InitTemplateAndLoad(templatePath, template.DefaultHttpClientApiTemplate)
		if err != nil {
			return err
		}

		return generator.GenClients(apiGroups, func(option *config.GenOption) {
			option.ClientsPath = clientPath
			option.ApiTemplate = apiTemplate
		})
	})
}
