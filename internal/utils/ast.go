package utils

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/spelens-gud/gsus/internal/errors"
)

// ParseFileAst function    解析 Go 文件为 AST.
func ParseFileAst(path string) (astF *ast.File, fileSet *token.FileSet, data []byte, err error) {
	data, err = os.ReadFile(path)
	if err != nil {
		return nil, nil, nil, errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("读取文件失败: %s", err))
	}
	fileSet = token.NewFileSet()
	astF, err = parser.ParseFile(fileSet, "", data, parser.ParseComments)
	if err != nil {
		return nil, nil, nil, errors.WrapWithCode(err, errors.ErrCodeParse, fmt.Sprintf("解析文件失败: %s", err))
	}
	return
}

// GetFuncCallerIdent function    获取函数调用者标识符.
func GetFuncCallerIdent(structName string) string {
	return string(strings.ToLower(structName)[0])
}

// FormatAst function    格式化 AST 节点为字符串.
func FormatAst(in interface{}, fileSet *token.FileSet) (ret string, err error) {
	bf := new(bytes.Buffer)
	if err = format.Node(bf, fileSet, in); err != nil {
		return "", errors.WrapWithCode(err, errors.ErrCodeParse, fmt.Sprintf("格式化 AST 失败: %s", err))
	}
	ret = bf.String()
	return
}
