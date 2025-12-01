package generator

import (
	"encoding/json"
	"go/ast"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/parser"
	"github.com/spelens-gud/gsus/internal/utils"
	"github.com/stoewer/go-strcase"
)

type Option struct {
	Path   string
	Struct string
}
type MatchStruct struct {
	Type              string
	FieldName         string
	Package           string
	AnnotationContent string
	Path              string
}

func Exec(cfg config.Mount) (err error) {
	if len(cfg.Scope) == 0 {
		cfg.Scope = "./"
	}

	if len(cfg.Name) == 0 {
		cfg.Name = "mount"
	}

	mountTargetStructs := matchFields(cfg.Scope, cfg.Name, false)
	if len(mountTargetStructs) == 0 {
		return
	}

	specName := make(map[string]bool)
	for _, n := range cfg.Args {
		specName[strings.TrimSpace(n)] = true
	}

	for _, st := range mountTargetStructs {
		sp := strings.Split(st.Type, ".")
		st.Type = sp[len(sp)-1]

		var match []MatchStruct

		for _, ident := range strings.Split(st.AnnotationContent, ",") {
			ident = strings.TrimSpace(ident)
			if len(specName) > 0 && !specName[ident] {
				continue
			}
			match = append(match, matchFields(cfg.Scope, ident, true)...)
		}
		if len(match) == 0 {
			continue
		}
		if err = ExecFields(Option{
			Path:   st.Path,
			Struct: st.Type,
		}, match); err != nil {
			return
		}
	}
	return
}

func matchFields(scope string, ident string, funcParams bool) (fields []MatchStruct) {
	regexConfig, err := regexp.Compile(`@` + ident + `\((.*?)\)`)
	if err != nil {
		return
	}
	mu := sync.Mutex{}
	if err = utils.ExecFiles(scope, func(path string) (err error) {
		astFile, _, data, err := utils.ParseFileAst(path)
		if err != nil || !regexConfig.Match(data) {
			return
		}
		dirPkg, _ := utils.GetPathModPkg(filepath.Dir(path))
		fields2 := matchAnnotations(regexConfig, astFile, dirPkg, funcParams)
		for i := range fields2 {
			f := &fields2[i]
			if f.Package == dirPkg {
				f.Path = path
			}
		}

		mu.Lock()
		fields = append(fields, fields2...)
		mu.Unlock()
		return
	}); err != nil {
		return
	}
	return
}

func ExecFields(cfg Option, fields []MatchStruct) (err error) {
	if err = utils.FixFilepathByProjectDir(&cfg.Path); err != nil {
		return
	}

	removeDuplicate(fields)

	pathPkg, err := utils.GetPathModPkg(filepath.Dir(cfg.Path))
	if err != nil {
		return
	}

	mounter, err := parser.NewStructMounter(cfg.Path, cfg.Struct)
	if err != nil {
		return
	}

	sort.Slice(fields, func(i, j int) bool {
		stri, _ := json.Marshal(fields[i])
		strj, _ := json.Marshal(fields[i])
		return string(stri) > string(strj)
	})

	for _, field := range fields {
		if field.Package == pathPkg {
			// 去除本身就属于挂载目录结构体的import前缀
			tmp := strings.Split(field.Type, ".")
			field.Type = tmp[len(tmp)-1]
			field.Package = ""
		}

		if err = mounter.MountTypeField(field.Type, strcase.UpperCamelCase(field.FieldName), field.Package); err != nil {
			return
		}
	}

	return mounter.Write()
}

func removeDuplicate(arr []MatchStruct) []MatchStruct {
	resArr := make([]MatchStruct, 0)
	tmpMap := make(map[string]bool)
	for _, val := range arr {
		k := val.Type + "#" + val.FieldName
		if tmpMap[k] {
			continue
		}
		resArr = append(resArr, val)
		tmpMap[k] = true
	}
	return resArr
}

func matchAnnotations(re *regexp.Regexp, astF *ast.File, dirPkg string, fieldParams bool) (fields []MatchStruct) {
	for _, d := range astF.Decls {
		switch t := d.(type) {
		case *ast.FuncDecl:
			if !fieldParams {
				continue
			}
			fields = append(fields, getFuncParamsFields(re, astF, t, dirPkg)...)
		case *ast.GenDecl:
			for _, s := range t.Specs {
				switch spec := s.(type) {
				case *ast.TypeSpec:
					if t.Doc == nil || t.Doc.List == nil {
						continue
					}
					var match []string
					for _, l := range t.Doc.List {
						if match = re.FindStringSubmatch(strings.TrimPrefix(l.Text, "//")); len(match) == 2 {
							break
						}
					}

					if len(match) == 0 {
						continue
					}

					fields = append(fields, MatchStruct{
						Type:              astF.Name.Name + "." + spec.Name.Name,
						FieldName:         spec.Name.Name,
						Package:           dirPkg,
						AnnotationContent: match[1],
					})
				}
			}
		}
	}
	return
}

func getFuncParamsFields(re *regexp.Regexp, f *ast.File, fd *ast.FuncDecl, pkg string) (fields []MatchStruct) {
	if fd.Doc == nil || len(fd.Doc.List) == 0 || fd.Type == nil || fd.Type.Params == nil || len(fd.Type.Params.List) == 0 {
		return
	}

	paramsMap := make(map[string]string)
	for _, doc := range fd.Doc.List {
		find := re.FindStringSubmatch(doc.Text)
		if len(find) != 2 {
			continue
		}
		for _, v := range strings.Split(find[1], ",") {
			if kvs := strings.Split(v, "="); len(kvs) == 2 {
				paramsMap[kvs[0]] = kvs[1]
			} else {
				paramsMap[kvs[0]] = ""
			}
		}
	}

	// 从参数提取类型
	for _, param := range fd.Type.Params.List {
		if len(param.Names) != 1 {
			continue
		}
		paramType, ok := paramsMap[param.Names[0].Name]
		if !ok {
			continue
		}
		var nf *MatchStruct
		switch se := param.Type.(type) {
		// 从其他包导入的结构体
		case *ast.SelectorExpr:
			id, ok := se.X.(*ast.Ident)
			if !ok {
				continue
			}
			// 检查导入名
			if fullPkg, ok := getImportFullPkgByName(f.Imports, id.Name); ok {
				nf = &MatchStruct{
					Type:      id.Name + "." + se.Sel.Name,
					FieldName: se.Sel.Name,
					Package:   fullPkg,
				}
			}
		// 同个包的配置
		case *ast.Ident:
			if se.Obj == nil {
				continue
			}
			nf = &MatchStruct{
				Type:      f.Name.Name + "." + se.Name,
				Package:   pkg,
				FieldName: se.Name,
			}
		}
		// 无法匹配的类型
		if nf == nil {
			continue
		}
		if len(paramType) > 0 {
			nf.FieldName = paramType
		}
		fields = append(fields, *nf)
	}
	return
}

func getImportFullPkgByName(imports []*ast.ImportSpec, name string) (pkg string, ok bool) {
	for _, imp := range imports {
		v := strings.Trim(imp.Path.Value, `"`)
		if imp.Name != nil && imp.Name.Name == name {
			return v, true
		} else if imp.Name == nil && (strings.HasSuffix(v, "/"+name) || v == name) {
			return v, true
		}
	}
	return
}
