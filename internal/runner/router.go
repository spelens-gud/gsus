package runner

import (
	"fmt"
	"path/filepath"

	"github.com/spelens-gud/gsus/apis/constant"
	"github.com/spelens-gud/gsus/apis/helpers"
	"github.com/spelens-gud/gsus/apis/helpers/executor"
	"github.com/spelens-gud/gsus/apis/httpgen"
	"github.com/spelens-gud/gsus/apis/httpgen/routergen"
	"github.com/spelens-gud/gsus/basetmpl"
	"github.com/spelens-gud/gsus/internal/config"
)

type RouterOptions struct {
	Args string // 命令参数
}

func RunAutoRouter(opts *RouterOptions) {
	executor.ExecuteWithConfig(func(_ config.Option) (err error) {
		if len(opts.Args) == 0 {
			return fmt.Errorf("client path is required")
		}
		routerPath := opts.Args
		if len(routerPath) == 0 {
			routerPath = "./api"
		}

		if err = helpers.FixFilepathByProjectDir(&routerPath); err != nil {
			return err
		}

		svc, err := SearchServices("./")
		if err != nil {
			return err
		}

		apiGroups, err := httpgen.ParseApiFromService(svc)
		if err != nil {
			return fmt.Errorf("parse http api annotation error: %v", err)
		}

		templatePath := filepath.Join(routerPath, ".gsus.router"+constant.TemplateSuffix)

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
