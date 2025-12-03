// Package template 提供模板加载和管理功能.
package template

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/errors"
	"github.com/spelens-gud/gsus/internal/logger"
	"github.com/spelens-gud/gsus/internal/utils"
)

// Manager struct    模板管理器.
type Manager struct {
	templates map[string]*template.Template
}

// NewManager function    创建模板管理器.
func NewManager() *Manager {
	return &Manager{
		templates: make(map[string]*template.Template),
	}
}

// InitAndLoad function    初始化并加载模板.
func InitAndLoad(templatePath string, defaultTemplate string) (*template.Template, string, error) {
	if err := initTemplate(templatePath, defaultTemplate); err != nil {
		return nil, "", err
	}
	return Load(templatePath)
}

// initTemplate function    初始化模板文件.
func initTemplate(templatePath string, defaultTemplate string) error {
	if _, err := os.Stat(templatePath); err != nil {
		if !os.IsNotExist(err) {
			return errors.WrapWithCode(err, errors.ErrCodeTemplate, "failed to check template file")
		}

		logger.Info("template not found, initializing: %s", templatePath)
		if err := os.MkdirAll(filepath.Dir(templatePath), 0755); err != nil {
			return errors.WrapWithCode(err, errors.ErrCodeFile, "failed to create template directory")
		}

		if err := os.WriteFile(templatePath, []byte(defaultTemplate), 0644); err != nil {
			return errors.WrapWithCode(err, errors.ErrCodeFile, "failed to write template file")
		}
	}
	return nil
}

// Load function    加载模板文件.
func Load(templatePath string) (*template.Template, string, error) {
	// 添加模板后缀
	if !strings.HasSuffix(templatePath, config.GsusTemplateSuffix) {
		templatePath += config.GsusTemplateSuffix
	}

	// 处理相对路径
	if !filepath.IsAbs(templatePath) {
		templatePath = filepath.Join(config.GsusTemplateDir, templatePath)
		if err := utils.FixFilepathByProjectDir(&templatePath); err != nil {
			return nil, "", errors.WrapWithCode(err, errors.ErrCodeTemplate, "failed to resolve template path")
		}
	}

	// 读取模板文件
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, "", errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("failed to read template: %s", templatePath))
	}

	// 计算模板哈希
	tmplHash := fmt.Sprintf("%x", md5.Sum(data))

	// 解析模板
	tmpl, err := template.New(filepath.Base(templatePath)).Parse(string(data))
	if err != nil {
		return nil, "", errors.WrapWithCode(err, errors.ErrCodeTemplate, "failed to parse template")
	}

	return tmpl, tmplHash, nil
}

// LoadByName method    根据名称加载模板.
func (m *Manager) LoadByName(name, path string) error {
	tmpl, _, err := Load(path)
	if err != nil {
		return err
	}
	m.templates[name] = tmpl
	return nil
}

// Get method    获取模板.
func (m *Manager) Get(name string) (*template.Template, bool) {
	tmpl, ok := m.templates[name]
	return tmpl, ok
}

// Render method    渲染模板.
func (m *Manager) Render(name string, data interface{}) (string, error) {
	tmpl, ok := m.Get(name)
	if !ok {
		return "", errors.New(errors.ErrCodeTemplate, fmt.Sprintf("template not found: %s", name))
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", errors.WrapWithCode(err, errors.ErrCodeTemplate, "failed to render template")
	}

	return buf.String(), nil
}
