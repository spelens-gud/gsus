package executor

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spelens-gud/gsus/apis/helpers"
	"github.com/spelens-gud/gsus/internal/fileconfig"
	"golang.org/x/sync/errgroup"
)

func Execute(f func() (err error)) {
	var err error
	defer func() {
		if e := recover(); e != nil {
			panic(e)
		}
		if err != nil {
			switch e := err.(type) {
			case *exec.ExitError:
				log.Fatalf("%v: %s", err, e.Stderr)
			default:
				log.Fatalf("%+v", err)
			}
		}
	}()
	err = f()
}

func ExecuteWithConfig(f func(cfg fileconfig.Config) (err error)) {
	Execute(func() (err error) {
		cfg, _ := fileconfig.Get()
		return f(cfg)
	})
}

const goGenerateEnv = "GOFILE"

func ExecFiles(scope string, f func(path string) (err error)) (err error) {
	if gofile := os.Getenv(goGenerateEnv); len(gofile) > 0 {
		gofile, err = filepath.Abs(gofile)
		if err != nil {
			return
		}
		return f(gofile)
	}

	if err = helpers.FixFilepathByProjectDir(&scope); err != nil {
		return
	}

	// walk files
	wg := new(errgroup.Group)
	_ = filepath.Walk(scope, func(path string, info os.FileInfo, _ error) (_ error) {
		if info.IsDir() || strings.HasSuffix(path, "_test.go") || !strings.HasSuffix(path, ".go") {
			return
		}
		wg.Go(func() error {
			return f(path)
		})
		return
	})
	return wg.Wait()
}
