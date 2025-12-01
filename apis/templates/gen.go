package templates

import (
	"errors"
	"go/ast"
	"path/filepath"
	"strings"

	"github.com/spelens-gud/gsus/apis/helpers"
	"github.com/spelens-gud/gsus/internal/fileconfig"
	"github.com/stoewer/go-strcase"
)

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

func getTableFile(tableName, modelPath string) (data []byte, astFile *ast.File, err error) {
	path := filepath.Join(modelPath, tableName+".go")
	astFile, _, data, err = helpers.ParseFileAst(path)
	return
}

type Config struct {
	ModelName string
	ModelPath string
	Overwrite bool
	Templates []fileconfig.Template
}

func Gen(cfg Config, mainConfig fileconfig.Config) (err error) {
	tableBytes, astFile, err := getTableFile(cfg.ModelName, cfg.ModelPath)
	if err != nil {
		return
	}

	templateStruct, err := parseTemplates(cfg.ModelName, tableBytes, astFile)
	if err != nil || templateStruct == nil {
		return
	}

	if len(templateStruct.ServicePkgPath) == 0 {
		err = errors.New("service pkg config not found,please check implements config")
		return
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
