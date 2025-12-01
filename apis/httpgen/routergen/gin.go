package routergen

import (
	"text/template"

	"github.com/spelens-gud/gsus/basetmpl"
)

var defaultRouterTemplate = template.Must(template.New("svc").Parse(basetmpl.DefaultHttpRouterTemplate))
