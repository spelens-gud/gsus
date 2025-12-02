package config

// Mount struct    挂载配置.
// 用于配置代码挂载相关的参数.
type Mount struct {
	Scope string   `yaml:"scope"` // 扫描范围
	Name  string   `yaml:"name"`  // 挂载名称
	Args  []string `yaml:"args"`  // 挂载参数列表
}
