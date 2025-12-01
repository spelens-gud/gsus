package initial

import (
	"github.com/spelens-gud/gsus/basetmpl"
)

var templateMap = map[string]string{
	"impl":             basetmpl.DefaultImplTemplate,
	"http_router":      basetmpl.DefaultHttpRouterTemplate,
	"http_client_api":  basetmpl.DefaultHttpClientApiTemplate,
	"http_client_base": basetmpl.DefaultHttpClientBaseTemplate,
	"dao":              basetmpl.DefaultDaoTemplate,
	"dao_impl":         basetmpl.DefaultDaoImplTemplate,
	"service":          basetmpl.DefaultServiceTemplate,
	"service_impl":     basetmpl.DefaultServiceImplTemplate,
	"model_cast":       basetmpl.DefaultModelCastTemplate,
	"model_generic":    basetmpl.DefaultModelGenericTemplate,
}
