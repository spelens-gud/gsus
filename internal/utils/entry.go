package utils

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"
)

const goGenerateEnv = "GOFILE"

type IOption interface {
	Get() (IOption, error)
}

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

func ExecFiles(scope string, f func(path string) (err error)) (err error) {
	if gofile := os.Getenv(goGenerateEnv); len(gofile) > 0 {
		gofile, err = filepath.Abs(gofile)
		if err != nil {
			return
		}
		return f(gofile)
	}

	if err = FixFilepathByProjectDir(&scope); err != nil {
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
