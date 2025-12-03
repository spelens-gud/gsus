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
	// loadError 配置加载错误信息.
	loadError error
	// config 全局配置信息.
	config Option
	// once 确保配置只加载一次的同步锁.
	once sync.Once
	// configPath 配置文件相对路径.
	configPath = ".gsus/config.yaml"
)

// Get function    获取全局配置.
// 使用单例模式确保配置只加载一次，返回配置对象和可能的错误.
func Get() (Option, error) {
	once.Do(func() {
		var content []byte
		if content, loadError = Load(configPath); loadError != nil {
			loadError = errors.WrapWithCode(loadError, errors.ErrCodeConfig, "加载项目 gsus 配置失败，请运行 [gsus init] 来初始化项目 gsus 配置")
			return
		}
		if loadError = yaml.Unmarshal(content, &config); loadError != nil {
			loadError = errors.WrapWithCode(loadError, errors.ErrCodeConfig, "反序列化 gsus 配置错误")
			return
		}
	})
	return config, loadError
}

// Load function    加载指定路径的配置文件.
// 参数 relativePath 是相对于项目根目录的路径.
// 返回文件内容字节数组和可能的错误.
func Load(relativePath string) (content []byte, err error) {
	dir, err := utils.GetProjectDir()
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeConfig, "获取项目目录失败")
	}
	path := filepath.Join(dir, relativePath)
	content, err = os.ReadFile(path)
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("读取配置文件失败: %s", path))
	}
	return content, nil
}

// ExecuteWithConfig function    执行需要配置的函数.
// 自动加载配置并传递给回调函数，处理错误日志.
func ExecuteWithConfig(fn func(cfg Option) error) {
	utils.Execute(func() error {
		cfg, err := Get()
		if err != nil {
			logger.Error("加载配置失败: %v", err)
			return errors.WrapWithCode(err, errors.ErrCodeConfig, fmt.Sprintf("加载配置失败: %s", err))
		}
		return fn(cfg)
	})
}

// Validate method    验证配置的有效性.
// 检查配置项是否符合要求，返回验证错误.
func (o *Option) Validate() error {
	// 添加配置验证逻辑
	return nil
}
