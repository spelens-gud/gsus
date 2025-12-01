package runner

import (
	"fmt"
	"sync"

	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/generator"
	"github.com/spelens-gud/gsus/internal/parser"
	"github.com/spelens-gud/gsus/internal/utils"
)

func SearchServices(scope string) (services []parser.Service, err error) {
	mu := sync.Mutex{}

	err = utils.ExecFiles(scope, func(path string) (err error) {
		svcs, err := generator.GetAllService(path, config.WithIdent("service"))
		if err != nil {
			return fmt.Errorf("get service annotation error: %v", err)
		}
		mu.Lock()
		services = append(services, svcs...)
		mu.Unlock()
		return
	})
	return
}
