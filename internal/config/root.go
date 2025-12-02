package config

// Option struct    全局配置选项.
// 包含 gsus 工具的所有配置项，从 YAML 配置文件中加载.
type Option struct {
	Gsus      Gsus      `yaml:"gsus"`      // gsus 基础配置
	Db2struct Db2struct `yaml:"db2struct"` // 数据库转结构体配置
	Enum      Enum      `yaml:"enum"`      // 枚举生成配置
	Templates Templates `yaml:"templates"` // 模板配置
}

// Gsus struct    gsus 基础配置.
type Gsus struct {
	Origin string // 源代码仓库地址
}
