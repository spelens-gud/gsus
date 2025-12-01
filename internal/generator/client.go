package generator

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/errors"
	"github.com/spelens-gud/gsus/internal/parser"
	template2 "github.com/spelens-gud/gsus/internal/template"
	"github.com/spelens-gud/gsus/internal/utils"
)

type clientApi struct {
	*parser.Api
	Param      string
	Return     string
	MethodSign string
}
type clientGroup struct {
	parser.ApiGroup
	ClientApis []clientApi
}

var defaultApiTemplate = template.Must(template.New("api").Parse(template2.DefaultHttpClientApiTemplate))
var defaultBaseTemplate = template.Must(template.New("base").Parse(template2.DefaultHttpClientBaseTemplate))

// GenClients function    生成 HTTP 客户端代码.
func GenClients(apiGroups []parser.ApiGroup, opts ...func(*config.GenOption)) (err error) {
	if len(apiGroups) == 0 {
		return errors.WrapWithCode(errors.New(errors.ErrCodeGenerate, "没有可用的 API"), errors.ErrCodeGenerate, "没有可用的 API")
	}

	o := &config.GenOption{
		ApiTemplate:  defaultApiTemplate,
		BaseTemplate: defaultBaseTemplate,
	}

	for _, opt := range opts {
		opt(o)
	}

	_ = os.MkdirAll(o.ClientsPath, 0775)

	for _, group := range apiGroups {
		var apis []clientApi

		handlerGenned := make(map[string]bool)
		for _, api := range group.Apis {
			if strings.ToUpper(api.Method) == http.MethodOptions {
				continue
			}
			if handlerGenned[api.Handler] {
				continue
			}

			if strings.ToUpper(api.Method) == "ANY" {
				api.Method = "POST"
			}

			client := clientApi{
				Api: api,
			}

			var param, ret string
			for _, p := range api.Params {
				if p == "context.Context" {
					continue
				}
				client.Param = p
				param = ",param " + p
				break
			}

			for _, p := range api.Returns {
				if p == "error" {
					continue
				}
				client.Return = p
				ret = "ret " + p + ","
				break
			}
			client.MethodSign = fmt.Sprintf(`(ctx context.Context%s) (%serr error)`, param, ret)
			handlerGenned[api.Handler] = true
			apis = append(apis, client)
		}

		client := clientGroup{
			ApiGroup:   group,
			ClientApis: apis,
		}

		clientDir := filepath.Join(o.ClientsPath, "client_"+client.GroupName, "client_"+client.GroupName+".go")
		if err = utils.ExecuteTemplateAndWrite(o.ApiTemplate, &client, clientDir); err != nil {
			return errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("生成客户端 API 文件失败:%s", err))
		}
	}

	baseClientDir := filepath.Join(o.ClientsPath, "client.go")
	if _, err = os.Stat(baseClientDir); err != nil {
		if err = utils.ExecuteTemplateAndWrite(o.BaseTemplate, struct{}{}, baseClientDir); err != nil {
			return errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("生成基础客户端文件失败:%s", err))
		}
	}
	return nil
}
