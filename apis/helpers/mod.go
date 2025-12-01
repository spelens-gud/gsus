package helpers

import (
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var (
	modFilepath *string
)

func GetModBase() (path string, err error) {
	ret, err := exec.Command("go", "list", "-m").Output()
	if err != nil {
		return
	}
	path = strings.TrimSpace(string(ret))
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

var (
	modPkgTmp     = map[string]string{}
	modPkgTmpLock = sync.Mutex{}
)

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
