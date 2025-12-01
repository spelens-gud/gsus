package syncimpl

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/spelens-gud/gsus/apis/helpers"
	"github.com/spelens-gud/gsus/apis/helpers/executor"
)

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
		return
	}
	mu := sync.Mutex{}
	if err = executor.ExecFiles(scope, func(path string) (err error) {
		astFile, fileSet, data, err := helpers.ParseFileAst(path)
		if err != nil || !regexConfig.Match(data) {
			return
		}
		packagePath, _ := helpers.GetPathModPkg(filepath.Dir(path))
		fs := matchAnnotations(regexConfig, astFile)
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

func matchAnnotations(re *regexp.Regexp, astF *ast.File) (iface []Interface) {
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
