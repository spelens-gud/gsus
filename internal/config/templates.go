package config

// Templates struct    模板配置集合.
// 包含模型路径和多个模板配置.
type Templates struct {
	ModelPath string     `yaml:"modelPath"` // 模型文件路径
	Templates []Template `yaml:"templates"` // 模板配置列表
}

// Template struct    单个模板配置.
// 定义模板的名称、路径和生成选项.
type Template struct {
	Name      string `yaml:"name"`      // 模板名称
	Path      string `yaml:"path"`      // 生成代码的输出路径
	Template  string `yaml:"template"`  // 模板文件路径
	Overwrite bool   `yaml:"overwrite"` // 是否覆盖已存在的文件
}
