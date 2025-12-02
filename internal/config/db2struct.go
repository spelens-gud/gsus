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
