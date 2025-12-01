package runner

import (
	"context"
	"path/filepath"

	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/errors"
	"github.com/spelens-gud/gsus/internal/generator"
	"github.com/spelens-gud/gsus/internal/parser"
	"github.com/spelens-gud/gsus/internal/template"
	"github.com/spelens-gud/gsus/internal/utils"
	"github.com/spelens-gud/gsus/internal/validator"
)

// RouterOptions struct    HTTP 路由生成选项.
type RouterOptions struct {
	RouterPath string // 路由路径
}

// Router function    执行 HTTP 路由代码生成.
func Router(ctx context.Context, opts *RouterOptions) error {
	// 验证参数
	if err := validator.ValidateRequired(opts.RouterPath, "router path"); err != nil {
		return err
	}

	routerPath := opts.RouterPath
	if len(routerPath) == 0 {
		routerPath = "./api"
	}

	// 修正路径
	if err := utils.FixFilepathByProjectDir(&routerPath); err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeFile, "failed to resolve router path")
	}

	// 搜索服务
	svc, err := SearchServices("./")
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeParse, "failed to search services")
	}

	// 解析 API
	apiGroups, err := parser.ParseApiFromService(svc)
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeParse, "failed to parse API from service")
	}

	// 加载模板
	templatePath := filepath.Join(routerPath, ".gsus.router"+config.TemplateSuffix)
	customTemplate, hash, err := template.InitAndLoad(templatePath, template.DefaultHttpRouterTemplate)
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeTemplate, "failed to load router template")
	}

	// 生成路由代码
	if err := generator.GenApiRouterGroups(apiGroups, routerPath, func(options *parser.GenOptions) {
		options.Template = customTemplate
		options.TemplateHash = hash
	}); err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeGenerate, "failed to generate router code")
	}

	return nil
}

// RunAutoRouter function    执行 HTTP 路由代码生成（兼容旧接口）.
func RunAutoRouter(opts *RouterOptions) {
	config.ExecuteWithConfig(func(_ config.Option) error {
		return Router(context.Background(), opts)
	})
}
