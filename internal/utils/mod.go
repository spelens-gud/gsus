package utils

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spelens-gud/gsus/internal/errors"
)

var (
	modPkgTmp     = map[string]string{}
	modPkgTmpLock = sync.Mutex{}
)

var (
	modFilepath *string
)

// FixFilepathByProjectDir function    将相对路径转换为基于项目根目录的绝对路径.
func FixFilepathByProjectDir(fp ...*string) (err error) {
	dir, err := GetProjectDir()
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeConfig, fmt.Sprintf("获取项目目录失败: %s", err))
	}
	for _, p := range fp {
		if filepath.IsAbs(*p) {
			continue
		}
		*p = filepath.Join(dir, *p)
	}
	return
}

// GetProjectDir function    获取项目根目录.
func GetProjectDir() (path string, err error) {
	ret, err := GetModFilepath()
	if err != nil {
		return "", errors.WrapWithCode(err, errors.ErrCodeConfig, fmt.Sprintf("获取 go.mod 路径失败: %s", err))
	}
	path = filepath.Dir(ret)
	return
}

// GetModFilepath function    获取 go.mod 文件路径.
func GetModFilepath() (path string, err error) {
	if modFilepath != nil {
		return *modFilepath, nil
	}
	ret, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		return "", errors.WrapWithCode(err, errors.ErrCodeConfig, "执行 go env GOMOD 失败")
	}
	path = strings.TrimSpace(string(ret))
	modFilepath = &path
	return
}

// GetModBase function    获取模块基础路径.
func GetModBase() (path string, err error) {
	ret, err := exec.Command("go", "list", "-m").Output()
	if err != nil {
		return "", errors.WrapWithCode(err, errors.ErrCodeConfig, "执行 go list -m 失败")
	}
	path = strings.TrimSpace(string(ret))
	return
}

// GetPathModPkg function    获取路径对应的包名.
func GetPathModPkg(fp string) (pkg string, err error) {
	modPkgTmpLock.Lock()
	defer modPkgTmpLock.Unlock()
	if len(modPkgTmp[fp]) > 0 {
		return modPkgTmp[fp], nil
	}
	if err = FixFilepathByProjectDir(&fp); err != nil {
		return "", errors.WrapWithCode(err, errors.ErrCodeConfig, fmt.Sprintf("修正文件路径失败: %s", err))
	}
	ret, err := exec.Command("go", "list", fp).Output()
	if err != nil {
		return "", errors.WrapWithCode(err, errors.ErrCodeConfig, fmt.Sprintf("执行 go list 失败: %s", err))
	}
	pkg = strings.TrimSpace(string(ret))
	modPkgTmp[fp] = pkg
	return
}
