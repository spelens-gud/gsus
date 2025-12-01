package mount

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path"
	"strings"

	"github.com/spelens-gud/gsus/apis/helpers"
	"golang.org/x/tools/go/ast/astutil"
)

type structMounter struct {
	fileSet    *token.FileSet
	astFile    *ast.File
	data       []byte
	structPath string
	structName string
	// object     *ast.Object
	typeSpec *ast.TypeSpec // 替代 ast.Object
}

func NewStructMounter(structPath, structName string) (set *structMounter, err error) {
	b, err := os.ReadFile(structPath)
	if err != nil {
		return
	}
	set = &structMounter{
		fileSet:    token.NewFileSet(),
		data:       b,
		structPath: structPath,
		structName: structName,
	}
	set.astFile, err = parser.ParseFile(set.fileSet, "", set.data, parser.ParseComments)
	if err != nil {
		return
	}
	// set.object = set.astFile.Scope.Lookup(set.structName)
	// if set.object == nil {
	// 	err = errors.New("struct not found")
	// 	return
	// }
	obj := set.astFile.Scope.Lookup(set.structName)
	if obj == nil {
		err = errors.New("struct not found")
		return
	}
	// t, ok := set.object.Decl.(*ast.TypeSpec)
	// if !ok {
	// 	err = fmt.Errorf("%s is not type spec", set.structName)
	// 	return
	// }
	t, ok := obj.Decl.(*ast.TypeSpec)
	if !ok {
		err = fmt.Errorf("%s is not type spec", set.structName)
		return
	}
	if _, ok := t.Type.(*ast.StructType); !ok {
		err = fmt.Errorf("%s is not struct type", set.structName)
		return
	}
	set.typeSpec = t // 存储 TypeSpec 而不是 Object
	return
}

func (sSet *structMounter) Write() (err error) {
	return helpers.ImportAndWrite(sSet.data, sSet.structPath)
}

func (sSet *structMounter) MountTypeField(fieldType, fieldName, pkgPath string) (err error) {
	var (
		// typ    = sSet.object.Decl.(*ast.TypeSpec)
		typ    = sSet.typeSpec // 直接使用 typeSpec
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
			err = errors.New("invalid type with import package path")
			return
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

	log.Printf("mount %s [ %s ] as %s on [ %s.%s ]", pkgPath, fieldType, fieldName, sSet.astFile.Name.Name, sSet.structName)

	if err = sSet.insertField(fieldType, fieldName, fields.Fields); err != nil {
		return
	}

	// append new import
	if len(importAs) > 0 {
		if err = sSet.importPkg(pkgPath, importAs); err != nil {
			return
		}
	}
	return
}

func (sSet *structMounter) importPkg(pkgPath, importAs string) (err error) {
	if path.Base(pkgPath) == importAs {
		importAs = ""
	}
	_ = astutil.AddNamedImport(sSet.fileSet, sSet.astFile, importAs, pkgPath)
	r, err := helpers.FormatAst(sSet.astFile, sSet.fileSet)
	if err != nil {
		return
	}
	return sSet.freshFile([]byte(r))
}

func (sSet *structMounter) freshFile(data []byte) (err error) {
	sSet.data = data
	sSet.fileSet = token.NewFileSet()
	if sSet.astFile, err = parser.ParseFile(sSet.fileSet, "", sSet.data, parser.ParseComments); err != nil {
		fmt.Printf("%s", sSet.data)
		return
	}
	// sSet.object = sSet.astFile.Scope.Lookup(sSet.structName)
	obj := sSet.astFile.Scope.Lookup(sSet.structName)
	if obj != nil {
		if t, ok := obj.Decl.(*ast.TypeSpec); ok {
			if _, ok := t.Type.(*ast.StructType); ok {
				sSet.typeSpec = t
			}
		}
	}
	return
}

func (sSet *structMounter) getImportPkgName(pkgPath string) (name string, imported bool) {
	if len(pkgPath) == 0 {
		return
	}
	defaultImportName := path.Base(pkgPath)
	name = defaultImportName

	usedPkgName := make(map[string]bool)
	for _, imp := range sSet.astFile.Imports {
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

func (sSet *structMounter) insertField(fieldType, name string, list *ast.FieldList) (err error) {
	bf := &bytes.Buffer{}
	bf.Write(sSet.data[:list.End()-2])
	if sSet.data[list.End()-3] == '\n' {
		bf.WriteString(fmt.Sprintf(`    %s %s`, name, fieldType) + "\n")
	} else {
		bf.WriteString("\n" + fmt.Sprintf(`    %s %s`, name, fieldType))
	}
	bf.Write(sSet.data[list.End()-2:])
	return sSet.freshFile(bf.Bytes())
}
