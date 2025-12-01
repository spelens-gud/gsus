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

type RouterOptions struct {
	Args string // 命令参数
}

func RunAutoRouter(opts *RouterOptions) {
	config.ExecuteWithConfig(func(_ config.Option) (err error) {
		if len(opts.Args) == 0 {
			return fmt.Errorf("client path is required")
		}
		routerPath := opts.Args
		if len(routerPath) == 0 {
			routerPath = "./api"
		}

		if err = utils.FixFilepathByProjectDir(&routerPath); err != nil {
			return err
		}

		svc, err := SearchServices("./")
		if err != nil {
			return err
		}

		apiGroups, err := parser.ParseApiFromService(svc)
		if err != nil {
			return fmt.Errorf("parse http api annotation error: %v", err)
		}

		templatePath := filepath.Join(routerPath, ".gsus.router"+config.TemplateSuffix)

		customTemplate, hash, err := config.InitTemplateAndLoad(templatePath, template.DefaultHttpRouterTemplate)
		if err != nil {
			return fmt.Errorf("init router template error: %v", err)
		}

		return generator.GenApiRouterGroups(apiGroups, routerPath, func(options *parser.GenOptions) {
			options.Template = customTemplate
			options.TemplateHash = hash
		})
	})
}
