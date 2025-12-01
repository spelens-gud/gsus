package utils

import (
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var (
	modPkgTmp     = map[string]string{}
	modPkgTmpLock = sync.Mutex{}
)

var (
	modFilepath *string
)

// FixFilepathByProjectDir 将相对路径转换为基于项目根目录的绝对路径
// 参数 fp 是一个或多个字符串指针，指向需要修正的路径
// 如果路径已经是绝对路径，则跳过处理；否则将其与项目根目录拼接
// 返回值 err 表示处理过程中可能出现的错误.
func FixFilepathByProjectDir(fp ...*string) (err error) {
	dir, err := GetProjectDir()
	if err != nil {
		return
	}
	for _, p := range fp {
		if filepath.IsAbs(*p) {
			continue
		}
		*p = filepath.Join(dir, *p)
	}
	return
}

func GetProjectDir() (path string, err error) {
	ret, err := GetModFilepath()
	if err != nil {
		return
	}
	path = filepath.Dir(ret)
	return
}

func GetModFilepath() (path string, err error) {
	if modFilepath != nil {
		return *modFilepath, nil
	}
	ret, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		return
	}
	path = strings.TrimSpace(string(ret))
	modFilepath = &path
	return
}
func GetModBase() (path string, err error) {
	ret, err := exec.Command("go", "list", "-m").Output()
	if err != nil {
		return
	}
	path = strings.TrimSpace(string(ret))
	return
}
func GetPathModPkg(fp string) (pkg string, err error) {
	modPkgTmpLock.Lock()
	defer modPkgTmpLock.Unlock()
	if len(modPkgTmp[fp]) > 0 {
		return modPkgTmp[fp], nil
	}
	if err = FixFilepathByProjectDir(&fp); err != nil {
		return
	}
	ret, err := exec.Command("go", "list", fp).Output()
	if err != nil {
		return
	}
	pkg = strings.TrimSpace(string(ret))
	modPkgTmp[fp] = pkg
	return
}
