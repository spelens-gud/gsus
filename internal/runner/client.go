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

// ClientOptions struct    HTTP 客户端生成选项.
type ClientOptions struct {
	ServicePath string // 服务路径
}

// Client function    执行 HTTP 客户端代码生成.
func Client(ctx context.Context, opts *ClientOptions) error {
	log := logger.WithPrefix("[client]")
	log.Info("开始执行 HTTP 客户端代码生成")
	// 验证参数
	if err := validator.ValidateRequired(opts.ServicePath, "service path"); err != nil {
		log.Error("验证服务路径失败")
		return errors.WrapWithCode(err, errors.ErrCodeParse, fmt.Sprintf("未能验证服务路径: %s", err))
	}

	// 修正路径
	clientPath := opts.ServicePath
	if err := utils.FixFilepathByProjectDir(&clientPath); err != nil {
		log.Error("无法解析客户端路径")
		return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("无法解析客户端路径: %s", err))
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
	templatePath := filepath.Join(clientPath, ".gsus.client_api"+config.TemplateSuffix)
	apiTemplate, _, err := template.InitAndLoad(templatePath, template.DefaultHttpClientApiTemplate)
	if err != nil {
		log.Error("加载客户端API模板失败")
		return errors.WrapWithCode(err, errors.ErrCodeTemplate, fmt.Sprintf("加载客户端API模板失败: %s", err))
	}

	// 生成客户端代码
	if err := generator.GenClients(apiGroups, func(option *config.GenOption) {
		option.ClientsPath = clientPath
		option.ApiTemplate = apiTemplate
	}); err != nil {
		log.Error("生成客户端代码失败")
		return errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("生成客户端代码失败: %s", err))
	}

	log.Info("生成http客户端代码成功")
	return nil
}

// RunAutoClient function    执行 HTTP 客户端代码生成（兼容旧接口）.
func RunAutoClient(opts *ClientOptions) {
	config.ExecuteWithConfig(func(_ config.Option) error {
		return Client(context.Background(), opts)
	})
}
