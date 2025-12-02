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

// SearchServices function    搜索服务定义.
// 在指定范围内扫描所有 Go 文件，查找服务接口定义.
// 返回找到的所有服务列表和可能的错误.
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
