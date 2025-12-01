package helpers

import (
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spelens-gud/gsus/apis/constant"
)

func LoadConfig(relativePath string) (content []byte, err error) {
	dir, err := GetProjectDir()
	if err != nil {
		return
	}
	path := filepath.Join(dir, relativePath)
	content, err = os.ReadFile(path)
	return
}

func LoadTemplate(templatePath string) (tmpl *template.Template, tmplHash string, err error) {
	if !strings.HasSuffix(templatePath, constant.TemplateSuffix) {
		templatePath += constant.TemplateSuffix
	}

	if !filepath.IsAbs(templatePath) {
		templatePath = filepath.Join(constant.TemplateDir, templatePath)
		if err = FixFilepathByProjectDir(&templatePath); err != nil {
			return
		}
	}
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return
	}
	tmplHash = fmt.Sprintf("%x", md5.Sum(data))
	tmpl, err = template.New(filepath.Base(templatePath)).Parse(string(data))
	return
}

func InitTemplateAndLoad(templatePath string, template string) (tmpl *template.Template, tmplHash string, err error) {
	if err = initTemplate(templatePath, template); err != nil {
		return
	}
	return LoadTemplate(templatePath)
}

func initTemplate(templatePath string, template string) (err error) {
	if _, e := os.Stat(templatePath); e != nil {
		log.Printf("template not found ,init [ %s ]", templatePath)
		_ = os.MkdirAll(filepath.Dir(templatePath), 0775)
		if err = os.WriteFile(templatePath, []byte(template), 0664); err != nil {
			return fmt.Errorf("init template error: %v", err)
		}
	}
	return nil
}
