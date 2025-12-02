package generator

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/errors"
	"github.com/spelens-gud/gsus/internal/parser"
	"github.com/spelens-gud/gsus/internal/utils"
	"github.com/stoewer/go-strcase"
)

// Option 配置选项结构体，用于指定挂载目标的路径和结构体名称.
type Option struct {
	Path   string // 结构体所在的文件路径
	Struct string // 要挂载的目标结构体名称
}

// MatchStruct 匹配到的结构体信息，包含类型、字段名、包名等信息.
type MatchStruct struct {
	Type              string // 结构体类型，格式为 包名.结构体名
	FieldName         string // 字段名称
	Package           string // 包名
	AnnotationContent string // 注解内容
	Path              string // 文件路径
}

// Exec 执行挂载操作的主要入口函数
// 根据提供的配置信息，查找并处理带有特定注解的结构体，然后执行字段挂载操作.
func Exec(cfg config.Mount) (err error) {
	// 如果作用域未设置，默认为当前目录
	if len(cfg.Scope) == 0 {
		cfg.Scope = "./"
	}

	// 如果名称未设置，默认为"mount"
	if len(cfg.Name) == 0 {
		cfg.Name = "mount"
	}

	// 查找所有匹配的挂载目标结构体
	mountTargetStructs := matchFields(cfg.Scope, cfg.Name, false)
	if len(mountTargetStructs) == 0 {
		return nil
	}

	// 构建需要特殊处理的名称映射表
	specName := make(map[string]bool)
	for _, n := range cfg.Args {
		specName[strings.TrimSpace(n)] = true
	}

	// 遍历所有匹配到的结构体并执行挂载操作
	for _, st := range mountTargetStructs {
		sp := strings.Split(st.Type, ".")
		st.Type = sp[len(sp)-1]

		var match []MatchStruct

		// 根据注解内容中的标识符查找对应的字段
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

		// 执行具体的字段挂载操作
		if err = ExecFields(Option{
			Path:   st.Path,
			Struct: st.Type,
		}, match); err != nil {
			return errors.WrapWithCode(err, errors.ErrCodeExecute, fmt.Sprintf("执行挂载失败: %s", err))
		}
	}
	return
}

// matchFields 根据给定的作用域和标识符匹配相应的字段
// scope: 搜索范围（目录路径）
// ident: 要匹配的标识符
// funcParams: 是否匹配函数参数.
func matchFields(scope string, ident string, funcParams bool) (fields []MatchStruct) {
	// 编译正则表达式，用于匹配 @ident(...) 格式的注解
	regexConfig, err := regexp.Compile(`@` + ident + `\\((.*?)\\)`)
	if err != nil {
		return
	}

	// 使用互斥锁保证并发安全
	mu := sync.Mutex{}

	// 遍历指定作用域下的所有文件
	if err = utils.ExecFiles(scope, func(path string) (err error) {
		// 解析文件的AST信息和原始数据
		astFile, _, data, err := utils.ParseFileAst(path)
		if err != nil || !regexConfig.Match(data) {
			return
		}

		// 获取文件所在目录的包名
		dirPkg, _ := utils.GetPathModPkg(filepath.Dir(path))

		// 匹配注解并获取相关字段信息
		fields2 := matchAnnotations(regexConfig, astFile, dirPkg, funcParams)
		for i := range fields2 {
			f := &fields2[i]
			if f.Package == dirPkg {
				f.Path = path
			}
		}

		// 添加到结果集中
		mu.Lock()
		fields = append(fields, fields2...)
		mu.Unlock()
		return
	}); err != nil {
		return
	}
	return
}

// ExecFields 执行具体的字段挂载操作
// cfg: 配置选项，包括目标文件路径和结构体名称
// fields: 需要挂载的字段列表
func ExecFields(cfg Option, fields []MatchStruct) (err error) {
	// 修复文件路径为项目内的相对路径
	if err = utils.FixFilepathByProjectDir(&cfg.Path); err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeExecute, fmt.Sprintf("修复文件路径失败: %s", err))
	}

	// 移除重复的字段
	removeDuplicate(fields)

	// 获取路径对应的包名
	pathPkg, err := utils.GetPathModPkg(filepath.Dir(cfg.Path))
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeExecute, fmt.Sprintf("获取路径包失败: %s", err))
	}

	// 创建结构体挂载器实例
	mounter, err := parser.NewStructMounter(cfg.Path, cfg.Struct)
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeExecute, fmt.Sprintf("创建结构体失败: %s", err))
	}

	// 对字段进行排序
	sort.Slice(fields, func(i, j int) bool {
		stri, _ := json.Marshal(fields[i])
		strj, _ := json.Marshal(fields[i])
		return string(stri) > string(strj)
	})

	// 遍历所有字段并执行挂载
	for _, field := range fields {
		if field.Package == pathPkg {
			// 去除本身就属于挂载目录结构体的import前缀
			tmp := strings.Split(field.Type, ".")
			field.Type = tmp[len(tmp)-1]
			field.Package = ""
		}

		// 执行字段类型的挂载
		if err = mounter.MountTypeField(field.Type, strcase.UpperCamelCase(field.FieldName), field.Package); err != nil {
			return errors.WrapWithCode(err, errors.ErrCodeExecute, fmt.Sprintf("挂载字段失败: %s", err))
		}
	}

	// 将变更写入文件
	return mounter.Write()
}

// removeDuplicate 移除匹配结果中的重复项
// arr: 包含可能重复的MatchStruct的切片
// 返回值: 去重后的MatchStruct切片.
func removeDuplicate(arr []MatchStruct) []MatchStruct {
	resArr := make([]MatchStruct, 0)
	tmpMap := make(map[string]bool)

	// 使用map来去重，key为 Type#FieldName 的格式
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

// matchAnnotations 匹配注解信息并提取相关字段
// re: 用于匹配注解的正则表达式
// astF: AST文件节点
// dirPkg: 当前目录的包名
// fieldParams: 是否处理函数参数字段.
func matchAnnotations(re *regexp.Regexp, astF *ast.File, dirPkg string, fieldParams bool) (fields []MatchStruct) {
	// 遍历所有的声明语句
	for _, d := range astF.Decls {
		switch t := d.(type) {
		case *ast.FuncDecl:
			// 如果不处理函数参数，则跳过函数声明
			if !fieldParams {
				continue
			}
			// 处理函数参数中的字段
			fields = append(fields, getFuncParamsFields(re, astF, t, dirPkg)...)
		case *ast.GenDecl:
			// 处理一般声明（如type声明）
			for _, s := range t.Specs {
				switch spec := s.(type) {
				case *ast.TypeSpec:
					// 如果没有文档注释，则跳过
					if t.Doc == nil || t.Doc.List == nil {
						continue
					}

					// 查找匹配的注解
					var match []string
					for _, l := range t.Doc.List {
						if match = re.FindStringSubmatch(strings.TrimPrefix(l.Text, "//")); len(match) == 2 {
							break
						}
					}

					if len(match) == 0 {
						continue
					}

					// 添加匹配到的结构体信息
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

// getFuncParamsFields 从函数参数中提取字段信息
// re: 用于匹配注解的正则表达式
// f: AST文件节点
// fd: 函数声明节点
// pkg: 包名
func getFuncParamsFields(re *regexp.Regexp, f *ast.File, fd *ast.FuncDecl, pkg string) (fields []MatchStruct) {
	// 检查函数是否有文档注释以及参数列表
	if fd.Doc == nil || len(fd.Doc.List) == 0 || fd.Type == nil || fd.Type.Params == nil || len(fd.Type.Params.List) == 0 {
		return
	}

	// 解析注解中的参数映射关系
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

// getImportFullPkgByName 根据导入名称获取完整的包路径
// imports: 导入规范列表
// name: 要查找的导入名称
// 返回值: 完整的包路径和是否找到的布尔值.
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
