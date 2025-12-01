package gengeneric

import (
	"text/template"

	"github.com/spelens-gud/gsus/apis/helpers"
	"github.com/spelens-gud/gsus/basetmpl"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var defaultTemplate *template.Template

type T struct {
	MapTypes []MapType
	Type     string
	Package  string
}

type MapType struct {
	MapType  string
	MapBType string
}

func init() {
	defaultTemplate = template.Must(template.New("").Parse(basetmpl.DefaultModelGenericTemplate))
}

type Options struct {
	MapTypes []string
	Template *template.Template
}

func NewType(typeName, pkg string, opts ...func(o *Options)) (res string, err error) {
	t := T{
		Type:    typeName,
		Package: pkg,
	}
	o := &Options{
		MapTypes: []string{"int", "string"},
		Template: defaultTemplate,
	}
	for _, opt := range opts {
		opt(o)
	}

	for _, typ := range o.MapTypes {
		caser := cases.Title(language.English)
		t.MapTypes = append(t.MapTypes, MapType{
			MapType:  caser.String(typ),
			MapBType: typ,
		})
	}

	ret, err := helpers.ExecuteTemplate(o.Template, t)
	if err != nil {
		return
	}
	res = string(ret)
	return
}
