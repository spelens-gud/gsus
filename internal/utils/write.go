package utils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/Just-maple/xtoolinternal/gocommand"
	"github.com/Just-maple/xtoolinternal/imports"
	"github.com/spelens-gud/gsus/internal/errors"
)

var importMu sync.Mutex

var localPrefix = func() string {
	path, _ := GetModBase()
	return path
}()

var opt2 = &imports.Options{Comments: true, TabIndent: true, TabWidth: 8}
var opt = &imports.Options{
	Env: &imports.ProcessEnv{
		GocmdRunner: &gocommand.Runner{},
	},
	LocalPrefix: localPrefix,
	AllErrors:   opt2.AllErrors,
	Comments:    opt2.Comments,
	FormatOnly:  opt2.FormatOnly,
	Fragment:    opt2.Fragment,
	TabIndent:   opt2.TabIndent,
	TabWidth:    opt2.TabWidth,
}

// ImportAndWrite function    处理导入并写入文件.
func ImportAndWrite(b []byte, path string) (err error) {
	var data []byte
	_ = os.MkdirAll(filepath.Dir(path), 0775)
	data, err = ImportProcess(b)
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("处理导入失败: %s", err))
	}
	if err = os.WriteFile(path, data, 0664); err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("写入文件失败: %s", err))
	}
	return
}

// ImportProcess function    处理 Go 文件的导入.
func ImportProcess(bytes []byte) (ret []byte, err error) {
	importMu.Lock()
	defer importMu.Unlock()
	ret, err = imports.Process("", bytes, opt)
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("处理导入失败: %s", err))
	}
	return
}

// ExecuteTemplateAndWrite function    执行模板并写入文件.
func ExecuteTemplateAndWrite(temp *template.Template, iface interface{}, path string) (err error) {
	data, err := ExecuteTemplate(temp, iface)
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeTemplate, fmt.Sprintf("执行模板失败: %s", err))
	}
	if len(path) == 0 {
		fmt.Printf("\n%v\n", data)
	}
	return ImportAndWrite(data, path)
}

// ExecuteTemplate function    执行模板.
func ExecuteTemplate(temp *template.Template, iface interface{}) (data []byte, err error) {
	var bf bytes.Buffer
	if err = temp.Execute(&bf, iface); err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeTemplate, fmt.Sprintf("执行模板失败: %s", err))
	}
	data = bf.Bytes()
	return
}
