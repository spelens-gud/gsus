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

func ImportAndWrite(b []byte, path string) (err error) {
	var data []byte
	_ = os.MkdirAll(filepath.Dir(path), 0775)
	data, err = ImportProcess(b)
	if err != nil {
		return err
	}
	if err = os.WriteFile(path, data, 0664); err != nil {
		return err
	}
	return
}
func ImportProcess(bytes []byte) (ret []byte, err error) {
	importMu.Lock()
	defer importMu.Unlock()
	return imports.Process("", bytes, opt)
}
func ExecuteTemplateAndWrite(temp *template.Template, iface interface{}, path string) (err error) {
	data, err := ExecuteTemplate(temp, iface)
	if err != nil {
		return
	}
	if len(path) == 0 {
		fmt.Printf("\n%v\n", data)
	}
	return ImportAndWrite(data, path)
}

func ExecuteTemplate(temp *template.Template, iface interface{}) (data []byte, err error) {
	var bf bytes.Buffer
	if err = temp.Execute(&bf, iface); err != nil {
		return
	}
	data = bf.Bytes()
	return
}
