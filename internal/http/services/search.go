package services

import (
	"fmt"
	"sync"

	"github.com/spelens-gud/gsus/apis/helpers/executor"
	"github.com/spelens-gud/gsus/apis/httpgen"
)

func SearchServices(scope string) (services []httpgen.Service, err error) {
	mu := sync.Mutex{}

	err = executor.ExecFiles(scope, func(path string) (err error) {
		svcs, err := httpgen.GetAllService(path, httpgen.WithIdent("service"))
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
