// Package version 提供版本信息管理功能.
package version

import (
	"fmt"
	"runtime/debug"
)

var (
	// Version 版本号（构建时注入）.
	Version = "devel"
	// GitCommit Git 提交哈希（构建时注入）.
	GitCommit = "unknown"
	// BuildTime 构建时间（构建时注入）.
	BuildTime = "unknown"
)

// Info struct    版本信息.
type Info struct {
	Version   string
	GitCommit string
	BuildTime string
}

// init function    初始化版本信息.
// 当使用 go install 安装时，会从构建信息中读取版本号.
func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	mainVersion := info.Main.Version
	if mainVersion != "" && mainVersion != "(devel)" {
		Version = mainVersion
	}
}

// Get function    获取版本信息.
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildTime: BuildTime,
	}
}

// String method    格式化版本信息.
func (i Info) String() string {
	return fmt.Sprintf("Version: %s\nGit Commit: %s\nBuild Time: %s",
		i.Version, i.GitCommit, i.BuildTime)
}

// Short method    获取简短版本信息.
func (i Info) Short() string {
	return i.Version
}
