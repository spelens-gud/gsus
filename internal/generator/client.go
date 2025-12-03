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
	tmpl "github.com/spelens-gud/gsus/internal/template"
	"github.com/spelens-gud/gsus/internal/utils"
)

var genFilePrefix = "client.go"
var defaultApiTemplate = template.Must(template.New("api").Parse(tmpl.DefaultHttpClientApiTemplate))
var defaultBaseTemplate = template.Must(template.New("base").Parse(tmpl.DefaultHttpClientBaseTemplate))

// clientApi struct    HTTP 客户端 API 结构体.
type clientApi struct {
	*parser.Api        // 继承 Api 结构体
	Param       string // 参数类型
	Return      string // 返回值类型
	MethodSign  string // 方法签名
}

// clientGroup struct    HTTP 客户端组结构体.
type clientGroup struct {
	parser.ApiGroup             // 继承 ApiGroup 结构体
	ClientApis      []clientApi // HTTP 客户端 API 列表
}

// GenClients function    生成 HTTP 客户端代码.
func GenClients(apiGroups []parser.ApiGroup, opts ...func(*config.ClientOpt)) (err error) {
	if len(apiGroups) == 0 {
		return errors.WrapWithCode(errors.New(errors.ErrCodeGenerate, "没有可用的 API"), errors.ErrCodeGenerate, "没有可用的 API")
	}

	clientOpt := &config.ClientOpt{
		ApiTemplate:  defaultApiTemplate,
		BaseTemplate: defaultBaseTemplate,
	}

	for _, opt := range opts {
		opt(clientOpt)
	}

	_ = os.MkdirAll(clientOpt.ClientsPath, 0775)

	// 处理每个API组
	for _, group := range apiGroups {
		if err = generateClientGroup(group, clientOpt); err == nil {
			continue
		}
		return errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("生成 HTTP 客户端代码失败: %s", err))
	}

	// 生成基础客户端文件
	return generateBaseClient(clientOpt)
}

// generateBaseClient function    生成基础客户端代码.
func generateBaseClient(o *config.ClientOpt) error {
	baseClientDir := filepath.Join(o.ClientsPath, genFilePrefix)
	if _, err := os.Stat(baseClientDir); !os.IsNotExist(err) {
		if err = utils.ExecuteTemplateAndWrite(o.BaseTemplate, struct{}{}, baseClientDir); err == nil {
			return nil
		}
		return errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("生成基础客户端文件失败:%s", err))
	}
	return nil
}

// generateClientGroup function    提取生成客户端组的函数.
func generateClientGroup(group parser.ApiGroup, o *config.ClientOpt) error {
	var apis []clientApi
	handlerGenned := make(map[string]bool)

	// 处理单个API
	for _, api := range group.Apis {
		if client, shouldAdd := processApi(api, handlerGenned); shouldAdd {
			apis = append(apis, *client)
		}
	}

	client := clientGroup{
		ApiGroup:   group,
		ClientApis: apis,
	}

	clientDir := filepath.Join(o.ClientsPath, "client_"+client.GroupName, "client_"+client.GroupName+".go")
	if err := utils.ExecuteTemplateAndWrite(o.ApiTemplate, &client, clientDir); err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("生成客户端 API 文件失败:%s", err))
	}

	return nil
}

// processApi function    提取处理单个API的函数.
func processApi(api *parser.Api, handlerGenned map[string]bool) (*clientApi, bool) {
	if strings.ToUpper(api.Method) == http.MethodOptions {
		return nil, false
	}
	// 检查是否已经生成过该处理程序
	if handlerGenned[api.Handler] {
		return nil, false
	}

	if strings.ToUpper(api.Method) == "ANY" {
		api.Method = "POST"
	}

	client := &clientApi{
		Api: api,
	}

	// 获取参数和返回值类型
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

	// 构建方法签名
	client.MethodSign = fmt.Sprintf(`(ctx context.Context%s) (%serr error)`, param, ret)
	handlerGenned[api.Handler] = true

	return client, true
}
