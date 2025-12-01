package config

type Swagger struct {
	Path        string `yaml:"path"`
	MainApiPath string `yaml:"mainApiPath"`
	Success     string `yaml:"success"`
	Failed      string `yaml:"failed"`
	ProduceType string `yaml:"produceType"`
}
