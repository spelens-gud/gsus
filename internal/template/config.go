package template

// language=yaml
const DefaultConfigYaml = `# go-gsus config
gsus:
  origin:

# 接口实现同步配置
# 会在 项目根路径/${scope} 搜索带@${name}注解的interface
# 并在${path}生成 名为${structName}的接口实现
# 可以指定${template}自定义实现文件生成模板
impls:
  - name: service
    scope: service
    path: internal/service_impls/{{ .ImplPackage }}/{{ .SnakeIfaceName }}.go
    structName: Service
    template: impl
  - name: dao
    scope: internal/dao
    path: internal/dao/dao_impls/{{ .ImplPackage }}/{{ .SnakeIfaceName }}.go
    structName: DaoImpl
    template: impl


# http配置
# 包含三个功能
# router 路由生成
# client 客户端生成
# swagger 接口文档生成
# ${scope}为搜索目录 将会搜索所有 带@service 的interface以下带@http注解的方法
http:
  scope: service

  # 客户端生成能够快速生成近似本地调用的调用桩代码
  client:
    # base为所有调用的最终发起 用户应修改base接口以接收调用参数 可以通过修改${apiTemplate}自定义生成模板
    apiTemplate: http_client_api
    # api为调用桩代码 本质为路由和参数类型的函数名代理调用 生成后不应该人工更改 可以通过修改${baseTemplate}自定义生成模板
    baseTemplate: http_client_base
    # ${path}指定生成客户端代码的目录
    path: clients
    # 路由层生成可以通过模板快速完成参数绑定和服务调用


  router:
    # 可以通过修改${template}自定义生成模板
    template: http_router
    # ${path}指定生成路由层代码的目录
    path: api


  # 接口文档生成遵循swagger(openapi2.0)规范
  # 详情请查阅https://github.com/swaggo/swag注解规范
  swagger:
    # ${path}指定生成swagger.json文件名
    path: docs/swagger.json
    # ${mainApiPath}指定swagger文档头部内容的文件 如无此文件 会自动生成
    mainApiPath: ./api/root.go
    # ${success}为成功返回时的应答格式 Response会自动填充为接口返回值类型
    success: 200 {object} object{data={{ .Response }},ok=bool}
    # ${failed}为失败返回时的应答格式
    failed: 400,500 {object} object{message=string,ok=bool,code=int} "failed"


# 表生成工具能够快速地将sql表结构生成为代码model结构体 并生成泛型调用方法
db2struct:
  # ${path}指定生成的model文件目录
  path: ./internal/model

  # ${typeMap}可以指定生成model结构体时的类型映射 kv格式
  typeMap:
    interface: string

  # ${genericMapTypes}可以指定生成泛型mapping方法时包含的类型 数组
  genericMapTypes: [int,string]

  # ${genericTemplate}指定泛型生成模板 可自定义
  genericTemplate: model_generic

  # 以下参数为数据库连接参数 请使用测试或本地库进行生成
  user:
  password:
  host:
  port:
  db:
  charset:


# 模板生成
# 可以从model(从表结构体) 根据模板快速生成应用各层代码
templates:
  # ${modelPath} 指定model类型的目录
  # 使用gsus template ${arg} 时将会通过此目录下查找名为 ${arg}.go 的文件 并使用 名为 ${arg}的 结构体进行生成
  modelPath:

  # 生成模板列表
  # ${name} 会影响生成模板和生成中的一些默认值
  # ${path} 指定生成的文件名 通过规则渲染而成
  # ${overwrite} 指定是否覆盖生成 如不使用覆盖生成 在已有相同文件时会使用 .bak后缀
  templates:
    - name: service
      path: service/{{ .PackageName }}.go
    - name: dao
      path: internal/dao/{{ .PackageName }}.go
    - name: service_impl
      path: internal/service_impls/svc_{{.PackageName}}/{{ .PackageName }}.go
    - name: dao_impl
      path: internal/dao/dao_impls/dao_{{.PackageName}}/{{ .PackageName }}.go
    - name: model_cast
      path: internal/model/{{ .PackageName }}_cast.go

# autowire 会自动搜索 @autowire 注解并自动生成依赖注入
# ${path}指定生成依赖注入文件的目录
autowire:
  path: ./cmd/inject


# mount 会搜索 @${name} 注解的结构体或初始函数 并将对应类型挂载到指定结构体的字段
# ${path} 指定要挂载的目标文件
# ${struct} 指定要挂载的目标结构体
mounts:
  - name: config
    path: config/config.go
    struct: Config
  - name: service
    path: api/service.go
    struct: Service


enum:
  scope:
  path:
  template:
`
