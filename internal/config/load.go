// Package config 提供配置文件加载和管理功能.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/spelens-gud/gsus/internal/errors"
	"github.com/spelens-gud/gsus/internal/logger"
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
		if content, loadError = Load(configPath); loadError != nil {
			loadError = errors.WrapWithCode(
				loadError,
				errors.ErrCodeConfig,
				"load project gsus config failed, run [ gsus init ] to init project gsus config",
			)
			return
		}
		if loadError = yaml.Unmarshal(content, &config); loadError != nil {
			loadError = errors.WrapWithCode(loadError, errors.ErrCodeConfig, "unmarshal gsus config error")
			return
		}
	})
	return config, loadError
}

func Load(relativePath string) (content []byte, err error) {
	dir, err := utils.GetProjectDir()
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeConfig, "failed to get project directory")
	}
	path := filepath.Join(dir, relativePath)
	content, err = os.ReadFile(path)
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("failed to read config file: %s", path))
	}
	return content, nil
}

func ExecuteWithConfig(fn func(cfg Option) error) {
	utils.Execute(func() error {
		cfg, err := Get()
		if err != nil {
			logger.Error("failed to load config: %v", err)
			return err
		}
		return fn(cfg)
	})
}

func (o *Option) Validate() error {
	// 添加配置验证逻辑
	return nil
}
