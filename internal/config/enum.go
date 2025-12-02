package config

// Enum struct    枚举生成配置.
// 用于配置枚举类型代码生成的参数.
type Enum struct {
	Scope    string `yaml:"scope"`    // 扫描范围
	Path     string `yaml:"path"`     // 生成代码的输出路径
	Template string `yaml:"template"` // 模板文件路径
}
