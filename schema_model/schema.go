package schema_model

type GoRelSchema struct {
	Connection Connection `yaml:"connection,flow"`
	Models     []Model    `yaml:"models,flow"`
	Enums      []Enum     `yaml:"enums,flow"`
}
