package fileconfig

type Db2struct struct {
	User            string            `yaml:"user"`
	Password        string            `yaml:"password"`
	Host            string            `yaml:"host"`
	Port            int               `yaml:"port"`
	Db              string            `yaml:"db"`
	Charset         string            `yaml:"charset"`
	Path            string            `yaml:"path"`
	GenericTemplate string            `yaml:"genericTemplate"`
	GenericMapTypes []string          `yaml:"genericMapTypes"`
	TypeMap         map[string]string `yaml:"typeMap"`
}
