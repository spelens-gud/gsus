package generator

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"strings"

	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/errors"
	"github.com/spelens-gud/gsus/internal/utils"
	"github.com/stoewer/go-strcase"
)

type TemplateConfig struct {
	ModelName string
	ModelPath string
	Overwrite bool
	Templates []config.Template
}
type Template struct {
	Name           string
	ModelName      string
	ModelDesc      string
	PackageName    string
	ServicePkgPath string
	ServicePkg     string
	ModelPkg       string
	StructName     string
	CallerIdent    string
	Fields         []Field
}
type Field struct {
	Name    string
	Type    string
	Comment string
}

func GenTemplate(cfg TemplateConfig, mainConfig config.Option) (err error) {
	tableBytes, astFile, err := getTableFile(cfg.ModelName, cfg.ModelPath)
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeTemplate, fmt.Sprintf("获取表文件: %s", err))
	}

	templateStruct, err := parseTemplates(cfg.ModelName, tableBytes, astFile)
	if err != nil || templateStruct == nil {
		return
	}

	if len(templateStruct.ServicePkgPath) == 0 {
		return errors.New(errors.ErrCodeTemplate, "服务包配置未找到，请检查实现配置")
	}

	templateStruct.ServicePkg = filepath.Base(templateStruct.ServicePkgPath)

	for _, temp := range cfg.Templates {
		if cfg.Overwrite {
			temp.Overwrite = cfg.Overwrite
		}
	}
	return
}

func parseTemplates(tableName string, tableBytes []byte, astFile *ast.File) (tmpl *Template, err error) {
	object := astFile.Scope.Lookup(strcase.UpperCamelCase(tableName))
	if object == nil {
		return
	}

	typeSpec, ok := object.Decl.(*ast.TypeSpec)
	if !ok {
		return
	}
	structType := typeSpec.Type.(*ast.StructType)

	tmpl = &Template{
		ModelName:   strcase.UpperCamelCase(tableName),
		PackageName: strcase.SnakeCase(tableName),
		ModelPkg:    astFile.Name.Name,
	}

	if len(astFile.Doc.List) > 1 {
		tmpl.ModelDesc = strings.TrimSpace(strings.TrimPrefix(astFile.Doc.List[1].Text, "//"))
	}

	for _, l := range structType.Fields.List {
		var name string
		switch typ := l.Type.(type) {
		case *ast.Ident:
			name = typ.Name
		default:
			name = string(tableBytes[typ.Pos()-1 : typ.End()-1])
		}

		field := Field{
			Name: l.Names[0].Name,
			Type: name,
		}
		if l.Comment != nil {
			field.Comment = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(l.Comment.Text()), "//"))
		}
		tmpl.Fields = append(tmpl.Fields, field)
	}
	return
}

func getTableFile(tableName, modelPath string) (data []byte, astFile *ast.File, err error) {
	path := filepath.Join(modelPath, tableName+".go")
	astFile, _, data, err = utils.ParseFileAst(path)
	if err != nil {
		return data, astFile, errors.WrapWithCode(err, errors.ErrCodeTemplate, fmt.Sprintf("解析文件失败: %s", err))
	}
	return
}
