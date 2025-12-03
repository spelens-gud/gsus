package runner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/errors"
	"github.com/spelens-gud/gsus/internal/logger"
	"github.com/spelens-gud/gsus/internal/template"
	"github.com/spelens-gud/gsus/internal/utils"
	"gopkg.in/yaml.v3"
)

// InitOptions struct    初始化选项.
type InitOptions struct {
	// 预留扩展字段
}

// RunAutoInit function    执行项目初始化.
func RunAutoInit(opts *InitOptions) {
	log := logger.WithPrefix("[init]")
	log.Info("开始执行 init 代码生成")

	utils.Execute(func() (err error) {
		// 检查配置是否已存在
		if _, err := config.Get(); err == nil {
			log.Error("gsus 已在项目中初始化")
			return errors.New(errors.ErrCodeConfig, "gsus has already init in project,run [ gsus update ] to update templates")
		}

		// 获取项目目录
		dir, err := utils.GetProjectDir()
		if err != nil {
			log.Error("获取项目目录失败")
			return errors.WrapWithCode(err, errors.ErrCodeConfig, fmt.Sprintf("获取项目目录失败: %s", err))
		}

		// 创建 .gsus 目录
		if err = os.Mkdir(filepath.Join(dir, config.GsusConfigDir), 0775); err != nil {
			log.Error("创建 .gsus 目录失败")
			return errors.WrapWithCode(err, errors.ErrCodeConfig, fmt.Sprintf("创建 .gsus 目录失败: %s", err))
		}

		// 解析默认配置
		var tmp config.Option
		if err = yaml.Unmarshal([]byte(template.DefaultConfigYaml), &tmp); err != nil {
			log.Error("解析默认配置失败")
			return errors.WrapWithCode(err, errors.ErrCodeConfig, fmt.Sprintf("解析默认配置失败: %s", err))
		}
		bytes, err := yaml.Marshal(&tmp)
		if err != nil {
			log.Error("写入默认配置文件失败")
			return errors.WrapWithCode(err, errors.ErrCodeConfig, fmt.Sprintf("写入默认配置文件失败: %s", err))
		}

		// 写入默认配置文件
		if err = os.WriteFile(filepath.Join(dir, config.GsusConfigFile), bytes, 0664); err != nil {
			log.Error("写入默认配置文件失败")
			return errors.WrapWithCode(err, errors.ErrCodeConfig, fmt.Sprintf("写入默认配置文件失败: %s", err))
		}

		// 创建模板目录
		if err = os.Mkdir(filepath.Join(dir, config.GsusTemplateDir), 0775); err != nil {
			log.Error("创建模板目录失败")
			return errors.WrapWithCode(err, errors.ErrCodeConfig, fmt.Sprintf("创建模板目录失败: %s", err))
		}

		log.Info("开始写入模板文件")
		// 写入模板文件
		return writeTemplates(dir, getTemplateMap())
	})
}

func writeTemplates(dir string, contentMap map[string]string) error {
	for path, content := range contentMap {
		p := filepath.Join(dir, config.GsusTemplateDir, path+config.GsusTemplateSuffix)
		if err := os.WriteFile(p, []byte(content), 0664); err != nil {
			return errors.WrapWithCode(err, errors.ErrCodeTemplate, fmt.Sprintf("写入模板文件失败: %s", err))
		}
	}
	return nil
}

func getTemplateMap() map[string]string {
	return map[string]string{
		"impl":             template.DefaultImplTemplate,
		"http_router":      template.DefaultHttpRouterTemplate,
		"http_client_api":  template.DefaultHttpClientApiTemplate,
		"http_client_base": template.DefaultHttpClientBaseTemplate,
		"dao":              template.DefaultDaoTemplate,
		"dao_impl":         template.DefaultDaoImplTemplate,
		"service":          template.DefaultServiceTemplate,
		"service_impl":     template.DefaultServiceImplTemplate,
		"model_cast":       template.DefaultModelCastTemplate,
		"model_generic":    template.DefaultModelGenericTemplate,
	}
}
