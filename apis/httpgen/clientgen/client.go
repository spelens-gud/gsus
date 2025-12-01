package clientgen

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spelens-gud/gsus/apis/helpers"
	"github.com/spelens-gud/gsus/apis/httpgen"
	"github.com/spelens-gud/gsus/basetmpl"
)

var defaultApiTemplate = template.Must(template.New("api").Parse(basetmpl.DefaultHttpClientApiTemplate))

type GenOption struct {
	ClientsPath  string
	ApiTemplate  *template.Template
	BaseTemplate *template.Template
}

type clientApi struct {
	*httpgen.Api
	Param      string
	Return     string
	MethodSign string
}

type clientGroup struct {
	httpgen.ApiGroup
	ClientApis []clientApi
}

func GenClients(apiGroups []httpgen.ApiGroup, opts ...func(*GenOption)) (err error) {
	if len(apiGroups) == 0 {
		return
	}

	o := &GenOption{
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
		if err = helpers.ExecuteTemplateAndWrite(o.ApiTemplate, &client, clientDir); err != nil {
			return
		}
	}

	baseClientDir := filepath.Join(o.ClientsPath, "client.go")
	if _, err = os.Stat(baseClientDir); err != nil {
		if err = helpers.ExecuteTemplateAndWrite(o.BaseTemplate, struct{}{}, baseClientDir); err != nil {
			return
		}
	}
	return
}
