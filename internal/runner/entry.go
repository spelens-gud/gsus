package runner

import (
	"errors"
	"log"
	"os/exec"

	"github.com/spelens-gud/gsus/internal/config"
)

func Execute(f func() (err error)) {
	var err error

	defer func() {
		if e := recover(); e != nil {
			panic(e)
		}

		if err != nil {
			var e *exec.ExitError
			switch {
			case errors.As(err, &e):
				log.Fatalf("%v: %s", err, e.Stderr)
			default:
				log.Fatalf("%+v", err)
			}
		}
	}()

	err = f()
}

func ExecuteWithConfig(fn func(cfg config.Option) (err error)) {
	Execute(func() (err error) {
		cfg, _ := config.Get()
		return fn(cfg)
	})
}
