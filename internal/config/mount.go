package config

type Mount struct {
	Scope string   `yaml:"scope"`
	Name  string   `yaml:"name"`
	Args  []string `yaml:"args"`
}
