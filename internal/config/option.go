package config

import (
	"text/template"
)

type GenOption struct {
	ClientsPath  string
	ApiTemplate  *template.Template
	BaseTemplate *template.Template
}

type DbOpt struct {
	TypeReplace    map[string]string
	GormAnnotation bool
	CommentOutside bool
	SqlInfo        bool
	Json           string
	SqlTag         string
	PkgName        string
	GenericOption  []func(*Options)
}
type DbOption func(*DbOpt)

type HttpOpt struct {
	Ident string
}
type HttpOption func(*HttpOpt)
type HttpOptions []HttpOption
type Options struct {
	MapTypes []string
	Template *template.Template
}

func (os HttpOptions) Apply() *HttpOpt {
	op := &HttpOpt{
		Ident: "service",
	}
	for _, o := range os {
		o(op)
	}
	return op
}

func WithIdent(ident string) HttpOption {
	return func(o *HttpOpt) {
		o.Ident = ident
	}
}

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

func WithGormAnnotation() DbOption {
	return func(o *DbOpt) {
		o.GormAnnotation = true
	}
}

func WithGenericOption(opts ...func(*Options)) DbOption {
	return func(o *DbOpt) {
		o.GenericOption = opts
	}
}

func WithSQLTag(tag string) DbOption {
	return func(o *DbOpt) {
		o.SqlTag = tag
	}
}

func WithPkgName(pkg string) DbOption {
	return func(o *DbOpt) {
		o.PkgName = pkg
	}
}

func WithCamelJson() DbOption {
	return func(o *DbOpt) {
		o.Json = "camel"
	}
}

func WithSnakeJson() DbOption {
	return func(o *DbOpt) {
		o.Json = "snake"
	}
}

func WithTypeReplace(old, new string) DbOption {
	return func(o *DbOpt) {
		o.TypeReplace[old] = new
	}
}

func WithCommentOutside() DbOption {
	return func(o *DbOpt) {
		o.CommentOutside = true
	}
}

func WithSQLInfo() DbOption {
	return func(o *DbOpt) {
		o.SqlInfo = true
	}
}
