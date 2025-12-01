package helpers

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

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
