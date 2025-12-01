package utils

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

func ParseFileAst(path string) (astF *ast.File, fileSet *token.FileSet, data []byte, err error) {
	data, err = os.ReadFile(path)
	if err != nil {
		return
	}
	fileSet = token.NewFileSet()
	astF, err = parser.ParseFile(fileSet, "", data, parser.ParseComments)
	if err != nil {
		return
	}
	return
}

func GetFuncCallerIdent(structName string) string {
	return string(strings.ToLower(structName)[0])
}

func FormatAst(in interface{}, fileSet *token.FileSet) (ret string, err error) {
	bf := new(bytes.Buffer)
	if err = format.Node(bf, fileSet, in); err != nil {
		return
	}
	ret = bf.String()
	return
}
