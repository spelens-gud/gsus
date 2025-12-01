package syncimpl

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"strings"
)

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

func addPkg2type(typ *ast.Expr, itfPkg string) {
	if len(itfPkg) == 0 {
		return
	}

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
