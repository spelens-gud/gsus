package config

// Swagger struct    Swagger 文档配置.
// 用于配置 Swagger API 文档生成的参数.
type Swagger struct {
	Path        string `yaml:"path"`        // Swagger 文档路径
	MainApiPath string `yaml:"mainApiPath"` // 主 API 路径
	Success     string `yaml:"success"`     // 成功响应模板
	Failed      string `yaml:"failed"`      // 失败响应模板
	ProduceType string `yaml:"produceType"` // 响应内容类型
}
