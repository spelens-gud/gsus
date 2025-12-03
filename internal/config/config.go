package config

// Db2struct struct    数据库转结构体配置.
// 包含数据库连接信息和代码生成相关配置.
type Db2struct struct {
	User            string            `yaml:"user"`            // 数据库用户名
	Password        string            `yaml:"password"`        // 数据库密码
	Host            string            `yaml:"host"`            // 数据库主机地址
	Port            int               `yaml:"port"`            // 数据库端口
	Db              string            `yaml:"db"`              // 数据库名称
	Charset         string            `yaml:"charset"`         // 字符集
	Path            string            `yaml:"path"`            // 生成代码的输出路径
	GenericTemplate string            `yaml:"genericTemplate"` // 泛型模板路径
	GenericMapTypes []string          `yaml:"genericMapTypes"` // 泛型映射类型列表
	TypeMap         map[string]string `yaml:"typeMap"`         // 数据库类型到 Go 类型的映射
}

// Enum struct    枚举生成配置.
// 用于配置枚举类型代码生成的参数.
type Enum struct {
	Scope    string `yaml:"scope"`    // 扫描范围
	Path     string `yaml:"path"`     // 生成代码的输出路径
	Template string `yaml:"template"` // 模板文件路径
}

// Swagger struct    Swagger 文档配置.
// 用于配置 Swagger API 文档生成的参数.
type Swagger struct {
	Path        string `yaml:"path"`        // Swagger 文档路径
	MainApiPath string `yaml:"mainApiPath"` // 主 API 路径
	Success     string `yaml:"success"`     // 成功响应模板
	Failed      string `yaml:"failed"`      // 失败响应模板
	ProduceType string `yaml:"produceType"` // 响应内容类型
}

// Mount struct    挂载配置.
// 用于配置代码挂载相关的参数.
type Mount struct {
	Scope string   `yaml:"scope"` // 扫描范围
	Name  string   `yaml:"name"`  // 挂载名称
	Args  []string `yaml:"args"`  // 挂载参数列表
}

// Gsus struct    gsus 基础配置.
type Gsus struct {
	Origin string // 源代码仓库地址
}

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
