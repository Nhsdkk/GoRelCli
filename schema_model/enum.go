package schema_model

type Enum struct {
	Name   string   `yaml:"name"`
	Values []string `yaml:"values,flow"`
}
