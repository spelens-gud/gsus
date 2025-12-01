package config

import (
	"fmt"
	"sync"

	"github.com/spelens-gud/gsus/apis/helpers"
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
		if content, loadError = helpers.LoadConfig(configPath); loadError != nil {
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
