package runner

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/errors"
	"github.com/spelens-gud/gsus/internal/generator"
	"github.com/spelens-gud/gsus/internal/logger"
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
	log := logger.WithPrefix("[router]")
	log.Info("开始执行 HTTP 路由代码生成")
	// 验证参数
	if err := validator.ValidateRequired(opts.RouterPath, "router path"); err != nil {
		log.Error("未能验证路由器路径")
		return errors.WrapWithCode(err, errors.ErrCodeParse, fmt.Sprintf("未能验证路由器路径: %s", err))
	}

	routerPath := opts.RouterPath
	if len(routerPath) == 0 {
		routerPath = "./api"
	}

	// 修正路径
	if err := utils.FixFilepathByProjectDir(&routerPath); err != nil {
		log.Error("无法解析路由器路径")
		return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("无法解析路由器路径: %s", err))
	}

	// 搜索服务
	svc, err := SearchServices("./")
	if err != nil {
		log.Error("搜索服务失败")
		return errors.WrapWithCode(err, errors.ErrCodeParse, fmt.Sprintf("搜索服务失败: %s", err))
	}

	// 解析 API
	apiGroups, err := parser.ParseApiFromService(svc)
	if err != nil {
		log.Error("无法从服务解析API")
		return errors.WrapWithCode(err, errors.ErrCodeParse, fmt.Sprintf("无法从服务解析API: %s", err))
	}

	// 加载模板
	templatePath := filepath.Join(routerPath, ".gsus.router"+config.TemplateSuffix)
	customTemplate, hash, err := template.InitAndLoad(templatePath, template.DefaultHttpRouterTemplate)
	if err != nil {
		log.Error("加载路由器模板失败")
		return errors.WrapWithCode(err, errors.ErrCodeTemplate, fmt.Sprintf("加载路由器模板失败: %s", err))
	}

	// 生成路由代码
	if err := generator.GenApiRouterGroups(apiGroups, routerPath, func(options *parser.GenOptions) {
		options.Template = customTemplate
		options.TemplateHash = hash
	}); err != nil {
		log.Error("生成路由代码失败")
		return errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("生成路由代码失败: %s", err))
	}

	log.Info("生成路由代码成功")
	return nil
}

// RunAutoRouter function    执行 HTTP 路由代码生成（兼容旧接口）.
func RunAutoRouter(opts *RouterOptions) {
	config.ExecuteWithConfig(func(_ config.Option) error {
		return Router(context.Background(), opts)
	})
}
