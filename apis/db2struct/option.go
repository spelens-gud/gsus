package db2struct

import (
	"github.com/spelens-gud/gsus/apis/gengeneric"
)

type (
	opt struct {
		typeReplace    map[string]string
		gormAnnotation bool
		commentOutside bool
		sqlInfo        bool
		json           string
		sqlTag         string
		pkgName        string
		GenericOption  []func(*gengeneric.Options)
	}

	Option func(*opt)
)

func newOpt(options ...Option) *opt {
	o := &opt{
		sqlTag:      "gorm",
		pkgName:     "model",
		typeReplace: map[string]string{},
	}
	for _, option := range options {
		option(o)
	}
	return o
}

func WithGormAnnotation() Option {
	return func(o *opt) {
		o.gormAnnotation = true
	}
}

func WithGenericOption(opts ...func(*gengeneric.Options)) Option {
	return func(o *opt) {
		o.GenericOption = opts
	}
}

func WithSQLTag(tag string) Option {
	return func(o *opt) {
		o.sqlTag = tag
	}
}

func WithPkgName(pkg string) Option {
	return func(o *opt) {
		o.pkgName = pkg
	}
}

func WithCamelJson() Option {
	return func(o *opt) {
		o.json = "camel"
	}
}

func WithSnakeJson() Option {
	return func(o *opt) {
		o.json = "snake"
	}
}

func WithTypeReplace(old, new string) Option {
	return func(o *opt) {
		o.typeReplace[old] = new
	}
}

func WithCommentOutside() Option {
	return func(o *opt) {
		o.commentOutside = true
	}
}
func WithSQLInfo() Option {
	return func(o *opt) {
		o.sqlInfo = true
	}
}
