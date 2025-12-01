package parser

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"strings"

	"github.com/spelens-gud/gsus/internal/errors"
	"github.com/spelens-gud/gsus/internal/logger"
	"github.com/spelens-gud/gsus/internal/utils"
	"golang.org/x/tools/go/ast/astutil"
)

type StructMounter struct {
	FileSet    *token.FileSet
	AstFile    *ast.File
	Data       []byte
	StructPath string
	StructName string
	TypeSpec   *ast.TypeSpec
}

func NewStructMounter(structPath, structName string) (set *StructMounter, err error) {
	b, err := os.ReadFile(structPath)
	if err != nil {
		return set, errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("读取配置文件失败: %s", structPath))
	}
	set = &StructMounter{
		FileSet:    token.NewFileSet(),
		Data:       b,
		StructPath: structPath,
		StructName: structName,
	}
	set.AstFile, err = parser.ParseFile(set.FileSet, "", set.Data, parser.ParseComments)
	if err != nil {
		return set, errors.WrapWithCode(err, errors.ErrCodeParse, fmt.Sprintf("解析文件失败: %s", err))
	}

	obj := set.AstFile.Scope.Lookup(set.StructName)
	if obj == nil {
		return set, errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("结构未找到: %s", err))
	}

	t, ok := obj.Decl.(*ast.TypeSpec)
	if !ok {
		err = fmt.Errorf("%s is not type spec", set.StructName)
		return set, errors.New(errors.ErrCodeGenerate, err.Error())
	}
	if _, ok := t.Type.(*ast.StructType); !ok {
		err = fmt.Errorf("%s is not struct type", set.StructName)
		return set, errors.New(errors.ErrCodeGenerate, err.Error())
	}
	set.TypeSpec = t
	return
}

func (sSet *StructMounter) Write() (err error) {
	return utils.ImportAndWrite(sSet.Data, sSet.StructPath)
}

func (sSet *StructMounter) MountTypeField(fieldType, fieldName, pkgPath string) (err error) {
	var (
		// typ    = sSet.object.Decl.(*ast.TypeSpec)
		typ    = sSet.TypeSpec // 直接使用 typeSpec
		fields = typ.Type.(*ast.StructType)
		list   = fields.Fields.List
	)

	pkgPath = strings.Trim(pkgPath, `"`)
	splitIdent := strings.Split(fieldType, ".")

	if len(fieldName) == 0 {
		if len(splitIdent) > 1 {
			fieldName = splitIdent[1]
		} else {
			fieldName = splitIdent[0]
		}
	}

	var importAs string
	// check imports and fix import pkg name
	if len(pkgPath) > 0 {
		if len(splitIdent) != 2 {
			return errors.New(errors.ErrCodeParse, "invalid type with import package path")
		}
		importPkgName, imported := sSet.getImportPkgName(pkgPath)
		if !imported {
			importAs = importPkgName
		}

		if len(splitIdent) > 1 && splitIdent[0] != importPkgName {
			splitIdent[0] = importPkgName
			fieldType = strings.Join(splitIdent, ".")
		}
	}

	getExprTypeName := func(expr ast.Expr) string {
		// 检查是否已经挂载
		switch t := expr.(type) {
		case *ast.SelectorExpr:
			fpkg, ok := t.X.(*ast.Ident)
			if !ok {
				return ""
			}
			return fpkg.Name + "." + t.Sel.Name
		case *ast.Ident:
			return t.String()
		default:
			return ""
		}
	}

	usedFieldName := make(map[string]bool)

	for _, f := range list {
		fType := getExprTypeName(f.Type)
		// 类型已挂载
		if len(fType) > 0 && fType == fieldType {
			return
		}
		if len(f.Names) == 0 {
			usedFieldName[fType] = true
		} else {
			for _, name := range f.Names {
				usedFieldName[name.String()] = true
			}
		}
	}

	for {
		if usedFieldName[fieldName] {
			fieldName += "2"
		} else {
			break
		}
	}

	logger.Info("mount %s [ %s ] as %s on [ %s.%s ]", pkgPath, fieldType, fieldName, sSet.AstFile.Name.Name, sSet.StructName)

	if err = sSet.insertField(fieldType, fieldName, fields.Fields); err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("挂载字段失败: %s", err))
	}

	// append new import
	if len(importAs) > 0 {
		if err = sSet.importPkg(pkgPath, importAs); err != nil {
			return errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("导入包: %s", err))
		}
	}
	return
}

func (sSet *StructMounter) importPkg(pkgPath, importAs string) (err error) {
	if path.Base(pkgPath) == importAs {
		importAs = ""
	}
	_ = astutil.AddNamedImport(sSet.FileSet, sSet.AstFile, importAs, pkgPath)
	r, err := utils.FormatAst(sSet.AstFile, sSet.FileSet)
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("格式化文件失败: %s", err))
	}
	return sSet.freshFile([]byte(r))
}

func (sSet *StructMounter) freshFile(data []byte) (err error) {
	sSet.Data = data
	sSet.FileSet = token.NewFileSet()
	if sSet.AstFile, err = parser.ParseFile(sSet.FileSet, "", sSet.Data, parser.ParseComments); err != nil {
		fmt.Printf("%s", sSet.Data)
		return errors.WrapWithCode(err, errors.ErrCodeParse, fmt.Sprintf("解析文件失败: %s", err))
	}
	// sSet.object = sSet.astFile.Scope.Lookup(sSet.structName)
	obj := sSet.AstFile.Scope.Lookup(sSet.StructName)
	if obj != nil {
		if t, ok := obj.Decl.(*ast.TypeSpec); ok {
			if _, ok := t.Type.(*ast.StructType); ok {
				sSet.TypeSpec = t
			}
		}
	}
	return
}

func (sSet *StructMounter) getImportPkgName(pkgPath string) (name string, imported bool) {
	if len(pkgPath) == 0 {
		return
	}
	defaultImportName := path.Base(pkgPath)
	name = defaultImportName

	usedPkgName := make(map[string]bool)
	for _, imp := range sSet.AstFile.Imports {
		if v := strings.Trim(imp.Path.Value, `"`); v == pkgPath {
			if imp.Name != nil {
				return imp.Name.Name, true
			} else {
				return defaultImportName, true
			}
		} else {
			if imp.Name != nil {
				usedPkgName[imp.Name.Name] = true
			} else {
				usedPkgName[path.Base(v)] = true
			}
		}
	}
	for {
		if usedPkgName[name] {
			name += "2"
		} else {
			break
		}
	}
	return name, false
}

func (sSet *StructMounter) insertField(fieldType, name string, list *ast.FieldList) (err error) {
	bf := &bytes.Buffer{}
	bf.Write(sSet.Data[:list.End()-2])
	if sSet.Data[list.End()-3] == '\n' {
		bf.WriteString(fmt.Sprintf(`    %s %s`, name, fieldType) + "\n")
	} else {
		bf.WriteString("\n" + fmt.Sprintf(`    %s %s`, name, fieldType))
	}
	bf.Write(sSet.Data[list.End()-2:])
	return sSet.freshFile(bf.Bytes())
}
