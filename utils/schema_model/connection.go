package schema_model

type Connection struct {
	Provider Provider `yaml:"provider"`
	Url      string   `yaml:"url"`
}

type Provider string

const (
	PostgreSQL Provider = "postgresql"
	MySQL               = "mysql"
)
