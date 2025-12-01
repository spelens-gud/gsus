package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spelens-gud/gsus/internal/errors"
	"github.com/spelens-gud/gsus/internal/logger"
	"golang.org/x/sync/errgroup"
)

const goGenerateEnv = "GOFILE"

type IOption interface {
	Get() (IOption, error)
}

// Execute function    执行函数并处理错误.
func Execute(f func() (err error)) {
	var err error

	defer func() {
		if e := recover(); e != nil {
			panic(e)
		}

		if err != nil {
			var e *exec.ExitError
			switch {
			case errors.As(err, e):
				logger.Fatal("%v: %s", err, e.Stderr)
			default:
				logger.Fatal("%+v", err)
			}
		}
	}()

	err = f()
}

// ExecFiles function    遍历文件并执行函数.
func ExecFiles(scope string, f func(path string) (err error)) (err error) {
	if gofile := os.Getenv(goGenerateEnv); len(gofile) > 0 {
		gofile, err = filepath.Abs(gofile)
		if err != nil {
			return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("获取文件绝对路径失败: %s", err))
		}
		return f(gofile)
	}

	if err = FixFilepathByProjectDir(&scope); err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("修正文件路径失败: %s", err))
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
