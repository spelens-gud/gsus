package constant

import (
	"path/filepath"
	"reflect"
	"strings"
)

type t struct{}

func init() {
	GoGetUrl = strings.TrimSuffix(reflect.TypeOf(t{}).PkgPath(), "/internal/constant") + "/cmd/gsus"
}

var (
	GoGetUrl string
)

const (
	ConfigDir      = ".gsus"
	ConfigFile     = ConfigDir + string(filepath.Separator) + "config.yaml"
	TemplateDir    = ConfigDir + string(filepath.Separator) + "templates"
	TemplateSuffix = ".tmpl"
	BackupSuffix   = ".bak"
)
