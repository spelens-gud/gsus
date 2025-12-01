package httpgen

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	apiAnnotateRegex             = regexp.MustCompile(`@(!?[A-Za-z0-9_.:]+?)\((.+?)\)`)
	serviceAnnotateRegexTemplate = `@%s\((.+?)\)`

	logger = log.New(os.Stdout, "[SVC] ", 0)
)

type (
	ApiAnnotateItem struct {
		Handler string
		Params  []string
		Returns []string
		Title   string
		Method  string
		Args    []string
		Options map[string]string
		Doc     []string
	}

	Service struct {
		InterfaceName string
		ServiceName   string
		Pkg           string
		OtherOptions  map[string]string
		ApiAnnotates  map[string]*ApiAnnotate
	}

	ApiAnnotate struct {
		Interface string
		Apis      []ApiAnnotateItem
	}

	AnnotateParser struct {
		m           map[string]*ApiAnnotate
		namespace   string
		serviceName string
		fileData    []byte
		file        *ast.File
	}

	Option func(*opt)

	opt struct {
		Ident string
	}

	Options []Option
)

func WithIdent(ident string) Option {
	return func(o *opt) {
		o.Ident = ident
	}
}

func (os Options) Apply() *opt {
	op := &opt{
		Ident: "service",
	}
	for _, o := range os {
		o(op)
	}
	return op
}

func (r Service) GetPath(apiDir string) (groupRoute, dir, pkg string) {
	group := r.OtherOptions["group"]
	if len(group) == 0 {
		group = r.ServiceName
	}
	group = strings.Trim(group, `"`)
	// 组路由前缀
	groupRoute = r.OtherOptions["route"]
	if len(groupRoute) == 0 {
		groupRoute = fmt.Sprintf(`"%s"`, group)
	}
	// 创建组目录
	dir = filepath.Join(apiDir, group)
	// package名
	pkg = filepath.Base(dir)
	return
}

func GetAllService(file string, opts ...Option) (res []Service, err error) {
	fileData, err := os.ReadFile(file)
	if err != nil {
		return
	}
	f, err := parser.ParseFile(token.NewFileSet(), "", fileData, parser.ParseComments)
	if err != nil {
		return
	}

	o := Options(opts).Apply()
	serviceAnnotateRegex := regexp.MustCompile(fmt.Sprintf(serviceAnnotateRegexTemplate, o.Ident))

	for i := range f.Decls {
		decl, ok := f.Decls[i].(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spt := range decl.Specs {
			sp, ok := spt.(*ast.TypeSpec)
			if !ok {
				continue
			}
			_, ok = sp.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}
			var doc *ast.CommentGroup
			if len(decl.Specs) == 1 {
				doc = decl.Doc
			} else {
				doc = sp.Doc
			}
			if doc == nil {
				continue
			}
			svcNameMap := map[string]struct{}{}
			for _, cm := range doc.List {
				match := serviceAnnotateRegex.FindStringSubmatch(cm.Text)
				if len(match) != 2 {
					continue
				}
				annotate := strings.Split(match[1], ",")
				serviceName := strings.TrimSpace(annotate[0])
				if _, dup := svcNameMap[serviceName]; dup {
					err = errors.New("duplicate service name")
					return nil, err
				}
				svcNameMap[serviceName] = struct{}{}
				apis, err := AnalysisServiceWithFileToken(fileData, sp.Name.String(), serviceName)
				if err != nil {
					return nil, err
				}
				svc := Service{
					InterfaceName: sp.Name.String(),
					ServiceName:   serviceName,
					ApiAnnotates:  apis,
					Pkg:           f.Name.Name,
				}
				if len(annotate) > 1 {
					_, svc.OtherOptions, err = parseKV(strings.Join(annotate[1:], ","))
					if err != nil {
						return nil, err
					}
				}
				res = append(res, svc)
			}
		}
	}
	return
}

func AnalysisServiceWithFileToken(fileData []byte, serviceName, namespace string) (apiAnnotate map[string]*ApiAnnotate, err error) {
	f, err := parser.ParseFile(token.NewFileSet(), "", fileData, parser.ParseComments)
	if err != nil {
		return
	}
	aParser := AnnotateParser{
		m:           make(map[string]*ApiAnnotate),
		namespace:   namespace,
		serviceName: serviceName,
		fileData:    fileData,
		file:        f,
	}
	for i := range f.Decls {
		decl, ok := f.Decls[i].(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spt := range decl.Specs {
			sp, ok := spt.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if sp.Name.Name != serviceName {
				continue
			}
			itf, ok := sp.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}
			for _, method := range itf.Methods.List {
				err = aParser.parseMethod(method)
				if err != nil {
					return
				}
			}
		}
	}
	apiAnnotate = aParser.m
	logger.Printf("analysis finished: %s", serviceName)
	return
}

func (p AnnotateParser) parseMethod(method *ast.Field) (err error) {
	// check doc
	if method.Doc == nil {
		return
	}
	var params, results []string
	if ft, ok := method.Type.(*ast.FuncType); ok {
		// get param results
		if ft.Params != nil {
			collectList(p.fileData, &params, ft.Params.List, p.file)
		}
		if ft.Results != nil {
			collectList(p.fileData, &results, ft.Results.List, p.file)
		}
	} else {
		return
	}
	var (
		title     string
		docOffset int
	)
	for i, cm := range method.Doc.List {
		// match annotation like
		// @namespace:type.method(opt1=xxx,opt2=xxx,opt3=xxx)
		match := apiAnnotateRegex.FindStringSubmatch(cm.Text)
		if len(match) != 3 {
			if i == 0 {
				text := strings.TrimPrefix(cm.Text, "//")
				title = strings.TrimSpace(text)
				docOffset = 1
			} else if len(strings.TrimSpace(strings.TrimPrefix(cm.Text, "//"))) == 0 {
				// split empty doc
				docOffset = i + 1
			}
			continue
		}
		// new api item
		newApi := ApiAnnotateItem{
			Options: make(map[string]string),
			Handler: method.Names[0].Name,
			Title:   title,
			Params:  params,
			Returns: results,
		}

		// item doc
		if docOffset < i {
			for _, d := range method.Doc.List[docOffset:i] {
				newApi.Doc = append(newApi.Doc, d.Text)
			}
		}

		// parse options
		newApi.Args, newApi.Options, err = parseKV(match[2])
		if err != nil {
			return err
		}
		apiType := match[1]
		// check namespace
		if strings.ContainsRune(apiType, ':') {
			sp := strings.Split(apiType, ":")
			if (strings.HasPrefix(sp[0], "!") && sp[0] == p.namespace) ||
				(!strings.HasPrefix(sp[0], "!") && sp[0] != p.namespace) {
				docOffset = i + 1
				continue
			}
			apiType = sp[1]
		}
		if strings.ContainsRune(apiType, '.') {
			sp := strings.SplitN(apiType, ".", 2)
			apiType = sp[0]
			newApi.Method = sp[1]
		}
		// split method
		if p.m[apiType] == nil {
			p.m[apiType] = &ApiAnnotate{
				Interface: p.file.Name.String() + "." + p.serviceName,
			}
		}
		p.m[apiType].Apis = append(p.m[apiType].Apis, newApi)

		docOffset = i + 1
	}
	return
}

func parseKV(raw string) (args []string, resMap map[string]string, err error) {
	options := strings.Split(raw, ",")
	if len(options) == 0 {
		return
	}
	resMap = make(map[string]string)
	for _, o := range options {
		res := strings.Split(o, "=")
		if len(res) == 1 {
			args = append(args, o)
			continue
		}
		if len(res) != 2 {
			err = errors.New("invalid options key")
			return

		}
		if len(res[0]) == 0 {
			err = errors.New("invalid options key")
			return
		}
		if _, dup := resMap[res[0]]; dup {
			err = errors.New("duplicate options key")
			return
		}
		resMap[res[0]] = res[1]
	}
	return
}

func collectList(fileData []byte, collectList *[]string, fl []*ast.Field, f *ast.File) {
	for _, l := range fl {
		addPkg2type(&l.Type, f.Name.String())
		var bf bytes.Buffer
		_ = format.Node(&bf, token.NewFileSet(), l.Type)
		*collectList = append(*collectList, bf.String())
	}
}

func addPkg2type(typ *ast.Expr, itfPkg string) {
	switch s := (*typ).(type) {
	case *ast.StarExpr:
		s.Star = 0
		addPkg2type(&s.X, itfPkg)
	case *ast.ArrayType:
		addPkg2type(&s.Elt, itfPkg)
	case *ast.MapType:
		addPkg2type(&s.Key, itfPkg)
		addPkg2type(&s.Value, itfPkg)
	case *ast.SelectorExpr:
		if s.Sel.Obj != nil {
			s.Sel.Obj = ast.NewObj(s.Sel.Obj.Kind, s.Sel.Obj.Name)
		}
		s.Sel.NamePos = 0
		if s.X != nil {
			switch xt := s.X.(type) {
			case *ast.Ident:
				s.X = ast.NewIdent(xt.Name)
			}
		}
	case *ast.Ident:
		if s.Obj == nil {
			if s.IsExported() {
				s.NamePos = 0
				*typ = &ast.SelectorExpr{
					X:   ast.NewIdent(itfPkg),
					Sel: s,
				}
			} else {
				s.NamePos = 0
			}
		} else if s.Obj.Kind == ast.Typ {
			s.NamePos = 0
			*typ = &ast.SelectorExpr{
				X:   ast.NewIdent(itfPkg),
				Sel: s,
			}
		}
	}
}
