package parser

import (
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	modFilepath *string
)

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
