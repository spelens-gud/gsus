package config

import (
	"text/template"
)

// GenOption struct    代码生成选项.
// 用于配置代码生成的路径和模板.
type GenOption struct {
	ClientsPath  string             // 客户端代码输出路径
	ApiTemplate  *template.Template // API 模板
	BaseTemplate *template.Template // 基础模板
}

// DbOpt struct    数据库转结构体选项.
// 配置数据库表转 Go 结构体的各种参数.
type DbOpt struct {
	TypeReplace    map[string]string // 类型替换映射
	GormAnnotation bool              // 是否生成 GORM 注解
	CommentOutside bool              // 注释是否放在字段外部
	SqlInfo        bool              // 是否包含 SQL 信息
	Json           string            // JSON 标签格式（camel/snake）
	SqlTag         string            // SQL 标签名称
	PkgName        string            // 包名
	GenericOption  []func(*Options)  // 通用选项函数列表
}

// DbOption type    数据库选项函数类型.
type DbOption func(*DbOpt)

// HttpOpt struct    HTTP 代码生成选项.
// 配置 HTTP 客户端和路由代码生成的参数.
type HttpOpt struct {
	Ident string // 服务标识符
}

// HttpOption type    HTTP 选项函数类型.
type HttpOption func(*HttpOpt)

// HttpOptions type    HTTP 选项函数列表.
type HttpOptions []HttpOption

// Options struct    通用生成选项.
// 用于配置代码生成的通用参数.
type Options struct {
	MapTypes []string           // 映射类型列表
	Template *template.Template // 代码模板
}

// Apply method    应用 HTTP 选项.
// 创建默认的 HttpOpt 并应用所有选项函数.
func (os HttpOptions) Apply() *HttpOpt {
	op := &HttpOpt{
		Ident: "service",
	}
	for _, o := range os {
		o(op)
	}
	return op
}

// WithIdent function    设置服务标识符.
// 返回一个设置 Ident 字段的选项函数.
func WithIdent(ident string) HttpOption {
	return func(o *HttpOpt) {
		o.Ident = ident
	}
}

// NewDbOpt function    创建数据库选项.
// 使用默认值创建 DbOpt 并应用提供的选项函数.
func NewDbOpt(options ...DbOption) *DbOpt {
	o := &DbOpt{
		SqlTag:      "gorm",
		PkgName:     "model",
		TypeReplace: map[string]string{},
	}
	for _, option := range options {
		option(o)
	}
	return o
}

// WithGormAnnotation function    启用 GORM 注解.
// 返回一个启用 GormAnnotation 的选项函数.
func WithGormAnnotation() DbOption {
	return func(o *DbOpt) {
		o.GormAnnotation = true
	}
}

// WithGenericOption function    设置通用选项.
// 返回一个设置 GenericOption 的选项函数.
func WithGenericOption(opts ...func(*Options)) DbOption {
	return func(o *DbOpt) {
		o.GenericOption = opts
	}
}

// WithSQLTag function    设置 SQL 标签名称.
// 返回一个设置 SqlTag 的选项函数.
func WithSQLTag(tag string) DbOption {
	return func(o *DbOpt) {
		o.SqlTag = tag
	}
}

// WithPkgName function    设置包名.
// 返回一个设置 PkgName 的选项函数.
func WithPkgName(pkg string) DbOption {
	return func(o *DbOpt) {
		o.PkgName = pkg
	}
}

// WithCamelJson function    使用驼峰命名的 JSON 标签.
// 返回一个设置 JSON 格式为 camel 的选项函数.
func WithCamelJson() DbOption {
	return func(o *DbOpt) {
		o.Json = "camel"
	}
}

// WithSnakeJson function    使用蛇形命名的 JSON 标签.
// 返回一个设置 JSON 格式为 snake 的选项函数.
func WithSnakeJson() DbOption {
	return func(o *DbOpt) {
		o.Json = "snake"
	}
}

// WithTypeReplace function    设置类型替换.
// 返回一个添加类型替换映射的选项函数.
func WithTypeReplace(old, new string) DbOption {
	return func(o *DbOpt) {
		o.TypeReplace[old] = new
	}
}

// WithCommentOutside function    将注释放在字段外部.
// 返回一个启用 CommentOutside 的选项函数.
func WithCommentOutside() DbOption {
	return func(o *DbOpt) {
		o.CommentOutside = true
	}
}

// WithSQLInfo function    包含 SQL 信息.
// 返回一个启用 SqlInfo 的选项函数.
func WithSQLInfo() DbOption {
	return func(o *DbOpt) {
		o.SqlInfo = true
	}
}
