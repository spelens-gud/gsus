package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"text/template"

	"github.com/spelens-gud/gsus/internal/errors"
	"github.com/spelens-gud/gsus/internal/logger"
	template2 "github.com/spelens-gud/gsus/internal/template"
	"github.com/spelens-gud/gsus/internal/utils"
	"github.com/stoewer/go-strcase"
	"golang.org/x/sync/errgroup"
)

var defaultImplTemplate = template.Must(template.New("impl").Parse(template2.DefaultImplTemplate))

type Impl struct {
	InterfaceName  string
	ImplPackage    string
	SetName        string
	ImplStructName string
}

type implsSync struct {
	ImplStructName       string
	ImplPackage          string
	InterfacePackagePath string
	InterfacePackageName string
	InterfaceName        string
	SetName              string

	implDir            string
	ifaceAstType       *ast.InterfaceType
	implStructTemplate *template.Template
	interfaceFileSet   *token.FileSet
}

type Config struct {
	ImplBaseTemplate *template.Template
	SetName          string
	Scope            string
	ImplementsDir    string
	Prefix           string
}

// SyncInterfaceImpls method    同步接口实现.
func (cfg Config) SyncInterfaceImpls() (err error) {
	if len(cfg.SetName) == 0 {
		return errors.New(errors.ErrCodeConfig, "无效的实现集名称")
	}

	if cfg.ImplBaseTemplate == nil {
		cfg.ImplBaseTemplate = defaultImplTemplate
	}

	if len(cfg.Scope) == 0 {
		cfg.Scope = "./"
	}

	if len(cfg.Prefix) == 0 {
		cfg.Prefix = cfg.SetName
	}

	if len(cfg.ImplementsDir) == 0 {
		cfg.ImplementsDir = filepath.Join("./", "internal", fmt.Sprintf("%s_impls", strcase.SnakeCase(cfg.Prefix)))
	}

	interfaces := matchInterface(cfg.Scope, cfg.SetName)
	wg := errgroup.Group{}
	for _, item := range interfaces {
		item := item
		wg.Go(func() (err error) {
			implPackage := fmt.Sprintf("%s_%s", strcase.SnakeCase(cfg.Prefix), strcase.SnakeCase(item.Name))
			targetDir := filepath.Join(cfg.ImplementsDir, implPackage)
			syncer := implsSync{
				InterfacePackagePath: item.PackagePath,
				InterfacePackageName: item.PackageName,
				InterfaceName:        item.Name,
				ImplPackage:          implPackage,
				ImplStructName:       strcase.UpperCamelCase(cfg.SetName),
				SetName:              cfg.SetName,
				ifaceAstType:         item.IfaceType,
				implDir:              targetDir,
				implStructTemplate:   cfg.ImplBaseTemplate,
				interfaceFileSet:     item.fileSet,
			}
			var updated int
			updated, err = syncer.sync()
			if err != nil {
				logger.Error("sync interface [ %s.%s ] err: %v", item.PackageName, item.Name, err)
				return errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("同步接口实现失败%s", err))
			}
			logger.Info("interface [ %s.%s ] sync finished,[ %d ] functions updated", item.PackageName, item.Name, updated)
			return nil
		})
	}
	return wg.Wait()
}

func (s implsSync) walkImplDir(f func(fp string) (err error)) (err error) {
	return filepath.WalkDir(s.implDir, func(fp string, d fs.DirEntry, _ error) (err error) {
		if d.IsDir() && fp != s.implDir {
			return filepath.SkipDir
		}
		if !strings.HasSuffix(fp, ".go") || strings.HasSuffix(fp, "_test.go") {
			return nil
		}
		return f(fp)
	})
}

// 获取结构体已实现的方法
func (s implsSync) getImplementFunc(astF *ast.File) map[string]*ast.FuncDecl {
	res := make(map[string]*ast.FuncDecl)
	for _, decl := range astF.Decls {
		f, ok := decl.(*ast.FuncDecl)
		if !ok || f.Recv == nil || len(f.Recv.List) != 1 || f.Recv.List[0].Type == nil {
			continue
		}
		typ := f.Recv.List[0].Type
		if st, isStarExpr := typ.(*ast.StarExpr); isStarExpr {
			typ = st.X
		}
		structIdent, ok := typ.(*ast.Ident)
		if !ok || structIdent.Name != s.ImplStructName {
			continue
		}
		res[f.Name.Name] = f
	}
	return res
}

func (s implsSync) getImportedName(astF *ast.File) (string, bool) {
	interfacePkgName := s.InterfacePackageName
	imported := false
	// 接口import名
	for _, ip := range astF.Imports {
		if strings.Trim(ip.Doc.Text(), `"`) == s.InterfacePackagePath {
			if ip.Name != nil {
				interfacePkgName = ip.Name.Name
			}
			imported = true
			break
		}
	}
	return interfacePkgName, imported
}

func (s implsSync) sync() (updated int, err error) {
	_ = os.MkdirAll(s.implDir, 0775)

	implStructDeclPath := ""

	if err = s.walkImplDir(func(fp string) (err error) {
		astF, _, data, err := utils.ParseFileAst(fp)
		if err != nil {
			return errors.Wrap(err, "解析文件失败")
		}
		interfacePkgName, _ := s.getImportedName(astF)
		interfaceTypeName := interfacePkgName + "." + s.InterfaceName
		match := regexp.MustCompile(fmt.Sprintf(`_ %s.+= (.+)`, interfaceTypeName)).FindStringSubmatch(string(data))
		if len(match) == 2 {
			implStructDeclPath = fp
			implStructName := match[1]
			implStructName = strings.TrimPrefix(implStructName, "new(")
			implStructName = strings.TrimFunc(implStructName, func(r rune) bool {
				switch r {
				case '{', '}', '&', '(', ')':
					return true
				}
				return false
			})
			s.ImplStructName = implStructName
		}
		s.ImplPackage = astF.Name.Name
		return
	}); err != nil {
		return updated, errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("解析文件失败%s", err))
	}

	// 未有实现结构体 创建
	if len(implStructDeclPath) == 0 {
		fp := filepath.Join(s.implDir, "init.go")
		logger.Info("implement for [ %s.%s ] not found,create in [ %s ]", s.InterfacePackageName, s.InterfaceName, fp)
		if err = utils.ExecuteTemplateAndWrite(s.implStructTemplate, s, fp); err != nil {
			return 0, errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("生成实现结构体文件失败:%s", err))
		}
	}

	if s.ifaceAstType.Methods.List == nil {
		return updated, nil
	}

	ifaceFuncMap := s.getInterfaceFuncMap()
	mu := sync.Mutex{}
	wg := new(errgroup.Group)

	if err = s.walkImplDir(func(fp string) (err error) {
		wg.Go(func() (err error) {
			update, err := s.updateFileImplements(fp, ifaceFuncMap, &mu)
			if err == nil && update > 0 {
				mu.Lock()
				updated += update
				mu.Unlock()
			}
			return nil
		})
		return
	}); err != nil {
		return updated, errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("同步接口实现失败%s", err))
	}

	if err = wg.Wait(); err != nil {
		return updated, errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("同步接口实现失败%s", err))
	}

	if len(ifaceFuncMap) == 0 {
		return updated, nil
	}

	for name, f := range ifaceFuncMap {
		wg.Go(func() (err error) {
			if err = s.appendNewFunc(name, f); err == nil {
				mu.Lock()
				updated += 1
				mu.Unlock()
			}
			return nil
		})
	}
	if err = wg.Wait(); err != nil {
		return updated, errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("同步接口实现失败%s", err))
	}
	return updated, nil
}

func (s implsSync) updateFileImplements(fp string, ifaceFuncMap map[string]ifaceFunc, mutex *sync.Mutex) (edited int, err error) {
	bf := bytes.Buffer{}
	astF, fileSet, data, err := utils.ParseFileAst(fp)
	if err != nil {
		return edited, errors.WrapWithCode(err, errors.ErrCodeParse, fmt.Sprintf("解析文件失败: %s", err))
	}
	implements := s.getImplementFunc(astF)
	if len(implements) == 0 {
		return edited, nil
	}

	for name, f := range implements {
		mutex.Lock()
		interfaceFunc, ok := ifaceFuncMap[name]
		if ok {
			delete(ifaceFuncMap, name)
		}
		mutex.Unlock()

		if !ok {
			continue
		}

		newFunc, err := utils.FormatAst(f.Type, fileSet)
		if err != nil {
			return 0, errors.WrapWithCode(err, errors.ErrCodeParse, fmt.Sprintf("格式化函数失败: %s", err))
		}

		if newFunc = strings.TrimPrefix(newFunc, "func"); newFunc == interfaceFunc.string {
			continue
		}

		logger.Info("update [ func(%s)%s%s ] => [ func%s ] in [ %s ]", s.ImplStructName, name, newFunc, interfaceFunc.string, fp)

		bf.Reset()
		bf.WriteString(string(data[:f.Type.Params.Pos()-1]))
		bf.WriteString(strings.TrimPrefix(interfaceFunc.string, "func"))
		bf.Write(data[f.Type.End()-1:])
		data = bf.Bytes()
		edited += 1
	}
	if edited == 0 {
		return edited, nil
	}
	if err = utils.ImportAndWrite(data, fp); err != nil {
		fmt.Printf("%s", data)
		return 0, errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("写入文件失败: %s", err))
	}
	return edited, nil
}

func (s implsSync) appendNewFunc(name string, f ifaceFunc) (err error) {
	bf := bytes.Buffer{}
	fp := filepath.Join(s.implDir, strcase.SnakeCase(name)+".go")

	astF, _, data, err := utils.ParseFileAst(fp)
	if err == nil {
		interfacePkgName, _ := s.getImportedName(astF)
		funcBody, funcStr, err := s.getFunc(f.FuncType, interfacePkgName, name)
		if err != nil {
			logger.Error("get func body error: %v", name)
			return errors.WrapWithCode(err, errors.ErrCodeParse, fmt.Sprintf("获取函数数失败: %s", err))
		}
		bf.Write(data)
		bf.WriteString("\n\n\n" + funcBody)
		logger.Info("sync [ %s ] in [ %s ]", funcStr, fp)
	} else {
		bf.WriteString(`package ` + s.ImplPackage)
		bf.WriteString("\n\nimport " + fmt.Sprintf(`%s "%s"`, s.InterfacePackageName, s.InterfacePackagePath))
		funcBody, funcStr, err := s.getFunc(f.FuncType, s.InterfacePackageName, name)
		if err != nil {
			logger.Error("get func body error: %v", name)
			return errors.WrapWithCode(err, errors.ErrCodeParse, fmt.Sprintf("获取函数数失败: %s", err))
		}
		bf.WriteString(funcBody)
		logger.Info("sync [ %s ] in [ %s ]", funcStr, fp)
	}
	if err = utils.ImportAndWrite(bf.Bytes(), fp); err != nil {
		logger.Error("%v", bf.Bytes())
		return errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("写入文件失败: %s", err))
	}
	return nil
}

func (s implsSync) getFunc(f *ast.FuncType, interfacePackageName, name string) (ret, funcStr string, err error) {
	str, err := utils.FormatAst(&ast.FuncType{
		Params:  copyFieldList(interfacePackageName, f.Params),
		Results: copyFieldList(interfacePackageName, f.Results),
	}, token.NewFileSet())

	if err != nil {
		return "", "", errors.WrapWithCode(err, errors.ErrCodeParse, fmt.Sprintf("格式化函数失败: %s", err))
	}
	str = strings.ReplaceAll(str, "\n", "")
	str = strings.TrimPrefix(str, "func")
	funcStr = fmt.Sprintf(`func (%s *%s) %s`, utils.GetFuncCallerIdent(s.ImplStructName), s.ImplStructName, name+str)
	ret += "\n\n" + funcStr + ` {
panic("implement me")
} `
	return ret, funcStr, nil
}

func copyFieldList(itfPkg string, src *ast.FieldList) *ast.FieldList {
	if src == nil {
		return nil
	}
	dst := new(ast.FieldList)
	*dst = *src
	dst.Opening = 0
	dst.Closing = 0
	for i := range dst.List {
		for j := range dst.List[i].Names {
			n := dst.List[i].Names[j]
			o := n.Obj
			dst.List[i].Names[j].NamePos = 0
			dst.List[i].Names[j].Obj = ast.NewObj(o.Kind, o.Name)
		}
		addPkg2type(&dst.List[i].Type, itfPkg)
	}
	return dst
}

type ifaceFunc struct {
	*ast.FuncType
	string
}

func (s implsSync) getInterfaceFuncMap() map[string]ifaceFunc {
	ifaceFuncMap := make(map[string]ifaceFunc)
	bf := new(bytes.Buffer)
	for _, m := range s.ifaceAstType.Methods.List {
		ft, ok := m.Type.(*ast.FuncType)
		if !ok {
			continue
		}
		nft := &ast.FuncType{
			Params:  copyFieldList(s.InterfacePackageName, ft.Params),
			Results: copyFieldList(s.InterfacePackageName, ft.Results),
		}
		bf.Reset()
		_ = format.Node(bf, token.NewFileSet(), nft)
		ifaceFuncMap[m.Names[0].Name] = ifaceFunc{
			FuncType: nft,
			string:   strings.TrimPrefix(bf.String(), "func"),
		}
	}
	return ifaceFuncMap
}

type Interface struct {
	PackagePath       string
	PackageName       string
	Name              string
	AnnotationContent string

	IfaceType *ast.InterfaceType
	fileSet   *token.FileSet
}

func matchInterface(scope string, ident string) (fields []Interface) {
	regexConfig, err := regexp.Compile(`@` + ident + `\((.*?)\)`)
	if err != nil {
		logger.Error("编译正则表达式失败: %v", err)
		return
	}
	mu := sync.Mutex{}
	if err = utils.ExecFiles(scope, func(path string) (err error) {
		astFile, fileSet, data, err := utils.ParseFileAst(path)
		if err != nil || !regexConfig.Match(data) {
			return
		}
		packagePath, _ := utils.GetPathModPkg(filepath.Dir(path))
		fs := matchAnnotationsByImpl(regexConfig, astFile)
		for i := range fs {
			fs[i].PackageName = astFile.Name.Name
			fs[i].PackagePath = packagePath
			fs[i].fileSet = fileSet
		}
		mu.Lock()
		fields = append(fields, fs...)
		mu.Unlock()
		return
	}); err != nil {
		return
	}
	return
}

func matchAnnotationsByImpl(re *regexp.Regexp, astF *ast.File) (iface []Interface) {
	for _, d := range astF.Decls {
		switch t := d.(type) {
		case *ast.GenDecl:
			for _, s := range t.Specs {
				switch spec := s.(type) {
				case *ast.TypeSpec:
					if t.Doc == nil || t.Doc.List == nil {
						continue
					}

					ifaceType, ok := spec.Type.(*ast.InterfaceType)
					if !ok {
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

					iface = append(iface, Interface{
						Name:              spec.Name.Name,
						AnnotationContent: match[1],
						IfaceType:         ifaceType,
					})
				}
			}
		}
	}
	return
}
