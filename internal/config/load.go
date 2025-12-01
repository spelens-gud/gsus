package config

import (
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/spelens-gud/gsus/internal/utils"
	"gopkg.in/yaml.v3"
)

var (
	loadError  error                 // 错误信息
	config     Option                // 配置信息
	once       sync.Once             // 锁
	configPath = ".gsus/config.yaml" // 配置文件路径
)

func Get() (Option, error) {
	once.Do(func() {
		var content []byte
		if content, loadError = LoadConfig(configPath); loadError != nil {
			loadError = fmt.Errorf("load project gsus config failed,run [ gsus init ] to init project gsus config.\nerror: %w",
				loadError)
			return
		}
		if loadError = yaml.Unmarshal(content, &config); loadError != nil {
			loadError = fmt.Errorf("unmarshal gsus config error: %w", loadError)
			return
		}
	})
	return config, loadError
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
func LoadTemplate(templatePath string) (tmpl *template.Template, tmplHash string, err error) {
	if !strings.HasSuffix(templatePath, TemplateSuffix) {
		templatePath += TemplateSuffix
	}

	if !filepath.IsAbs(templatePath) {
		templatePath = filepath.Join(TemplateDir, templatePath)
		if err = utils.FixFilepathByProjectDir(&templatePath); err != nil {
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
func LoadConfig(relativePath string) (content []byte, err error) {
	dir, err := utils.GetProjectDir()
	if err != nil {
		return
	}
	path := filepath.Join(dir, relativePath)
	content, err = os.ReadFile(path)
	return
}

func ExecuteWithConfig(fn func(cfg Option) (err error)) {
	utils.Execute(func() (err error) {
		cfg, _ := Get()
		return fn(cfg)
	})
}
