package runner

import (
	"fmt"
	"sync"

	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/generator"
	"github.com/spelens-gud/gsus/internal/logger"
	"github.com/spelens-gud/gsus/internal/parser"
	"github.com/spelens-gud/gsus/internal/utils"
)

func SearchServices(scope string) (services []parser.Service, err error) {
	mu := sync.Mutex{}

	err = utils.ExecFiles(scope, func(path string) (err error) {
		log := logger.WithPrefix("[http]")
		log.Info("开始执行 http 代码生成")

		svcs, err := generator.GetAllService(path, config.WithIdent("service"))
		if err != nil {
			log.Error("获取服务注解错误")
			return fmt.Errorf("获取服务注解错误: %v", err)
		}
		mu.Lock()
		services = append(services, svcs...)
		mu.Unlock()
		return
	})
	return
}
