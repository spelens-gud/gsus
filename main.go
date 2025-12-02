// Package main 是 gsus 工具的程序入口.
// gsus 是一个强大的 Go 代码生成工具，支持从数据库生成结构体、生成 HTTP 客户端/路由代码、生成接口实现等功能.
package main

import "github.com/spelens-gud/gsus/cmd"

// main function    程序入口函数.
func main() {
	cmd.Execute()
}
