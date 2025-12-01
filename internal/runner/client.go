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

// ClientOptions struct    HTTP 客户端生成选项.
type ClientOptions struct {
	ServicePath string // 服务路径
}

// Client function    执行 HTTP 客户端代码生成.
func Client(ctx context.Context, opts *ClientOptions) error {
	// 验证参数
	if err := validator.ValidateRequired(opts.ServicePath, "service path"); err != nil {
		return err
	}

	// 修正路径
	clientPath := opts.ServicePath
	if err := utils.FixFilepathByProjectDir(&clientPath); err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeFile, "failed to resolve client path")
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
	templatePath := filepath.Join(clientPath, ".gsus.client_api"+config.TemplateSuffix)
	apiTemplate, _, err := template.InitAndLoad(templatePath, template.DefaultHttpClientApiTemplate)
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeTemplate, "failed to load client API template")
	}

	// 生成客户端代码
	if err := generator.GenClients(apiGroups, func(option *config.GenOption) {
		option.ClientsPath = clientPath
		option.ApiTemplate = apiTemplate
	}); err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeGenerate, "failed to generate client code")
	}

	return nil
}

// RunAutoClient function    执行 HTTP 客户端代码生成（兼容旧接口）.
func RunAutoClient(opts *ClientOptions) {
	config.ExecuteWithConfig(func(_ config.Option) error {
		return Client(context.Background(), opts)
	})
}
