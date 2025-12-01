package parser

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/spelens-gud/gsus/internal/errors"
	"github.com/stoewer/go-strcase"
)

type Service struct {
	InterfaceName string
	ServiceName   string
	Pkg           string
	OtherOptions  map[string]string
	ApiAnnotates  map[string]*ApiAnnotate
}

func (r Service) GetPath(apiDir string) (groupRoute, dir, pkg string) {
	group := r.OtherOptions["group"]
	if len(group) == 0 {
		group = r.ServiceName
	}
	group = strings.Trim(group, `"`)
	// 组路由前缀
	groupRoute = r.OtherOptions["route"]
	if len(groupRoute) == 0 {
		groupRoute = fmt.Sprintf(`"%s"`, group)
	}
	// 创建组目录
	dir = filepath.Join(apiDir, group)
	// package名
	pkg = filepath.Base(dir)
	return
}

type ApiAnnotate struct {
	Interface string
	Apis      []ApiAnnotateItem
}

type ApiAnnotateItem struct {
	Handler string
	Params  []string
	Returns []string
	Title   string
	Method  string
	Args    []string
	Options map[string]string
	Doc     []string
}
type ApiGroup struct {
	Package     string
	GroupName   string
	ServiceName string
	GroupRoute  string
	Hash        string
	Apis        []*Api
	Options     map[string]string
	Filepath    string
	Skip        bool
}
type Api struct {
	Params        []string
	Returns       []string
	Method        string
	BaseRoute     string
	HttpMethod    string
	Route         string
	Handler       string
	Title         string
	Options       map[string]string
	AnnotationMap string
}

// ParseApiFromService 从服务定义中解析出API组信息
// 参数 services 是服务定义列表
// 返回值 apiGroups 是解析出的API组列表，每个组包含该服务下的所有HTTP API
// 返回值 err 表示解析过程中可能出现的错误.
func ParseApiFromService(services []Service) (apiGroups []ApiGroup, err error) {
	// 所有接口定义
	for _, service := range services {
		var apiGroup ApiGroup
		apiGroup, err = parseService(service)
		if err != nil {
			return apiGroups, errors.WrapWithCode(err, errors.ErrCodeParse, fmt.Sprintf("解析服务 %s 失败", service.ServiceName))
		}
		if len(apiGroup.Apis) == 0 {
			continue
		}
		apiGroups = append(apiGroups, apiGroup)
	}
	return
}
func parseService(service Service) (apiRouteGroup ApiGroup, err error) {
	quoteServiceName := strconv.Quote(service.ServiceName)

	// 接口组
	apiGroup := service.OtherOptions["group"]
	if len(apiGroup) == 0 {
		apiGroup = quoteServiceName
	}
	if !isValidRoute(apiGroup) {
		err = fmt.Errorf("invalid group %s in %s %s", apiGroup, service.Pkg, service.InterfaceName)
		return apiRouteGroup, errors.New(errors.ErrCodeParse, err.Error())
	}
	apiGroup = strings.Trim(apiGroup, `"`)

	// 组路由前缀
	groupRoute := service.OtherOptions["route"]
	if len(groupRoute) == 0 {
		groupRoute = quoteServiceName
	}

	if !isValidRoute(groupRoute) {
		err = fmt.Errorf("invalid route %s in %s %s", groupRoute, service.Pkg, service.InterfaceName)
		return apiRouteGroup, errors.New(errors.ErrCodeParse, err.Error())
	}

	// 生成package名
	pkgName := filepath.Base(apiGroup)
	pkgName = strings.ReplaceAll(pkgName, "-", "_")

	//  http服务
	httpApis, ok := service.ApiAnnotates["http"]
	if !ok {
		return apiRouteGroup, errors.New(errors.ErrCodeParse, fmt.Sprintf("服务 %s 没有定义http接口", service.ServiceName))
	}

	// 生成文件名
	filename := service.OtherOptions["filename"]
	if len(filename) == 0 {
		filename = service.ServiceName
	}

	// 路由组
	apiRouteGroup = ApiGroup{
		Package:     pkgName,
		GroupName:   strcase.UpperCamelCase(service.ServiceName),
		ServiceName: httpApis.Interface,
		GroupRoute:  groupRoute,
		Options:     service.OtherOptions,
	}

	// 创建组目录
	apiRouteGroup.Filepath = filepath.Join(apiGroup, filename+".go")

	// 所有http注解
	for _, api := range httpApis.Apis {
		var ginApi *Api
		ginApi, err = parseHttp(api, service.ServiceName, groupRoute)
		if err != nil {
			return
		}
		if ginApi == nil {
			continue
		}
		ginApi.AnnotationMap = renderAnnotationMap(api.Title, api.Options, apiRouteGroup.Options)
		apiRouteGroup.Apis = append(apiRouteGroup.Apis, ginApi)
	}
	return
}
func isValidRoute(route string) bool {
	return strings.HasPrefix(route, `"`) && strings.HasSuffix(route, `"`)
}

func parseHttp(api ApiAnnotateItem, serviceName, groupRoute string) (ginApi *Api, err error) {
	method := api.Options["method"]
	if len(api.Method) > 0 {
		method = api.Method
		api.Options["method"] = api.Method
	}
	// 命名空间过滤
	namespace := api.Options["ns"]
	if len(namespace) > 0 && namespace != serviceName {
		return
	}

	baseRoute := api.Options["route"]
	// 路由名
	if len(baseRoute) == 0 && len(api.Args) > 0 {
		baseRoute = api.Args[0]
	}

	if !isValidRoute(baseRoute) {
		err = fmt.Errorf("invalid route %s in %s %s", baseRoute, serviceName, api.Handler)
		return ginApi, errors.New(errors.ErrCodeParse, err.Error())
	}

	baseRoute = strings.Trim(baseRoute, `"`)

	// 修正后缀
	fullRoutePath := path.Join(strings.Trim(groupRoute, `"`), baseRoute)
	if fullRoutePath = strings.Trim(fullRoutePath, "/"); strings.HasSuffix(baseRoute, "/") {
		fullRoutePath += "/"
	}
	fullRoutePath = strconv.Quote(fullRoutePath)

	ginApi = &Api{
		Method:     method,
		BaseRoute:  baseRoute,
		HttpMethod: strings.ToUpper(method),
		Route:      fullRoutePath,
		Handler:    api.Handler,
		Params:     api.Params,
		Returns:    api.Returns,
		Title:      api.Title,
		Options:    api.Options,
	}
	return
}

func renderAnnotationMap(title string, m map[string]string, groupM map[string]string) (ret string) {
	var kv []string
	var tmp = make(map[string]string)
	for k, v := range groupM {
		tmp[k] = v
	}
	for k, v := range m {
		tmp[k] = v
	}

	tmp["title"] = title
	for k, v := range tmp {
		kv = append(kv, fmt.Sprintf(`"%s": "%s",`, k, strings.Trim(v, `"`)))
	}
	sort.Strings(kv)

	return fmt.Sprintf(`map[string]string{
		%s
	}`, strings.Join(kv, "\n"))
}
