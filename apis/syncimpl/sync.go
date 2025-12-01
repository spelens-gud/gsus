package syncimpl

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"text/template"

	"github.com/spelens-gud/gsus/apis/helpers"
	"github.com/spelens-gud/gsus/basetmpl"
	"golang.org/x/sync/errgroup"

	"github.com/stoewer/go-strcase"
)

var defaultImplTemplate = template.Must(template.New("impl").Parse(basetmpl.DefaultImplTemplate))

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

func (cfg Config) SyncInterfaceImpls() (err error) {
	if len(cfg.SetName) == 0 {
		return errors.New("invalid impl set name")
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
				log.Printf("sync interface [ %s.%s ] err: %v", item.PackageName, item.Name, err)
				return
			}
			log.Printf("interface [ %s.%s ] sync finished,[ %d ] functions updated", item.PackageName, item.Name, updated)
			return
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
		astF, _, data, err := helpers.ParseFileAst(fp)
		if err != nil {
			return err
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
		return
	}

	// 未有实现结构体 创建
	if len(implStructDeclPath) == 0 {
		fp := filepath.Join(s.implDir, "init.go")
		log.Printf("implement for [ %s.%s ] not found,create in [ %s ]", s.InterfacePackageName, s.InterfaceName, fp)
		if err = helpers.ExecuteTemplateAndWrite(s.implStructTemplate, s, fp); err != nil {
			return
		}
	}

	if s.ifaceAstType.Methods.List == nil {
		return
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
			return
		})
		return
	}); err != nil {
		return
	}

	if err = wg.Wait(); err != nil {
		return
	}

	if len(ifaceFuncMap) == 0 {
		return
	}

	for name, f := range ifaceFuncMap {
		name := name
		f := f
		wg.Go(func() (err error) {
			if err = s.appendNewFunc(name, f); err == nil {
				mu.Lock()
				updated += 1
				mu.Unlock()
			}
			return
		})
	}
	err = wg.Wait()
	return
}

func (s implsSync) updateFileImplements(fp string, ifaceFuncMap map[string]ifaceFunc, mutex *sync.Mutex) (edited int, err error) {
	bf := bytes.Buffer{}
	astF, fileSet, data, err := helpers.ParseFileAst(fp)
	if err != nil {
		return
	}
	implements := s.getImplementFunc(astF)
	if len(implements) == 0 {
		return
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

		newFunc, err := helpers.FormatAst(f.Type, fileSet)
		if err != nil {
			return 0, err
		}

		if newFunc = strings.TrimPrefix(newFunc, "func"); newFunc == interfaceFunc.string {
			continue
		}

		log.Printf("update [ func(%s)%s%s ] => [ func%s ] in [ %s ]", s.ImplStructName, name, newFunc, interfaceFunc.string, fp)

		bf.Reset()
		bf.WriteString(string(data[:f.Type.Params.Pos()-1]))
		bf.WriteString(strings.TrimPrefix(interfaceFunc.string, "func"))
		bf.Write(data[f.Type.End()-1:])
		data = bf.Bytes()
		edited += 1
	}
	if edited == 0 {
		return
	}
	if err = helpers.ImportAndWrite(data, fp); err != nil {
		fmt.Printf("%s", data)
		return
	}
	return
}

func (s implsSync) appendNewFunc(name string, f ifaceFunc) (err error) {
	bf := bytes.Buffer{}
	fp := filepath.Join(s.implDir, strcase.SnakeCase(name)+".go")

	astF, _, data, err := helpers.ParseFileAst(fp)
	if err == nil {
		interfacePkgName, _ := s.getImportedName(astF)
		funcBody, funcStr, err := s.getFunc(f.FuncType, interfacePkgName, name)
		if err != nil {
			log.Printf("get func body error: %v", name)
			return err
		}
		bf.Write(data)
		bf.WriteString("\n\n\n" + funcBody)
		log.Printf("sync [ %s ] in [ %s ]", funcStr, fp)
	} else {
		bf.WriteString(`package ` + s.ImplPackage)
		bf.WriteString("\n\nimport " + fmt.Sprintf(`%s "%s"`, s.InterfacePackageName, s.InterfacePackagePath))
		funcBody, funcStr, err := s.getFunc(f.FuncType, s.InterfacePackageName, name)
		if err != nil {
			log.Printf("get func body error: %v", name)
			return err
		}
		bf.WriteString(funcBody)
		log.Printf("sync [ %s ] in [ %s ]", funcStr, fp)
	}
	if err = helpers.ImportAndWrite(bf.Bytes(), fp); err != nil {
		log.Printf("%v", bf.Bytes())
		return
	}
	return
}

func (s implsSync) getFunc(f *ast.FuncType, interfacePackageName, name string) (ret, funcStr string, err error) {
	str, err := helpers.FormatAst(&ast.FuncType{
		Params:  copyFieldList(interfacePackageName, f.Params),
		Results: copyFieldList(interfacePackageName, f.Results),
	}, token.NewFileSet())

	if err != nil {
		return
	}
	str = strings.ReplaceAll(str, "\n", "")
	str = strings.TrimPrefix(str, "func")
	funcStr = fmt.Sprintf(`func (%s *%s) %s`, helpers.GetFuncCallerIdent(s.ImplStructName), s.ImplStructName, name+str)
	ret += "\n\n" + funcStr + ` {
panic("implement me")
} `
	return
}
