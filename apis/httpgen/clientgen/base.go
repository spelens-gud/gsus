package clientgen

import (
	"text/template"

	"github.com/spelens-gud/gsus/basetmpl"
)

var defaultBaseTemplate = template.Must(template.New("base").Parse(basetmpl.DefaultHttpClientBaseTemplate))
