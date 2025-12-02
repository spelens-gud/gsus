package config

import (
	"path/filepath"
	"reflect"
	"strings"
)

// empty struct    用于获取包路径的空结构体.
type empty struct{}

// init function    初始化包级变量.
// 通过反射获取当前包的导入路径，用于确定 go get 的 URL.
func init() {
	GoGetUrl = strings.TrimSuffix(reflect.TypeOf(empty{}).PkgPath(), "/internal/config")
}

var (
	// GoGetUrl var    Go 模块的导入路径.
	GoGetUrl string
)

const (
	// ConfigDir 配置目录名称.
	ConfigDir = ".gsus"
	// ConfigFile 配置文件路径.
	ConfigFile = ConfigDir + string(filepath.Separator) + "config.yaml"
	// TemplateDir 模板目录路径.
	TemplateDir = ConfigDir + string(filepath.Separator) + "templates"
	// TemplateSuffix 模板文件后缀.
	TemplateSuffix = ".tmpl"
	// BackupSuffix 备份文件后缀.
	BackupSuffix = ".bak"
)
