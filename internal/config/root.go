package config

type Option struct {
	Gsus      Gsus      `yaml:"gsus"`
	Db2struct Db2struct `yaml:"db2struct"`
	Enum      Enum      `yaml:"enum"`
	Templates Templates `yaml:"templates"`
}

type Gsus struct {
	Origin string
}
