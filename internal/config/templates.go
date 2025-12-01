package config

type Templates struct {
	ModelPath string     `yaml:"modelPath"`
	Templates []Template `yaml:"templates"`
}

type Template struct {
	Name      string `yaml:"name"`
	Path      string `yaml:"path"`
	Template  string `yaml:"template"`
	Overwrite bool   `yaml:"overwrite"`
}
